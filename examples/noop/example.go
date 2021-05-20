package noop

import (
	"context"
)

type Example interface {
	Init(ctx context.Context) error
}
