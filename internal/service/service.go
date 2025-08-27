package service

import (
	"Wb_Test_L0/internal/config"
	"Wb_Test_L0/internal/models"
	"Wb_Test_L0/internal/repository"
	"log"
	"sync"
)

type Service struct {
	repo  *repository.Repository
	cache map[string]models.Order
	mu    sync.RWMutex
}

func NewService(repo *repository.Repository, cfg *config.Config) *Service {
	s := &Service{
		repo:  repo,
		cache: make(map[string]models.Order),
	}
	orders, err := repo.GetLastOrders(cfg.Cache.Size)
	if err == nil {
		for _, o := range orders {
			s.cache[o.OrderUID] = o
		}
	}
	return s
}

func (s *Service) SaveOrder(order models.Order) error {
	log.Printf("service: saving order %s", order.OrderUID)
	if err := s.repo.SaveOrder(order); err != nil {
		log.Printf("service: failed to save order %s: %v", order.OrderUID, err)
		return err
	}
	log.Printf("service: saved order %s to db and cache", order.OrderUID)
	return nil
}

func (s *Service) GetOrder(orderUID string) (*models.Order, error) {
	s.mu.RLock()
	if order, ok := s.cache[orderUID]; ok {
		s.mu.RUnlock()
		return &order, nil
	}
	s.mu.RUnlock()

	order, err := s.repo.GetOrderByID(orderUID)
	if err != nil {
		return nil, err
	}

	s.mu.Lock()
	s.cache[order.OrderUID] = *order
	s.mu.Unlock()
	return order, nil
}
