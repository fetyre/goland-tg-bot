// internal/bot/handler.go
package bot

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	// "github.com/robfig/cron/v3"
	tele "gopkg.in/telebot.v4"

	"tg-bot/internal/reminders"
	"tg-bot/internal/services"
	// resty "resty.dev/v3"
)

// BotApp хранит зависимости (бот, локацию, хранилище, сервисы)
type BotApp struct {
	bot         *tele.Bot
	location    *time.Location
	storage     reminders.Storage
	weatherSvc  *services.WeatherService
	currencySvc *services.CurrencyService
}


func InitBot( botToken string, location *time.Location, storage reminders.Storage, weatherSvc *services.WeatherService, currencySvc *services.CurrencyService ) ( *BotApp, error ) {
	bot, err := tele.NewBot( 
		tele.Settings{
			Token:  botToken,
			Poller: &tele.LongPoller{ Timeout: 10 * time.Second },
		},
	)

	if err != nil {
		log.Fatal( "Ошибка создания бота: %", err )
		return nil, err
	}

	app := &BotApp{
		bot:         bot,
		location:    location,
		storage:     storage,
		weatherSvc:  weatherSvc,
		currencySvc: currencySvc,
	}

	app.registerHandlers()

	return app, nil
}

// Start запускает Long Polling
func ( app *BotApp ) StartBot() {
	app.bot.Start()
}

// registerHandlers настраивает все команды и колбеки
func ( app *BotApp ) registerHandlers() {

	var (
		keyboardMenu = &tele.ReplyMarkup{ResizeKeyboard: true}
		lineMenu = &tele.ReplyMarkup{}

		currencyBtn = keyboardMenu.Text( "Узнать курс валюты" )
		currencyUSDBtn = keyboardMenu.Text( "USD" )
		currencyRUBBtn = keyboardMenu.Text( "RU" )
		currencyEURBtn = keyboardMenu.Text( "EUR" )
		currencyALLBtn = keyboardMenu.Text( "USD, RU, EUR" )
		weatherBtn = keyboardMenu.Text( "Узнать погоду" )
		weatherCurrentDayBtn = keyboardMenu.Text( "Узнать погоду на день" )
		weatherCurrentBtn = keyboardMenu.Text( "Узнать текущую погоду" )

		moneyBtn = keyboardMenu.Text( "Посмотреть затраты" )
	)

	setDeafultKeyboard := func() {
		keyboardMenu.Reply( keyboardMenu.Row( weatherBtn, moneyBtn, currencyBtn ) )
	}

	getRate := func( t services.CurrencyType, c tele.Context, sendMsg bool ) (string, error) {
		res, err := app.currencySvc.GetCurrency( t )
		if err != nil {
			c.Send( err )
			// keyboardMenu.Reply( keyboardMenu.Row( weatherBtn, moneyBtn, currencyBtn ) )
			return "", c.Send( "Чем хотите возпользоваться?", keyboardMenu )
		}

		rateMsg := fmt.Sprintf( "Текущий курс %v: %v BYN", t, res )

		if sendMsg {
			c.Send( rateMsg )
			keyboardMenu.Reply( keyboardMenu.Row( currencyUSDBtn, currencyRUBBtn, currencyEURBtn ), keyboardMenu.Row( currencyALLBtn ) )
			return "", c.Send( "Выберите курс:", keyboardMenu )
		}

		return rateMsg, nil
	}

	app.bot.Handle("/hello", func(c tele.Context) error {
    helpBtn := lineMenu.Data( "⚙ Settings111", "setting" )

    keyboardMenu.Reply( keyboardMenu.Row( weatherBtn, moneyBtn, currencyBtn ) )
    lineMenu.Inline( lineMenu.Row(helpBtn) )     // одна строка Inline-кнопок
		c.Send( "С возвращением!\nМожет нужны настройки?", lineMenu )
		return c.Send( "Чем хотите возпользоваться?", keyboardMenu )
	})

	app.bot.Handle( &currencyBtn, func(c tele.Context) error {
		keyboardMenu.Reply( keyboardMenu.Row( currencyUSDBtn, currencyRUBBtn, currencyEURBtn ), keyboardMenu.Row( currencyALLBtn ) )
    return c.Send( "Выберите курс:", keyboardMenu )
	})

	app.bot.Handle( &currencyUSDBtn, func(c tele.Context)  error {
		_, err := getRate("USD", c, true)
		return err
	})

	app.bot.Handle( &currencyRUBBtn, func(c tele.Context)  error {
		_, err := getRate("RUB", c, true)
		return err
	})

	app.bot.Handle( &currencyEURBtn, func(c tele.Context)  error {
		_, err := getRate("EUR", c, true)
		return err
	})

	app.bot.Handle( &currencyALLBtn, func(c tele.Context)  error {
		eur, errEur := getRate( "EUR", c, false )
		usd, errUsd := getRate( "USD", c, false )
		rub, errRub := getRate( "RUB", c, false )
		if errEur != nil || errUsd != nil || errRub != nil {
			return errEur
		}

		msg := fmt.Sprintf( "%v\n%v\n%v\n", eur, usd, rub )
		c.Send(msg)
		setDeafultKeyboard()
		return c.Send( "Чем хотите возпользоваться?", keyboardMenu )
	})

	app.bot.Handle( &weatherBtn, func( c tele.Context ) error {
		keyboardMenu.Reply( keyboardMenu.Row( weatherCurrentDayBtn, weatherCurrentBtn ) )
		return c.Send( "Выберете промежуток", keyboardMenu )
	})

	app.bot.Handle( &weatherCurrentDayBtn, func(c tele.Context) error {

		data, err := app.weatherSvc.GetWeather("55.139235", "27.6845787")
    if err != nil {
			c.Send(err.Error())
			return c.Send("Чем хотите воспользоваться?", keyboardMenu)
    }

		type Weather struct {
			Daily string   `json:"daily"`
			Age  int      `json:"age"`
			Tags []string `json:"tags"`
		}

    // Сформируем JSON
    jsonBytes, err := json.MarshalIndent(data, "", "  ")
    if err != nil {
        return c.Send("Ошибка форматирования JSON")
    }
    full := string(jsonBytes)

    // Разбиваем на куски по 3800 р/з>4000 на всякий и отправляем
    const maxMsgLen = 3800
    for _, part := range splitInChunks(full, maxMsgLen) {
        if err := c.Send(part); err != nil {
            return err
        }
    }
    return nil
	})

	// --------------- 2) /help ---------------
	app.bot.Handle("/help", func(c tele.Context) error {
		m := c.Message()
		helpText := `/remind YYYY-MM-DD HH:MM текст
    — установить напоминание (пример: /remind 2025-06-20 15:30 Купить цветы)

 /subscribe
    — подписаться на ежедневную утреннюю сводку (08:00 Europe/Vilnius)

 /unsubscribe
    — отписаться от утренней сводки`
		return c.Send(m.Sender, helpText)
	})

	// --------------- 3) /remind ---------------
	app.bot.Handle("/remind", func(c tele.Context) error {
		m := c.Message()
		payload := m.Payload // всё после "/remind "

		if payload == "" {
			return c.Send(m.Sender, "Неверный формат. Используйте:\n/remind YYYY-MM-DD HH:MM текст")
		}

		// Разбиваем payload на 3 части: [date, time, текст]
		parts := splitNSpaces(payload, 3)
		if len(parts) < 3 {
			return c.Send(m.Sender,
				"Нужно указать: дату, время и текст. Пример:\n/remind 2025-06-20 15:30 Купить цветы")
		}
		dateStr := parts[0] // "2025-06-20"
		timeStr := parts[1] // "15:30"
		text := parts[2]    // "Купить цветы"

		// Объединяем и парсим в time.Time
		datetimeStr := fmt.Sprintf("%s %s", dateStr, timeStr)
		remindTime, err := time.ParseInLocation("2006-01-02 15:04", datetimeStr, app.location)
		if err != nil {
			// Уже отправили ответ пользователю, поэтому возвращаем nil
			c.Send(m.Sender, "Не удалось распарсить дату/время. Формат: YYYY-MM-DD HH:MM.")
			return nil
		}
		if remindTime.Before(time.Now().In(app.location)) {
			c.Send(m.Sender, "Эта дата уже прошла. Укажите время в будущем.")
			return nil
		}

		// Создаём напоминание и сохраняем
		rem := reminders.Reminder{
			ChatID: m.Chat.ID,
			Text:   text,
			Time:   remindTime,
		}
		if err := app.storage.Add(rem); err != nil {
			log.Printf("Ошибка при добавлении напоминания в хранилище: %v", err)
			c.Send(m.Sender, "Не удалось сохранить напоминание. Попробуйте позже.")
			return nil
		}

		confirm := fmt.Sprintf("Напоминание установлено на %s:\n«%s»",
			remindTime.Format("2006-01-02 15:04"), text)
		return c.Send(m.Sender, confirm)
	})

	// --------------- 4) /subscribe ---------------
	app.bot.Handle("/subscribe", func(c tele.Context) error {
		m := c.Message()
		subscribers[m.Chat.ID] = struct{}{}
		return c.Send(m.Sender, "Вы подписаны на утреннюю сводку (08:00 Europe/Vilnius).")
	})

	// --------------- 5) /unsubscribe ---------------
	app.bot.Handle("/unsubscribe", func(c tele.Context) error {
		m := c.Message()
		delete(subscribers, m.Chat.ID)
		return c.Send(m.Sender, "Вы отписаны от утренней сводки.")
	})
}

// splitNSpaces разбивает строку s на N полей по пробелам, склеивая остаток в последний элемент.
func splitNSpaces(s string, n int) []string {
	fields := strings.Fields(s)
	if len(fields) <= n {
		return fields
	}
	result := make([]string, n)
	copy(result, fields[:n-1])
	// Всё, что после N-1-го пробела, склеиваем в одну строку
	result[n-1] = strings.Join(fields[n-1:], " ")
	return result
}

// ========================================
// Глобальная карта подписчиков утренней сводки
// ========================================
var subscribers = make(map[int64]struct{})

// StartReminderChecker запускает горутину, которая каждую минуту проверяет,
// есть ли «сработавшие» напоминания, и отправляет их пользователям.
func (app *BotApp) StartReminderChecker() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for tickTime := range ticker.C {
		now := tickTime.In(app.location)
		due := app.storage.FetchDue(now) // берём все напоминания, у которых r.Time <= now

		for _, r := range due {
			_, err := app.bot.Send(&tele.Chat{ID: r.ChatID}, fmt.Sprintf("⌛ Напоминание: %s", r.Text))
			if err != nil {
				log.Printf("Не удалось отправить напоминание пользователю %d: %v", r.ChatID, err)
			}
		}
	}
}

// // StartMorningBriefCron настраивает cron-задачу «каждый день в 08:00 Europe/Vilnius»
// func (app *BotApp) StartMorningBriefCron() {
// 	c := cron.New(
// 		cron.WithLocation(app.location),
// 		cron.WithParser(cron.NewParser(
// 			cron.SecondOptional|cron.Minute|cron.Hour|cron.Dom|cron.Month|cron.Dow,
// 		)),
// 	)

// 	// “0 8 * * *” — каждый день в 08:00:00
// 	_, err := c.AddFunc("0 8 * * *", func() {
// 		app.sendMorningBrief()
// 	})
// 	if err != nil {
// 		log.Fatalf("Не удалось добавить cron-задачу: %v", err)
// 	}
// 	c.Start()
// }

// sendMorningBrief формирует и рассылает утреннюю сводку всем подписчикам
// func (app *BotApp) sendMorningBrief() {
// 	if len(subscribers) == 0 {
// 		return
// 	}

// 	dateStr := time.Now().In(app.location).Format("2006-01-02")

// 	// 1) Погода в Вильнюсе
// 	weatherDesc, temp, err := app.weatherSvc.Forecast("Vilnius")
// 	if err != nil {
// 		log.Printf("Ошибка при получении погоды: %v", err)
// 		weatherDesc = "нет данных"
// 		temp = 0
// 	}

// 	// 2) Курс EUR → USD
// 	rate, err := app.currencySvc.RateEURtoUSD()
// 	if err != nil {
// 		log.Printf("Ошибка при получении курса валют: %v", err)
// 		rate = 0
// 	}

// 	body := fmt.Sprintf(
// 		"🌞 Доброе утро! Сегодня %s\n\n"+
// 			"🌡 Погода в Вильнюсе: %s, %.1f°C\n"+
// 			"💱 Курс EUR→USD: %.4f\n",
// 		dateStr, weatherDesc, temp, rate,
// 	)

// 	for chatID := range subscribers {
// 		_, err := app.bot.Send(&tele.Chat{ID: chatID}, body)
// 		if err != nil {
// 			log.Printf("Не удалось отправить утреннюю сводку пользователю %d: %v", chatID, err)
// 		}
// 	}
// }


// func( app *BotApp ) getRate( t services.CurrencyType, msg string, c tele.Context, keyboardMenu *tele.ReplyMarkup, sendMsg bool ) (error, string){
// 	res, err := app.currencySvc.GetCurrency( t )
// 	if err != nil {
// 		c.Send( err )
// 		// keyboardMenu.Reply( keyboardMenu.Row( weatherBtn, moneyBtn, currencyBtn ) )
// 		return c.Send( "Чем хотите возпользоваться?", keyboardMenu ), ""
// 	}

// 	rateMsg := fmt.Sprintf( "Текущий курс %v: %v BYN", t, res )

// 	if sendMsg == true {
// 		c.Send( rateMsg )
// 		keyboardMenu.Reply( keyboardMenu.Row( currencyUSDBtn, currencyRUBBtn, currencyEURBtn ), keyboardMenu.Row( currencyALLBtn ) )
// 		return c.Send( "Выберите курс:", keyboardMenu ), ""
// 	}

// 	return nil, rateMsg
// }

func splitInChunks(s string, maxLen int) []string {
    var chunks []string
    for len(s) > maxLen {
        // ищем последний перенос строки в пределах maxLen,
        // чтобы не резать текст внутри строки
        cut := strings.LastIndex(s[:maxLen], "\n")
        if cut <= 0 {
            cut = maxLen
        }
        chunks = append(chunks, s[:cut])
        s = s[cut:]
    }
    if len(s) > 0 {
        chunks = append(chunks, s)
    }
    return chunks
}