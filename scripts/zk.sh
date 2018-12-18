#!/bin/bash
USAGE="./zk.sh myid ip_array\r\n
eg:./zk.sh 1 127.0.0.1:2181,127.0.0.2:2181"
MYID=$1
str=$2
IP_ARRAY=(${str//,/ })
IP_NUM=${#IP_ARRAY[@]}
if [ ! -n "$1" ];then
echo -e $USAGE
exit 1
fi
if [ ! -n "$2" ];then
echo -e $USAGE
exit 1
fi 
echo $MYID
echo ${IP_ARRAY[@]}

ZK_VER=3.4.12
JDK=/usr/local/jdk8
ZK_PATH=/usr/local/zookeeper-${ZK_VER}
WORK_DIR=/data/server/zookeeper
LOG_DIR=/data/log/zookeeper
SUPERVISOR_PATH=/etc/supervisor/conf.d/zookeeper.conf

set -e

curl -L http://apache.website-solution.net/zookeeper/zookeeper-3.4.12/zookeeper-3.4.12.tar.gz -o zookeeper-${ZK_VER}.tar.gz
tar zxf zookeeper-${ZK_VER}.tar.gz
mkdir -p $ZK_PATH
mkdir -p $LOG_DIR
mkdir -p $WORK_DIR
echo $MYID>$WORK_DIR/myid
mv zookeeper-${ZK_VER}/* $ZK_PATH
echo "tickTime=2000" > $ZK_PATH/conf/zoo.cfg
echo "initLimit=10">> $ZK_PATH/conf/zoo.cfg 
echo "syncLimit=5" >>$ZK_PATH/conf/zoo.cfg
echo "dataDir=/data/server/zookeeper" >>$ZK_PATH/conf/zoo.cfg
echo "clientPort=2181" >>$ZK_PATH/conf/zoo.cfg
echo "autopurge.snapRetainCount=5" >>$ZK_PATH/conf/zoo.cfg
echo "autopurge.purgeInterval=24" >>$ZK_PATH/conf/zoo.cfg
for ((index=0;index<$IP_NUM;index++))
do
    tmp=$[$index+1]
    echo "server.$tmp=${IP_ARRAY[index]}:2888:3888" >>$ZK_PATH/conf/zoo.cfg 
done

### start by supervisor
apt-get install -y supervisor

echo "[program:zookeeper]">$SUPERVISOR_PATH
echo "command=/usr/local/zookeeper-${ZK_VER}/bin/zkServer.sh start-foreground">>$SUPERVISOR_PATH
echo "directory=/usr/local/zookeeper-${ZK_VER}">>$SUPERVISOR_PATH
echo 'user=root
autostart=true
autorestart=true
stopsignal=KILL
startsecs=10
startretries=3
stdout_logfile = /data/log/zookeeper/stdout.log
stdout_logfile_backups = 3
stderr_logfile = /data/log/zookeeper/stderr.log
stderr_logfile_backups = 3
logfile_maxbytes=20MB'>>$SUPERVISOR_PATH
echo "environment=JAVA_HOME=\"${JDK}\",JRE_HOME=\"${JDK}\"/jre\"">>$SUPERVISOR_PATH
supervisorctl update