// Package counter accumulates per-IP request counts in a plain in-memory map.
//
// This is intentionally the naive implementation: every unique IP is kept in
// RAM for the whole run.
package counter

// Counter counts occurrences keyed by IP address using a plain map.
type Counter struct {
	data map[string]int
}

// New returns an empty Counter.
func New() *Counter {
	return &Counter{data: make(map[string]int)}
}

// Add increments the counter for the given IP address.
func (c *Counter) Add(ip string) {
	c.data[ip]++
}

// Len reports the number of distinct keys currently held in memory.
func (c *Counter) Len() int {
	return len(c.data)
}

// Snapshot returns a copy of the accumulated counts.
func (c *Counter) Snapshot() map[string]int {
	out := make(map[string]int, len(c.data))
	for ip, n := range c.data {
		out[ip] = n
	}
	return out
}
