package config

import (
	"log"
	"os"
	"time"
)

type Config struct {
	BotToken          string
	OpenWeatherAPIKey string
	Location          *time.Location
}

func LoadConfig() ( *Config, error ) {

	botToken := os.Getenv("TG_BOT_TOKEN")
	openWeatherAPIKey := os.Getenv("OPEN_API_KEY")

	loc, err := time.LoadLocation("Europe/Moscow")
	if err != nil {
		log.Fatal("Ошибка получения локали")
		return nil, err
	}

	return &Config{
		BotToken:          botToken,
		OpenWeatherAPIKey: openWeatherAPIKey,
		Location:          loc,
	}, nil
}