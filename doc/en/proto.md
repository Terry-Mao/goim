# comet and clients protocols
comet supports two protocols to communicate with client: WebSocket, TCP

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

