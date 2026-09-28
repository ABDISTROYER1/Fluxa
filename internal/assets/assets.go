package assets

import (
	"strings"
	"sync"
)

type Asset struct {
	Code   string `json:"code"`
	Issuer string `json:"issuer"`
	Type   string `json:"type"` // "native", "credit_alphanum4", "credit_alphanum12"
}

type Registry struct {
	mu     sync.RWMutex
	assets map[string]Asset
}

func NewRegistry(usdcIssuer, eurcIssuer string) *Registry {
	r := &Registry{
		assets: make(map[string]Asset),
	}

	// Register default curated assets
	r.Register(Asset{Code: "XLM", Issuer: "", Type: "native"})
	if usdcIssuer != "" {
		r.Register(Asset{Code: "USDC", Issuer: usdcIssuer, Type: "credit_alphanum4"})
	}
	if eurcIssuer != "" {
		r.Register(Asset{Code: "EURC", Issuer: eurcIssuer, Type: "credit_alphanum4"})
	}

	return r
}

func (r *Registry) Register(a Asset) {
	r.mu.Lock()
	defer r.mu.Unlock()
	key := strings.ToUpper(a.Code)
	r.assets[key] = a
}

func (r *Registry) Get(code string) (Asset, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	a, ok := r.assets[strings.ToUpper(code)]
	return a, ok
}

func (r *Registry) IsSupported(code string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	_, ok := r.assets[strings.ToUpper(code)]
	return ok
}

func (r *Registry) List() []Asset {
	r.mu.RLock()
	defer r.mu.RUnlock()
	list := make([]Asset, 0, len(r.assets))
	for _, a := range r.assets {
		list = append(list, a)
	}
	return list
}
