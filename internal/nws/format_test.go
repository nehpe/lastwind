package nws

import (
	"math"
	"testing"
)

func floatPtr(f float64) *float64 {
	return &f
}

func TestCToF(t *testing.T) {
	tests := []struct {
		celsius    float64
		fahrenheit float64
	}{
		{0, 32},
		{100, 212},
		{-40, -40},
		{37, 98.6},
	}
	for _, tt := range tests {
		got := CToF(tt.celsius)
		if math.Abs(got-tt.fahrenheit) > 0.1 {
			t.Errorf("CToF(%v) = %v, want %v", tt.celsius, got, tt.fahrenheit)
		}
	}
}

func TestKmhToMph(t *testing.T) {
	tests := []struct {
		kmh, mph float64
	}{
		{0, 0},
		{100, 62.1371},
		{1.60934, 1.0},
	}
	for _, tt := range tests {
		got := KmhToMph(tt.kmh)
		if math.Abs(got-tt.mph) > 0.01 {
			t.Errorf("KmhToMph(%v) = %v, want %v", tt.kmh, got, tt.mph)
		}
	}
}

func TestMetersToMiles(t *testing.T) {
	got := MetersToMiles(1609.34)
	if math.Abs(got-1.0) > 0.001 {
		t.Errorf("MetersToMiles(1609.34) = %v, want 1.0", got)
	}
}

func TestPaToInHg(t *testing.T) {
	got := PaToInHg(101325)
	if math.Abs(got-29.92) > 0.01 {
		t.Errorf("PaToInHg(101325) = %v, want ~29.92", got)
	}
}

func TestCompassDir(t *testing.T) {
	tests := []struct {
		deg  *float64
		want string
	}{
		{nil, ""},
		{floatPtr(0), "N"},
		{floatPtr(90), "E"},
		{floatPtr(180), "S"},
		{floatPtr(270), "W"},
		{floatPtr(45), "NE"},
		{floatPtr(225), "SW"},
		{floatPtr(315), "NW"},
		{floatPtr(360), "N"},
		{floatPtr(350), "N"},
		{floatPtr(170), "S"},
	}
	for _, tt := range tests {
		got := CompassDir(tt.deg)
		if got != tt.want {
			t.Errorf("CompassDir(%v) = %q, want %q", tt.deg, got, tt.want)
		}
	}
}

func TestFormatWind(t *testing.T) {
	tests := []struct {
		name       string
		dir        *float64
		speed      *float64
		gust       *float64
		want       string
	}{
		{"calm - all nil", nil, nil, nil, "Calm"},
		{"calm - zero speed", nil, floatPtr(0), nil, "Calm"},
		{"speed only with dir", floatPtr(270), floatPtr(30), nil, "W 19"},
		{"speed and gust", floatPtr(180), floatPtr(20), floatPtr(40), "S 12 G 25"},
		{"gust only no dir", nil, nil, floatPtr(63), "G 39"},
		{"speed no dir (vrbl)", nil, floatPtr(10), nil, "Vrbl 6"},
		{"speed no dir with gust", nil, floatPtr(10), floatPtr(30), "Vrbl 6 G 19"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FormatWind(tt.dir, tt.speed, tt.gust)
			if got != tt.want {
				t.Errorf("FormatWind() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestFormatTime(t *testing.T) {
	// Invalid timestamp returns as-is
	got := FormatTime("not-a-timestamp")
	if got != "not-a-timestamp" {
		t.Errorf("FormatTime(invalid) = %q, want %q", got, "not-a-timestamp")
	}

	// Valid RFC3339 returns local formatted time
	got = FormatTime("2026-02-17T18:15:00+00:00")
	if got == "" || got == "2026-02-17T18:15:00+00:00" {
		t.Errorf("FormatTime(valid) should format the time, got %q", got)
	}
}

func TestFmtVal(t *testing.T) {
	// nil returns dash
	got := FmtVal(nil, func(v float64) string { return "x" })
	if got != "-" {
		t.Errorf("FmtVal(nil) = %q, want %q", got, "-")
	}

	// non-nil calls formatter
	got = FmtVal(floatPtr(42), func(v float64) string { return "42" })
	if got != "42" {
		t.Errorf("FmtVal(42) = %q, want %q", got, "42")
	}
}

func TestTruncate(t *testing.T) {
	tests := []struct {
		input  string
		maxLen int
		want   string
	}{
		{"short", 10, "short"},
		{"exactly10!", 10, "exactly10!"},
		{"this is too long", 10, "this is t…"},
		{"", 5, ""},
	}
	for _, tt := range tests {
		got := Truncate(tt.input, tt.maxLen)
		if got != tt.want {
			t.Errorf("Truncate(%q, %d) = %q, want %q", tt.input, tt.maxLen, got, tt.want)
		}
	}
}

func TestWordWrap(t *testing.T) {
	tests := []struct {
		name  string
		input string
		width int
		want  []string
	}{
		{"empty", "", 10, nil},
		{"single word", "hello", 10, []string{"hello"}},
		{"fits on one line", "hello world", 20, []string{"hello world"}},
		{"wraps", "the quick brown fox jumps", 15, []string{"the quick brown", "fox jumps"}},
		{"each word on own line", "aaa bbb ccc", 3, []string{"aaa", "bbb", "ccc"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := WordWrap(tt.input, tt.width)
			if len(got) != len(tt.want) {
				t.Fatalf("WordWrap() = %v, want %v", got, tt.want)
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("WordWrap()[%d] = %q, want %q", i, got[i], tt.want[i])
				}
			}
		})
	}
}
