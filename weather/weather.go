package weather

import "errors"

// For is the main function of the endpoint, but after we've already parsed all necessary data out of the request, so it is easier to test than a handler.
func For(requestor Requestor, lat, lon float64) (Output, error) {
	url, err := requestor.GetForecastUrl(lat, lon)
	if err != nil {
		return Output{}, errors.Join(errors.New("request 1 failed"), err)
	}
	resp, err := requestor.GetPointInfo(url)
	if err != nil {
		return Output{}, errors.Join(errors.New("request 2 failed"), err)
	}
	if len(resp.Properties.Periods) == 0 {
		return Output{}, errors.Join(errors.New("no results found on response"), err)
	}
	return NewOutput(resp)
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
