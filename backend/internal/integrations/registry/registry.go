package registry

import (
	"fmt"
	"sort"
	"sync"
)

// Registry holds registered integration providers.
type Registry struct {
	mu        sync.RWMutex
	providers map[string]Provider
	order     []string
}

func New() *Registry {
	return &Registry{
		providers: make(map[string]Provider),
		order:     make([]string, 0, 8),
	}
}

func (r *Registry) Register(provider Provider) error {
	if r == nil || provider == nil {
		return fmt.Errorf("invalid provider registration")
	}
	name := provider.Name()
	if name == "" {
		return fmt.Errorf("provider name is required")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.providers[name]; exists {
		return fmt.Errorf("provider %q is already registered", name)
	}
	r.providers[name] = provider
	r.order = append(r.order, name)
	return nil
}

func (r *Registry) Get(name string) (Provider, bool) {
	if r == nil {
		return nil, false
	}
	r.mu.RLock()
	defer r.mu.RUnlock()
	provider, ok := r.providers[name]
	return provider, ok
}

func (r *Registry) MustGet(name string) Provider {
	provider, ok := r.Get(name)
	if !ok {
		return nil
	}
	return provider
}

func (r *Registry) List() []ProviderDescriptor {
	if r == nil {
		return nil
	}
	r.mu.RLock()
	defer r.mu.RUnlock()

	descriptors := make([]ProviderDescriptor, 0, len(r.order))
	for _, name := range r.order {
		provider := r.providers[name]
		descriptors = append(descriptors, ProviderDescriptor{
			Name:         provider.Name(),
			DisplayName:  provider.DisplayName(),
			Description:  provider.Description(),
			Available:    provider.Available(),
			Capabilities: provider.Capabilities(),
		})
	}
	return descriptors
}

func (r *Registry) ListNames() []string {
	if r == nil {
		return nil
	}
	r.mu.RLock()
	defer r.mu.RUnlock()
	names := make([]string, len(r.order))
	copy(names, r.order)
	return names
}

func (r *Registry) Capabilities(name string) (ProviderCapabilities, bool) {
	provider, ok := r.Get(name)
	if !ok {
		return ProviderCapabilities{}, false
	}
	return provider.Capabilities(), true
}

func (r *Registry) SyncableProviderNames() []string {
	if r == nil {
		return nil
	}
	r.mu.RLock()
	defer r.mu.RUnlock()

	names := make([]string, 0, len(r.order))
	for _, name := range r.order {
		if r.providers[name].Capabilities().SupportsScheduling {
			names = append(names, name)
		}
	}
	sort.Strings(names)
	return names
}

func (r *Registry) AsOAuth(name string) (OAuthProvider, bool) {
	provider, ok := r.Get(name)
	if !ok {
		return nil, false
	}
	oauth, ok := provider.(OAuthProvider)
	return oauth, ok
}

func (r *Registry) AsPasswordCreate(name string) (PasswordCreateProvider, bool) {
	provider, ok := r.Get(name)
	if !ok {
		return nil, false
	}
	password, ok := provider.(PasswordCreateProvider)
	return password, ok
}

func (r *Registry) AsPrepareSync(name string) (PrepareSyncProvider, bool) {
	provider, ok := r.Get(name)
	if !ok {
		return nil, false
	}
	prepare, ok := provider.(PrepareSyncProvider)
	return prepare, ok
}

func (r *Registry) Has(name string) bool {
	_, ok := r.Get(name)
	return ok
}
