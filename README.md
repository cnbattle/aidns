# AIDNS

[简体中文](./README_ZH.md)

A lightweight DNS server that provides HTTP management interface, based on [CoreDNS](https://github.com/coredns/coredns)

## For learning and research purposes only

## Characteristics

- Based on CoreDNS development
- Data stored in MySQL
- Provide concise HttpApi ( Authentication can be configured )

## TODO

- [x] Add `read-through` cache processing solution
- [ ] Add web management page

## AIDNS config

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

- `dsn` DSN for MySQL as per https://github.com/go-sql-driver/mysql#dsn-data-source-name examples. You can
  use `$ENV_NAME` format in the DSN,
  and it will be replaced with the environment variable value.
- `table_prefix` Prefix for the MySQL tables. Defaults to `aidns_`.
- `max_lifetime` Duration (in Golang format) for a SQL connection. Default is 1 minute.
- `max_open_connections` Maximum number of open connections to the database server. Default is 10.
- `max_idle_connections` Maximum number of idle connections in the database connection pool. Default is 10.
- `ttl` Default TTL for records without a specified TTL in seconds. Default is 360 (seconds)
- `zone_update_interval` Maximum time interval between loading all the zones from the database. Default is 10 minutes.
- `http_token` Http Api Server authorization token. Default is empty, is no authorization.
- `http_addr` Http Api Server Addr. Default is :8888.
- `redis_url` URL for Redis as per https://github.com/redis/go-redis#connecting-via-a-redis-url examples. Default is
  empty, not cache.
- `redis_ttl` Redis cache time. Default is 10 minutes.

#### CoreDNS full config example

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
        redis_url redis://:123456@localhost:6379/0?dial_timeout=3&read_timeout=6s&max_retries=2
        redis_ttl 10m
    }
    loop
    reload
    loadbalance
}
```

## Supported Record Types

`A`, `AAAA`, `CNAME`, `SOA`, `TXT`, `NS`, `MX`, `CAA` and `SRV`. Wildcard records are supported as well. This backend
doesn't support `AXFR` requests.

## Build

```shell script
$ make
```

## Database Setup

This plugin doesn't create or migrate database schema for its use yet. To create the database and tables, use the
following table structure (note the table name prefix):

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

## Management Records API

[API document](./docs/APIS.md)

### Acknowledgements and Credits

- https://github.com/coredns/coredns
- https://github.com/cloud66-oss/coredns_mysql

