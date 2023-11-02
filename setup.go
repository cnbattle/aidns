package aidns

import (
	"context"
	"database/sql"
	"os"
	"strconv"
	"time"

	"github.com/coredns/caddy"
	"github.com/redis/go-redis/v9"

	"github.com/coredns/coredns/core/dnsserver"
	"github.com/coredns/coredns/plugin"
	clog "github.com/coredns/coredns/plugin/pkg/log"
)

var log = clog.NewWithPlugin("aidns")

const (
	defaultTtl                = 360
	defaultMaxLifeTime        = 1 * time.Minute
	defaultMaxOpenConnections = 10
	defaultMaxIdleConnections = 10
	defaultZoneUpdateTime     = 10 * time.Minute
	defaultRedisTtl           = 10 * time.Minute
)

func init() {
	caddy.RegisterPlugin("aidns", caddy.Plugin{
		ServerType: "dns",
		Action:     setup,
	})
}

func setup(c *caddy.Controller) error {
	r, err := mysqlParse(c)
	if err != nil {
		return plugin.Error("mysql", err)
	}

	dnsserver.GetConfig(c).AddPlugin(func(next plugin.Handler) plugin.Handler {
		r.Next = next
		return r
	})

	return nil
}

func mysqlParse(c *caddy.Controller) (*AiDNS, error) {
	aiDNS := AiDNS{
		TablePrefix: "aidns_",
		Ttl:         300,
		HttpAddr:    ":8888",
	}
	var err error

	//c.OnFirstStartup(func() error {
	//	log.Info("OnFirstStartup message")
	//	return nil
	//})

	c.OnStartup(func() error { return aiDNS.Server() })
	//c.OnRestartFailed(func() error {
	//	log.Info("OnRestartFailed message")
	//	return nil
	//})

	//c.OnRestart(func() error {
	//	log.Info("OnRestart message")
	//	return nil
	//})
	//c.OnRestartFailed(func() error {
	//	log.Info("OnRestartFailed message")
	//	return nil
	//})

	c.OnShutdown(func() error { return aiDNS.db.Close() })

	c.Next()
	if c.NextBlock() {
		for {
			switch c.Val() {
			case "dsn":
				if !c.NextArg() {
					return &AiDNS{}, c.ArgErr()
				}
				aiDNS.Dsn = c.Val()
			case "table_prefix":
				if !c.NextArg() {
					return &AiDNS{}, c.ArgErr()
				}
				aiDNS.TablePrefix = c.Val()
			case "max_lifetime":
				if !c.NextArg() {
					return &AiDNS{}, c.ArgErr()
				}
				var val time.Duration
				val, err = time.ParseDuration(c.Val())
				if err != nil {
					val = defaultMaxLifeTime
				}
				aiDNS.MaxLifetime = val
			case "max_open_connections":
				if !c.NextArg() {
					return &AiDNS{}, c.ArgErr()
				}
				var val int
				val, err = strconv.Atoi(c.Val())
				if err != nil {
					val = defaultMaxOpenConnections
				}
				aiDNS.MaxOpenConnections = val
			case "max_idle_connections":
				if !c.NextArg() {
					return &AiDNS{}, c.ArgErr()
				}
				var val int
				val, err = strconv.Atoi(c.Val())
				if err != nil {
					val = defaultMaxIdleConnections
				}
				aiDNS.MaxIdleConnections = val
			case "zone_update_interval":
				if !c.NextArg() {
					return &AiDNS{}, c.ArgErr()
				}
				var val time.Duration
				val, err = time.ParseDuration(c.Val())
				if err != nil {
					val = defaultZoneUpdateTime
				}
				aiDNS.zoneUpdateTime = val
			case "ttl":
				if !c.NextArg() {
					return &AiDNS{}, c.ArgErr()
				}
				var val int
				val, err = strconv.Atoi(c.Val())
				if err != nil {
					val = defaultTtl
				}
				aiDNS.Ttl = uint32(val)
			case "http_token":
				if !c.NextArg() {
					return &AiDNS{}, c.ArgErr()
				}
				aiDNS.HttpToken = c.Val()
			case "http_addr":
				if !c.NextArg() {
					return &AiDNS{}, c.ArgErr()
				}
				aiDNS.HttpAddr = c.Val()
			case "redis_url":
				if !c.NextArg() {
					return &AiDNS{}, c.ArgErr()
				}
				aiDNS.RedisURL = c.Val()
			case "redis_ttl":
				if !c.NextArg() {
					return &AiDNS{}, c.ArgErr()
				}
				var val time.Duration
				val, err = time.ParseDuration(c.Val())
				if err != nil {
					val = defaultRedisTtl
				}
				aiDNS.RedisTTL = val
			default:
				if c.Val() != "}" {
					return &AiDNS{}, c.Errf("unknown property '%s'", c.Val())
				}
			}

			if !c.Next() {
				break
			}
		}

	}

	aiDNS.tableName = aiDNS.TablePrefix + "records"

	db, err := sql.Open("mysql", os.ExpandEnv(aiDNS.Dsn))
	if err != nil {
		return nil, err
	}

	db.SetConnMaxLifetime(aiDNS.MaxLifetime)
	db.SetMaxOpenConns(aiDNS.MaxOpenConnections)
	db.SetMaxIdleConns(aiDNS.MaxIdleConnections)

	err = db.Ping()
	if err != nil {
		return nil, err
	}

	aiDNS.db = db

	if aiDNS.RedisURL != "" {
		redisOpt, err := redis.ParseURL(aiDNS.RedisURL)
		if err != nil {
			return nil, err
		}
		aiDNS.locker = NewLocker(context.Background(), aiDNS, redis.NewClient(redisOpt))
	}

	return &aiDNS, nil
}
