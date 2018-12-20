#!/bin/bash
# config parameters
JDK_HOME=/usr/local/jdk8/
KAFKA_HOME=/data/app/kafka/
KAFKA_CONF=/data/app/kafka/config/server.properties 
KAFKA_DATA=/data/kafka_data/
KAFKA_SUPERVISOR=/etc/supervisor/conf.d/kafka.conf

USAGE="./kafka.sh {broker.id} {current.addr} {zk.addr1,zk.addr2,zk.addr3}\r\neg: ./kafka.sh 1 127.0.0.1 127.0.0.1:2181,127.0.0.2:2181,127.0.0.3:2181"
# check parammeters
BROKER_ID=$1
CURRENT_ADDR=$2
ZK_ADDRS=$3
if [ ! "$BROKER_ID" ]; then
echo "kafka {broker.id} can not be null"
echo -e  $USAGE
exit 1
fi
if [ ! "$CURRENT_ADDR" ]; then
echo "current addr can not be null"
exit 1
fi
if [ ! "$ZK_ADDRS" ]; then
echo "zookeeper addrs can not be null"
exit 1
fi
if [ ! -d "$JDK_HOME" ]; then
echo "jdk not exist down jdk8"
./jdk8.sh
fi

# download kafka
curl "http://apache.stu.edu.tw/kafka/2.1.0/kafka_2.11-2.1.0.tgz"  -o kafka_2.11-2.1.0.tgz
tar zxf kafka_2.11-2.1.0.tgz
# install kafka
mkdir -p $KAFKA_HOME
mkdir -p $KAFKA_DATA
mv kafka_2.11-2.1.0/* $KAFKA_HOME
# kafka config
echo "broker.id=$BROKER_ID">$KAFKA_CONF
echo "host.name=$CURRENT_ADDR">>$KAFKA_CONF
echo "zookeeper.connect=$ZK_ADDRS/kafka">>$KAFKA_CONF
echo "num.network.threads=4
num.io.threads=4
socket.send.buffer.bytes=1024000
socket.receive.buffer.bytes=1024000
socket.request.max.bytes=52428800
log.dirs=$KAFKA_DATA
num.partitions=9
num.recovery.threads.per.data.dir=1
log.cleanup.policy=delete
log.retention.hours=24
log.segment.bytes=536870912
log.retention.check.interval.ms=300000
log.cleaner.enable=false
zookeeper.connection.timeout.ms=6000
default.replication.factor=2
delete.topic.enable=false
auto.create.topics.enable=false">>$KAFKA_CONF

# install supervisor
if ! type "supervisorctl" > /dev/null; then
    apt-get install -y supervisor
fi
# kafka supervisor config
mkdir -p /etc/supervisor/conf.d/
mkdir -p /data/log/kafka/
echo "[program:kafka]
command=$KAFKA_HOME/bin/kafka-server-start.sh $KAFKA_CONF
user=root
autostart=true
autorestart=true
exitcodes=0
startsecs=10
startretries=10
stopwaitsecs=10
stopsignal=KILL
stdout_logfile=/data/log/kafka/stdout.log
stderr_logfile=/data/log/kafka/stderr.log
stdout_logfile_maxbytes=100MB
stdout_logfile_backups=5
stderr_logfile_maxbytes=100MB
stderr_logfile_backups=5
environment=JAVA_HOME=$JDK_HOME,JRE_HOME='$JDK_HOME/jre',KAFKA_HEAP_OPTS='-Xmx6g -Xms6g -XX:MetaspaceSize=96m -XX:G1HeapRegionSize=16M -XX:MinMetaspaceFreeRatio=50 -XX:MaxMetaspaceFreeRatio=80'">$_KAFKA_SUPERVISOR

supervisorctl update
