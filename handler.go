package btc

import (
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"
)

// global average variable for sharing between requests
var average float64

// global indicator of whether a write to the average variable is complete or in progress
var writeInProgress bool

// Conf is the configuration and depedencies for our AverageHandler handler
type Conf struct {
	// used to queue requests until global average is written to
	Lock sync.Mutex
	// ring structure which is keeping track of moving average. used to grab current moving average
	Ring *AvgRing
	// ticker to evaluate if we need to refresh global average variable. this interval decides
	// how often a client sees a new moving average.
	T *time.Ticker
}

// refreshAverage holds the instructions to queue requests, safely write to the average variable
// and route requests into and out of the queueing lock
func refreshAverage(conf Conf) {
	// obtain lock to queue requests
	conf.Lock.Lock()
	// change global to route request into queueing lock
	writeInProgress = true
	// update our average and log
	average = conf.Ring.Average()
	log.Printf("refreshed client average to %f", average)
	// route requests away from queuing lock
	writeInProgress = false
	// unblock queued requests
	conf.Lock.Unlock()
}

// AverageHandler launches a background refresh worker to update the global average variable
// it then returns our handler which is equiped to queue requests as the average is being updated
func AverageHandler(conf Conf) http.HandlerFunc {
	// launch background refresh worker
	go func(conf Conf) {
		log.Printf("launched background refresh worker")
		// spin on average until we see atleast one datapoint ensuring we don't send 0s
		// to the client until RefreshInterval is up
		for conf.Ring.Average() == float64(0) {
			time.Sleep(time.Duration(500) * time.Millisecond)
		}
		refreshAverage(conf)

		// now continue on our normal RefreshInterval ticker
		for _ = range conf.T.C {
			refreshAverage(conf)
		}
	}(conf)

	return func(w http.ResponseWriter, r *http.Request) {

		if !writeInProgress {
			s := strconv.FormatFloat(average, 'f', 2, 64)
			w.Write([]byte(s))
			return
		} else {
			conf.Lock.Lock()   // requests will queue here
			conf.Lock.Unlock() // queued requests will immediately unlock freeing others
			s := strconv.FormatFloat(average, 'f', 2, 64)
			w.Write([]byte(s))
			return
		}

	}
}
