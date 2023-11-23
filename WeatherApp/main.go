package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

type apiConfigData struct {
	OpenWeatherMapApiKey string `json:"OpenWeatherMapApiKey"`
}

type weatherData struct {
	Name string `json:"name"`
	Sys  struct {
		Country string `json:"country"`
	} `json:"sys"`
	Main struct {
		TemperatureCelsius float64 `json:"temp"`
	} `json:"main"`
}

func loadApiConfig(filename string) (apiConfigData, error) {
	bytes, err := os.ReadFile(filename)
	if err != nil {
		return apiConfigData{}, err
	}

	var c apiConfigData
	err = json.Unmarshal(bytes, &c)
	if err != nil {
		return apiConfigData{}, err
	}
	return c, nil
}

func hello(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("hello from go! \n"))
}

func query(city string) (weatherData, error) {
	apiConfig, err := loadApiConfig(".apiConfig")
	if err != nil {
		return weatherData{}, err
	}

	url := "http://api.openweathermap.org/data/2.5/weather?appid=" + apiConfig.OpenWeatherMapApiKey + "&q=" + city
	resp, err := http.Get(url)
	if err != nil {
		return weatherData{}, err
	}
	defer resp.Body.Close()

	// Логирование статуса ответа
	fmt.Printf("Response Status: %s\n", resp.Status)

	// Логирование тела ответа
	body, _ := io.ReadAll(resp.Body)
	fmt.Printf("Response Body: %s\n", body)

	var d weatherData
	if err := json.NewDecoder(bytes.NewReader(body)).Decode(&d); err != nil {
		return weatherData{}, err
	}
	d.Main.TemperatureCelsius = d.Main.TemperatureCelsius - 273.15

	fmt.Printf("URL: %s\n", url)
	fmt.Printf("Response: %+v\n", d)

	return d, nil
}

func main() {
	http.HandleFunc("/hello", hello)

	http.HandleFunc("/weather/",
		func(w http.ResponseWriter, r *http.Request) {
			city := strings.SplitN(r.URL.Path, "/", 3)[2]
			city = strings.ToLower(city) // Приводим к нижнему регистру
			data, err := query(city)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			json.NewEncoder(w).Encode(data)
		})

	http.ListenAndServe(":8080", nil)
}
