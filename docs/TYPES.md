| record type| content |
|---|---|
| A | {"ip":"10.11.12.13"} |
| AAAA | {"ip":"2001:db8::3"} |
| CNAME | {"host":"http://example.org"} |
| TXT | {"text":"hello txt"} |
| MX | {"host":"http://example.org","preference": 10} |
| NS | {"host":"http://example.org"} |
| SRV | {"target":"http://example.org","priority": 10,"weight": 10,"port": 8080} |
| SOA | {"ns":"ns1.www","MBox": "hostmaster.www","refresh":86400,"retry":7200,"expire":3600,"minttl":60} |
| CAA | {"flag":0,"tag":"issue","value":"example.org"} |
