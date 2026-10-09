package controller

import(
	"money-savior-telegram-bot/app/service"
)

type ExpenseHandler struct {
	expenseService *service.ExpenseService
}

type QueryHandler struct {
	expenseService *service.ExpenseService
}

type DeleteHandler struct {
	service *service.ExpenseService
}

func NewExpenseHandler(expenseService *service.ExpenseService) *ExpenseHandler {
	return &ExpenseHandler{expenseService: expenseService}
}

func NewQueryHandler(expenseService *service.ExpenseService) *QueryHandler {
	return &QueryHandler{expenseService: expenseService}
}

func NewDeleteHandler(service *service.ExpenseService) *DeleteHandler {
	return &DeleteHandler{service: service}
}
