package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
)

var (
	OPEN_WEATHER_API_KEY string
)

type WeatherResponse struct {
	Main struct {
		Temp float64 `json:"temp"`
	} `json:"main"`
	Weather []struct {
		Main string `json:"main"`
	} `json:"weather"`
}

func init() {
	OPEN_WEATHER_API_KEY = os.Getenv("OPEN_WEATHER_API_KEY")
	if OPEN_WEATHER_API_KEY == "" {
		panic("OpenWeather API key not providedr")
	}
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

	url := fmt.Sprintf("https://api.openweathermap.org/data/2.5/weather?lat=%s&lon=%s&appid=%s", lat, lon, OPEN_WEATHER_API_KEY)

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
