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

## Protocol
![proto](https://github.com/Terry-Mao/goim/blob/master/doc/protocol.png)

## Document
[English](./README_en.md)

[中文](./README_cn.md)

## Examples
Websocket: [Websocket Client Demo](https://github.com/Terry-Mao/goim/tree/master/examples/javascript)
Android: [Android](https://github.com/roamdy/goim-sdk)
iOS: [iOS](https://github.com/roamdy/goim-oc-sdk)

## LICENSE
goim is is distributed under the terms of the GNU General Public License, version 3.0 [GPLv3](http://www.gnu.org/licenses/gpl.txt)
