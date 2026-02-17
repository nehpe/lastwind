package config

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	Station   string  `json:"station"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

var Default = Config{
	Station:   "KEIK",
	Latitude:  40.0388,
	Longitude: -105.0412,
}

type geoIPResponse struct {
	Status  string  `json:"status"`
	City    string  `json:"city"`
	Region  string  `json:"regionName"`
	Lat     float64 `json:"lat"`
	Lon     float64 `json:"lon"`
}

type nwsPointsResponse struct {
	Properties struct {
		ObservationStations string `json:"observationStations"`
	} `json:"properties"`
}

type nwsStationsResponse struct {
	Features []struct {
		Properties struct {
			StationIdentifier string `json:"stationIdentifier"`
			Name              string `json:"name"`
		} `json:"properties"`
	} `json:"features"`
}

// DetectedLocation holds auto-detected location info.
type DetectedLocation struct {
	City      string
	Region    string
	Latitude  float64
	Longitude float64
	Station   string
	StationName string
}

// DetectLocation uses IP geolocation and the NWS API to find the user's
// nearest weather station. Returns Default values on any failure.
func DetectLocation(client *http.Client) DetectedLocation {
	result := DetectedLocation{
		Latitude:  Default.Latitude,
		Longitude: Default.Longitude,
		Station:   Default.Station,
	}

	// 1. IP geolocation
	geo, err := fetchGeoIP(client)
	if err != nil || geo.Status != "success" {
		return result
	}
	result.City = geo.City
	result.Region = geo.Region
	result.Latitude = geo.Lat
	result.Longitude = geo.Lon

	// 2. Find nearest NWS station
	station, name, err := fetchNearestStation(client, geo.Lat, geo.Lon)
	if err != nil {
		return result
	}
	result.Station = station
	result.StationName = name

	return result
}

func fetchGeoIP(client *http.Client) (geoIPResponse, error) {
	var geo geoIPResponse
	resp, err := client.Get("http://ip-api.com/json/")
	if err != nil {
		return geo, err
	}
	defer resp.Body.Close()
	err = json.NewDecoder(resp.Body).Decode(&geo)
	return geo, err
}

func fetchNearestStation(client *http.Client, lat, lon float64) (id, name string, err error) {
	url := fmt.Sprintf("https://api.weather.gov/points/%.4f,%.4f", lat, lon)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", "", err
	}
	req.Header.Set("User-Agent", "(lastwind)")
	req.Header.Set("Accept", "application/geo+json")

	resp, err := client.Do(req)
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()

	var points nwsPointsResponse
	if err := json.NewDecoder(resp.Body).Decode(&points); err != nil {
		return "", "", err
	}

	if points.Properties.ObservationStations == "" {
		return "", "", fmt.Errorf("no stations URL")
	}

	req2, err := http.NewRequest("GET", points.Properties.ObservationStations, nil)
	if err != nil {
		return "", "", err
	}
	req2.Header.Set("User-Agent", "(lastwind)")
	req2.Header.Set("Accept", "application/geo+json")

	resp2, err := client.Do(req2)
	if err != nil {
		return "", "", err
	}
	defer resp2.Body.Close()

	var stations nwsStationsResponse
	if err := json.NewDecoder(resp2.Body).Decode(&stations); err != nil {
		return "", "", err
	}

	if len(stations.Features) == 0 {
		return "", "", fmt.Errorf("no stations found")
	}

	s := stations.Features[0].Properties
	return s.StationIdentifier, s.Name, nil
}

func Dir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".config", "lastwind"), nil
}

func Path() (string, error) {
	dir, err := Dir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "config.json"), nil
}

func Load() (Config, error) {
	path, err := Path()
	if err != nil {
		return Default, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return Default, err
		}
		return Default, err
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return Default, fmt.Errorf("invalid config file %s: %w", path, err)
	}
	return cfg, nil
}

func Save(cfg Config) error {
	dir, err := Dir()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')

	path := filepath.Join(dir, "config.json")
	return os.WriteFile(path, data, 0644)
}

// RunSetup interactively prompts the user to create a config file.
// It reads from the provided reader (typically os.Stdin).
func RunSetup(reader *bufio.Reader) (Config, error) {
	fmt.Println()
	fmt.Println("  ── lastwind configuration ──")
	fmt.Println()
	fmt.Println("  No configuration file found. Let's set one up.")
	fmt.Print("  Detecting your location...")

	client := &http.Client{Timeout: 10 * time.Second}
	detected := DetectLocation(client)

	cfg := Config{
		Station:   detected.Station,
		Latitude:  detected.Latitude,
		Longitude: detected.Longitude,
	}

	if detected.City != "" {
		fmt.Printf(" found %s, %s\n", detected.City, detected.Region)
	} else {
		fmt.Println(" could not detect, using defaults")
	}
	if detected.StationName != "" {
		fmt.Printf("  Nearest station: %s (%s)\n", detected.StationName, detected.Station)
	}
	fmt.Println()

	cfg.Station = prompt(reader, "ICAO station code", cfg.Station)
	cfg.Latitude = promptFloat(reader, "Latitude", cfg.Latitude)
	cfg.Longitude = promptFloat(reader, "Longitude", cfg.Longitude)

	if err := Save(cfg); err != nil {
		return cfg, fmt.Errorf("failed to save config: %w", err)
	}

	path, _ := Path()
	fmt.Printf("\n  Config saved to %s\n\n", path)
	return cfg, nil
}

func prompt(reader *bufio.Reader, label string, defaultVal string) string {
	fmt.Printf("  %s [%s]: ", label, defaultVal)
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)
	if input == "" {
		return defaultVal
	}
	return strings.ToUpper(input)
}

func promptFloat(reader *bufio.Reader, label string, defaultVal float64) float64 {
	fmt.Printf("  %s [%.4f]: ", label, defaultVal)
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)
	if input == "" {
		return defaultVal
	}
	val, err := strconv.ParseFloat(input, 64)
	if err != nil {
		fmt.Printf("  Invalid number, using default %.4f\n", defaultVal)
		return defaultVal
	}
	return val
}

// LoadOrSetup loads the config, or runs interactive setup if it doesn't exist.
func LoadOrSetup() (Config, error) {
	cfg, err := Load()
	if err == nil {
		return cfg, nil
	}

	if !os.IsNotExist(err) {
		return cfg, err
	}

	return RunSetup(bufio.NewReader(os.Stdin))
}
