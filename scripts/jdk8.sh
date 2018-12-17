#!/bin/bash
curl -L https://download.oracle.com/otn-pub/java/jdk/8u191-b12/2787e4a523244c269598db4e85c51e0c/jdk-8u191-linux-x64.tar.gz -o jdk-8u191-linux-x64.tar.gz
t
tar zxf jdk-8u191-linux-x64.tar.gz 
mkdir -p /usr/local/jdk8/
mv jdk1.8.0/* /usr/local/jdk8/
update-alternatives --install "/usr/bin/java" "java" "/usr/local/jdk8/bin/java" 1500
update-alternatives --install "/usr/bin/javac" "javac" "/usr/local/jdk8/bin/javac" 1500
update-alternatives --install "/usr/bin/javaws" "javaws" "/usr/local/jdk8/bin/javaws" 1500