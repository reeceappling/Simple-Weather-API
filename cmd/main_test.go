package main

import (
	"appli.ng/simple_weather_api/weather"
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

type mockRequestor struct{}

func (r mockRequestor) GetForecastUrl(lat, lon float64) (string, error) {
	return "mockUrl", nil
}
func (r mockRequestor) GetPointInfo(url string) (weather.Data, error) {
	wd := weather.Data{}
	wd.Properties = struct {
		Periods []weather.Period
	}{
		Periods: []weather.Period{
			{
				ShortForecast:   "a short forecast",
				Temperature:     70,
				TemperatureUnit: "F",
			},
		},
	}
	return wd, nil
}

const (
	shortForecast = "a short forecast"
	tempMod       = 70
	tempUnits     = "F"
	catMod        = "moderate"
)

func TestServer(t *testing.T) {
	lat, lon := 35.7596, -79.0193 // Valid lats and lons (Somewhere in NC)
	exp := weather.Output{
		ShortForecast:       shortForecast,
		Temperature:         tempMod,
		TemperatureUnits:    tempUnits,
		TemperatureCategory: catMod,
	}
	mr := mockRequestor{}
	t.Run("more succinct testing format", func(t *testing.T) {
		act, err := GetFor(mr, lat, lon)
		require.NoError(t, err)
		require.Equal(t, exp.ShortForecast, act.ShortForecast)
		require.Equal(t, exp.Temperature, act.Temperature)
		require.Equal(t, exp.TemperatureUnits, act.TemperatureUnits)
		require.Equal(t, exp.TemperatureCategory, act.TemperatureCategory)
	})

	t.Run("More thorough testing format)", func(t *testing.T) {
		act := weather.Output{}

		handler := getWeatherHandler(mr)
		testReq := &http.Request{}
		testReq.SetPathValue("latLon", fmt.Sprintf(`%.2f,%.2f`, lat, lon))
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, testReq)
		resp := w.Result()
		require.Equal(t, 200, resp.StatusCode)
		bs, err := io.ReadAll(resp.Body)
		require.NoError(t, err)
		require.NoError(t, json.Unmarshal(bs, &act))
		require.Equal(t, exp.ShortForecast, act.ShortForecast)
		require.Equal(t, exp.Temperature, act.Temperature)
		require.Equal(t, exp.TemperatureUnits, act.TemperatureUnits)
		require.Equal(t, exp.TemperatureCategory, act.TemperatureCategory)
	})
}
