package cache

import (
	"testing"
	"time"
)

func TestCacheSetAndGet(t *testing.T) {
	cache := NewCache()
	key := "test-key"
	data := []byte("test-data")
	headers := map[string][]string{
		"Content-Type": {"application/json"},
	}
	statusCode := 200
	ttl := 1 * time.Minute

	cache.Set(key, data, headers, statusCode, ttl)

	retrievedData, retrievedHeaders, retrievedStatus, found := cache.Get(key)
	if !found {
		t.Fatal("expected cache entry to be found")
	}

	if string(retrievedData) != "test-data" {
		t.Fatalf("expected data 'test-data', got '%s'", string(retrievedData))
	}

	if retrievedStatus != 200 {
		t.Fatalf("expected status code 200, got %d", retrievedStatus)
	}

	if len(retrievedHeaders) == 0 || retrievedHeaders["Content-Type"][0] != "application/json" {
		t.Fatalf("expected headers to be preserved")
	}
}

func TestCacheReturnsNotFoundForMissingKey(t *testing.T) {
	cache := NewCache()

	_, _, _, found := cache.Get("non-existent-key")
	if found {
		t.Fatal("expected cache entry to not be found")
	}
}

func TestCacheExpiresEntries(t *testing.T) {
	cache := NewCache()
	key := "expiring-key"
	data := []byte("expiring-data")
	headers := map[string][]string{}
	statusCode := 200
	ttl := 10 * time.Millisecond

	cache.Set(key, data, headers, statusCode, ttl)

	// Verify entry exists immediately
	_, _, _, found := cache.Get(key)
	if !found {
		t.Fatal("expected cache entry to be found immediately after set")
	}

	// Wait for expiration
	time.Sleep(50 * time.Millisecond)

	// Verify entry is expired
	_, _, _, found = cache.Get(key)
	if found {
		t.Fatal("expected cache entry to be expired")
	}
}

func TestCacheClear(t *testing.T) {
	cache := NewCache()
	cache.Set("key1", []byte("data1"), map[string][]string{}, 200, 1*time.Minute)
	cache.Set("key2", []byte("data2"), map[string][]string{}, 200, 1*time.Minute)

	_, _, _, found1 := cache.Get("key1")
	_, _, _, found2 := cache.Get("key2")
	if !found1 || !found2 {
		t.Fatal("expected both cache entries to exist")
	}

	cache.Clear()

	_, _, _, found1 = cache.Get("key1")
	_, _, _, found2 = cache.Get("key2")
	if found1 || found2 {
		t.Fatal("expected cache to be empty after Clear()")
	}
}

func TestCacheCleanupExpired(t *testing.T) {
	cache := NewCache()
	cache.Set("expired-key", []byte("data"), map[string][]string{}, 200, 10*time.Millisecond)
	cache.Set("active-key", []byte("data"), map[string][]string{}, 200, 1*time.Minute)

	// Wait for first entry to expire
	time.Sleep(50 * time.Millisecond)

	cache.CleanupExpired()

	_, _, _, expiredFound := cache.Get("expired-key")
	_, _, _, activeFound := cache.Get("active-key")

	if expiredFound {
		t.Fatal("expected expired entry to be removed")
	}

	if !activeFound {
		t.Fatal("expected active entry to remain")
	}
}

func TestCacheMultipleHeaderValues(t *testing.T) {
	cache := NewCache()
	key := "test-key"
	data := []byte("test-data")
	headers := map[string][]string{
		"Set-Cookie": {"cookie1=value1", "cookie2=value2"},
	}
	statusCode := 200
	ttl := 1 * time.Minute

	cache.Set(key, data, headers, statusCode, ttl)

	_, retrievedHeaders, _, found := cache.Get(key)
	if !found {
		t.Fatal("expected cache entry to be found")
	}

	cookies := retrievedHeaders["Set-Cookie"]
	if len(cookies) != 2 || cookies[0] != "cookie1=value1" || cookies[1] != "cookie2=value2" {
		t.Fatalf("expected multiple header values to be preserved, got %v", cookies)
	}
}

func TestCacheThreadSafety(t *testing.T) {
	cache := NewCache()
	done := make(chan bool, 10)

	// Concurrent writes
	for i := 0; i < 5; i++ {
		go func(index int) {
			for j := 0; j < 100; j++ {
				key := "key-" + string(rune(index)) + "-" + string(rune(j))
				cache.Set(key, []byte("data"), map[string][]string{}, 200, 1*time.Minute)
			}
			done <- true
		}(i)
	}

	// Concurrent reads
	for i := 0; i < 5; i++ {
		go func() {
			for j := 0; j < 100; j++ {
				key := "key-0-" + string(rune(j))
				cache.Get(key)
			}
			done <- true
		}()
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}

	t.Log("Concurrent operations completed successfully")
}
