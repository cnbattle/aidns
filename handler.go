package aidns

import (
	"database/sql"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/miekg/dns"
	"golang.org/x/net/context"

	"github.com/coredns/coredns/plugin"
	"github.com/coredns/coredns/request"
)

type AiDNS struct {
	Next               plugin.Handler
	Dsn                string
	TablePrefix        string
	MaxLifetime        time.Duration
	MaxOpenConnections int
	MaxIdleConnections int
	Ttl                uint32
	HttpToken          string
	HttpAddr           string
	RedisURL           string
	RedisTTL           time.Duration

	tableName      string
	lastZoneUpdate time.Time
	zoneUpdateTime time.Duration
	zones          []string
	db             *sql.DB
	locker         *Locker
}

// ServeDNS implements the plugin.Handler interface.
func (handler *AiDNS) ServeDNS(ctx context.Context, w dns.ResponseWriter, r *dns.Msg) (int, error) {
	state := request.Request{W: w, Req: r}
	qName := state.Name()
	qType := state.Type()
	if time.Since(handler.lastZoneUpdate) > handler.zoneUpdateTime {
		err := handler.loadZones()
		if err != nil {
			return handler.errorResponse(state, dns.RcodeServerFailure, err)
		}
	}
	// 检查是否属于当前 Zone（如 example.com. 的查询才处理）
	qZone := plugin.Zones(handler.zones).Matches(qName)
	if qZone == "" {
		return plugin.NextOrFailure(handler.Name(), handler.Next, ctx, w, r)
	}
	// 查询数据库
	records, err := handler.findRecord(qZone, qName, qType)
	if err != nil {
		return handler.errorResponse(state, dns.RcodeServerFailure, err)
	}
	var appendSOA bool
	if len(records) == 0 {
		appendSOA = true
		// no record found but we are going to return a SOA
		recs, err := handler.findRecord(qZone, "", "SOA")
		if err != nil {
			return handler.errorResponse(state, dns.RcodeServerFailure, err)
		}
		records = append(records, recs...)
	}
	if qType == "SOA" {
		recsNs, err := handler.findRecord(qZone, qName, "NS")
		if err != nil {
			return handler.errorResponse(state, dns.RcodeServerFailure, err)
		}
		records = append(records, recsNs...)
	}
	if qType == "AXFR" {
		return handler.errorResponse(state, dns.RcodeNotImplemented, nil)
	}
	// 1. 提取查询名称和类型（如 "www.example.com." -> "www", "A"）
	answers := make([]dns.RR, 0, 10) // www
	extras := make([]dns.RR, 0, 10)  // A
	for _, record := range records {
		var answer dns.RR
		switch record.RecordType {
		case "A":
			answer, extras, err = record.AsARecord()
		case "AAAA":
			answer, extras, err = record.AsAAAARecord()
		case "CNAME":
			answer, extras, err = record.AsCNAMERecord()
		case "SOA":
			answer, extras, err = record.AsSOARecord()
		case "SRV":
			answer, extras, err = record.AsSRVRecord()
		case "NS":
			answer, extras, err = record.AsNSRecord()
		case "MX":
			answer, extras, err = record.AsMXRecord()
		case "TXT":
			answer, extras, err = record.AsTXTRecord()
		case "CAA":
			answer, extras, err = record.AsCAARecord()
		default:
			return handler.errorResponse(state, dns.RcodeNotImplemented, nil)
		}

		if err != nil {
			return handler.errorResponse(state, dns.RcodeServerFailure, err)
		}
		if answer != nil {
			answers = append(answers, answer)
		}
	}
	if len(answers) == 0 {
		// 无记录则转发到上游 DNS
		return plugin.NextOrFailure(handler.Name(), handler.Next, ctx, w, r)
	}
	//  构建 DNS 响应
	m := new(dns.Msg)
	m.SetReply(r)
	m.Authoritative = true
	m.RecursionAvailable = false
	m.Compress = true

	if !appendSOA {
		m.Answer = append(m.Answer, answers...)
	} else {
		m.Ns = append(m.Ns, answers...)
	}
	m.Extra = append(m.Extra, extras...)

	state.SizeAndDo(m)
	m = state.Scrub(m)
	_ = w.WriteMsg(m)
	return dns.RcodeSuccess, nil
}

// Name implements the Handler interface.
func (handler *AiDNS) Name() string { return "handler" }

func (handler *AiDNS) errorResponse(state request.Request, rCode int, err error) (int, error) {
	m := new(dns.Msg)
	m.SetRcode(state.Req, rCode)
	m.Authoritative, m.RecursionAvailable, m.Compress = true, false, true

	state.SizeAndDo(m)
	_ = state.W.WriteMsg(m)
	// Return success as the rCode to signal we have written to the client.
	return dns.RcodeSuccess, err
}
