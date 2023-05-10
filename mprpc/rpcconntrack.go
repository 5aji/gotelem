package mprpc

import (
	"errors"
	"math/rand"
	"sync"
)

// RPCConntrack is a request-response tracker that is used to connect
// the response to the appropriate caller.
type rpcConnTrack struct {
	ct map[uint32]chan Response
	mu sync.RWMutex
}

// Get attempts to get a random mark from the mutex.
func (c *rpcConnTrack) Claim() (uint32, chan Response) {
	var val uint32
	for {

		//
		newVal := rand.Uint32()

		// BUG(saji): rpcConnTrack collisions are inefficient.

		// collision is *rare* - so we just try again.
		// I hope to god you don't saturate this tracker.
		c.mu.RLock()
		if _, exist := c.ct[newVal]; !exist {
			val = newVal
			c.mu.RUnlock()
			break
		}
		c.mu.RUnlock()
	}

	// claim it
	// the channel should be buffered. We only expect one value to go through.
	// so the size is fixed to 1.
	ch := make(chan Response, 1)
	c.mu.Lock()
	c.ct[val] = ch
	c.mu.Unlock()

	return val, ch
}

// Clear deletes the connection from the tracker and returns the channel
// associated with it. The caller can use the channel afterwards
// to send the response. It is the caller's responsibility to close the channel.
func (c *rpcConnTrack) Clear(val uint32) (chan Response, error) {
	c.mu.RLock()
	ch, ok := c.ct[val]
	c.mu.RUnlock()
	if !ok {
		return nil, errors.New("invalid msg id")
	}
	c.mu.Lock()
	delete(c.ct, val)
	c.mu.Unlock()
	return ch, nil
}
