package btc

import (
	"fmt"
	"log"
)

// AvgRing utilizes a ring buffer and keeps a running average of it's contents.
// Capacity field is used to set the bounds of the ring buffer.
// A channel is listened to in order to add items into the ring.
type AvgRing struct {
	// array to act as our ring buffer
	ring []float64
	// current sum of all values in ring
	sum float64
	// current average of all values in ring
	avg float64
	// number of elements allowed in ring
	capacity int
	// next index to write to in our ring
	head int
	// channel to retrieve items from and add to ring.
	pollChan chan float64
}

// NewAvgRing is a constructor for an AvgRing
func NewAvgRing(capacity int, pollChan chan float64) *AvgRing {
	a := &AvgRing{
		ring:     []float64{},
		capacity: capacity,
		pollChan: pollChan,
	}

	go a.startPoll()

	return a
}

// AddVal is a public method which enqueues a value onto the pollChannel
// in the case of slow consumption we will drop messages
func (a *AvgRing) AddVal(v float64) error {
	select {
	case a.pollChan <- v:
		return nil
	default:
		return fmt.Errorf("failed to add value. buffer is full")
	}
}

// addVal adds a value to our ring buffer. The ring is implemented as a slice in order
// to use the length of the slice as the divisor simplifying sma logic
func (a *AvgRing) addVal(v float64) {

	var overwritten float64

	if len(a.ring) < a.capacity {
		a.ring = append(a.ring, v)
	} else {
		overwritten = a.ring[a.head]
		a.ring[a.head] = v
	}

	// subtract overwritten from sum, add new value
	a.sum -= overwritten
	a.sum += v

	// recompute average
	a.avg = a.sum / float64(len(a.ring))

	// increment index
	a.head = (a.head + 1) % a.capacity

	// log
	log.Printf("received value %f and updated moving average to %f", v, a.avg)

}

// Average returns the current average of all values in the ring buffer
func (a *AvgRing) Average() float64 {
	return a.avg
}

// startPoll is intended to be ran as a go routine. It begins pulling from the
// pollChan and adding items to the ring
func (a *AvgRing) startPoll() {
	for value := range a.pollChan {
		a.addVal(value)
	}
}
