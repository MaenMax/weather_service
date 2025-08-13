package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"weather_service/internal/models"
	"weather_service/internal/services"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

// WeatherHandler handles weather-related HTTP requests
type WeatherHandler struct {
	weatherService *services.WeatherService
	logger         *logrus.Logger
}

// NewWeatherHandler creates a new weather handler instance
func NewWeatherHandler(weatherService *services.WeatherService, logger *logrus.Logger) *WeatherHandler {
	return &WeatherHandler{
		weatherService: weatherService,
		logger:         logger,
	}
}

// GetWeatherForecast handles GET requests for weather forecasts
func (h *WeatherHandler) GetWeatherForecast(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	// Extract query parameters
	latStr := r.URL.Query().Get("lat")
	lonStr := r.URL.Query().Get("lon")

	if latStr == "" || lonStr == "" {
		h.sendErrorResponse(w, http.StatusBadRequest, "MISSING_PARAMETERS",
			"Latitude and longitude parameters are required",
			"Please provide both 'lat' and 'lon' query parameters")
		return
	}

	// Parse coordinates
	lat, err := strconv.ParseFloat(latStr, 64)
	if err != nil {
		h.sendErrorResponse(w, http.StatusBadRequest, "INVALID_LATITUDE",
			"Invalid latitude parameter",
			"Latitude must be a valid number between -90 and 90")
		return
	}

	lon, err := strconv.ParseFloat(lonStr, 64)
	if err != nil {
		h.sendErrorResponse(w, http.StatusBadRequest, "INVALID_LONGITUDE",
			"Invalid longitude parameter",
			"Longitude must be a valid number between -180 and 180")
		return
	}

	// Validate coordinate ranges
	if lat < -90 || lat > 90 {
		h.sendErrorResponse(w, http.StatusBadRequest, "INVALID_LATITUDE_RANGE",
			"Latitude out of valid range",
			"Latitude must be between -90 and 90 degrees")
		return
	}

	if lon < -180 || lon > 180 {
		h.sendErrorResponse(w, http.StatusBadRequest, "INVALID_LONGITUDE_RANGE",
			"Longitude out of valid range",
			"Longitude must be between -180 and 180 degrees")
		return
	}

	// Log the request
	h.logger.WithFields(logrus.Fields{
		"lat": lat, "lon": lon, "user_agent": r.UserAgent(), "ip": r.RemoteAddr,
	}).Info("Weather forecast request received")

	// Get weather forecast
	ctx := r.Context()
	forecast, err := h.weatherService.GetWeatherForecast(ctx, lat, lon)
	if err != nil {
		h.logger.WithFields(logrus.Fields{
			"lat": lat, "lon": lon, "error": err.Error(),
		}).Error("Failed to get weather forecast")

		h.sendErrorResponse(w, http.StatusInternalServerError, "WEATHER_SERVICE_ERROR",
			"Failed to retrieve weather data",
			"Please try again later or contact support if the problem persists")
		return
	}

	// Send successful response
	h.sendJSONResponse(w, http.StatusOK, forecast)

	// Log response time
	h.logger.WithFields(logrus.Fields{
		"lat": lat, "lon": lon, "response_time_ms": time.Since(start).Milliseconds(),
	}).Info("Weather forecast request completed successfully")
}

// sendJSONResponse sends a JSON response with proper headers
func (h *WeatherHandler) sendJSONResponse(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "public, max-age=900") // 15 minutes
	w.WriteHeader(statusCode)

	if err := json.NewEncoder(w).Encode(data); err != nil {
		h.logger.WithError(err).Error("Failed to encode JSON response")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

// sendErrorResponse sends a structured error response
func (h *WeatherHandler) sendErrorResponse(w http.ResponseWriter, statusCode int, code, message, details string) {
	errorResp := models.ErrorResponse{
		Error:   code,
		Message: message,
		Details: details,
	}

	h.sendJSONResponse(w, statusCode, errorResp)
}

// RegisterRoutes registers the weather handler routes
func (h *WeatherHandler) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/weather", h.GetWeatherForecast).Methods("GET")
}
