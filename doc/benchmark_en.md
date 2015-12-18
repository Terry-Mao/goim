## Benchmark Chart
![benchmark](benchmark.jpg)

### Benchmark Server

| CPU | Memory | Instance |
| :---- | :---- | :---- |
| Intel(R) Xeon(R) CPU E5-2630 v2 @ 2.60GHz  | DDR3 32GB | 2 |

### Benchmark Case

* Online: 500,000(250,000 per server)
* Duration: 15min
* Push Speed: 50/s (broadcast room)
* Received: 2,440,000/s
* Push Message: {"test":1}
* Received calc mode: 30s per times, total 30 times

### Benchmark Resource

* CPU: 1400%~2340%
* Memory: 4.22GB
* GC Pause: 77ms
* Network: Incoming(302MBit/s), Outgoing(3.19GBit/s)

### Benchmark Result

560 million/second message received with 5 24c server, 120 million/second per server.


## comet
![benchmark-comet](benchmark-comet.png)

## network traffic
![benchmark-flow](benchmark-flow.png)

## heap (include GC)
![benchmark-flow](benchmark-heap.png)