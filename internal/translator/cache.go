package translator

import (
	"context"
	"fmt"
	"sync"
)

// NewCachedTranslator wraps a Translator with an in-memory LRU-style cache
// keyed by word|source_lang|target_lang.
func NewCachedTranslator(inner Translator) Translator {
	return &cachedTranslator{
		inner: inner,
		cache: make(map[string]*TranslateResponse),
	}
}

type cachedTranslator struct {
	inner Translator
	mu    sync.RWMutex
	cache map[string]*TranslateResponse
}

func cacheKey(req TranslateRequest) string {
	return req.Word + "|" + req.SourceLang + "|" + req.TargetLang
}

func (c *cachedTranslator) Translate(ctx context.Context, req TranslateRequest) (*TranslateResponse, error) {
	if req.Word == "" {
		return nil, fmt.Errorf("word cannot be empty")
	}

	key := cacheKey(req)

	// Read lock — check cache
	c.mu.RLock()
	if res, ok := c.cache[key]; ok {
		c.mu.RUnlock()
		return res, nil
	}
	c.mu.RUnlock()

	// Miss — call inner translator
	res, err := c.inner.Translate(ctx, req)
	if err != nil {
		return nil, err
	}

	// Write lock — store in cache
	c.mu.Lock()
	c.cache[key] = res
	c.mu.Unlock()

	return res, nil
}
