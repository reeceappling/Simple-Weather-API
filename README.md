# Simple-Weather-API
A simple weather API which I was requested to create.
The endpoint of interest takes a latitude and a longitude, then queries [api.weather.gov](https://api.weather.gov) for the current meteorological data near that location in the USA.
A valid response will contain a ShortForecast explaining the forecast, a Temperature with its associated unit of measure, and a TemperatureCategory string of either "hot", "cold", or "moderate"

## Requirements
- Go __MUST__ be installed in order to compile and/or run this server from source
- An API client such as Bruno or Postman, or curl (for utilizing locally).
- Curl is required only if running the integration tests

# Running Locally
There is an easy way with less user flexibility (see the [Integration tests](#integration-tests) section for the easy way), and the regular way (as described here),

The default port is 9000, but can be changed by providing the -port flag

To run the server from source on port 9001:
```bash
go run . -port 9001
```

The API exposes two endpoints
- __/*__ : the default/root endpoint, which acts as a health check endpoint
- __/weather/{lat},{lon}__ : which allows the user to query for the weather near a specific coordinate.

### Examples
#### Local Server Terminal Output
![Terminal during a Local Run](images/runLocal.png)
#### Bruno Request and Response
![Bruno Results](images/brunoResult.png)

# Running Tests

## Local (Unit) Tests
To run the local tests where everything external is mocked:
```bash
go test ./...
```
#### Example output from unit tests
![Output from testing](images/tests.png)

## Integration Tests
I have also provided a convenience shell script which boots the server, curls the server, and then shuts down the server:
```bash
./run.sh
```
#### Example integration test output
![Output from run.sh](images/localScript.png) 