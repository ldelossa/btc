package main

import (
	"log"
	"net/http"
	"oden/btc"
	"oden/btc/internal/binance"
	"sync"
	"time"
)

const (
	// bitcoin symbol used in api call to get price
	BTCS = "BTCUSDT"
	// one minute refresh interval for moving average
	RefreshInterval = time.Duration(1) * time.Minute
	// one second poll interval for querying binance API
	PollInterval = time.Duration(1) * time.Second
	// default address this server will listen on
	DefaultServiceAddr = "0.0.0.0:8080"
	// capacity of our ring buffer; set to 60 to keep one minute moving average when PollInterval @ 1 second
	RingCap = 60
)

func main() {
	// --- Ring Buffer and Binance Client Init ---

	// create channel for client -> ring communication
	pollChan := make(chan float64, 1024)

	// create client
	bc := binance.NewClient()

	// create ring buffer. capacity will be 60 to provide one minute moving average
	ring := btc.NewAvgRing(RingCap, pollChan)

	// launch go routine which will poll binance API every second and push the stock
	// price into our ring buffer
	go func(bc *binance.Client, ring *btc.AvgRing) {
		log.Printf("Launched Binance client")
		// create ticker for making a binance api request every second
		binanceTicker := time.NewTicker(PollInterval)

		for _ = range binanceTicker.C {
			tickerPrice, err := bc.GetPrice(BTCS)
			if err != nil {
				log.Printf(err.Error())
				continue
			}

			err = ring.AddVal(tickerPrice.Price)
			if err != nil {
				log.Printf(err.Error())
				continue
			}
		}
	}(bc, ring)

	// --- HTTP Service Init ---

	// create our service configuration
	conf := btc.Conf{
		Lock: sync.Mutex{},
		Ring: ring,
		T:    time.NewTicker(RefreshInterval),
	}

	// create our mux and add route
	mux := http.NewServeMux()
	mux.Handle("/price", btc.AverageHandler(conf))

	// create our server
	s := &http.Server{
		Addr:    DefaultServiceAddr,
		Handler: mux,
	}

	// listen and serve
	log.Printf("launching http service on %s", DefaultServiceAddr)
	s.ListenAndServe()

}
