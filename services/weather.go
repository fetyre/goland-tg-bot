package services

import (
	"errors"
	"fmt"
	"time"

	resty "resty.dev/v3"
)

const maxDaily = 999

type WeatherService struct {
	client *resty.Client
	apiKey   string
	location *time.Location
	day      string
	requestCount  int
}

// NewWeatherService создаёт новый экземпляр
func NewWeatherService(apiKey string, loc *time.Location) *WeatherService {
	return &WeatherService{
		apiKey:   apiKey,
		location: loc,
		client: resty.New().SetTimeout( 5 * time.Second ).SetRetryCount( 1 ),
	}
}

// Forecast возвращает описание погоды (строка) и температуру (°C) для заданного города
func (s *WeatherService) GetWeather(lat string, lon string, exclude string, units string) ([]byte, error) {
	if units == "" {
		units = "metric"
	}

	today := time.Now().Format( time.DateOnly )

	if today != s.day {
		s.requestCount = 0
		s.day = time.Now().Format( time.DateOnly )
	}

	if s.requestCount >= maxDaily {
		return nil, errors.New("достигнут дневной лимит запросов (999)")
	}
	s.requestCount++

	url := "https://api.openweathermap.org/data/3.0/onecall"
	res, err := s.client.R().SetQueryParam( "lat", lat ).SetQueryParam( "lon", lon ).SetQueryParam( "appid", s.apiKey ).SetQueryParam( "appid", s.apiKey ).SetQueryParam("exclude", "minutely,hourly,alerts").SetQueryParam("units", units).SetQueryParam( "lang","ru" ).Get( url )
	if err != nil {
		return nil, fmt.Errorf("ошибка выполнения запроса: %w", err)
	}
	if res.IsError() {
		return nil, fmt.Errorf("API вернул статус %s", res.Status())
	}

	return res.Bytes(), nil
}
