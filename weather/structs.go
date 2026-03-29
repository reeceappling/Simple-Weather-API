package weather

import "errors"

type PointsResponseBody struct {
	ForecastUrl string `json:"forecast"`
	// unused response fields not added to struct
	// See full schema at https://www.weather.gov/documentation/services-web-api#/default/point
}

type Data struct {
	Properties Properties
	// unused response fields not added to struct
	// See full schema at https://www.weather.gov/documentation/services-web-api#/default/gridpoint_forecast
	// Note: This is the GEOJSON version, unlike from the points response
}

type Properties struct {
	Periods []Period
}

type Period struct {
	ShortForecast   string
	Temperature     int
	TemperatureUnit string // F or C (I added K for fun)
	// unused response fields not added to struct. See note on Data
}

type Output struct {
	ShortForecast       string `json:"shortForecast"`
	Temperature         int    `json:"temperature"`
	TemperatureUnits    string `json:"temperatureUnits"`
	TemperatureCategory string `json:"temperatureCategory"`
}

func NewOutput(inp Data) (out Output, err error) {
	data := inp.Properties.Periods[0]
	out = Output{
		ShortForecast:    data.ShortForecast,
		Temperature:      data.Temperature,
		TemperatureUnits: data.TemperatureUnit,
	}
	out.TemperatureCategory, err = categorizeTemp(float64(data.Temperature), data.TemperatureUnit)
	return
}

func categorizeTemp(temp float64, units string) (string, error) {
	// Instead of using a temporary variable for TempF, we just overwrite the temp variable (copied by value when passed in)
	switch units {
	case "K":
		// If in Kelvin, then we first convert it to Celsius
		temp -= 273.3
		fallthrough
	case "C":
		temp = (temp * 9.0 / 5.0) + 32
	case "F":
		// Already in the correct units
	default:
		return "", errors.New("unhandled units found on response")
	}
	if temp > 80 {
		return "hot", nil
	}
	if temp < 45 {
		return "cold", nil
	}
	return "moderate", nil
}
