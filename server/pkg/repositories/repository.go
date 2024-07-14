package repositories

import (
	"udp-hole-punch/pkg/repositories/adapters"
)

var repository IRepository

func GetRepository() IRepository {
	//TODO: change by config and singleton
	if repository == nil {
		repository = adapters.NewInMemoryRepository()
	}
	return repository
}
