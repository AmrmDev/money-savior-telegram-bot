package handler

import (
    "context"

    "money-savior-telegram-bot/app/src/adapter/controller"
)

type Handler struct {
    startController   *controller.StartController
    expenseController *controller.ExpenseController
    helpController    *controller.HelpController
}

func NewHandler(
    start *controller.StartController,
    expense *controller.ExpenseController,
    help *controller.HelpController,
) *Handler {
    return &Handler{
        startController:   start,
        expenseController: expense,
        helpController:    help,
    }
}

func (h *Handler) Handle(
    ctx context.Context,
    message Message,
) error {
    switch message.Command {
    case "/start":
        return h.startController.Handle(ctx, message)

    case "/help":
        return h.helpController.Handle(ctx, message)

    case "/gastei":
        return h.expenseController.Handle(ctx, message)

        //default value for unknown commands is the help handle
    default:
        return h.helpController.Handle(ctx, message)
    }
}