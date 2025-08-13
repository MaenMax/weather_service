package services

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"weather_service/internal/cache"
	"weather_service/internal/models"

	"github.com/sirupsen/logrus"
)

// WeatherService handles weather data retrieval and processing
type WeatherService struct {
	client  *http.Client
	cache   cache.Cache
	logger  *logrus.Logger
	baseURL string
}

// NewWeatherService creates a new weather service instance
func NewWeatherService(baseURL string, timeout time.Duration, cache cache.Cache, logger *logrus.Logger) *WeatherService {
	client := &http.Client{
		Timeout: timeout,
		Transport: &http.Transport{
			MaxIdleConns:        100,
			MaxIdleConnsPerHost: 100,
			IdleConnTimeout:     90 * time.Second,
		},
	}

	return &WeatherService{
		client:  client,
		cache:   cache,
		logger:  logger,
		baseURL: baseURL,
	}
}

// GetWeatherForecast retrieves weather forecast for given coordinates
func (ws *WeatherService) GetWeatherForecast(ctx context.Context, lat, lon float64) (*models.WeatherResponse, error) {
	// Check cache first
	cacheKey := fmt.Sprintf("weather:%f:%f", lat, lon)
	if cached, found := ws.cache.Get(cacheKey); found {
		if response, ok := cached.(*models.WeatherResponse); ok {
			ws.logger.WithFields(logrus.Fields{
				"lat": lat, "lon": lon, "source": "cache",
			}).Debug("Weather data retrieved from cache")
			return response, nil
		}
	}

	// Check if HTTP client is available
	if ws.client == nil {
		return nil, fmt.Errorf("HTTP client not initialized")
	}

	// Get grid points first
	gridPoints, err := ws.getGridPoints(ctx, lat, lon)
	if err != nil {
		return nil, fmt.Errorf("failed to get grid points: %w", err)
	}

	// Get forecast data
	forecast, err := ws.getForecast(ctx, gridPoints)
	if err != nil {
		return nil, fmt.Errorf("failed to get forecast: %w", err)
	}

	// Process and transform the data
	response := ws.transformForecast(lat, lon, forecast)

	// Cache the response
	ws.cache.Set(cacheKey, response, 15*time.Minute)

	return response, nil
}

// getGridPoints retrieves the grid points for given coordinates
func (ws *WeatherService) getGridPoints(ctx context.Context, lat, lon float64) (string, error) {
	url := fmt.Sprintf("%s/points/%.4f,%.4f", ws.baseURL, lat, lon)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("User-Agent", "WeatherService/1.0")
	req.Header.Set("Accept", "application/geo+json")

	resp, err := ws.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("grid points request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var gridResponse struct {
		Properties struct {
			GridID string `json:"gridId"`
			GridX  int    `json:"gridX"`
			GridY  int    `json:"gridY"`
		} `json:"properties"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&gridResponse); err != nil {
		return "", fmt.Errorf("failed to decode grid response: %w", err)
	}

	gridURL := fmt.Sprintf("%s/gridpoints/%s/%d,%d/forecast",
		ws.baseURL, gridResponse.Properties.GridID,
		gridResponse.Properties.GridX, gridResponse.Properties.GridY)

	return gridURL, nil
}

// getForecast retrieves the forecast data from the grid URL
func (ws *WeatherService) getForecast(ctx context.Context, gridURL string) (*models.NationalWeatherServiceResponse, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", gridURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create forecast request: %w", err)
	}

	req.Header.Set("User-Agent", "WeatherService/1.0")
	req.Header.Set("Accept", "application/geo+json")

	resp, err := ws.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make forecast request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("forecast request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var forecast models.NationalWeatherServiceResponse
	if err := json.NewDecoder(resp.Body).Decode(&forecast); err != nil {
		return nil, fmt.Errorf("failed to decode forecast response: %w", err)
	}

	return &forecast, nil
}

// transformForecast transforms NWS data to our response format
func (ws *WeatherService) transformForecast(lat, lon float64, nwsData *models.NationalWeatherServiceResponse) *models.WeatherResponse {
	response := &models.WeatherResponse{
		Location: models.Location{
			Latitude:  lat,
			Longitude: lon,
		},
		GeneratedAt: time.Now(),
	}

	if len(nwsData.Properties.Periods) == 0 {
		return response
	}

	// Get current weather (first period)
	current := nwsData.Properties.Periods[0]
	response.CurrentWeather = models.CurrentWeather{
		Temperature:  float64(current.Temperature),
		TemperatureF: float64(current.Temperature),
		Condition:    ws.normalizeCondition(current.ShortForecast),
		Description:  current.DetailedForecast,
	}

	// Determine temperature type
	response.TemperatureType = ws.determineTemperatureType(float64(current.Temperature))

	// Process forecast periods
	var forecasts []models.Forecast
	for _, period := range nwsData.Properties.Periods {
		startTime, _ := time.Parse(time.RFC3339, period.StartTime)
		endTime, _ := time.Parse(time.RFC3339, period.EndTime)

		forecast := models.Forecast{
			Period:      period.Name,
			StartTime:   startTime,
			EndTime:     endTime,
			Temperature: float64(period.Temperature),
			Condition:   ws.normalizeCondition(period.ShortForecast),
			Description: period.DetailedForecast,
		}

		if period.Probability.Value > 0 {
			forecast.Probability = period.Probability.Value
		}

		forecasts = append(forecasts, forecast)
	}

	response.Forecast = forecasts

	return response
}

// normalizeCondition normalizes weather conditions to standard format
func (ws *WeatherService) normalizeCondition(condition string) string {
	condition = strings.TrimSpace(condition)

	// Check if we have a direct mapping
	if normalized, exists := models.WeatherConditions[condition]; exists {
		return normalized
	}

	// Fallback logic for conditions not in our mapping
	condition = strings.ToLower(condition)
	switch {
	case strings.Contains(condition, "sunny") || strings.Contains(condition, "clear"):
		return "Clear"
	case strings.Contains(condition, "cloudy") || strings.Contains(condition, "overcast"):
		return "Cloudy"
	case strings.Contains(condition, "rain") || strings.Contains(condition, "shower"):
		return "Rain"
	case strings.Contains(condition, "snow") || strings.Contains(condition, "wintry"):
		return "Snow"
	case strings.Contains(condition, "thunder") || strings.Contains(condition, "storm"):
		return "Thunderstorms"
	case strings.Contains(condition, "fog") || strings.Contains(condition, "mist"):
		return "Fog"
	case strings.Contains(condition, "wind"):
		return "Windy"
	default:
		return "Partly Cloudy"
	}
}

// determineTemperatureType categorizes temperature
func (ws *WeatherService) determineTemperatureType(tempF float64) string {
	switch {
	case tempF >= 80:
		return models.TemperatureTypeHot
	case tempF <= 32:
		return models.TemperatureTypeCold
	default:
		return models.TemperatureTypeModerate
	}
}
