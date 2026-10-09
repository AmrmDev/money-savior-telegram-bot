package http

import (
	"context"
	"log"

	"money-savior-telegram-bot/app/controller"
	"money-savior-telegram-bot/app/repository"
	"money-savior-telegram-bot/app/service"
	"money-savior-telegram-bot/app/utils"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type BotController struct {
	Expense *controller.ExpenseHandler
	Query   *controller.QueryHandler
	Delete  *controller.DeleteHandler
}

func RouteUpdate(bot *tgbotapi.BotAPI,update tgbotapi.Update,h *BotController,) {
	msg := update.Message
	if msg == nil {
		msg = update.EditedMessage
	}

	if msg == nil {
		log.Println(utils.DebugUpdateWithNoMessage)
		return
	}

	log.Printf(utils.InfoMessageReceived, msg.Text)

	if msg.IsCommand() {
		command := msg.Command()
		log.Printf(utils.InfoCommandReceived, command)

		switch command {
		case "start":
			handlers.HandleStart(bot, msg)

		case "help":
			handlers.HandleHelp(bot, msg)

		case "gastei":
			h.Expense.Handle(bot, msg)

		case "consulta":
			h.Query.Handle(bot, msg)

		case "deletar":
			h.Delete.HandleDelete(bot, msg)

		case "deletartudo":
			h.Delete.HandleDeleteAll(bot, msg)

		default:
			log.Printf(utils.WarnUnknownCommand, command)
			handlers.HandleInvalidCommand(bot, msg)
		}
	}
}