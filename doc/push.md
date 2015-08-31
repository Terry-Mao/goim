<h3>Terry-Mao/goim push http协议文档</h3>
push http接口文档，用于推送接口接入

<h3>接口汇总</h3>
(head). | 接口名 | URL | 访问方式 |
| "单个推送":Push | /1/push | POST |
| "批量推送":Pushs    | /1/pushs    | POST |
| "广播":Broadcast | /1/push/all   | POST |

<h3>公共返回码</h3>

(head). | 错误码 | 描述 |
| 1 | 成功 |
| 65535 | 内部错误 |

<h3>基本返回结构</h3>
<pre>
{
    "ret": 1  //错误码
}
</pre>


<h3>单个推送</h3>
 * 请求例子

```sh
curl -d "{\"test\":1}" http://127.0.0.1:7172/1/push
```

 * 返回

<pre>
{
    "ret": 1
}
</pre>

<h3>批量推送</h3>
 * 请求例子

```sh
curl -d "{\"u\":[1,2,3,4,5],\"m\":{\"test\":1}}" http://127.0.0.1:7172/1/pushs
```

 * 返回

<pre>
{
    "ret": 1
}
</pre>

<h3>广播</h3>
 * 请求例子

```sh
curl -d "{\"test\": 1}" http://172.16.0.239:7172/1/push/all
```

 * 返回

<pre>
{
    "ret": 1
}
</pre>

[Push]#单个推送
[Pushs]#批量推送
[Broadcast]#广播
