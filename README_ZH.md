# AIDNS

[English](./README.md)

一个轻量级的DNS服务器，提供HTTP管理接口，基于 [CoreDNS](https://github.com/coredns/coredns)

## 仅供学习研究

## 特点

- 基于 CoreDNS 开发
- 数据存储于 MySQL
- 提供简洁的 HttpApi （可以配置身份验证）

## TODO

- [x] 增加 `read-through` 缓存处理方案
- [ ] 增加 web 管理页面

## AIDNS配置

```
aidns {
    dsn DSN
    [table_prefix TABLE_PREFIX]
    [max_lifetime MAX_LIFETIME]
    [max_open_connections MAX_OPEN_CONNECTIONS]
    [max_idle_connections MAX_IDLE_CONNECTIONS]
    [ttl DEFAULT_TTL]
    [zone_update_interval ZONE_UPDATE_INTERVAL]
    [zone_update_interval ZONE_UPDATE_INTERVAL]
    [http_token HTTP_TOKEN]
    [http_addr HTTP_ADDR]
    [redis_url REDIS_URL]
}
```

- `dsn` MySQL 的 DSN，按照 https://github.com/go-sql-driver/mysql#dsn-data-source-name 示例。 您可以在 DSN
  中使用 `$ENV_NAME` 格式，并将其替换为环境变量值。
- `table_prefix` MySQL 表的前缀。 默认为 `aidns_`。
- `max_lifetime` SQL 连接的持续时间（Golang 格式）。 默认值为 1 分钟。
- `max_open_connections` 数据库服务器的最大打开连接数。默认值为 10。
- `max_idle_connections` 数据库连接池中的最大空闲连接数。默认值为 10。
- `ttl` 没有指定 TTL（以秒为单位）的记录的默认 TTL。默认值为 360（秒）。
- `zone_update_interval` 从数据库加载所有区域之间的最大时间间隔。 默认值为 10 分钟。
- `http_token` Http API 服务器授权Token。 默认为空，不需要授权。
- `http_addr` Http API 服务器地址。 默认值为 :8888。
- `redis_url` Redis 的 URL，按照 https://github.com/redis/go-redis#connecting-via-a-redis-url 示例。默认为空，不缓存。
- `redis_ttl` Redis 缓存时间，默认值为 10 分钟。

#### 完整 CoreDNS 配置示例

```Corefile
.:53 {
    log
    health {
       lameduck 15s
    }
    ready
    aidns {
        dsn root:123456@(localhost:3306)/dev?charset=utf8mb4&parseTime=True&loc=Local
        http_token aidns
        http_addr :8888
        redis_url redis://:123456@localhost:30603/0?dial_timeout=3&read_timeout=6s&max_retries=2
        redis_ttl 10m
    }
    loop
    reload
    loadbalance
}
```

## 支持的记录类型

`A`, `AAAA`, `CNAME`, `SOA`, `TXT`, `NS`, `MX`, `CAA` and `SRV` 还支持通配符记录，不支持`AXFR`请求。

## 编译

```shell script
$ make
```

## 数据库设置

该插件尚未创建或迁移数据库架构以供其使用。 要创建数据库和表，请使用表结构如下（注意表名前缀）：

```sql
CREATE TABLE `aidns_records`
(
    `id`          INT          NOT NULL AUTO_INCREMENT,
    `zone`        VARCHAR(255) NOT NULL,
    `name`        VARCHAR(255) NOT NULL,
    `ttl`         INT DEFAULT NULL,
    `content`     TEXT,
    `record_type` VARCHAR(255) NOT NULL,
    PRIMARY KEY (`id`)
) ENGINE = INNODB
  AUTO_INCREMENT = 6
  DEFAULT CHARSET = utf8mb4
  COLLATE = utf8mb4_0900_ai_ci;
```

## 管理记录接口

[API 文档](./docs/APIS_ZH.md)

### 致谢

- https://github.com/coredns/coredns
- https://github.com/cloud66-oss/coredns_mysql
