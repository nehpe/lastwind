package nws

import (
	"fmt"
	"math"
	"strings"
	"time"
)

func CToF(c float64) float64 {
	return c*9.0/5.0 + 32.0
}

func KmhToMph(kmh float64) float64 {
	return kmh * 0.621371
}

func MetersToMiles(m float64) float64 {
	return m / 1609.34
}

func PaToInHg(pa float64) float64 {
	return pa / 3386.39
}

func CompassDir(deg *float64) string {
	if deg == nil {
		return ""
	}
	dirs := []string{"N", "NNE", "NE", "ENE", "E", "ESE", "SE", "SSE",
		"S", "SSW", "SW", "WSW", "W", "WNW", "NW", "NNW"}
	idx := int(math.Round(*deg/22.5)) % 16
	return dirs[idx]
}

func FormatWind(dir, speed, gust *float64) string {
	hasSpeed := speed != nil && *speed > 0
	hasGust := gust != nil && *gust > 0

	if !hasSpeed && !hasGust {
		return "Calm"
	}

	var parts []string
	dirStr := CompassDir(dir)
	if dirStr != "" {
		parts = append(parts, dirStr)
	} else if hasSpeed {
		parts = append(parts, "Vrbl")
	}
	if hasSpeed {
		parts = append(parts, fmt.Sprintf("%.0f", KmhToMph(*speed)))
	}
	if hasGust {
		parts = append(parts, fmt.Sprintf("G %.0f", KmhToMph(*gust)))
	}
	return strings.Join(parts, " ")
}

func FormatTime(ts string) string {
	t, err := time.Parse(time.RFC3339, ts)
	if err != nil {
		return ts
	}
	return t.Local().Format("Jan 02 15:04")
}

func FmtVal(v *float64, fn func(float64) string) string {
	if v == nil {
		return "-"
	}
	return fn(*v)
}

func Truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-1] + "…"
}

func WordWrap(s string, width int) []string {
	words := strings.Fields(s)
	if len(words) == 0 {
		return nil
	}
	var lines []string
	current := words[0]
	for _, w := range words[1:] {
		if len(current)+1+len(w) > width {
			lines = append(lines, current)
			current = w
		} else {
			current += " " + w
		}
	}
	lines = append(lines, current)
	return lines
}
