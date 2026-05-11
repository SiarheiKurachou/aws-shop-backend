package importfileparser

import (
	"context"
	"reflect"
	"strings"
	"testing"

	"github.com/aws/aws-lambda-go/events"
)

func TestParseCSVRecordsWithHeaders(t *testing.T) {
	input := strings.NewReader("title,price,count\nBook,10,3\nPen,2\nNotebook,7,15,unexpected\n")

	records, err := parseCSVRecordsWithHeaders(input)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	expected := []map[string]string{
		{"title": "Book", "price": "10", "count": "3"},
		{"title": "Pen", "price": "2", "count": ""},
		{"title": "Notebook", "price": "7", "count": "15", "extra_1": "unexpected"},
	}

	if !reflect.DeepEqual(expected, records) {
		t.Fatalf("expected records %+v, got %+v", expected, records)
	}
}

func TestHandleImportFileParser_EmptyEvent(t *testing.T) {
	err := HandleImportFileParser(context.Background(), events.S3Event{})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
}
