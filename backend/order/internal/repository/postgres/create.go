package postgres

import (
	"context"

	sq "github.com/Masterminds/squirrel"
	"github.com/google/uuid"

	"github.com/aidar-khabibullin/MSA_Rocket_Factory/order/internal/model"
	"github.com/aidar-khabibullin/MSA_Rocket_Factory/order/internal/repository/converter"
)

func (r *repository) CreateOrder(ctx context.Context, order *model.Order) (*model.Order, error) {
	repoOrder := converter.ToRepoOrderPostgres(order)

	columns := []string{"user_uuid", "part_uuid", "total_price", "transaction_uuid", "payment_method", "status"}
	values := []any{repoOrder.UserUUID, repoOrder.PartUUIDs, repoOrder.TotalPrice, repoOrder.TransactionUUID, repoOrder.PaymentMethod, repoOrder.Status}

	// UUID может быть задан заранее (под него уже сделан резерв деталей), иначе его выдаст база по умолчанию.
	if order.OrderUUID != uuid.Nil {
		columns = append([]string{"order_uuid"}, columns...)
		values = append([]any{order.OrderUUID}, values...)
	}

	builderInsert := sq.Insert("orders").
		PlaceholderFormat(sq.Dollar).
		Columns(columns...).
		Values(values...).
		Suffix("RETURNING order_uuid")

	query, args, err := builderInsert.ToSql()
	if err != nil {
		return nil, model.ErrFailedToBuildQuery
	}

	err = r.db.QueryRow(ctx, query, args...).Scan(&repoOrder.OrderUUID)
	if err != nil {
		return nil, model.ErrFailedToInsertOrder
	}

	order.OrderUUID = repoOrder.OrderUUID

	return order, nil
}
