package mongo

import (
	"context"
	"errors"
	"fmt"
	"log"
	"sort"
	"time"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/aidar-khabibullin/MSA_Rocket_Factory/inventory/internal/model"
)

const (
	reservationsCollection = "reservations"
	// partIDField — поле с идентификатором детали в документе parts (исторически оно названо order_uuid).
	partIDField = "order_uuid"
)

type reservationItem struct {
	PartUUID string `bson:"part_uuid"`
	Quantity int64  `bson:"quantity"`
}

type reservation struct {
	OrderUUID string            `bson:"order_uuid"`
	Items     []reservationItem `bson:"items"`
	CreatedAt time.Time         `bson:"created_at"`
}

// ensureReservationIndex создаёт уникальный индекс по заказу: он не даёт двум параллельным
// резервам одного заказа списать остаток дважды.
func (r *repository) ensureReservationIndex(ctx context.Context) {
	_, err := r.db.Collection(reservationsCollection).Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "order_uuid", Value: 1}},
		Options: options.Index().SetUnique(true),
	})
	if err != nil {
		log.Printf("Предупреждение: не удалось создать индекс резервов: %v", err)
	}
}

// Идемпотентность гарантирована для последовательных повторов (таймаут клиента, повтор gRPC). Параллельные дубликаты
// одного заказа при тесном остатке могут получить ErrInsufficientStock: каждый списывает остаток до проверки
// уникального индекса, но итоговое списание всегда одно, а перепродажи нет.
//
// Reserve списывает остатки по каждой детали условным UPDATE (`stock_quantity >= n`), поэтому остаток не уходит в минус
// даже при параллельных заказах. Транзакции не нужны (они требуют replica set): если одной детали не хватило,
// уже списанные откатываются компенсирующим $inc.
func (r *repository) Reserve(ctx context.Context, orderUUID uuid.UUID, partUUIDs []uuid.UUID) error {
	if len(partUUIDs) == 0 {
		return model.ErrEmptyReservation
	}

	items := countItems(partUUIDs)
	parts := r.db.Collection(partsCollection)
	reservations := r.db.Collection(reservationsCollection)

	reserved, err := reservations.CountDocuments(ctx, bson.M{"order_uuid": orderUUID.String()})
	if err != nil {
		return fmt.Errorf("check reservation: %w", err)
	}
	if reserved > 0 {
		return nil // идемпотентность: заказ уже зарезервирован
	}

	applied := make([]reservationItem, 0, len(items))
	for _, item := range items {
		result, err := parts.UpdateOne(ctx,
			bson.M{partIDField: item.PartUUID, "stock_quantity": bson.M{"$gte": item.Quantity}},
			bson.M{"$inc": bson.M{"stock_quantity": -item.Quantity}},
		)
		if err != nil {
			r.rollback(ctx, applied)
			return fmt.Errorf("reserve part %s: %w", item.PartUUID, err)
		}
		if result.MatchedCount == 0 {
			r.rollback(ctx, applied)
			return r.whyNotReserved(ctx, item.PartUUID)
		}
		applied = append(applied, item)
	}

	_, err = reservations.InsertOne(ctx, reservation{OrderUUID: orderUUID.String(), Items: items, CreatedAt: time.Now()})
	if err != nil {
		r.rollback(ctx, applied)
		if mongo.IsDuplicateKeyError(err) {
			return nil // параллельный резерв того же заказа победил: остаток уже списан им
		}
		return fmt.Errorf("save reservation: %w", err)
	}

	return nil
}

// Release возвращает остатки и удаляет резерв. Сначала резерв удаляется атомарно (FindOneAndDelete), поэтому
// параллельные вызовы не вернут остаток дважды; если возврат остатка не удался, резерв записывается обратно.
func (r *repository) Release(ctx context.Context, orderUUID uuid.UUID) error {
	var doc reservation
	err := r.db.Collection(reservationsCollection).
		FindOneAndDelete(ctx, bson.M{"order_uuid": orderUUID.String()}).
		Decode(&doc)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("find reservation: %w", err)
	}

	if err := r.restore(ctx, doc.Items); err != nil {
		if _, insertErr := r.db.Collection(reservationsCollection).InsertOne(context.WithoutCancel(ctx), doc); insertErr != nil {
			log.Printf("Ошибка: резерв заказа %s не удалось вернуть и восстановить, требуется ручная правка остатков: %v", doc.OrderUUID, insertErr)
		}
		return fmt.Errorf("release reservation: %w", err)
	}

	return nil
}

func (r *repository) whyNotReserved(ctx context.Context, partUUID string) error {
	exists, err := r.db.Collection(partsCollection).CountDocuments(ctx, bson.M{partIDField: partUUID})
	if err != nil {
		return fmt.Errorf("check part %s: %w", partUUID, err)
	}
	if exists == 0 {
		return model.ErrPartNotFound
	}
	return model.ErrInsufficientStock
}

// rollback возвращает уже списанные остатки. Контекст отвязан от отмены запроса, чтобы откат дошёл до конца.
func (r *repository) rollback(ctx context.Context, applied []reservationItem) {
	if err := r.restore(context.WithoutCancel(ctx), applied); err != nil {
		log.Printf("Ошибка отката резерва, остатки деталей могли остаться заниженными: %v", err)
	}
}

func (r *repository) restore(ctx context.Context, items []reservationItem) error {
	var firstErr error
	for _, item := range items {
		_, err := r.db.Collection(partsCollection).UpdateOne(ctx,
			bson.M{partIDField: item.PartUUID},
			bson.M{"$inc": bson.M{"stock_quantity": item.Quantity}},
		)
		if err != nil && firstErr == nil {
			firstErr = fmt.Errorf("restore part %s: %w", item.PartUUID, err)
		}
	}
	return firstErr
}

// countItems сворачивает повторяющиеся uuid в «деталь × количество» в детерминированном порядке.
func countItems(partUUIDs []uuid.UUID) []reservationItem {
	counts := make(map[string]int64, len(partUUIDs))
	for _, id := range partUUIDs {
		counts[id.String()]++
	}

	items := make([]reservationItem, 0, len(counts))
	for id, quantity := range counts {
		items = append(items, reservationItem{PartUUID: id, Quantity: quantity})
	}
	sort.Slice(items, func(i, j int) bool { return items[i].PartUUID < items[j].PartUUID })

	return items
}
