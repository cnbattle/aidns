## 管理记录接口

##### 全局请求头 ( 如果你配置了 `http_token` )

| name          | type     | data type | description                    |
|---------------|----------|-----------|--------------------------------|
| Authorization | optional | string    | example: `Bearer <http_token>` |

<details>
 <summary><code>POST</code> <code><b>[host:port]/records</b></code> <code>增加/更新记录</code></summary>

##### Parameters

| name        | type     | data type | description                                                 |
|-------------|----------|-----------|-------------------------------------------------------------|
| id          | optional | uint      | create: `0`, update: `record id`, default: `0`              |
| zone        | required | string    | domain zone, example: `example.org.`                        |
| name        | required | string    | domain name, example: `www`, primary domain: `empty string` |
| record_type | required | string    | record_type                                                 |
| content     | required | string    | content json                                                |
| ttl         | required | uint      | ttl                                                         |

##### Responses

```json
{
  "code": 0
}
```
</details>

<details>
 <summary><code>GET</code> <code><b>[host:port]/records</b></code> <code>查询记录列表</code></summary>

##### Parameters

| name | type     | data type | description                          |
|------|----------|-----------|--------------------------------------|
| zone | required | string    | domain zone, example: `example.org.` |

##### Responses

```json
{
  "code": 0,
  "records": [
    {
      "id": 1,
      "zone": "example.org.",
      "name": "foo",
      "record_type": "A",
      "ttl": 30,
      "content": "{\"ip\": \"100.11.12.12\"}"
    },
    {
      "id": 2,
      "zone": "example.org.",
      "name": "foo",
      "record_type": "TXT",
      "ttl": 30,
      "content": "{\"text\": \"hello\"}"
    },
    {
      "id": 3,
      "zone": "example.org.",
      "name": "foo",
      "record_type": "MX",
      "ttl": 30,
      "content": "{\"host\" : \"foo.example.org.\",\"priority\" : 10}"
    }
  ]
}
```
</details>

<details>
 <summary><code>DELETE</code> <code><b>[host:port]/records</b></code> <code>删除记录</code></summary>

##### Parameters

| name | type     | data type | description                          |
|------|----------|-----------|--------------------------------------|
| zone | required | string    | domain zone, example: `example.org.` |
| id   | required | uint      | record data id                       |

##### Responses

```json
{
  "code": 0
}
```
</details>

使用 `dig` 查询，如下所示：

```shell script
$ dig @host A foo.example.org 
```
