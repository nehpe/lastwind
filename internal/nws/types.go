package nws

type NullFloat64 struct {
	Value *float64 `json:"value"`
}

type Observation struct {
	Timestamp        string      `json:"timestamp"`
	TextDescription  string      `json:"textDescription"`
	Temperature      NullFloat64 `json:"temperature"`
	Dewpoint         NullFloat64 `json:"dewpoint"`
	WindDirection    NullFloat64 `json:"windDirection"`
	WindSpeed        NullFloat64 `json:"windSpeed"`
	WindGust         NullFloat64 `json:"windGust"`
	Visibility       NullFloat64 `json:"visibility"`
	RelativeHumidity NullFloat64 `json:"relativeHumidity"`
	Barometer        NullFloat64 `json:"barometricPressure"`
	WindChill        NullFloat64 `json:"windChill"`
}

type StationResponse struct {
	Properties struct {
		Name string `json:"name"`
	} `json:"properties"`
}

type ObservationsResponse struct {
	Features []struct {
		Properties Observation `json:"properties"`
	} `json:"features"`
}

type ObservationResponse struct {
	Properties Observation `json:"properties"`
}

type PointsResponse struct {
	Properties struct {
		RelativeLocation struct {
			Properties struct {
				City  string `json:"city"`
				State string `json:"state"`
			} `json:"properties"`
		} `json:"relativeLocation"`
		Forecast            string `json:"forecast"`
		ObservationStations string `json:"observationStations"`
	} `json:"properties"`
}

type StationsResponse struct {
	Features []struct {
		Properties struct {
			StationIdentifier string `json:"stationIdentifier"`
			Name              string `json:"name"`
		} `json:"properties"`
	} `json:"features"`
}

type ForecastResponse struct {
	Properties struct {
		Periods []ForecastPeriod `json:"periods"`
	} `json:"properties"`
}

type ForecastPeriod struct {
	Name             string `json:"name"`
	Temperature      int    `json:"temperature"`
	TemperatureUnit  string `json:"temperatureUnit"`
	WindSpeed        string `json:"windSpeed"`
	WindDirection    string `json:"windDirection"`
	ShortForecast    string `json:"shortForecast"`
	DetailedForecast string `json:"detailedForecast"`
	IsDaytime        bool   `json:"isDaytime"`
}
