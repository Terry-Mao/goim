# comet 客户端通讯协议文档                                                     
comet支持三种协议和客户端通讯，http long polling, websocket， tcp。
## http long polling      
   
**请求URL**

http://DOMAIN/sub?param=value

**HTTP请求方式**

GET

**请求参数**

| 参数名     | 必选  | 类型 | 说明       |
| :-----     | :---  | :--- | :---       |
| ver        | true  | int | 协议版本号 |
| op         | true  | int    | 指令 |
| seq        | true  | int    | 序列号（服务端返回和客户端发送一一对应） |
| t          | true | string | 授权令牌，用于检验获取用户真实用户Id |
| cb         | false | string | jsonp callback |

**返回结果**

```json
{
    "ver": 102,
    "op": 10,
    "seq": 10,
    "body": {"data": "xxx"}
}
```

**字段说明**

| 返回值字段  | 字段类型     |  字段说明        |
| :----:      | :---:        | :-----:          |
| ver        | int          | 协议版本          |
| op        | int          | 指令          |
| seq        | int          | 序列号          |
| body        | json          | 业务方推送数据（必须是json）          |

**http状态吗说明**

| 返回码      | 说明         |
| :----       | :---         |
| 200           | 请求成功     |
| 403           |  认证失败     |
| 500           |  内部错误     |

## websocket                                                                   
**请求URL**

ws://DOMAIN/sub

**HTTP请求方式**

Websocket（JSON Frame），请求和返回协议一致

**请求和返回json**

```json
{
    "ver": 102,
    "op": 10,
    "seq": 10,
    "body": {"data": "xxx"}
}
```

**请求和返回参数说明**

| 参数名     | 必选  | 类型 | 说明       |
| :-----     | :---  | :--- | :---       |
| ver        | true  | int | 协议版本号 |
| op         | true  | int    | 指令 |
| seq        | true  | int    | 序列号（服务端返回和客户端发送一一对应） |
| body          | true | string | 授权令牌，用于检验获取用户真实用户Id |

## tcp                                                                         
**请求URL**

tcp://DOMAIN

**协议格式**

二进制，请求和返回协议一致

**请求&返回参数**

| 参数名     | 必选  | 类型 | 说明       |
| :-----     | :---  | :--- | :---       |
| package length        | true  | int32 bigendian | 包长度 |
| header Length         | true  | int16 bigendian    | 包头长度 |
| ver        | true  | int16 bigendian    | 协议版本 |
| operation          | true | int32 bigendian | 协议指令 |
| seq         | true | int32 bigendian | jsonp callback |
| body         | false | binary | $(package lenth) - $(header length) |

## 指令
| 指令     | 说明  | 
| :-----     | :---  |
| 2 | 客户端请求心跳 |
| 3 | 服务端心跳答复 |
| 7 | auth认证 |
| 8 | auth认证返回 |

