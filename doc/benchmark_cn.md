## 压测图表
![benchmark](https://github.com/Terry-Mao/goim/blob/master/doc/benchmark.png)

<h4>服务端配置</h4>
| CPU | 内存 | 数量 |
| :---- | :---- | :---- |
| Intel(R) Xeon(R) CPU E5-2630 v2 @ 2.60GHz  | DDR3 32GB | 5台 |

<h4>压测参数</h4>

* 不同UID同房间在线人数: 500,000
* 持续推送时长: 10分钟
* 持续推送数量: 20条/秒
* 推送到达: 560万/秒左右
* 推送内容: {"test":1}
* 推送类型: 单房间推送

<h4>资源使用</h4>

* 每台服务端CPU使用: 2340%左右(满)
* 每台服务端内存使用: 2.65GB左右
* GC耗时: 41毫秒左右
* 流量使用: Incoming(500MBit左右), Outgoing(780MBit左右)



## comet模块
![benchmark-comet](https://github.com/Terry-Mao/goim/blob/master/doc/benchmark-comet.png)

## 流量
![benchmark-flow](https://github.com/Terry-Mao/goim/blob/master/doc/benchmark-flow.png)
