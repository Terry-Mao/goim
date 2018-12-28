## goim push API

### error codes
```
// ok
OK = 0

// request error
RequestErr = -400

// server error
ServerErr = -500
```

### push keys
[POST] /goim/push/keys

| Name            | Type     | Remork                 |
|:----------------|:--------:|:-----------------------|
| [url]:operation | int32    | operation for response |
| [url]:keys      | []string | multiple client keys   |
| [Body]          | []byte   | http request body      |

response:
```
{
    "code": 0
}
```

### push mids
[POST] /goim/push/mids

| Name            | Type     | Remork                 |
|:----------------|:--------:|:-----------------------|
| [url]:operation | int32    | operation for response |
| [url]:mids      | []int64  | multiple user mids     |
| [Body]          | []byte   | http request body      |

response:
```
{
    "code": 0
}
```

### push room
[POST] /goim/push/room

| Name            | Type     | Remork                 |
|:----------------|:--------:|:-----------------------|
| [url]:operation | int32    | operation for response |
| [url]:type      | string   | room type              |
| [url]:room      | string   | room id                |
| [Body]          | []byte   | http request body      |

response:
```
{
    "code": 0
}
```

### push all
[POST] /goim/push/all

| Name            | Type     | Remork                 |
|:----------------|:--------:|:-----------------------|
| [url]:operation | int32    | operation for response |
| [url]:speed     | int32    | push speed             |
| [Body]          | []byte   | http request body      |

response:
```
{
    "code": 0
}
```

### online top
[GET] /goim/online/top

| Name    | Type     | Remork                 |
|:--------|:--------:|:-----------------------|
| type    | string   | room type              |
| limit   | string   | online limit           |

response:
```
{
    "code": 0,
    "message": "",
    "data": [
        {
            "room_id": "1000",
            "count": 100
        },
        {
            "room_id": "2000",
            "count": 200
        },
        {
            "room_id": "3000",
            "count": 300
        }
    ]
}
```

### online room
[GET] /goim/online/room

| Name    | Type     | Remork                 |
|:--------|:--------:|:-----------------------|
| type    | string   | room type              |
| rooms   | []string | room ids               |

response:
```
{
    "code": 0,
    "message": "",
    "data": {
        "1000": 100,
        "2000": 200,
        "3000": 300
    }
}
```
### online total
[GET] /goim/online/total

response:
```
{
    "code": 0,
    "message": "",
    "data": {
        "conn_count": 1,
        "ip_count": 1
    }
}
```

### nodes weighted
[GET] /goim/nodes/weighted

| Name     | Type     | Remork                 |
|:---------|:--------:|:-----------------------|
| platform | string   | web/android/ios        |

response:
```
{
    "code": 0,
    "message": "",
    "data": {
        "domain": "conn.goim.io",
        "tcp_port": 3101,
        "ws_port": 3102,
        "wss_port": 3103,
        "heartbeat": 30,    // heartbeat seconds
        "heartbeat_max": 3  // heartbeat tries
        "nodes": [
            "47.89.10.97"
        ],
        "backoff": {
            "max_delay": 300,
            "base_delay": 3,
            "factor": 1.8,
            "jitter": 0.3
        },
        
    }
}
```

### nodes instances
[GET] /nodes/instances

response:
```
{
    "code": 0,
    "message": "",
    "data": [
        {
            "region": "sh",
            "zone": "sh001",
            "env": "dev",
            "appid": "goim.comet",
            "hostname": "test",
            "addrs": [
                "grpc://192.168.1.30:3109"
            ],
            "version": "",
            "latest_timestamp": 1545750122311688676,
            "metadata": {
                "addrs": "47.89.10.97",
                "conn_count": "1",
                "ip_count": "1",
                "offline": "false",
                "weight": "10"
            }
        }
    ]
}
`
