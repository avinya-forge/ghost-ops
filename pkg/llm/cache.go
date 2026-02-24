package llm

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"sync"

	"ghost-ops/pkg/protocol"
)

// Cache defines the interface for caching LLM responses.
type Cache interface {
	Get(key string) (*CachedResponse, bool)
	Set(key string, response *CachedResponse)
}

// CachedResponse holds the content and token usage of a cached LLM response.
type CachedResponse struct {
	Content    string               `json:"content"`
	TokenUsage *protocol.TokenUsage `json:"token_usage"`
}

// InMemoryCache implements Cache using a map.
type InMemoryCache struct {
	mu       sync.RWMutex
	store    map[string]*CachedResponse
	maxItems int
}

// NewInMemoryCache creates a new InMemoryCache.
func NewInMemoryCache() *InMemoryCache {
	return &InMemoryCache{
		store:    make(map[string]*CachedResponse),
		maxItems: 100, // Reasonable default limit
	}
}

// Get retrieves a response from the cache.
func (c *InMemoryCache) Get(key string) (*CachedResponse, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	val, ok := c.store[key]
	return val, ok
}

// Set stores a response in the cache.
func (c *InMemoryCache) Set(key string, response *CachedResponse) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Simple eviction if full
	if len(c.store) >= c.maxItems {
		// Go map iteration is random, so this evicts a random item
		for k := range c.store {
			delete(c.store, k)
			break
		}
	}

	c.store[key] = response
}

// CachedLLMProvider wraps an LLMProvider with caching logic.
type CachedLLMProvider struct {
	provider protocol.LLMProvider
	cache    Cache
}

// NewCachedLLMProvider creates a new CachedLLMProvider.
func NewCachedLLMProvider(provider protocol.LLMProvider, cache Cache) *CachedLLMProvider {
	return &CachedLLMProvider{
		provider: provider,
		cache:    cache,
	}
}

// GenerateCode generates code using the underlying provider, checking the cache first.
func (p *CachedLLMProvider) GenerateCode(ctx context.Context, blueprint protocol.Blueprint) (string, *protocol.TokenUsage, error) {
	// Generate cache key
	key, err := p.generateCacheKey(blueprint)
	if err != nil {
		// If key generation fails, fallback to direct call
		return p.provider.GenerateCode(ctx, blueprint)
	}

	// Check cache
	if cached, ok := p.cache.Get(key); ok {
		return cached.Content, cached.TokenUsage, nil
	}

	// Call provider
	content, usage, err := p.provider.GenerateCode(ctx, blueprint)
	if err != nil {
		return "", nil, err
	}

	// Cache result
	p.cache.Set(key, &CachedResponse{
		Content:    content,
		TokenUsage: usage,
	})

	return content, usage, nil
}

func (p *CachedLLMProvider) generateCacheKey(blueprint protocol.Blueprint) (string, error) {
	// Include intent and constraints in the hash
	data := struct {
		Intent      string                 `json:"intent"`
		Constraints map[string]interface{} `json:"constraints"`
	}{
		Intent:      blueprint.Intent,
		Constraints: blueprint.Constraints,
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		return "", err
	}

	// Hash it
	hash := sha256.Sum256(jsonData)
	return hex.EncodeToString(hash[:]), nil
}
