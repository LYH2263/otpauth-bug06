package otpauth

import "sync"

type CounterStore struct {
	mu   sync.Mutex
	cur  map[string]uint64
	path string
}

func NewCounterStore(path string) *CounterStore {
	return &CounterStore{cur: make(map[string]uint64), path: path}
}

func (c *CounterStore) Get(id string) uint64 {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.cur[id]
}

func (c *CounterStore) Consume(id string) (uint64, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	next := c.cur[id] + 1
	// Persist first; only advance the in-memory counter once the write is
	// durable. Otherwise a failed write still bumps c.cur[id], and after a
	// restart the loaded value lags the in-memory value so the HOTP window
	// drifts and every subsequent code is rejected.
	if err := persistCounter(c.path, id, next); err != nil {
		return 0, err
	}
	c.cur[id] = next
	return next, nil
}
