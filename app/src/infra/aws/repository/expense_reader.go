package repository

import (
	"context"
	"fmt"
	"strconv"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"

	"money-savior-telegram-bot/app/src/core/domain"
	"money-savior-telegram-bot/app/src/core/port"
)

const createdAtIndexName = "user_id-created_at-index"

var _ port.ExpenseReader = (*ExpenseRepository)(nil)

func (r *ExpenseRepository) FindByPeriod(ctx context.Context, userID int64, period domain.Period) ([]domain.Expense, error) {
	paginator := dynamodb.NewQueryPaginator(r.client, &dynamodb.QueryInput{
		TableName:              aws.String(r.tableName),
		IndexName:              aws.String(createdAtIndexName),
		KeyConditionExpression: aws.String("user_id = :uid AND created_at BETWEEN :start AND :end"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":uid":   &types.AttributeValueMemberN{Value: strconv.FormatInt(userID, 10)},
			":start": &types.AttributeValueMemberS{Value: formatCreatedAt(period.Start)},
			":end":   &types.AttributeValueMemberS{Value: formatCreatedAt(period.End)},
		},
	})

	var expenses []domain.Expense
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("query expenses by period: %w", err)
		}

		var items []expenseItem
		if err := attributevalue.UnmarshalListOfMaps(page.Items, &items); err != nil {
			return nil, fmt.Errorf("unmarshal expenses: %w", err)
		}

		for _, item := range items {
			expense, err := item.toDomain()
			if err != nil {
				return nil, err
			}
			//BETWEEN is inclusive on both ends, but the period is [Start, End).
			if !expense.CreatedAt.Before(period.End) {
				continue
			}
			expenses = append(expenses, expense)
		}
	}

	return expenses, nil
}
