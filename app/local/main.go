package main

import (
	"context"
	"log/slog"
	"os"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"money-savior-telegram-bot/app/src/adapter/controller"
	"money-savior-telegram-bot/app/src/adapter/handler"
	"money-savior-telegram-bot/app/src/core/usecase"
	infraaws "money-savior-telegram-bot/app/src/infra/aws"
	"money-savior-telegram-bot/app/src/infra/aws/repository"
	"money-savior-telegram-bot/app/src/infra/config"
	infrahttp "money-savior-telegram-bot/app/src/infra/http"
	"money-savior-telegram-bot/app/src/infra/id"
	"money-savior-telegram-bot/app/src/infra/logger"
	"money-savior-telegram-bot/app/src/infra/groq"
)

func main() {
	ctx := context.Background()

	cfg, err := config.Load()
	if err != nil {
		fatal(slog.Default(), "load config", err)
	}

	log := logger.New(cfg.LogLevel, cfg.LogFormat, cfg.TelegramToken, cfg.GroqAPIKey)

	slog.SetDefault(log)

	bot, err := tgbotapi.NewBotAPI(cfg.TelegramToken)
	if err != nil {
		fatal(log, "create bot", err)
	}

	dynamoClient, err := infraaws.NewDynamoClient(ctx)
	if err != nil {
		fatal(log, "create dynamo client", err)
	}

	sender := infrahttp.NewTelegramSender(bot)
	expenseRepository := repository.NewExpenseRepository(dynamoClient, cfg.TableName)
	idGenerator := id.NewExpenseIDGenerator()

	registerExpense := usecase.NewRegisterExpense(expenseRepository, idGenerator)
	queryExpenses := usecase.NewQueryExpenses(expenseRepository)

	audioDownloader := infrahttp.NewTelegramAudioDownloader(bot)
	transcriber := groq.NewTranscriber(cfg.GroqAPIKey)
	intentParser := groq.NewIntentParser(cfg.GroqAPIKey)

	queryController := controller.NewQueryExpensesController(queryExpenses, sender)
	voiceController := controller.NewVoiceController(audioDownloader, transcriber, intentParser, queryController, sender)

	h := handler.NewHandler(
		controller.NewStartController(sender),
		controller.NewHelpController(sender),
		controller.NewRegisterExpenseController(registerExpense, sender),
		queryController,
		voiceController,
		cfg.AllowedUserID,
		log,
	)

	log.Info("bot started",
		slog.String("bot", bot.Self.UserName),
		slog.String("mode", "polling"),
		slog.String("table", cfg.TableName),
		slog.String("log_level", cfg.LogLevel),
	)

	updates := bot.GetUpdatesChan(tgbotapi.NewUpdate(0))
	for update := range updates {
		h.Handle(ctx, update)
	}
}

func fatal(log *slog.Logger, msg string, err error) {
	log.Error(msg, slog.String("error", err.Error()))
	os.Exit(1)
}
