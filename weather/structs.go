package weather

type PointsResponseBody struct {
	ForecastUrl string `json:"forecast"`
}

type Data struct {
	Properties Properties
	// unused response fields not added to struct
}

type Properties struct {
	Periods []Period
}

type Period struct {
	ShortForecast   string
	Temperature     int
	TemperatureUnit string // F or C (I added K for fun)
	// unused response fields not added to struct
}

type Output struct {
	ShortForecast       string `json:"shortForecast"`
	Temperature         int    `json:"temperature"`
	TemperatureUnits    string `json:"temperatureUnits"`
	TemperatureCategory string `json:"temperatureCategory"`
}
