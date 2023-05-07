package xbee

import (
	"errors"
	"sync"
)

// A connTrack is a simple frame mark utility for xbee packets. The xbee api frame
// takes a mark that is used when sending the response - this allows to coordinate
// the sent packet and the response, since there may be other packets emitted
// between them. The data that is stored in the tag is a channel that contains an error - this
// is sent by the reader.
type connTrack struct {
	mu       sync.RWMutex // use RW mutex to allow for multiple readers
	internal map[uint8]chan []byte
	// the map is set when writing a frame, and deleted when recieving a matching frame.
}

func NewConnTrack() *connTrack {
	return &connTrack{
		internal: make(map[uint8]chan []byte),
	}
}

// GetMark finds the next available marker and takes it, returning the value of
// the mark as well as a channel to use as a semaphore when the mark is cleared.
// If no mark can be acquired, it returns an error.
func (ct *connTrack) GetMark() (uint8, <-chan []byte, error) {
	// get a read lock.
	ct.mu.RLock()

	// NOTE: we start at one. This is because 0 will not return a frame ever - it's
	// the "silent mode" mark
	for i := 1; i < 256; i++ {
		if _, ok := ct.internal[uint8(i)]; !ok {
			// it's free.
			// discard our read lock and lock for write.
			ct.mu.RUnlock()

			ct.mu.Lock()
			// create the channel, makeit buffered so that we don't
			// block when we write the error when freeing the mark later.
			ct.internal[uint8(i)] = make(chan []byte, 1)
			ct.mu.Unlock()
			return uint8(i), ct.internal[uint8(i)], nil
		}
	}
	ct.mu.RUnlock()
	return 0, nil, errors.New("no available marks")
}

// ClearMark removes a given mark from the set if it exists, or returns an error.
// it takes an error (which can be nil) to send to the channel. this is used to free
// whatever command wrote that packet - be it a write() call or a custom AT command that is
// tracked.
func (ct *connTrack) ClearMark(mark uint8, data []byte) error {
	ct.mu.RLock()

	val, ok := ct.internal[mark]
	if !ok {
		ct.mu.RUnlock()
		return errors.New("mark was not set")
	}

	ct.mu.RUnlock()
	ct.mu.Lock()
	val <- data
	close(val)
	delete(ct.internal, mark)
	ct.mu.Unlock()
	return nil
}
