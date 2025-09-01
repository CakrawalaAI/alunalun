package auth

import (
	"context"
	"errors"
	"fmt"
	"sync"
)

// ProviderRegistry manages authentication providers
type ProviderRegistry struct {
	providers map[string]Provider
	mu        sync.RWMutex
}

// NewProviderRegistry creates a new provider registry
func NewProviderRegistry() *ProviderRegistry {
	return &ProviderRegistry{
		providers: make(map[string]Provider),
	}
}

// Register adds a provider to the registry
func (r *ProviderRegistry) Register(provider Provider) error {
	if provider == nil {
		return errors.New("provider cannot be nil")
	}
	
	name := provider.Name()
	if name == "" {
		return errors.New("provider name cannot be empty")
	}
	
	// Validate provider configuration
	if err := provider.ValidateConfig(); err != nil {
		return fmt.Errorf("provider %s configuration invalid: %w", name, err)
	}
	
	r.mu.Lock()
	defer r.mu.Unlock()
	
	if _, exists := r.providers[name]; exists {
		return fmt.Errorf("provider %s already registered", name)
	}
	
	r.providers[name] = provider
	return nil
}

// Unregister removes a provider from the registry
func (r *ProviderRegistry) Unregister(name string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	if _, exists := r.providers[name]; !exists {
		return fmt.Errorf("provider %s not found", name)
	}
	
	delete(r.providers, name)
	return nil
}

// Get retrieves a provider by name
func (r *ProviderRegistry) Get(name string) (Provider, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	provider, exists := r.providers[name]
	if !exists {
		return nil, fmt.Errorf("provider %s not found", name)
	}
	
	return provider, nil
}

// Has checks if a provider is registered
func (r *ProviderRegistry) Has(name string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	_, exists := r.providers[name]
	return exists
}

// List returns all registered provider names
func (r *ProviderRegistry) List() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	names := make([]string, 0, len(r.providers))
	for name := range r.providers {
		names = append(names, name)
	}
	
	return names
}

// ListByType returns all providers of a specific type
func (r *ProviderRegistry) ListByType(providerType string) []Provider {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	var providers []Provider
	for _, provider := range r.providers {
		if provider.Type() == providerType {
			providers = append(providers, provider)
		}
	}
	
	return providers
}

// Authenticate attempts to authenticate using the specified provider
func (r *ProviderRegistry) Authenticate(ctx context.Context, providerName, credential string) (*UserInfo, error) {
	provider, err := r.Get(providerName)
	if err != nil {
		return nil, err
	}
	
	return provider.Authenticate(ctx, credential)
}

// Clear removes all providers from the registry
func (r *ProviderRegistry) Clear() {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	r.providers = make(map[string]Provider)
}

// Count returns the number of registered providers
func (r *ProviderRegistry) Count() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	return len(r.providers)
}

// ProviderInfo provides information about a registered provider
type ProviderInfo struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

// GetProviderInfo returns information about all registered providers
func (r *ProviderRegistry) GetProviderInfo() []ProviderInfo {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	info := make([]ProviderInfo, 0, len(r.providers))
	for _, provider := range r.providers {
		info = append(info, ProviderInfo{
			Name: provider.Name(),
			Type: provider.Type(),
		})
	}
	
	return info
}