package payment

import (
	"sync"

	"github.com/google/uuid"

	def "github.com/aidar-khabibullin/MSA_Rocket_Factory/payment/internal/service"
)

var _ def.PaymentService = (*service)(nil)

// record — платёж по заказу. refundUUID пуст, пока платёж не возвращён.
type record struct {
	transactionUUID uuid.UUID
	refundUUID      uuid.UUID
}

// service хранит платежи в памяти процесса: сервис имитирует платёжную систему и своей БД не имеет.
// Идемпотентность поэтому действует, пока процесс жив; в боевой системе записи лежали бы в хранилище.
type service struct {
	mu      sync.Mutex
	records map[uuid.UUID]*record // ключ — order_uuid, он же ключ идемпотентности
}

func NewService() *service {
	return &service{records: make(map[uuid.UUID]*record)}
}
