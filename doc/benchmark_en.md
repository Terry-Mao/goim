## Benchmark Chart
![benchmark](https://github.com/Terry-Mao/goim/blob/master/doc/benchmark.png)

### Benchmark Server

| CPU | Memory | Instance |
| :---- | :---- | :---- |
| Intel(R) Xeon(R) CPU E5-2630 v2 @ 2.60GHz  | DDR3 32GB | 5 |

### Benchmark Case

* Online: 500,000
* Duration: 10min
* Push Speed: 20/s (broadcast room)
* Push Message: {"test":1}

### Benchmark Resource

* CPU: 2340% (almost all busy)
* Memory: 2.65GB
* GC Pause: 41ms
* Network: Incoming(500MBit/s), Outgoing(780MBit/s)

### Benchmark Result

560 million/second message received with 5 24c server, 120 million/second per server.


## comet
![benchmark-comet](https://github.com/Terry-Mao/goim/blob/master/doc/benchmark-comet.png)

## network traffic
![benchmark-flow](https://github.com/Terry-Mao/goim/blob/master/doc/benchmark-flow.png)
