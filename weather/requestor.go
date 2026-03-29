package weather

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
)

type Requestor interface {
	GetPointInfo(url string) (Data, error)
	GetForecastUrl(lat, lon float64) (string, error)
}

// ActualRequestor is the default requestor the endpoint uses. Some tests use a mock requestor
type ActualRequestor struct {
	Client *http.Client
}

func (r ActualRequestor) GetForecastUrl(lat, lon float64) (string, error) {
	firstReq, err := http.NewRequest("GET", fmt.Sprintf("https://api.weather.gov/points/%f,%f", lat, lon), nil)
	if err != nil {
		return "", err
	}
	// We want the ld+json response format. Add the appropriate header
	firstReq.Header.Add("Accept", "application/ld+json")
	firstResult, err := r.Client.Do(firstReq)
	if err != nil {
		return "", errors.Join(errRequesting, err)
	}
	if firstResult.StatusCode != 200 {
		return "", errNon200
	}
	defer firstResult.Body.Close()
	bs, err := io.ReadAll(firstResult.Body)
	if err != nil {
		return "", errors.Join(errReading, err)
	}
	firstBody := PointsResponseBody{}
	err = json.Unmarshal(bs, &firstBody)
	if err != nil {
		return "", errors.Join(errUnmarshalling, err)
	}
	return firstBody.ForecastUrl, err
}

func ParseLatLonFromString(s string) (float64, float64, error) {
	vals := strings.Split(s, ",")
	if len(vals) != 2 {
		return 0, 0, errors.New("invalid latlon format: " + s)
	}
	lat, err := strconv.ParseFloat(vals[0], 64)
	if err != nil {
		return 0, 0, errors.New("invalid lat format: " + s)
	}
	lon, err := strconv.ParseFloat(vals[1], 64)
	if err != nil {
		return 0, 0, errors.New("invalid lon format: " + vals[1])
	}
	return lat, lon, nil
}

var (
	errNon200        = errors.New("non-200 response")
	errRequesting    = errors.New("failed when making request")
	errReading       = errors.New("failed ro read response")
	errUnmarshalling = errors.New("failed to unmarshal response")
)

func (r ActualRequestor) GetPointInfo(url string) (Data, error) {
	results := Data{}

	resp, err := r.Client.Get(url)
	if err != nil {
		return results, errors.Join(errRequesting, err)
	}
	if resp.StatusCode != 200 {
		return results, errNon200
	}
	defer resp.Body.Close()
	bs, err := io.ReadAll(resp.Body)
	if err != nil {
		return results, errors.Join(errReading, err)
	}
	err = json.Unmarshal(bs, &results)
	if err != nil {
		return results, errors.Join(errUnmarshalling, err)
	}
	return results, nil
}
