package core

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

func SetJSONHeaders(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
}

func WriteJSON(w http.ResponseWriter, statusCode int, payload interface{}) {
	SetJSONHeaders(w)
	w.WriteHeader(statusCode)

	if err := json.NewEncoder(w).Encode(payload); err != nil {
		log.Printf("Error encoding response: %v", err)
	}
}

func WriteError(w http.ResponseWriter, statusCode int, message string) {
	WriteJSON(w, statusCode, map[string]string{"error": message})
}

func ProductIDFromRequest(r *http.Request) string {
	productID := strings.TrimPrefix(r.URL.Path, "/products/")
	if productID == r.URL.Path || productID == "" {
		return r.URL.Query().Get("id")
	}

	return productID
}

func FindProductByID(products []Product, id string) (Product, bool) {
	for _, product := range products {
		if product.ID == id {
			return product, true
		}
	}

	return Product{}, false
}

// LoadProducts loads product data from data.json.
// It first checks the PRODUCTS_DATA_FILE environment variable.
// If not set, it searches in standard locations.
func LoadProducts() ([]Product, error) {
	var dataPath string

	if envPath := os.Getenv("PRODUCTS_DATA_FILE"); envPath != "" {
		if _, err := os.Stat(envPath); err == nil {
			dataPath = envPath
		} else {
			return nil, fmt.Errorf("PRODUCTS_DATA_FILE path does not exist: %s", envPath)
		}
	} else {
		var candidates []string
		candidates = append(candidates, "src/data.json", "../data.json", "data.json")

		execPath := filepath.Join(filepath.Dir(os.Args[0]), "../data.json")
		candidates = append(candidates, execPath)

		if _, thisFile, _, ok := runtime.Caller(0); ok {
			candidates = append(candidates, filepath.Join(filepath.Dir(thisFile), "..", "..", "data.json"))
		}

		for _, p := range candidates {
			if _, err := os.Stat(p); err == nil {
				dataPath = p
				break
			}
		}

		if dataPath == "" {
			return nil, fmt.Errorf("data.json not found in expected locations")
		}
	}

	data, err := os.ReadFile(dataPath)
	if err != nil {
		log.Printf("Error reading data.json: %v", err)
		return nil, err
	}

	var dataFile DataFile
	if err := json.Unmarshal(data, &dataFile); err != nil {
		log.Printf("Error unmarshaling data.json: %v", err)
		return nil, err
	}

	return dataFile.Products, nil
}
