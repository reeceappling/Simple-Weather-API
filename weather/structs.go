package weather

// PointsResponseBody is the format of the response retrieved within Requestor.GetForecastUrl
type PointsResponseBody struct {
	ForecastUrl string `json:"forecast"`
	// unused response fields not added to struct
	// See full schema at https://www.weather.gov/documentation/services-web-api#/default/point
}

// Data is the format of the response retrieved within Requestor.GetPointInfo
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

// Output is the format of the non-erroneous output from our weather api endpoint
type Output struct {
	ShortForecast       string `json:"shortForecast"`
	Temperature         int    `json:"temperature"`
	TemperatureUnits    string `json:"temperatureUnits"`
	TemperatureCategory string `json:"temperatureCategory"`
}

func NewOutput(inp Data) (out Output, err error) {
	vals := inp.Properties.Periods[0]
	out = Output{
		ShortForecast:    vals.ShortForecast,
		Temperature:      vals.Temperature,
		TemperatureUnits: vals.TemperatureUnit,
		//TemperatureCategory is set next, but not on initialization
	}
	out.TemperatureCategory, err = categorizeTemp(float64(vals.Temperature), vals.TemperatureUnit)
	return
}
