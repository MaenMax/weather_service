package services

import (
	"context"
	"testing"
	"time"

	"weather_service/internal/models"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockCache is a mock implementation of the cache interface
type MockCache struct {
	mock.Mock
}

func (m *MockCache) Get(key string) (interface{}, bool) {
	args := m.Called(key)
	return args.Get(0), args.Bool(1)
}

func (m *MockCache) Set(key string, value interface{}, ttl time.Duration) {
	m.Called(key, value, ttl)
}

func (m *MockCache) Delete(key string) {
	m.Called(key)
}

func (m *MockCache) Clear() {
	m.Called()
}

func TestWeatherService_DetermineTemperatureType(t *testing.T) {
	tests := []struct {
		name     string
		tempF    float64
		expected string
	}{
		{"hot temperature", 85.0, models.TemperatureTypeHot},
		{"cold temperature", 25.0, models.TemperatureTypeCold},
		{"moderate temperature", 65.0, models.TemperatureTypeModerate},
		{"boundary hot", 80.0, models.TemperatureTypeHot},
		{"boundary cold", 32.0, models.TemperatureTypeCold},
	}

	service := &WeatherService{}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.determineTemperatureType(tt.tempF)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestWeatherService_NormalizeCondition(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"direct mapping", "Partly Sunny", "Partly Cloudy"},
		{"fallback sunny", "Sunny with clear skies", "Clear"},
		{"fallback cloudy", "Mostly Cloudy conditions", "Cloudy"},
		{"fallback rain", "Light rain showers", "Rain"},
		{"fallback snow", "Heavy snow", "Snow"},
		{"fallback thunder", "Severe thunderstorms", "Thunderstorms"},
		{"fallback fog", "Dense fog", "Fog"},
		{"fallback wind", "Very windy", "Windy"},
		{"default case", "Unknown condition", "Partly Cloudy"},
	}

	service := &WeatherService{}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.normalizeCondition(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestWeatherService_TransformForecast(t *testing.T) {
	service := &WeatherService{}

	// Mock NWS response
	nwsData := &models.NationalWeatherServiceResponse{
		Properties: struct {
			Periods []struct {
				Number           int    `json:"number"`
				Name             string `json:"name"`
				StartTime        string `json:"startTime"`
				EndTime          string `json:"endTime"`
				Temperature      int    `json:"temperature"`
				TemperatureUnit  string `json:"temperatureUnit"`
				ShortForecast    string `json:"shortForecast"`
				DetailedForecast string `json:"detailedForecast"`
				Probability      struct {
					Value int `json:"value"`
				} `json:"probability,omitempty"`
			} `json:"periods"`
		}{
			Periods: []struct {
				Number           int    `json:"number"`
				Name             string `json:"name"`
				StartTime        string `json:"startTime"`
				EndTime          string `json:"endTime"`
				Temperature      int    `json:"temperature"`
				TemperatureUnit  string `json:"temperatureUnit"`
				ShortForecast    string `json:"shortForecast"`
				DetailedForecast string `json:"detailedForecast"`
				Probability      struct {
					Value int `json:"value"`
				} `json:"probability,omitempty"`
			}{
				{
					Number:           1,
					Name:             "Today",
					StartTime:        "2024-01-01T06:00:00-05:00",
					EndTime:          "2024-01-01T18:00:00-05:00",
					Temperature:      75,
					TemperatureUnit:  "F",
					ShortForecast:    "Partly Sunny",
					DetailedForecast: "Partly sunny with a high near 75.",
				},
				{
					Number:           2,
					Name:             "Tonight",
					StartTime:        "2024-01-01T18:00:00-05:00",
					EndTime:          "2024-01-02T06:00:00-05:00",
					Temperature:      55,
					TemperatureUnit:  "F",
					ShortForecast:    "Partly Cloudy",
					DetailedForecast: "Partly cloudy with a low around 55.",
				},
			},
		},
	}

	result := service.transformForecast(40.7128, -74.0060, nwsData)

	// Assertions
	assert.NotNil(t, result)
	assert.Equal(t, 40.7128, result.Location.Latitude)
	assert.Equal(t, -74.0060, result.Location.Longitude)
	assert.Equal(t, models.TemperatureTypeModerate, result.TemperatureType)
	assert.Equal(t, 2, len(result.Forecast))
	assert.Equal(t, "Today", result.Forecast[0].Period)
	assert.Equal(t, "Partly Cloudy", result.CurrentWeather.Condition)
}

func TestWeatherService_GetWeatherForecast_CacheHit(t *testing.T) {
	mockCache := &MockCache{}
	logger := logrus.New()

	service := &WeatherService{
		cache:  mockCache,
		logger: logger,
	}

	// Mock cache hit
	cachedResponse := &models.WeatherResponse{
		Location:        models.Location{Latitude: 40.7128, Longitude: -74.0060},
		TemperatureType: models.TemperatureTypeModerate,
	}

	mockCache.On("Get", "weather:40.712800:-74.006000").Return(cachedResponse, true)

	result, err := service.GetWeatherForecast(context.Background(), 40.7128, -74.0060)

	assert.NoError(t, err)
	assert.Equal(t, cachedResponse, result)
	mockCache.AssertExpectations(t)
}

func TestWeatherService_GetWeatherForecast_CacheMiss(t *testing.T) {
	mockCache := &MockCache{}
	logger := logrus.New()

	// Create a service with a nil client to avoid HTTP calls in tests
	service := &WeatherService{
		cache:  mockCache,
		logger: logger,
		client: nil, // This will cause the HTTP call to fail, which is expected
	}

	// Mock cache miss
	mockCache.On("Get", "weather:40.712800:-74.006000").Return(nil, false)

	// This test expects an error due to nil HTTP client
	_, err := service.GetWeatherForecast(context.Background(), 40.7128, -74.0060)

	// We expect an error here since the HTTP client is nil
	assert.Error(t, err)
	mockCache.AssertExpectations(t)
}
