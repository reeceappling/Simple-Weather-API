package weather

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"slices"
	"strconv"
	"strings"
	"testing"
	"time"
)

var _ http.RoundTripper = &mockTransport{}
var _ io.ReadCloser = badReadCloser{}

type Result[T any] struct {
	Val T
	Err error
}

func (res Result[T]) Parts() (T, error) {
	return res.Val, res.Err
}

type badReadCloser struct{}

func (b badReadCloser) Read(p []byte) (n int, err error) {
	return 0, errors.New("read failure")
}

func (b badReadCloser) Close() error {
	return errors.New("close failure")
}

type mockTransport struct {
	mappedResults map[*http.Request]Result[*http.Response]
	variedResults func(req *http.Request) (*http.Response, error)
}

func (m mockTransport) RoundTrip(request *http.Request) (*http.Response, error) {
	if res, exists := m.mappedResults[request]; exists {
		return res.Parts()
	}
	return m.variedResults(request)
}

func TestRequestor(t *testing.T) {
	lat, lon := 35.7596, -79.0193 // Valid lats and lons (Somewhere in NC)
	_, badRequestLon, badCodeLon, badReaderLon, badBodyLon := lon, lon+0.0001, lon+0.0002, lon+0.0003, lon+0.0004
	_, latFirstRequestFail, latSecondRequestFail := lat, lat+0.0001, lat+0.0002
	secondRequestUrlTemplate := `%f,%f`
	mockTransporter := mockTransport{}
	mockTransporter.variedResults = func(req *http.Request) (*http.Response, error) {
		var response = Result[*http.Response]{
			Val: &http.Response{
				StatusCode: 200,
			},
			Err: nil,
		}
		if strings.HasPrefix(req.URL.String(), "https://api.weather.gov/points") {
			parts := strings.Split(req.URL.String(), "/")
			lastPart := parts[len(parts)-1]
			testLat, testLon, err := ParseLatLonFromString(lastPart)
			if err != nil {
				panic(err)
			}
			respBody := PointsResponseBody{ForecastUrl: fmt.Sprintf(secondRequestUrlTemplate, testLat, testLon)}
			bs, err := json.Marshal(respBody)
			if err != nil {
				panic(err)
			}
			response.Val.Body = io.NopCloser(bytes.NewReader(bs))
			if testLat == latFirstRequestFail {
				switch testLon {
				case badRequestLon:
					response.Err = errors.New("failed request in test")
				case badCodeLon:
					response.Val.StatusCode = 500
				case badReaderLon:
					response.Val.Body = badReadCloser{}
				case badBodyLon:
					response.Val.Body = io.NopCloser(strings.NewReader("wrong format response"))
				default:
					// Keep current response.
				}
			}
			return response.Parts()
		}
		// Second request
		testLat, testLon, err := ParseLatLonFromString(req.URL.String())
		if err != nil {
			panic(err)
		}
		respBody := Data{Properties: Properties{Periods: []Period{
			{
				ShortForecast:   "short",
				Temperature:     1,
				TemperatureUnit: "F",
			},
		}}}
		bs, err := json.Marshal(respBody)
		if err != nil {
			panic(err)
		}
		response.Val.Body = io.NopCloser(bytes.NewReader(bs))
		if testLat == latSecondRequestFail {
			switch testLon {
			case badRequestLon:
				response.Err = errors.New("failed request in test")
			case badCodeLon:
				response.Val.StatusCode = 500
			case badReaderLon:
				response.Val.Body = badReadCloser{}
			case badBodyLon:
				response.Val.Body = io.NopCloser(strings.NewReader("wrong format response"))
			default:
				// Keep current response.
			}
		}
		return response.Parts()

	}
	requestor := ActualRequestor{
		Client: &http.Client{
			Transport: mockTransporter,
			Timeout:   5 * time.Second,
		},
	}

	t.Run("ActualRequestor", func(t *testing.T) {
		t.Run("valid", func(t *testing.T) {
			url, err := requestor.GetForecastUrl(lat, lon)
			require.NoError(t, err)
			data, err := requestor.GetPointInfo(url)
			require.NoError(t, err)
			require.NotEmpty(t, data.Properties.Periods)
			firstPd := data.Properties.Periods[0]
			assert.True(t, slices.Contains([]string{"F", "C", "K"}, firstPd.TemperatureUnit))
			assert.NotEmpty(t, firstPd.ShortForecast)
		})
		t.Run("errors", func(t *testing.T) {
			for i, errLat := range []float64{latFirstRequestFail, latSecondRequestFail} {
				t.Run(strconv.Itoa(i), func(t *testing.T) {
					for errLon, expErr := range map[float64]error{badRequestLon: errRequesting, badCodeLon: errNon200, badReaderLon: errReading, badBodyLon: errUnmarshalling} {
						t.Run(expErr.Error(), func(t *testing.T) {
							url, err := requestor.GetForecastUrl(errLat, errLon)
							if i == 0 {
								require.Error(t, err)
								assert.ErrorIs(t, err, expErr)
							} else {
								require.NoError(t, err)
								_, errPointInfo := requestor.GetPointInfo(url)
								require.Error(t, errPointInfo)
								assert.ErrorIs(t, errPointInfo, expErr)
							}
						})
					}
				})
			}
		})
	})
}

func TestMisc(t *testing.T) {
	t.Run("ParseLatLonFromString", func(t *testing.T) {
		t.Run("valid", func(t *testing.T) {
			lat, lon, err := ParseLatLonFromString("-3.7,2.8")
			require.NoError(t, err)
			assert.Equal(t, -3.7, lat)
			assert.Equal(t, 2.8, lon)
		})
		for name, s := range map[string]string{
			"invalid number of commas": "-3.7,,2.8",
			"bad lat":                  "bad,2.8",
			"bad lon":                  "3.7,bad",
		} {
			t.Run(name, func(t *testing.T) {
				_, _, err := ParseLatLonFromString(s)
				require.Error(t, err)
			})
		}
	})
}
