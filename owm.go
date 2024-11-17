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

type CurrentWeather struct {
	Weather []struct {
		Main string `json:"main"`
	} `json:"weather"`
	Main struct {
		Actual    float64 `json:"temp"`
		FeelsLike float64 `json:"feels_like"`
	} `json:"main"`
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
	appid, lat, lon, err := readENV(home)
	if err != nil {
		logger.Fatal(err)
	}

	desc, temp, feel, err := owmGetCurrent(appid, lat, lon)
	if err != nil {
		logger.Fatal(err)
	}

	weather := fmt.Sprintf(
		"%s: %d°C %d°C", desc,
		int(math.Round(temp)), int(math.Round(feel)))
	err = os.WriteFile(
		fmt.Sprintf("%s/%s", home, OUT_FILE),
		[]byte(weather), 0600)
	if err != nil {
		logger.Fatal(err)
	}
}

func readENV(home string) (appid, lat, lon string, err error) {
	file, err := os.Open(fmt.Sprintf("%s/%s", home, ENV_FILE))
	if err != nil {
		return
	}
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		kv := strings.Split(scanner.Text(), "=")
		switch kv[0] {
		case "appid":
			appid = kv[1]
		case "lat":
			lat = kv[1]
		case "lon":
			lon = kv[1]
		}
	}
	if err := scanner.Err(); err != nil {
		return "", "", "", err
	}
	return
}

func owmGetCurrent(appid, lat, lon string) (desc string, temp, feel float64, err error) {
	params := url.Values{}
	params.Set("appid", appid)
	params.Set("lat", lat)
	params.Set("lon", lon)
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

	var cw CurrentWeather
	if err := json.NewDecoder(response.Body).Decode(&cw); err != nil {
		return "", 0, 0, err
	}

	desc = cw.Weather[0].Main
	temp = cw.Main.Actual
	feel = cw.Main.FeelsLike
	return
}
