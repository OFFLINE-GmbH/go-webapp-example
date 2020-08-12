package cache

import (
	"time"

	gocache "github.com/patrickmn/go-cache"
)

type Store struct {
	cache *gocache.Cache
}

func New() *Store {
	c := gocache.New(1*time.Hour, 30*time.Minute)

	return &Store{
		cache: c,
	}
}

func (s *Store) Set(key string, value interface{}) {
	s.cache.Set(key, value, gocache.DefaultExpiration)
}

func (s *Store) Get(key string) (value interface{}, found bool) {
	return s.cache.Get(key)
}

func (s *Store) Remember(key string, value interface{}, ttl time.Duration) {
	s.cache.Set(key, value, ttl)
}

func (s *Store) RememberForever(key string, value interface{}) {
	s.cache.Set(key, value, gocache.NoExpiration)
}
