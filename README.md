# BTC

BTC tracks bitcoin prices and keeps a configurable running average. There are 3 main components to this application's architecture:

* A http client that retrieves bitcoin stock (symbol BTCUSDT) on a specific interval. This interval is set by the PollInterval constant in cmd/btc/main.go
* A ring buffer which is responsible for keeping a running average. The ring buffer's capacity is configurable and the ring buffer provides an api wrapping a channel.
* A client facing http service which utilizes a request queueing mechanism. 

This application was ran with go's race detector to confirm it is deadlock free. 

Changing the RefreshInterval lower than a minute is recommended for testing

# Usage 

The server can simply be ran by it's binary. Any configuration can be changed by editing the constants cmd/btc/main.go 

```
~/git/go/src/oden/btc
❯ ./bin/macOS/btc
2018/11/03 20:48:00 Launched Binance Client
2018/11/03 20:48:00 launching http service on 0.0.0.0:8080
2018/11/03 20:48:02 received value 6375.700000 and updated moving average to 6375.700000
2018/11/03 20:48:02 received value 6375.700000 and updated moving average to 6375.700000
2018/11/03 20:48:03 received value 6375.700000 and updated moving average to 6375.700000
2018/11/03 20:48:04 received value 6375.700000 and updated moving average to 6375.700000
```

# Intervals
Two main intervals come into play with this application. 

* RefreshInterval: This is the interval in which we present updates to the moving average to clients. For example setting this interval to one minute will update the moving average for that time period.
* PollingInterval: This interval configures how often we query the Binance api and subsequently update our ring buffer and moving average. Every update to the moving average is logged to stderr

# Design choices 
For our purposes we set the PollingInterval to 1 second and the RefreshInterval to 1 minute. We set the capacity of the ring buffer to 60 elements. This allows us to approximately keep a moving 1 minute average once the buffer has seen 60 elements. We do not scan the entire ring buffer to compute the average, instead we keep track of the length of the buffer, the sum of all elements in the buffer, and the next index a write should occured in. When adding an element to the ring buffer we subtract the value at the saved index and add the new one, then recompute the average.

For client side updates we use a background worker. This background worker listens for ticks on the ticker configured with RefreshInterval. When a tick is encountered a bool is flipped and a lock is take. The bool routes requests into the lock thus queueing them while the update to average takes place. Once the worker updates the average the queue is unlocked and the bool is flipped; routing traffic away from the queueing lock.

# ToDos
* Remove the need to sleep in the refresh background worker
* Implement proper signal handling and graceful HTTP shutdown. 
* Close channels being ranged over to gracefully close go routines
* Move adhoc go routine functions to their own structs
