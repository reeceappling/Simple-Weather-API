package main

import (
	"appli.ng/simple_weather_api/logging"
	"appli.ng/simple_weather_api/weather"
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"github.com/google/uuid"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"
)

func main() {
	log := logging.LoggerFactoryFor("WeatherAPI")
	ctx := logging.SetLogger(context.Background(), log)
	port := flag.Int("port", 9000, "port to host the api on")
	flag.Parse()
	log.Infof("Starting WeatherAPI on port %d", *port)
	createAndRunServer(ctx, *port)
}

func createAndRunServer(ctx context.Context, port int) {
	log := logging.GetLogger(ctx)

	var err error = nil
	appName := "Weather API"

	// defer a panic-recovery gofunc, so we can properly log stacktraces if a panic were to occur
	defer func() {
		if r := recover(); r != nil {
			stackTrace := make([]byte, 1024)
			bytesRead := runtime.Stack(stackTrace, false)
			keyAndValues := []interface{}{
				"message", fmt.Sprint(r),
			}
			if bytesRead > 0 {
				keyAndValues = append(keyAndValues, "stackTrace", string(stackTrace[:bytesRead]))
			}
			log.Panicw("API Panic", keyAndValues...)
		}
	}()

	requestTimeout := 30 * time.Second
	srv := &http.Server{
		Addr:              fmt.Sprintf(`:%d`, port),
		ReadHeaderTimeout: requestTimeout,
	}
	sigterm := make(chan os.Signal, 1)
	signal.Notify(sigterm, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		switch v := <-sigterm; v {
		case syscall.SIGTERM:
			log.Info("Received SIGTERM Signal")
		case syscall.SIGINT:
			log.Info("Received SIGINT Signal")
		default:
			log.Warnw("Received unhandled Signal. Stopping", "signal", v)
		}
		log.Info("Shutting server down")
		// Typically I have a slight preference for using ECS clusters for hosting applications. 30 seconds is the default ecs shutdown timer
		srvCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		if err = srv.Shutdown(srvCtx); err != nil {
			log.Errorw("Error Shutting Down", "error", err)
		}
		if err = srv.Close(); err != nil {
			log.Errorw("Error Closing Server", "error", err)
		}
		// abnormal exit code so we can tell at a glance that it was spun down via signal
		os.Exit(42)
	}()

	http.Handle("/", middleware("/", rootHandler()))
	http.Handle("/weather/{latLon}", middleware("/weather/*", getWeatherHandler(nil)))

	log.Infow("Starting "+appName, "port", port, "cpus", runtime.NumCPU())
	err = srv.ListenAndServe()
	if errors.Is(err, http.ErrServerClosed) {
		log.Infow("Server Closed")
	} else {
		log.Fatal(appName+" Server Error", "error", err)
	}
}

// middleware is a middleware for adding request id and routeNames to the logger
func middleware(routeName string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		log := logging.GetLogger(ctx).
			WithRequestId(uuid.New().String()).
			WithRequestEndpointName(routeName)
		next.ServeHTTP(w, r.WithContext(logging.SetLogger(ctx, log)))
	})
}

// rootHandler is a basic health check endpoint which also functions as the default handler
func rootHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, err := w.Write([]byte("server is running")); err != nil {
			logging.GetLogger(r.Context()).Errorw("Error writing response", "error", err)
		}
	})
}

// getWeatherFor is the main function of the endpoint, but after we've already parsed all necessary data out of the request, so it is easier to test than a handler.
func getWeatherFor(requestor weather.Requestor, lat, lon float64) (weather.Output, error) {
	url, err := requestor.GetForecastUrl(lat, lon)
	if err != nil {
		return weather.Output{}, errors.Join(errors.New("request 1 failed"), err)
	}
	resp, err := requestor.GetPointInfo(url)
	if err != nil {
		return weather.Output{}, errors.Join(errors.New("request 2 failed"), err)
	}
	if len(resp.Properties.Periods) == 0 {
		return weather.Output{}, errors.Join(errors.New("no results found on response"), err)
	}
	data := resp.Properties.Periods[0]
	category, err := categorizeTemp(float64(data.Temperature), data.TemperatureUnit)
	if err != nil {
		return weather.Output{}, errors.Join(errors.New("failed to categorize temperature"), err)
	}
	return weather.Output{
		ShortForecast:       data.ShortForecast,
		Temperature:         data.Temperature,
		TemperatureUnits:    data.TemperatureUnit,
		TemperatureCategory: category,
	}, nil
}

// getWeatherHandler should only be passed a non-nil requestor if you don't want to use the default (for example in unit tests)
func getWeatherHandler(requestor weather.Requestor) http.Handler {
	// Use the default requestor if current is nil
	if requestor == nil {
		requestor = weather.ActualRequestor{Client: http.DefaultClient}
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log := logging.GetLogger(r.Context())
		lat, lon, err := weather.ParseLatLonFromString(r.PathValue("latLon"))
		if err != nil {
			log.Errorw("Failed to get latLon from request", "error", err)
			http.Error(w, "Failed to get latLon from request", http.StatusBadRequest)
			return
		}
		out, err := getWeatherFor(requestor, lat, lon)
		if err != nil {
			log.Errorw("Failed to get weather for latLon", "error", err)
			http.Error(w, "Failed to get weather for latLon", http.StatusInternalServerError)
			return
		}
		bs, err := json.Marshal(out)
		if err != nil {
			log.Errorw("Failed to marshal result", "error", err)
			http.Error(w, "Failed to marshal result", http.StatusInternalServerError)
			return
		}
		_, err = w.Write(bs)
		if err != nil {
			log.Errorw("failed to write response", "error", err)
			return
		}
	})
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
