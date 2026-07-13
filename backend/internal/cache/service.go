package cache

import (
	"context"
	"encoding/json"
	"time"
)

type Service struct {
	client Client
}

func NewService(client Client) *Service {
	return &Service{client: client}
}

func (s *Service) GetOrSet(ctx context.Context, key string, ttl time.Duration, loader func() ([]byte, error)) ([]byte, error) {
	if s == nil || s.client == nil {
		return loader()
	}
	if cached, ok, err := s.client.Get(ctx, key); err == nil && ok {
		return cached, nil
	} else if err != nil {
		// cache read failure should not break request
		return loader()
	}

	value, err := loader()
	if err != nil {
		return nil, err
	}
	_ = s.client.Set(ctx, key, value, ttl)
	return value, nil
}

func GetOrSetJSON[T any](ctx context.Context, service *Service, key string, ttl time.Duration, loader func() (T, error)) (T, error) {
	var zero T
	if service == nil {
		return loader()
	}
	raw, err := service.GetOrSet(ctx, key, ttl, func() ([]byte, error) {
		value, loadErr := loader()
		if loadErr != nil {
			return nil, loadErr
		}
		return json.Marshal(value)
	})
	if err != nil {
		return zero, err
	}
	if err := json.Unmarshal(raw, &zero); err != nil {
		return loader()
	}
	return zero, nil
}

func (s *Service) Delete(ctx context.Context, keys ...string) error {
	if s == nil || s.client == nil {
		return nil
	}
	return s.client.Delete(ctx, keys...)
}

func (s *Service) DeleteByPattern(ctx context.Context, pattern string) error {
	if s == nil || s.client == nil {
		return nil
	}
	return s.client.DeleteByPattern(ctx, pattern)
}

func (s *Service) GetRaw(ctx context.Context, key string) ([]byte, bool, error) {
	if s == nil || s.client == nil {
		return nil, false, nil
	}
	return s.client.Get(ctx, key)
}

func (s *Service) SetRaw(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	if s == nil || s.client == nil {
		return nil
	}
	return s.client.Set(ctx, key, value, ttl)
}
