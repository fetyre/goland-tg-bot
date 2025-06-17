package utils

import (
	"time"
)


type Utils struct {}

func NewUtilsService() *Utils {
	return &Utils{}
}

func (s *Utils) GetRusDayName(t time.Time) string {
	switch t.Weekday() {
		case time.Sunday:
			return "Воскресенье"
		case time.Monday:
			return "Понедельник"
		case time.Tuesday:
			return "Вторник"
		case time.Wednesday:
			return "Среда"
		case time.Thursday:
			return "Четверг"
		case time.Friday:
			return "Пятница"
		case time.Saturday:
			return "Суббота"
		default:
			return ""
	}
}

func (s *Utils) GetRusMonthName(t time.Time) string {
	switch t.Month() {
		case time.January:
			return "января"
		case time.February:
			return "февраля"
		case time.March:
			return "марта"
		case time.April:
			return "апреля"
		case time.May:
			return "мая"
		case time.June:
			return "июня"
		case time.July:
			return "июля"
		case time.August:
			return "августа"
		case time.September:
			return "сентября"
		case time.October:
			return "октября"
		case time.November:
			return "ноября"
		case time.December:
			return "декабря"
		default:
			return ""
	}
}