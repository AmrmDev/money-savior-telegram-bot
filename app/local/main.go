package main

import (
	"context"
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"money-savior-telegram-bot/app/src/adapter/controller"
	"money-savior-telegram-bot/app/src/adapter/handler"
	"money-savior-telegram-bot/app/src/core/usecase"
	infraaws "money-savior-telegram-bot/app/src/infra/aws"
	"money-savior-telegram-bot/app/src/infra/aws/repository"
	"money-savior-telegram-bot/app/src/infra/config"
	infrahttp "money-savior-telegram-bot/app/src/infra/http"
	"money-savior-telegram-bot/app/src/infra/id"
)

func main() {
	ctx := context.Background()

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	bot, err := tgbotapi.NewBotAPI(cfg.TelegramToken)
	if err != nil {
		log.Fatalf("create bot: %v", err)
	}

	dynamoClient, err := infraaws.NewDynamoClient(ctx)
	if err != nil {
		log.Fatalf("create dynamo client: %v", err)
	}

	sender := infrahttp.NewTelegramSender(bot)
	expenseRepository := repository.NewExpenseRepository(dynamoClient, cfg.TableName)
	idGenerator := id.NewExpenseIDGenerator()

	registerExpense := usecase.NewRegisterExpense(expenseRepository, idGenerator)
	queryExpenses := usecase.NewQueryExpenses(expenseRepository)

	h := handler.NewHandler(
		controller.NewStartController(sender),
		controller.NewHelpController(sender),
		controller.NewRegisterExpenseController(registerExpense, sender),
		controller.NewQueryExpensesController(queryExpenses, sender),
	)

	updates := bot.GetUpdatesChan(tgbotapi.NewUpdate(0))
	for update := range updates {
		if err := h.Handle(ctx, update); err != nil {
			log.Printf("handle update %d: %v", update.UpdateID, err)
		}
	}
}
