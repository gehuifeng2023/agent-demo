package knowledge

import "sync"

type Registry struct {
	mu    sync.RWMutex
	bases map[string]*KnowledgeBase
}

func NewRegistry() *Registry {
	return &Registry{bases: map[string]*KnowledgeBase{}}
}

func (r *Registry) Register(kb *KnowledgeBase) {
	if kb == nil || kb.ID == "" {
		return
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	r.bases[kb.ID] = kb
}

func (r *Registry) Get(id string) (*KnowledgeBase, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	kb, ok := r.bases[id]
	return kb, ok
}

func (r *Registry) GetMany(ids []string) []*KnowledgeBase {
	var out []*KnowledgeBase
	for _, id := range ids {
		if kb, ok := r.Get(id); ok {
			out = append(out, kb)
		}
	}
	return out
}

func (r *Registry) All() []*KnowledgeBase {
	r.mu.RLock()
	defer r.mu.RUnlock()

	out := make([]*KnowledgeBase, 0, len(r.bases))
	for _, kb := range r.bases {
		out = append(out, kb)
	}
	return out
}
