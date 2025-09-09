package service

import (
	"Wb_Test_L0/internal/config"
	"Wb_Test_L0/internal/models"
	"Wb_Test_L0/internal/repository"
	"container/list"
	"log"
	"sync"
)

type CacheItem struct {
	key   string
	value models.Order
}

type Service struct {
	repo     *repository.Repository
	cache    map[string]*list.Element
	order    *list.List
	mu       sync.RWMutex
	capacity int
}

func NewService(repo *repository.Repository, cfg *config.Config) *Service {
	s := &Service{
		repo:     repo,
		cache:    make(map[string]*list.Element),
		order:    list.New(),
		capacity: cfg.Cache.Capacity,
	}

	orders, err := repo.GetLastOrders(cfg.Cache.Size)
	if err == nil {
		for _, o := range orders {
			s.addToCache(o)
		}
	}
	return s
}

func (s *Service) addToCache(order models.Order) {
	if el, ok := s.cache[order.OrderUID]; ok {
		s.order.MoveToFront(el)
		el.Value.(*CacheItem).value = order
		return
	}

	if s.order.Len() >= s.capacity {
		last := s.order.Back()
		if last != nil {
			item := last.Value.(*CacheItem)
			delete(s.cache, item.key)
			s.order.Remove(last)
		}
	}

	el := s.order.PushFront(&CacheItem{key: order.OrderUID, value: order})
	s.cache[order.OrderUID] = el
}

func (s *Service) SaveOrder(order models.Order) error {
	log.Printf("service: saving order %s", order.OrderUID)
	if err := s.repo.SaveOrder(order); err != nil {
		log.Printf("service: failed to save order %s: %v", order.OrderUID, err)
		return err
	}
	log.Printf("service: saved order %s to db", order.OrderUID)

	s.mu.Lock()
	s.addToCache(order)
	s.mu.Unlock()

	return nil
}

func (s *Service) GetOrder(orderUID string) (*models.Order, error) {
	s.mu.RLock()
	if el, ok := s.cache[orderUID]; ok {
		s.mu.RUnlock()
		s.mu.Lock()
		s.order.MoveToFront(el)
		s.mu.Unlock()
		return &el.Value.(*CacheItem).value, nil
	}
	s.mu.RUnlock()

	order, err := s.repo.GetOrderByID(orderUID)
	if err != nil {
		return nil, err
	}

	s.mu.Lock()
	s.addToCache(*order)
	s.mu.Unlock()
	return order, nil
}
