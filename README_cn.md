gopush-cluster
==============
`Terry-Mao/goim` 是一个支持集群的im及实时推送服务（支持websocket，http和tcp协议）。

---------------------------------------
  * [特性](#特性)
  * [安装](#安装)
  * [配置](#配置)
  * [例子](#例子)
  * [文档](#文档)
  * [更多](#更多)

---------------------------------------

## 特性
 * 轻量级
 * 高性能
 * 纯Golang实现
 * 支持单个、多个以及广播消息推送
 * 支持单个Key多个订阅者（可限制订阅者最大人数）
 * 心跳支持（应用心跳和tcp、keepalive、http log pulling）
 * 支持安全验证（未授权用户不能订阅）
 * 多协议支持（websocket，tcp，http）
 * 可拓扑的架构（job、logic模块可动态无限扩展）
 * 基于Kafka做异步消息推送

## 安装
### 一、安装依赖
```sh
$ yum -y install java-1.7.0-openjdk
```

### 二、安装Kafka消息队列服务

kafka在官网已经描述的非常详细，在这里就不过多说明，安装、启动请查看[这里](http://kafka.apache.org/documentation.html#quickstart).

### 三、搭建golang环境
1.下载源码(根据自己的系统下载对应的[安装包](http://golang.org/dl/))
```sh
$ cd /data/programfiles
$ wget -c --no-check-certificate https://storage.googleapis.com/golang/go1.5.linux-amd64.tar.gz
$ tar -xvf go1.5.linux-amd64.tar.gz -C /usr/local
```
2.配置GO环境变量
(这里我加在/etc/profile.d/golang.sh)
```sh
$ vi /etc/profile.d/golang.sh
# 将以下环境变量添加到profile最后面
export GOROOT=/usr/local/go
export PATH=$PATH:$GOROOT/bin
export GOPATH=/data/apps/go
$ source /etc/profile
```

### 四、部署goim
1.下载goim及依赖包
```sh
$ yum install hg
$ go get -u github.com/Terry-Mao/goim
$ cd /data/apps/go/src/github.com/Terry-Mao/goim
$ go get ./...
```

2.安装router、logic、comet、job模块(配置文件请依据实际机器环境配置)
```sh
$ cd $GOPATH/src/github.com/Terry-Mao/goim/router
$ go install
$ cp router-example.conf $GOPATH/bin/router.conf
$ cp router-log.xml $GOPATH/bin/
$ cd ../logic/
$ go install
$ cp logic-example.conf $GOPATH/bin/logic.conf
$ cp logic-log.xml $GOPATH/bin/
$ cd ../comet/
$ go install
$ cp comet-example.conf $GOPATH/bin/comet.conf
$ cp comet-log.xml $GOPATH/bin/
$ cd ../logic/job/
$ go install
$ cp job-example.conf $GOPATH/bin/job.conf
$ cp job-log.xml $GOPATH/bin/
```
到此所有的环境都搭建完成！

### 五、启动goim
```sh
$ cd /$GOPATH/bin
$ nohup $GOPATH/bin/router -c $GOPATH/bin/router.conf 2>&1 > /data/logs/goim/panic-router.log &
$ nohup $GOPATH/bin/logic -c $GOPATH/bin/logic.conf 2>&1 > /data/logs/goim/panic-logic.log &
$ nohup $GOPATH/bin/comet -c $GOPATH/bin/comet.conf 2>&1 > /data/logs/goim/panic-comet.log &
$ nohup $GOPATH/bin/job -c $GOPATH/bin/job.conf 2>&1 > /data/logs/goim/panic-job.log &
```
如果启动失败，默认配置可通过查看panic-xxx.log日志文件来排查各个模块问题.

### 六、测试

推送协议可查看[push http协议文档](https://github.com/Terry-Mao/goim/blob/master/doc/push.md)

## 配置

TODO

## 例子

Websocket: [Websocket Client Demo](https://github.com/Terry-Mao/goim/tree/master/examples/javascript)

## 文档
[push http协议文档](https://github.com/Terry-Mao/goim/blob/master/doc/push.md)推送接口

##更多
TODO
