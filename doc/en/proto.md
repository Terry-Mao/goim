# comet and clients protocols
comet supports three protocols to communicate with client: HTTP long polling, WebSocket, TCP

## http long polling      
   
**Request URL**

http://DOMAIN/sub?param=value

**Request Method**

GET

**Request Parameters**

| parameter     | is required  | type | comment|
| :-----     | :---  | :--- | :---       |
| ver        | true  | int | Protocol version |
| op         | true  | int    | Operation |
| seq        | true  | int    | Sequence number (Server returned number maps to client sent) |
| t          | true | string | Authentication ticket. User to verify and get user's real ID |
| cb         | false | string | jsonp callback |

**Response Result**

```json
{
    "ver": 102,
    "op": 10,
    "seq": 10,
    "body": {"data": "xxx"}
}
```

**Response Fields**

| response key  | type     |  comment|
| :----:      | :---:        | :-----:|
| ver        | int          | Protocol version|
| op        | int          | Operation |
| seq        | int          | Sequence number|
| body        | json          | The JSON message pushed |

**HTTP response code**

| response code | comment         |
| :----       | :---         |
| 200           | Success |
| 403           |  Authenticate failed |
| 500           |  Internal Server Error|

## websocket                                                                   
**Request URL**

ws://DOMAIN/sub

**HTTP Request Method**

WebSocket (JSON Frame). Response is same as the request.

**Response Result**

```json
{
    "ver": 102,
    "op": 10,
    "seq": 10,
    "body": {"data": "xxx"}
}
```

**Request and Response Parameters**

| parameter     | is required  | type | comment|
| :-----     | :---  | :--- | :---       |
| ver        | true  | int | Protocol version |
| op         | true  | int    | Operation |
| seq        | true  | int    | Sequence number (Server returned number maps to client sent) |
| body        | json          | The JSON message pushed |

## tcp                                                                         
**Request URL**

tcp://DOMAIN

**Protocol**

Binary. Response is same as the request.

**Request and Response Parameters**

| parameter     | is required  | type | comment|
| :-----     | :---  | :--- | :---       |
| package length        | true  | int32 bigendian | package length |
| header Length         | true  | int16 bigendian    | header length |
| ver        | true  | int16 bigendian    | Protocol version |
| operation          | true | int32 bigendian | Operation |
| seq         | true | int32 bigendian | jsonp callback |
| body         | false | binary | $(package lenth) - $(header length) |

## Operations
| operation     | comment | 
| :-----     | :---  |
| 2 | Client send heartbeat|
| 3 | Server reply heartbeat|
| 7 | authentication request |
| 8 | authentication response |

