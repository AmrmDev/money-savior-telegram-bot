package main

import (
	"context"
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"money-savior-telegram-bot/app/src/adapter/controller"
	"money-savior-telegram-bot/app/src/adapter/handler"
	"money-savior-telegram-bot/app/src/infra/config"
	infrahttp "money-savior-telegram-bot/app/src/infra/http"
)

func main() {
	ctx := context.Background()

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Load config: %v", err)
	}

	bot, err := tgbotapi.NewBotAPI(cfg.TelegramToken)
	if err != nil {
		log.Fatalf("Create bot: %v", err)
	}

	sender := infrahttp.NewTelegramSender(bot)

	h := handler.NewHandler(
		controller.NewStartController(sender),
		controller.NewHelpController(sender),
		controller.NewRegisterExpenseController(sender),
	)

	updates := bot.GetUpdatesChan(tgbotapi.NewUpdate(0))
	for update := range updates {
		if err := h.Handle(ctx, update); err != nil {
			log.Printf("New update received %d: %v", update.UpdateID, err)
		}
	}
}
