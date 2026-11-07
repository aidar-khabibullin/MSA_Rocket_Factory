package inmemory

import (
	"sync"

	def "github.com/aidar-khabibullin/MSA_Rocket_Factory/inventory/internal/repository"
	repoModel "github.com/aidar-khabibullin/MSA_Rocket_Factory/inventory/internal/repository/model"
)

var _ def.InventoryRepository = (*repository)(nil)

type repository struct {
	mu   sync.RWMutex
	data map[string]*repoModel.RepositoryPart
	// reservations хранит, сколько каких деталей зарезервировано под заказ (ключ — order uuid).
	reservations map[string]map[string]int64
}

func NewRepository() def.InventoryRepository {
	repo := &repository{
		data:         make(map[string]*repoModel.RepositoryPart),
		reservations: make(map[string]map[string]int64),
	}

	repo.addTestData()

	return repo
}
