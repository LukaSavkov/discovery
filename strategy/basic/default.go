package basic

import (
	"context"
	"math/rand"
	"time"
)

type BasicStrategy struct {
}

func (s *BasicStrategy) Next(ctx context.Context, size int) (int, error) {
	rand.Seed(time.Now().Unix())
	return rand.Int() % size, nil
}

func NewStrategy() (*BasicStrategy, error) {
	return &BasicStrategy{}, nil
}
