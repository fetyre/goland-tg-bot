package services

import (
	"errors"
	"fmt"
	"log"
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
	log.Println( "Start get weather" )
	if units == "" {
		units = "metric"
	}

	today := time.Now().Format( time.DateOnly )

	if today != s.day {
		s.requestCount = 0
		s.day = time.Now().Format( time.DateOnly )
	}

	if s.requestCount >= maxDaily {
		log.Println( "Daily limit reached" )
		return nil, errors.New("достигнут дневной лимит запросов (999)")
	}
	s.requestCount++

	rrres, err := s.client.R().
    Get("https://httpbin.org/get")
	log.Println("🔗 → тестовый запрос к httpbin.org")
if err != nil {
    log.Printf("❌ тестовый запрос упал: %v", err)
} else {
    log.Printf("🔗 ← тестовый ответ httpbin: %v", rrres.StatusCode())
}


	url := "https://api.openweathermap.org/data/3.0/onecall"
	log.Println( "wearher: send get requests for info, key: %v", s.apiKey )
	res, err := s.client.R().SetQueryParam( "lat", lat ).SetQueryParam( "lon", lon ).SetQueryParam( "appid", s.apiKey ).SetQueryParam("exclude", "minutely,hourly,alerts").SetQueryParam("units", units).SetQueryParam( "lang","ru" ).Get( url )
	log.Println( "wearher: successfully end requests for info" )
	if err != nil {
		log.Printf( "Error http get req for weather info, err: %v", err )
		return nil, fmt.Errorf("ошибка выполнения запроса: %w", err)
	}
	if res.IsError() {
		log.Printf( "Http error get weather, err: %s", res.Status() )
		return nil, fmt.Errorf( "API вернул статус %s", res.Status() )
	}

	log.Println( "End get weathe, OK!" )
	return res.Bytes(), nil
}
