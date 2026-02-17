package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"lastwind/internal/config"
	"lastwind/internal/nws"
)

func main() {
	cfg, err := config.LoadOrSetup()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		os.Exit(1)
	}

	station := flag.String("station", cfg.Station, "ICAO station identifier (e.g. KEIK, KDEN)")
	count := flag.Int("n", 10, "number of recent observations to display")
	flag.Parse()

	stationID := strings.ToUpper(*station)

	// Fetch station name
	stationURL := fmt.Sprintf("https://api.weather.gov/stations/%s", stationID)
	stationInfo, err := nws.FetchJSON[nws.StationResponse](stationURL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error fetching station info: %v\n", err)
		os.Exit(1)
	}

	// Fetch observations (3 days worth)
	obsURL := fmt.Sprintf("https://api.weather.gov/stations/%s/observations?limit=500", stationID)
	obsResp, err := nws.FetchJSON[nws.ObservationsResponse](obsURL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error fetching observations: %v\n", err)
		os.Exit(1)
	}

	// Filter to last 3 days
	cutoff := time.Now().UTC().AddDate(0, 0, -3)
	var observations []nws.Observation
	for _, f := range obsResp.Features {
		t, err := time.Parse(time.RFC3339, f.Properties.Timestamp)
		if err != nil {
			continue
		}
		if t.After(cutoff) {
			observations = append(observations, f.Properties)
		}
	}

	if len(observations) == 0 {
		fmt.Fprintf(os.Stderr, "No observations found for station %s\n", stationID)
		os.Exit(1)
	}

	// Display header
	fmt.Printf("\n  Station: %s (%s)\n\n", stationInfo.Properties.Name, stationID)

	// Display recent observations table
	displayCount := *count
	if displayCount > len(observations) {
		displayCount = len(observations)
	}

	fmt.Printf("  ┌────────────────┬────────────────┬────────┬──────┬──────┬────────┬──────────────────────────────┐\n")
	fmt.Printf("  │ Time           │ Wind           │ Vis mi │ Temp │ Dwpt │ Hum    │ Weather                      │\n")
	fmt.Printf("  ├────────────────┼────────────────┼────────┼──────┼──────┼────────┼──────────────────────────────┤\n")

	for i := 0; i < displayCount; i++ {
		o := observations[i]
		ts := nws.FormatTime(o.Timestamp)
		wind := nws.FormatWind(o.WindDirection.Value, o.WindSpeed.Value, o.WindGust.Value)
		vis := nws.FmtVal(o.Visibility.Value, func(v float64) string { return fmt.Sprintf("%.1f", nws.MetersToMiles(v)) })
		temp := nws.FmtVal(o.Temperature.Value, func(v float64) string { return fmt.Sprintf("%.0f", nws.CToF(v)) })
		dwpt := nws.FmtVal(o.Dewpoint.Value, func(v float64) string { return fmt.Sprintf("%.0f", nws.CToF(v)) })
		hum := nws.FmtVal(o.RelativeHumidity.Value, func(v float64) string { return fmt.Sprintf("%.0f%%", v) })
		weather := nws.Truncate(o.TextDescription, 28)

		fmt.Printf("  │ %-14s │ %-14s │ %6s │ %4s │ %4s │ %6s │ %-28s │\n",
			ts, wind, vis, temp, dwpt, hum, weather)
	}

	fmt.Printf("  └────────────────┴────────────────┴────────┴──────┴──────┴────────┴──────────────────────────────┘\n")
	fmt.Printf("  Showing %d of %d observations (3 days)\n\n", displayCount, len(observations))

	// Find highest wind and gust
	maxSpeed, maxGust := 0.0, 0.0
	var maxSpeedObs, maxGustObs nws.Observation

	for _, o := range observations {
		if o.WindSpeed.Value != nil && *o.WindSpeed.Value > maxSpeed {
			maxSpeed = *o.WindSpeed.Value
			maxSpeedObs = o
		}
		if o.WindGust.Value != nil && *o.WindGust.Value > maxGust {
			maxGust = *o.WindGust.Value
			maxGustObs = o
		}
	}

	fmt.Printf("  ── 3-Day Extremes ─────────────────────────\n")
	if maxSpeed > 0 {
		fmt.Printf("  Highest Wind:  %.0f mph %s (%s)\n",
			nws.KmhToMph(maxSpeed), nws.CompassDir(maxSpeedObs.WindDirection.Value), nws.FormatTime(maxSpeedObs.Timestamp))
	} else {
		fmt.Printf("  Highest Wind:  No sustained winds recorded\n")
	}
	if maxGust > 0 {
		fmt.Printf("  Highest Gust:  %.0f mph %s (%s)\n",
			nws.KmhToMph(maxGust), nws.CompassDir(maxGustObs.WindDirection.Value), nws.FormatTime(maxGustObs.Timestamp))
	} else {
		fmt.Printf("  Highest Gust:  No gusts recorded\n")
	}
	fmt.Println()
}
