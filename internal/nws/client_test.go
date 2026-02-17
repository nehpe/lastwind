package nws

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestFetchJSON_Success(t *testing.T) {
	type testPayload struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("User-Agent") != userAgent {
			t.Errorf("expected User-Agent %q, got %q", userAgent, r.Header.Get("User-Agent"))
		}
		if r.Header.Get("Accept") != "application/geo+json" {
			t.Errorf("expected Accept %q, got %q", "application/geo+json", r.Header.Get("Accept"))
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(testPayload{Name: "test", Age: 42})
	}))
	defer server.Close()

	result, err := FetchJSON[testPayload](server.URL)
	if err != nil {
		t.Fatalf("FetchJSON() error = %v", err)
	}
	if result.Name != "test" || result.Age != 42 {
		t.Errorf("FetchJSON() = %+v, want {Name:test Age:42}", result)
	}
}

func TestFetchJSON_HTTPError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(404)
		w.Write([]byte("not found"))
	}))
	defer server.Close()

	type empty struct{}
	_, err := FetchJSON[empty](server.URL)
	if err == nil {
		t.Fatal("FetchJSON() expected error for 404, got nil")
	}
}

func TestFetchJSON_InvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("not json"))
	}))
	defer server.Close()

	type empty struct{}
	_, err := FetchJSON[empty](server.URL)
	if err == nil {
		t.Fatal("FetchJSON() expected error for invalid JSON, got nil")
	}
}

func TestFetchJSON_ConnectionError(t *testing.T) {
	type empty struct{}
	_, err := FetchJSON[empty]("http://localhost:1")
	if err == nil {
		t.Fatal("FetchJSON() expected connection error, got nil")
	}
}
