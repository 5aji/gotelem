package xbee
import (
	"sync"
	"errors"
)

// A connTrack is a simple frame mark utility for xbee packets. The xbee api frame
// takes a mark that is used when sending the response - this allows to coordinate
// the sent packet and the response, since there may be other packets emitted
// between them.
type connTrack struct {
	mu sync.RWMutex // use RW mutex to allow for multiple readers
	internal map[uint8]bool // map frame tag to if it's been used.
	// the map is set when writing a frame, and deleted when recieving a matching frame.
}

// GetMark finds the next available marker and takes it, returning the value of
// the mark. If no mark can be acquired, it returns an error.
func (ct *connTrack) GetMark() (uint8, error) {
	// get a read lock.
	ct.mu.RLock()

	// NOTE: we start at one. This is because 0 will not return a frame ever - it's 
	// the "silent mode" mark
	for i := 1; i < 256; i++ {
		if !ct.internal[uint8(i)] {
			// it's free.
			// discard our read lock and lock for write.
			ct.mu.RUnlock()

			ct.mu.Lock()
			// update the value to true.
			ct.internal[uint8(i)] = true
			ct.mu.Unlock()
			return uint8(i), nil
		}
	}
	ct.mu.RUnlock()
	return 0, errors.New("no available marks")
}
// ClearMark removes a given mark from the set if it exists, or returns an error.
func (ct *connTrack) ClearMark(mark uint8) error {
	ct.mu.RLock()
	// FIXME: should this be the other way around (swap if and normal execution
	if ct.internal[mark] {
		ct.mu.RUnlock()
		ct.mu.Lock()
		delete(ct.internal, mark)
		ct.mu.Unlock()
		return nil
	}
	ct.mu.RUnlock()
	return errors.New("mark was not set")
}
