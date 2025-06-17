// api/webhook.go
package handler

import (
	"log"
	"net/http"

	"github.com/fetyre/goland-tg-bot/bot"
	"github.com/fetyre/goland-tg-bot/config"
	"github.com/fetyre/goland-tg-bot/reminders"
	"github.com/fetyre/goland-tg-bot/services"
	"github.com/fetyre/goland-tg-bot/utils"
	"github.com/joho/godotenv"
)

// глобальный экземпляр приложения – инициализируется на cold-start
var app *bot.BotApp

func init() {

	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found or failed to load, assuming environment vars are set manually")
	}

	// === 1. Загрузка конфигурации ===
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Ошибка при загрузке конфигурации: %v", err)
	}

	// Проверяем, что обязательные переменные заданы
	if cfg.BotToken == "" {
		log.Fatal("TELEGRAM_BOT_TOKEN не задан в окружении")
	}
	if cfg.OpenWeatherAPIKey == "" {
		log.Fatal("Погода не задан в окружении")
	}

	storage   := reminders.NewMemoryStorage()
	weather   := services.NewWeatherService(cfg.OpenWeatherAPIKey, cfg.Location)
	currency  := services.NewCurrencyService()
	utilsSvc  := utils.NewUtilsService()

	app, err = bot.InitBot(
		cfg.BotToken,
		cfg.Location,
		storage,
		weather,
		currency,
		utilsSvc,
	)
	if err != nil {
		panic(err)         // Vercel покажет stack-trace в логах
	}
}

// Handle – единственная точка входа Vercel-функции
func Handle(w http.ResponseWriter, r *http.Request) {
	app.ServeHTTP(w, r)   // метод, который мы добавили в BotApp
}
