package nws

import (
	"encoding/json"
	"testing"
)

func TestNullFloat64_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantNil bool
		wantVal float64
	}{
		{"with value", `{"value": 42.5}`, false, 42.5},
		{"null value", `{"value": null}`, true, 0},
		{"zero value", `{"value": 0}`, false, 0},
		{"negative", `{"value": -10.5}`, false, -10.5},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var nf NullFloat64
			if err := json.Unmarshal([]byte(tt.input), &nf); err != nil {
				t.Fatalf("Unmarshal error: %v", err)
			}
			if tt.wantNil && nf.Value != nil {
				t.Errorf("expected nil Value, got %v", *nf.Value)
			}
			if !tt.wantNil {
				if nf.Value == nil {
					t.Fatal("expected non-nil Value, got nil")
				}
				if *nf.Value != tt.wantVal {
					t.Errorf("Value = %v, want %v", *nf.Value, tt.wantVal)
				}
			}
		})
	}
}

func TestObservation_UnmarshalJSON(t *testing.T) {
	input := `{
		"timestamp": "2026-02-17T18:15:00+00:00",
		"textDescription": "Clear",
		"temperature": {"value": 11.0},
		"dewpoint": {"value": -14.9},
		"windDirection": {"value": 270},
		"windSpeed": {"value": null},
		"windGust": {"value": 63},
		"visibility": {"value": 16090},
		"relativeHumidity": {"value": 14.7},
		"barometricPressure": {"value": 99800},
		"windChill": {"value": null}
	}`

	var obs Observation
	if err := json.Unmarshal([]byte(input), &obs); err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}

	if obs.Timestamp != "2026-02-17T18:15:00+00:00" {
		t.Errorf("Timestamp = %q", obs.Timestamp)
	}
	if obs.TextDescription != "Clear" {
		t.Errorf("TextDescription = %q", obs.TextDescription)
	}
	if obs.Temperature.Value == nil || *obs.Temperature.Value != 11.0 {
		t.Errorf("Temperature unexpected")
	}
	if obs.WindSpeed.Value != nil {
		t.Errorf("WindSpeed should be nil, got %v", *obs.WindSpeed.Value)
	}
	if obs.WindGust.Value == nil || *obs.WindGust.Value != 63 {
		t.Errorf("WindGust unexpected")
	}
}

func TestForecastPeriod_UnmarshalJSON(t *testing.T) {
	input := `{
		"name": "Today",
		"temperature": 53,
		"temperatureUnit": "F",
		"windSpeed": "24 to 31 mph",
		"windDirection": "W",
		"shortForecast": "Mostly Sunny",
		"detailedForecast": "Mostly sunny with a high near 53.",
		"isDaytime": true
	}`

	var fp ForecastPeriod
	if err := json.Unmarshal([]byte(input), &fp); err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}

	if fp.Name != "Today" {
		t.Errorf("Name = %q, want %q", fp.Name, "Today")
	}
	if fp.Temperature != 53 {
		t.Errorf("Temperature = %d, want 53", fp.Temperature)
	}
	if !fp.IsDaytime {
		t.Error("IsDaytime should be true")
	}
	if fp.DetailedForecast != "Mostly sunny with a high near 53." {
		t.Errorf("DetailedForecast = %q", fp.DetailedForecast)
	}
}
