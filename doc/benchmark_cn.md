## 压测图表
![benchmark](benchmark.jpg)

<h4>服务端配置</h4>
| CPU | 内存 | 数量 |
| :---- | :---- | :---- |
| Intel(R) Xeon(R) CPU E5-2630 v2 @ 2.60GHz  | DDR3 32GB | 2台 |

<h4>压测参数</h4>

* 不同UID同房间在线人数: 500,000(每台服务器250,000)
* 持续推送时长: 15分钟
* 持续推送数量: 50条/秒
* 推送到达: 2,440万/秒左右
* 推送内容: {"test":1}
* 推送类型: 单房间推送
* 到达收集方式: 30秒统计一次平均每秒到达,共30次统计

<h4>资源使用</h4>

* 每台服务端CPU使用: 1400%~2340%左右(刚好满负载)
* 每台服务端内存使用: 4.22GB左右
* GC耗时: 77毫秒左右
* 流量使用: Incoming(302MBit左右), Outgoing(3.19GBit左右)

## comet模块
![benchmark-comet](benchmark-comet.png)

## 流量
![benchmark-flow](benchmark-flow.png)

## heap信息(包含GC)
![benchmark-flow](benchmark-heap.png)
