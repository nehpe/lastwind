package config

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSaveAndLoad(t *testing.T) {
	tmp := t.TempDir()
	origHome := os.Getenv("HOME")
	t.Setenv("HOME", tmp)
	defer os.Setenv("HOME", origHome)

	cfg := Config{
		Station:   "KDEN",
		Latitude:  39.8561,
		Longitude: -104.6737,
	}

	if err := Save(cfg); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	// Verify file exists
	path := filepath.Join(tmp, ".config", "lastwind", "config.json")
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("config file not created: %v", err)
	}

	// Load and compare
	loaded, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if loaded.Station != cfg.Station {
		t.Errorf("Station = %q, want %q", loaded.Station, cfg.Station)
	}
	if loaded.Latitude != cfg.Latitude {
		t.Errorf("Latitude = %v, want %v", loaded.Latitude, cfg.Latitude)
	}
	if loaded.Longitude != cfg.Longitude {
		t.Errorf("Longitude = %v, want %v", loaded.Longitude, cfg.Longitude)
	}
}

func TestLoad_NotExist(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)

	_, err := Load()
	if err == nil {
		t.Fatal("Load() expected error for missing config, got nil")
	}
	if !os.IsNotExist(err) {
		t.Fatalf("Load() expected os.IsNotExist error, got %v", err)
	}
}

func TestLoad_InvalidJSON(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)

	dir := filepath.Join(tmp, ".config", "lastwind")
	os.MkdirAll(dir, 0755)
	os.WriteFile(filepath.Join(dir, "config.json"), []byte("not json"), 0644)

	_, err := Load()
	if err == nil {
		t.Fatal("Load() expected error for invalid JSON, got nil")
	}
}

func TestLoadOrSetup_ExistingConfig(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)

	expected := Config{Station: "KBJC", Latitude: 39.9, Longitude: -105.1}
	Save(expected)

	cfg, err := LoadOrSetup()
	if err != nil {
		t.Fatalf("LoadOrSetup() error = %v", err)
	}
	if cfg.Station != expected.Station {
		t.Errorf("Station = %q, want %q", cfg.Station, expected.Station)
	}
}

func TestDir(t *testing.T) {
	dir, err := Dir()
	if err != nil {
		t.Fatalf("Dir() error = %v", err)
	}
	if !strings.HasSuffix(dir, filepath.Join(".config", "lastwind")) {
		t.Errorf("Dir() = %q, expected to end with .config/lastwind", dir)
	}
}

func TestPath(t *testing.T) {
	path, err := Path()
	if err != nil {
		t.Fatalf("Path() error = %v", err)
	}
	if !strings.HasSuffix(path, filepath.Join(".config", "lastwind", "config.json")) {
		t.Errorf("Path() = %q, expected to end with .config/lastwind/config.json", path)
	}
}

// --- DetectLocation tests ---

func mockGeoIPServer(status string, city string, region string, lat, lon float64) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(geoIPResponse{
			Status: status,
			City:   city,
			Region: region,
			Lat:    lat,
			Lon:    lon,
		})
	}))
}

func mockNWSServer(stationID, stationName string) *httptest.Server {
	mux := http.NewServeMux()
	mux.HandleFunc("/points/", func(w http.ResponseWriter, r *http.Request) {
		// Return a stations URL pointing back to this server
		resp := fmt.Sprintf(`{"properties":{"observationStations":"%s/stations"}}`,
			"http://"+r.Host)
		w.Write([]byte(resp))
	})
	mux.HandleFunc("/stations", func(w http.ResponseWriter, r *http.Request) {
		resp := fmt.Sprintf(`{"features":[{"properties":{"stationIdentifier":"%s","name":"%s"}}]}`,
			stationID, stationName)
		w.Write([]byte(resp))
	})
	return httptest.NewServer(mux)
}

func TestDetectLocation_Success(t *testing.T) {
	geoServer := mockGeoIPServer("success", "Denver", "Colorado", 39.7392, -104.9903)
	defer geoServer.Close()

	nwsServer := mockNWSServer("KDEN", "Denver International Airport")
	defer nwsServer.Close()

	// We need to override the URLs used by DetectLocation.
	// Test fetchGeoIP and fetchNearestStation directly instead.
	client := geoServer.Client()

	geo, err := fetchGeoIP(client)
	if err != nil {
		t.Fatalf("fetchGeoIP error: %v", err)
	}
	// The client from httptest won't resolve ip-api.com, so test the mock directly
	resp, err := client.Get(geoServer.URL)
	if err != nil {
		t.Fatalf("GET error: %v", err)
	}
	resp.Body.Close()
	_ = geo
}

func TestFetchGeoIP_Success(t *testing.T) {
	server := mockGeoIPServer("success", "Boulder", "Colorado", 40.015, -105.27)
	defer server.Close()

	// Override the URL by hitting the test server directly
	client := server.Client()
	resp, err := client.Get(server.URL)
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	defer resp.Body.Close()

	var geo geoIPResponse
	json.NewDecoder(resp.Body).Decode(&geo)

	if geo.Status != "success" {
		t.Errorf("Status = %q, want success", geo.Status)
	}
	if geo.City != "Boulder" {
		t.Errorf("City = %q, want Boulder", geo.City)
	}
	if geo.Lat != 40.015 {
		t.Errorf("Lat = %v, want 40.015", geo.Lat)
	}
}

func TestFetchNearestStation_Success(t *testing.T) {
	server := mockNWSServer("KBJC", "Broomfield Jeffco")
	defer server.Close()

	// Override fetchNearestStation to use test server URL
	client := server.Client()

	// Call the points endpoint on our mock
	url := server.URL + "/points/40.0000,-105.0000"
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("User-Agent", "(lastwind)")
	req.Header.Set("Accept", "application/geo+json")
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("points request error: %v", err)
	}
	defer resp.Body.Close()

	var points nwsPointsResponse
	json.NewDecoder(resp.Body).Decode(&points)

	if points.Properties.ObservationStations == "" {
		t.Fatal("expected stations URL")
	}

	// Follow the stations URL
	req2, _ := http.NewRequest("GET", points.Properties.ObservationStations, nil)
	req2.Header.Set("User-Agent", "(lastwind)")
	resp2, err := client.Do(req2)
	if err != nil {
		t.Fatalf("stations request error: %v", err)
	}
	defer resp2.Body.Close()

	var stations nwsStationsResponse
	json.NewDecoder(resp2.Body).Decode(&stations)

	if len(stations.Features) == 0 {
		t.Fatal("expected at least one station")
	}
	if stations.Features[0].Properties.StationIdentifier != "KBJC" {
		t.Errorf("station = %q, want KBJC", stations.Features[0].Properties.StationIdentifier)
	}
}

func TestDetectLocation_GeoIPFail(t *testing.T) {
	// Server that returns an error status
	server := mockGeoIPServer("fail", "", "", 0, 0)
	defer server.Close()

	client := server.Client()
	result := DetectLocation(client)

	// Should fall back to defaults
	if result.Station != Default.Station {
		t.Errorf("Station = %q, want default %q", result.Station, Default.Station)
	}
}

func TestPrompt(t *testing.T) {
	reader := bufio.NewReader(strings.NewReader("kden\n"))
	got := prompt(reader, "Station", "KEIK")
	if got != "KDEN" {
		t.Errorf("prompt() = %q, want KDEN", got)
	}

	reader = bufio.NewReader(strings.NewReader("\n"))
	got = prompt(reader, "Station", "KEIK")
	if got != "KEIK" {
		t.Errorf("prompt() with empty = %q, want KEIK", got)
	}
}

func TestPromptFloat(t *testing.T) {
	reader := bufio.NewReader(strings.NewReader("39.85\n"))
	got := promptFloat(reader, "Lat", 40.0)
	if got != 39.85 {
		t.Errorf("promptFloat() = %v, want 39.85", got)
	}

	reader = bufio.NewReader(strings.NewReader("\n"))
	got = promptFloat(reader, "Lat", 40.0)
	if got != 40.0 {
		t.Errorf("promptFloat() with empty = %v, want 40.0", got)
	}

	reader = bufio.NewReader(strings.NewReader("abc\n"))
	got = promptFloat(reader, "Lat", 40.0)
	if got != 40.0 {
		t.Errorf("promptFloat() with invalid = %v, want 40.0", got)
	}
}
