package controller

import (
	"context"
	"log"

	"money-savior-telegram-bot/app/handlers"
	"money-savior-telegram-bot/app/repository"
	"money-savior-telegram-bot/app/service"
	"money-savior-telegram-bot/app/utils"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type BotController struct {
	Expense *handlers.ExpenseHandler
	Query   *handlers.QueryHandler
	Delete  *handlers.DeleteHandler
}
