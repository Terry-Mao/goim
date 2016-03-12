<h3>Terry-Mao/goim push HTTP protocols</h3>
push HTTP interface protocols for pusher

<h3>Interfaces</h3>
| Name | URL | HTTP method |
| :---- | :---- | :---- |
| [single push](#single push)  | /1/push       | POST |
| [multiple push](#multiple push) | /1/pushs      | POST |
| [room push](#room push) | /1/push/room   | POST |
| [broadcasting](#broadcasting) | /1/push/all   | POST |

<h3>Public response body</h3>

| response code | description |
| :---- | :---- |
| 1 | success |
| 65535 | internal error |

<h3>Response structure</h3>
<pre>
{
    "ret": 1  //response code
}
</pre>


##### single push
 * Example request

```sh
# uid is the user id pushing to?uid=0
curl -d "{\"test\":1}" http://127.0.0.1:7172/1/push?uid=0
```

 * Response

<pre>
{
    "ret": 1
}
</pre>

##### Multiple push
 * Example request

```sh
curl -d "{\"u\":[1,2,3,4,5],\"m\":{\"test\":1}}" http://127.0.0.1:7172/1/pushs
```

 * Response

<pre>
{
    "ret": 1
}
</pre>

##### room push
 * Example request

```sh
curl -d "{\"test\": 1}" http://127.0.0.1:7172/1/push/room?rid=1
```

 * Response

<pre>
{
    "ret": 1
}
</pre>

##### Broadcasting
 * Example request

```sh
curl -d "{\"test\": 1}" http://127.0.0.1:7172/1/push/all
```

 * Response

<pre>
{
    "ret": 1
}
</pre>
