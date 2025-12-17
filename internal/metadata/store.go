package metadata

import "sync"

// Store holds values per modelID keyed by Key. Safe for concurrent use.
type Store struct {
	mu      sync.RWMutex
	byModel map[string]map[Key]any
}

func NewStore() *Store {
	return &Store{byModel: make(map[string]map[Key]any)}
}

func (s *Store) View(modelID string) View {
	return View{store: s, modelID: modelID}
}

func (s *Store) Put(modelID string, key Key, value any) {
	if modelID == "" || key == "" {
		return
	}
	if value == nil {
		s.Delete(modelID, key)
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()

	m := s.byModel[modelID]
	if m == nil {
		m = make(map[Key]any)
		s.byModel[modelID] = m
	}
	m[key] = value
}

func (s *Store) Get(modelID string, key Key) (any, bool) {
	if modelID == "" || key == "" {
		return nil, false
	}
	s.mu.RLock()
	defer s.mu.RUnlock()

	m := s.byModel[modelID]
	if m == nil {
		return nil, false
	}
	v, ok := m[key]
	return v, ok
}

func (s *Store) Delete(modelID string, key Key) {
	if modelID == "" || key == "" {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()

	m := s.byModel[modelID]
	if m == nil {
		return
	}
	delete(m, key)
	if len(m) == 0 {
		delete(s.byModel, modelID)
	}
}
