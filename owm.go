package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"net/http"
	"net/url"
	"os"
	"strings"
)

var (
	logger *log.Logger
)

const (
	OUT_FILE = ".curwttr"
	LOG_FILE = ".curwttr_error"
	ENV_FILE = ".curwttr_env"
)

type appConfig struct {
	appid string
	lat   string
	lon   string
}

type weatherData struct {
	temp float64
	feel float64
	desc string
}

func main() {
	home, err := os.UserHomeDir()
	if err != nil {
		log.Fatal(err)
	}

	logFile, err := os.OpenFile(
		fmt.Sprintf("%s/%s", home, LOG_FILE),
		os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}
	defer logFile.Close()

	logger := log.New(logFile, "owm", log.LstdFlags)
	config, err := readENV(home)
	if err != nil {
		logger.Fatal(err)
	}

	data, err := owmGetCurrent(config)
	if err != nil {
		logger.Fatal(err)
	}

	weather := fmt.Sprintf(
		"%s: %d°C %d°C", data.desc,
		int(math.Round(data.temp)), int(math.Round(data.feel)))
	err = os.WriteFile(
		fmt.Sprintf("%s/%s", home, OUT_FILE),
		[]byte(weather), 0600)
	if err != nil {
		logger.Fatal(err)
	}
}

func readENV(home string) (config appConfig, err error) {
	file, err := os.Open(fmt.Sprintf("%s/%s", home, ENV_FILE))
	if err != nil {
		return
	}
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		kv := strings.Split(scanner.Text(), "=")
		switch kv[0] {
		case "appid":
			config.appid = kv[1]
		case "lat":
			config.lat = kv[1]
		case "lon":
			config.lon = kv[1]
		}
	}
	if err := scanner.Err(); err != nil {
		return appConfig{}, err
	}
	return
}

func owmGetCurrent(config appConfig) (data weatherData, err error) {
	params := url.Values{}
	params.Set("appid", config.appid)
	params.Set("lat", config.lat)
	params.Set("lon", config.lon)
	params.Set("units", "metric")

	baseURL := "https://api.openweathermap.org/data/2.5/weather"
	owmURL := fmt.Sprintf("%s?%s", baseURL, params.Encode())
	response, err := http.Get(owmURL)
	if err != nil {
		return
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		logger.Fatal(response.Status)
	}

	var cw struct {
		Weather []struct {
			Main string `json:"main"`
		} `json:"weather"`
		Main struct {
			Actual    float64 `json:"temp"`
			FeelsLike float64 `json:"feels_like"`
		} `json:"main"`
	}
	if err := json.NewDecoder(response.Body).Decode(&cw); err != nil {
		return data, err
	}

	data.desc = cw.Weather[0].Main
	data.temp = cw.Main.Actual
	data.feel = cw.Main.FeelsLike
	return
}
