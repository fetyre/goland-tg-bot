package services

import (
	"fmt"
	"time"
	resty "resty.dev/v3"
)


type CurrencyService struct {
	client *resty.Client
}


type CurrencyType string

const (
	USD CurrencyType = "USD"
	EUR CurrencyType = "EUR"
	RUB CurrencyType = "RUB"
)

func NewCurrencyService() *CurrencyService {
	return &CurrencyService{
		client: resty.New().SetTimeout( 5 * time.Second ).SetRetryCount( 1 ),
	}
}

func (s *CurrencyService) GetCurrency(t CurrencyType) ( float64, error ) {

	switch t {
    case USD, EUR, RUB:
    default:
			return 0.0, fmt.Errorf( "неизвестная валюта: %v", t )
  }

	url := fmt.Sprintf( "https://api.nbrb.by/exrates/rates/%v?parammode=2", t  )

	var data struct {
		CurOfficialRate float64 `json:"Cur_OfficialRate"`
	}

	res, err := s.client.R().SetResult( &data ).Get( url )
	if err != nil {
		return 0.0, fmt.Errorf( "ошибка выполнения запроса: %v", err )
	}

	if res.IsError() {
		return 0.0, fmt.Errorf( "сейчас почему-то не получается получить данные об курсе, код ошибки: %v", res.Status()  )
	}

	return data.CurOfficialRate, nil
}

