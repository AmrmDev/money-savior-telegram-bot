package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"

	"money-savior-telegram-bot/app/src/core/domain"
	"money-savior-telegram-bot/app/src/core/domain/exception"
	"money-savior-telegram-bot/app/src/core/port"
)

type ExpenseRepository struct {
	client    *dynamodb.Client
	tableName string
}

var _ port.ExpenseRepository = (*ExpenseRepository)(nil)

func NewExpenseRepository(client *dynamodb.Client, tableName string) *ExpenseRepository {
	return &ExpenseRepository{client: client, tableName: tableName}
}

func (r *ExpenseRepository) Save(ctx context.Context, expense domain.Expense) error {
	item, err := attributevalue.MarshalMap(toItem(expense))
	if err != nil {
		return fmt.Errorf("error to marshal expense: %w", err)
	}

	_, err = r.client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName:           aws.String(r.tableName),
		Item:                item,
		ConditionExpression: aws.String("attribute_not_exists(expense_id)"),
	})
	if err != nil {
		var conflict *types.ConditionalCheckFailedException
		if errors.As(err, &conflict) {
			return exception.ErrExpenseIDConflict
		}
		return fmt.Errorf("put expense: %w", err)
	}

	return nil
}
