package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
)

type WeatherResponse struct {
	Main struct {
		Temp float64 `json:"temp"`
	} `json:"main"`
	Weather []struct {
		Main string `json:"main"`
	} `json:"weather"`
}

func kelvinToCelsius(kelvin float64) float64 {
	return kelvin - 273.15
}

func determineWeather(tempCelsius float64) string {
	if tempCelsius > 30 {
		return "hot"
	} else if tempCelsius < 15 {
		return "cold"
	} else {
		return "moderate"
	}
}

func weatherHandler(w http.ResponseWriter, r *http.Request) {
	lat := r.URL.Query().Get("lat")
	lon := r.URL.Query().Get("lon")

	apiKey := os.Getenv("OPEN_WEATHER_API_KEY")
	if apiKey == "" {
		http.Error(w, "OpenWeather API key not provided", http.StatusInternalServerError)
		return
	}

	url := fmt.Sprintf("https://api.openweathermap.org/data/2.5/weather?lat=%s&lon=%s&appid=%s", lat, lon, apiKey)
	resp, err := http.Get(url)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var weatherResp WeatherResponse
	if err := json.Unmarshal(body, &weatherResp); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	tempCelsius := kelvinToCelsius(weatherResp.Main.Temp)
	weatherCondition := determineWeather(tempCelsius)

	response := map[string]string{
		"weather_condition": weatherResp.Weather[0].Main,
		"temperature":       fmt.Sprintf("%.2f°C %s", tempCelsius, weatherCondition),
	}

	jsonResponse, err := json.Marshal(response)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(jsonResponse)
}

func main() {
	http.HandleFunc("/weather", weatherHandler)
	http.ListenAndServe(":8080", nil)
}
