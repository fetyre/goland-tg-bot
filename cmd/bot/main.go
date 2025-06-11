// cmd/bot/main.go
package main

import (
	"log"

	"github.com/joho/godotenv" 
	"tg-bot/internal/bot"     // пакет с handler'ами
	"tg-bot/internal/config"  // пакет для загрузки конфигурации
	"tg-bot/internal/reminders" // in-memory хранилище напоминаний
	"tg-bot/internal/services" // пакеты для API (погода, курс)
)

func main() {

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

	// === 2. Инициализация «слоёв» приложения ===

	// 2.1. Хранилище напоминаний (in-memory)
	//      Заменить на БД: internal/reminders/storage.go
	remStorage := reminders.NewMemoryStorage()

	// 2.2. Клиент OpenWeatherMap
	weatherSvc := services.NewWeatherService(cfg.OpenWeatherAPIKey, cfg.Location)

	// 2.3. Клиент для курса валют
	currencySvc := services.NewCurrencyService()

	// 2.4. Инициализация Telebot с передачей зависимостей в handler-слой
	botApp, err := bot.InitBot(cfg.BotToken, cfg.Location, remStorage, weatherSvc, currencySvc)
	if err != nil {
		log.Fatalf("Ошибка при инициализации BotApp: %v", err)
	}

	// === 3. Запуск фоновых задач ===

	// 3.1. Ежеминутная проверка «календарных» напоминаний
	go botApp.StartReminderChecker()

	// 3.2. Cron-задача для утренней рассылки в 08:00 Europe/Vilnius
	//      Внутри BotApp есть метод для этого
	// botApp.StartMorningBriefCron()

	// === 4. Запуск Telebot (Long Polling) ===
	log.Println("Бот запущен, ожидаем обновления...")
	botApp.StartBot()
}
