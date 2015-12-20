goim
==============
goim is a im server writen by golang.

## Features
 * Light weight
 * High performance
 * Pure Golang
 * Supports single push, multiple push and broadcasting
 * Supports one key to multiple subscribers (Configurable maximum subscribers count)
 * Supports heartbeats (Application heartbeats, TCP, KeepAlive, HTTP long pulling)
 * Supports authentication (Unauthenticated user can't subscribe)
 * Supports multiple protocols (WebSocket，TCP，HTTP）
 * Scalable architecture (Unlimited dynamic job and logic modules)
 * Asynchronous push notification based on Kafka

## Architecture
![arch](https://github.com/Terry-Mao/goim/blob/master/doc/arch.png)

Protocol:

[proto](https://github.com/Terry-Mao/goim/blob/master/doc/protocol.png)

## Document
[English](./README_en.md)

[中文](./README_cn.md)

## Examples
Websocket: [Websocket Client Demo](https://github.com/Terry-Mao/goim/tree/master/examples/javascript)

Android: [Android](https://github.com/roamdy/goim-sdk)

iOS: [iOS](https://github.com/roamdy/goim-oc-sdk)

## Benchmark
![benchmark](./doc/benchmark.jpg)

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

[中文](./doc/benchmark_cn.md)

[English](./doc/benchmark_en.md)

## LICENSE
goim is is distributed under the terms of the GNU General Public License, version 3.0 [GPLv3](http://www.gnu.org/licenses/gpl.txt)
