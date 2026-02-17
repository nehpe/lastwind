package main

import (
	"flag"
	"fmt"
	"os"

	"lastwind/internal/config"
	"lastwind/internal/nws"
)

func main() {
	cfg, err := config.LoadOrSetup()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		os.Exit(1)
	}

	lat := flag.Float64("lat", cfg.Latitude, "latitude")
	lon := flag.Float64("lon", cfg.Longitude, "longitude")
	flag.Parse()

	// 1. Get point metadata
	pointsURL := fmt.Sprintf("https://api.weather.gov/points/%.4f,%.4f", *lat, *lon)
	points, err := nws.FetchJSON[nws.PointsResponse](pointsURL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error fetching point data: %v\n", err)
		os.Exit(1)
	}

	city := points.Properties.RelativeLocation.Properties.City
	state := points.Properties.RelativeLocation.Properties.State
	forecastURL := points.Properties.Forecast
	stationsURL := points.Properties.ObservationStations

	// 2. Get nearest station
	stations, err := nws.FetchJSON[nws.StationsResponse](stationsURL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error fetching stations: %v\n", err)
		os.Exit(1)
	}
	if len(stations.Features) == 0 {
		fmt.Fprintf(os.Stderr, "No observation stations found\n")
		os.Exit(1)
	}
	stationID := stations.Features[0].Properties.StationIdentifier
	stationName := stations.Features[0].Properties.Name

	// 3. Get current observation
	obsURL := fmt.Sprintf("https://api.weather.gov/stations/%s/observations/latest", stationID)
	obs, err := nws.FetchJSON[nws.ObservationResponse](obsURL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error fetching observations: %v\n", err)
		os.Exit(1)
	}

	// 4. Get forecast
	forecast, err := nws.FetchJSON[nws.ForecastResponse](forecastURL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error fetching forecast: %v\n", err)
		os.Exit(1)
	}

	// Display
	fmt.Printf("\n  %s, %s\n", city, state)
	fmt.Printf("  Station: %s (%s)\n\n", stationName, stationID)

	printCurrentConditions(obs)
	printForecast(forecast)
}

func printCurrentConditions(obs nws.ObservationResponse) {
	p := obs.Properties

	fmt.Printf("  ── Current Conditions (%s) ──\n\n", nws.FormatTime(p.Timestamp))
	fmt.Printf("    %s\n", p.TextDescription)

	if p.Temperature.Value != nil {
		fmt.Printf("    Temperature:  %.0f°F", nws.CToF(*p.Temperature.Value))
		if p.WindChill.Value != nil {
			fmt.Printf("  (Wind Chill: %.0f°F)", nws.CToF(*p.WindChill.Value))
		}
		fmt.Println()
	}
	if p.Dewpoint.Value != nil {
		fmt.Printf("    Dewpoint:     %.0f°F\n", nws.CToF(*p.Dewpoint.Value))
	}
	if p.RelativeHumidity.Value != nil {
		fmt.Printf("    Humidity:     %.0f%%\n", *p.RelativeHumidity.Value)
	}
	wind := nws.FormatWind(p.WindDirection.Value, p.WindSpeed.Value, p.WindGust.Value)
	if wind != "Calm" {
		wind += " mph"
	}
	fmt.Printf("    Wind:         %s\n", wind)
	if p.Visibility.Value != nil {
		fmt.Printf("    Visibility:   %.1f mi\n", nws.MetersToMiles(*p.Visibility.Value))
	}
	if p.Barometer.Value != nil {
		fmt.Printf("    Barometer:    %.2f in\n", nws.PaToInHg(*p.Barometer.Value))
	}
	fmt.Println()
}

func printForecast(forecast nws.ForecastResponse) {
	periods := forecast.Properties.Periods
	if len(periods) == 0 {
		return
	}

	maxPeriods := 4
	if len(periods) < maxPeriods {
		maxPeriods = len(periods)
	}

	fmt.Printf("  ── Forecast ───────────────────────────────\n\n")

	for i := 0; i < maxPeriods; i++ {
		p := periods[i]
		tempLabel := "High"
		if !p.IsDaytime {
			tempLabel = "Low"
		}
		fmt.Printf("    %-18s %s: %d°%s  Wind: %s %s\n", p.Name, tempLabel, p.Temperature, p.TemperatureUnit, p.WindDirection, p.WindSpeed)
		wrapped := nws.WordWrap(p.DetailedForecast, 60)
		for _, line := range wrapped {
			fmt.Printf("      %s\n", line)
		}
		fmt.Println()
	}
}
