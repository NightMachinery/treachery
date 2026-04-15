package app

import "sync"

type Hub struct {
	mu          sync.Mutex
	subscribers map[string]map[chan struct{}]struct{}
}

func NewHub() *Hub {
	return &Hub{subscribers: make(map[string]map[chan struct{}]struct{})}
}

func (h *Hub) Subscribe(gameID string) (<-chan struct{}, func()) {
	ch := make(chan struct{}, 1)

	h.mu.Lock()
	defer h.mu.Unlock()

	if _, ok := h.subscribers[gameID]; !ok {
		h.subscribers[gameID] = make(map[chan struct{}]struct{})
	}
	h.subscribers[gameID][ch] = struct{}{}

	cancel := func() {
		h.mu.Lock()
		defer h.mu.Unlock()
		if gameSubs, ok := h.subscribers[gameID]; ok {
			delete(gameSubs, ch)
			if len(gameSubs) == 0 {
				delete(h.subscribers, gameID)
			}
		}
		close(ch)
	}

	return ch, cancel
}

func (h *Hub) Publish(gameID string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	for ch := range h.subscribers[gameID] {
		select {
		case ch <- struct{}{}:
		default:
		}
	}
}
