#!/bin/bash
# config parameters
JDK_HOME=/usr/local/jdk8
ZK_HOME=/data/server/zookeeper
ZK_SUPERVISOR=/etc/supervisor/conf.d/zookeeper.conf

USAGE="./zk.sh {zk.id} {zk.addr1,zk.addr2,zk.addr3}\r\neg:./zk.sh 1 127.0.0.1:2181,127.0.0.2:2181,127.0.0.3:2181"
# check parammeters
ZK_ID=$1
str=$2
ZK_ADDRS=(${str//,/ })
ZK_ADDRS_SIZE=${#ZK_ADDRS[@]}
if [ ! -n "$1" ];then
echo -e $USAGE
exit 1
fi
if [ ! -n "$2" ];then
echo -e $USAGE
exit 1
fi 
echo $ZK_ID
echo ${ZK_ADDRS[@]}


set -e

curl -L http://apache.website-solution.net/zookeeper/zookeeper-3.4.12/zookeeper-3.4.12.tar.gz -o zookeeper-3.4.12.tar.gz
tar zxf zookeeper-3.4.12.tar.gz
mkdir -p $ZK_HOME
echo $ZK_ID>$ZK_HOME/myid
mv zookeeper-3.4.12/* $ZK_HOME
echo "tickTime=2000" > $ZK_HOME/conf/zoo.cfg
echo "initLimit=10" >> $ZK_HOME/conf/zoo.cfg 
echo "syncLimit=5" >> $ZK_HOME/conf/zoo.cfg
echo "dataDir=/data/server/zookeeper" >> $ZK_HOME/conf/zoo.cfg
echo "clientPort=2181" >> $ZK_HOME/conf/zoo.cfg
echo "autopurge.snapRetainCount=5" >> $ZK_HOME/conf/zoo.cfg
echo "autopurge.purgeInterval=24" >> $ZK_HOME/conf/zoo.cfg
for ((index=0;index<$ZK_ADDRS_SIZE;index++))
do
    id=$[$index+1]
    echo "server.$id=${ZK_ADDRS[index]}:2888:3888" >> $ZK_HOME/conf/zoo.cfg 
done

# install supervisor
if ! type "supervisorctl" > /dev/null; then
    apt-get install -y supervisor
fi
# zookeeper supervisor config
mkdir -p /etc/supervisor/conf.d/
mkdir -p /data/log/zookeeper/
echo "[program:zookeeper]
command=${ZK_HOME}/bin/zkServer.sh start-foreground
directory=${ZK_HOME}
user=root
autostart=true
autorestart=true
exitcodes=0
startsecs=10
startretries=10
stopwaitsecs=10
stopsignal=KILL
stdout_logfile=/data/log/zookeeper/stdout.log
stderr_logfile=/data/log/zookeeper/stderr.log
stdout_logfile_maxbytes=100MB
stdout_logfile_backups=5
stderr_logfile_maxbytes=100MB
stderr_logfile_backups=5
environment=JAVA_HOME=$JDK_HOME,JRE_HOME='$JDK_HOME/jre'
">$ZK_SUPERVISOR

supervisorctl update
