package translator

import (
	"context"
	"sync"
)

type callCounter struct {
	inner Translator
	count *int
	mu    *sync.Mutex
}

func (c *callCounter) Translate(ctx context.Context, req TranslateRequest) (*TranslateResponse, error) {
	c.mu.Lock()
	*c.count++
	c.mu.Unlock()
	return c.inner.Translate(ctx, req)
}
