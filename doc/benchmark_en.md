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
* Push Message: {"test":1}
* Received calc mode: 30s per times, total 30 times

### Benchmark Resource

* CPU: 1400%~2340%
* Memory: 4.22GB
* GC Pause: 77ms
* Network: Incoming(302MBit/s), Outgoing(3.19GBit/s)

### Benchmark Result
* Received: 2,440,000/s, 12,200,000/s per server.

## Comet
![benchmark-comet](benchmark-comet.png)

## Network traffic
![benchmark-flow](benchmark-flow.png)

## Heap (include GC)
![benchmark-flow](benchmark-heap.png)
