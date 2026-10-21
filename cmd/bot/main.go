package main

import (
	"context"
	"log"
	"net/http"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"payment_bot/internal/config"
	"payment_bot/internal/delivery/telegram"
	"payment_bot/internal/repository"
	"payment_bot/internal/repository/isdayoff"
	"payment_bot/internal/repository/postgres"
	"payment_bot/internal/usecase"
	"payment_bot/migrations"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	ctx := context.Background()

	db, err := postgres.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("database: %v", err)
	}
	defer db.Close()

	if err := postgres.Migrate(ctx, db, migrations.FS); err != nil {
		log.Fatalf("migrations: %v", err)
	}

	calendarCache := postgres.NewCalendarCache(db)
	userRepo := postgres.NewUserRepo(db)

	httpClient := &http.Client{Timeout: 10 * time.Second}
	isDayOffClient := isdayoff.New(httpClient)

	calendarRepo := repository.NewCachingCalendarRepository(calendarCache, isDayOffClient)
	calculator := usecase.NewSalaryCalculator(calendarRepo)

	handler := telegram.NewHandler(calculator, userRepo)

	bot, err := tgbotapi.NewBotAPI(cfg.TelegramToken)
	if err != nil {
		log.Fatalf("telegram bot: %v", err)
	}

	log.Printf("authorized as @%s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message != nil {
			handler.HandleMessage(bot, update.Message)
		}
		if update.CallbackQuery != nil {
			handler.HandleCallback(bot, update.CallbackQuery)
		}
	}
}
