package models

import "time"

// WeatherRequest represents the incoming weather request
type WeatherRequest struct {
	Latitude  float64 `json:"latitude" validate:"required,min=-90,max=90"`
	Longitude float64 `json:"longitude" validate:"required,min=-180,max=180"`
}

// WeatherResponse represents the weather service response
type WeatherResponse struct {
	Location        Location       `json:"location"`
	CurrentWeather  CurrentWeather `json:"current_weather"`
	Forecast        []Forecast     `json:"forecast"`
	TemperatureType string         `json:"temperature_type"`
	GeneratedAt     time.Time      `json:"generated_at"`
}

// Location represents geographic location information
type Location struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	City      string  `json:"city,omitempty"`
	State     string  `json:"state,omitempty"`
}

// CurrentWeather represents current weather conditions
type CurrentWeather struct {
	Temperature   float64 `json:"temperature_celsius"`
	TemperatureF  float64 `json:"temperature_fahrenheit"`
	Condition     string  `json:"condition"`
	Description   string  `json:"description"`
	Humidity      int     `json:"humidity"`
	WindSpeed     float64 `json:"wind_speed"`
	WindDirection string  `json:"wind_direction"`
	Pressure      float64 `json:"pressure"`
	Visibility    float64 `json:"visibility"`
}

// Forecast represents weather forecast for a specific period
type Forecast struct {
	Period      string    `json:"period"`
	StartTime   time.Time `json:"start_time"`
	EndTime     time.Time `json:"end_time"`
	Temperature float64   `json:"temperature_celsius"`
	Condition   string    `json:"condition"`
	Description string    `json:"description"`
	Probability int       `json:"probability,omitempty"`
}

// NationalWeatherServiceResponse represents the NWS API response
type NationalWeatherServiceResponse struct {
	Properties struct {
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
	} `json:"properties"`
}

// ErrorResponse represents error responses
type ErrorResponse struct {
	Error   string `json:"error"`
	Code    string `json:"code,omitempty"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

// TemperatureType constants
const (
	TemperatureTypeHot      = "hot"
	TemperatureTypeCold     = "cold"
	TemperatureTypeModerate = "moderate"
)

// Weather conditions mapping
var WeatherConditions = map[string]string{
	"Partly Sunny":  "Partly Cloudy",
	"Mostly Sunny":  "Partly Cloudy",
	"Mostly Clear":  "Clear",
	"Mostly Cloudy": "Cloudy",
	"Partly Cloudy": "Partly Cloudy",
	"Cloudy":        "Cloudy",
	"Rain":          "Rain",
	"Snow":          "Snow",
	"Thunderstorms": "Thunderstorms",
	"Fog":           "Fog",
	"Haze":          "Haze",
	"Windy":         "Windy",
}
