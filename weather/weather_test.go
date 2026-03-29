package weather

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestWeather(t *testing.T) {
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
	t.Run("categorize temp", func(t *testing.T) {
		c, err := categorizeTemp(0, "C")
		require.NoError(t, err)
		require.Equal(t, "cold", c)

		c, err = categorizeTemp(100, "C")
		require.NoError(t, err)
		require.Equal(t, "hot", c)

		c, err = categorizeTemp(20, "C")
		require.NoError(t, err)
		require.Equal(t, "moderate", c)

		c, err = categorizeTemp(20, "F")
		require.NoError(t, err)
		require.Equal(t, "cold", c)
	})
}
