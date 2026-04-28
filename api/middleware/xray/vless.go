package xray

import (
	"fmt"
	"net/url"
)

type Params struct {
	UUID string
	Host string
	Port int
	PBK  string
	SNI  string
	SID  string
	FP   string
	Flow string
	Name string
}

func Link(p Params) (string, error) {
	q := url.Values{}
	q.Set("type", "tcp")
	q.Set("security", "reality")
	q.Set("pbk", p.PBK)
	q.Set("fp", p.FP)
	q.Set("sni", p.SNI)
	q.Set("sid", p.SID)
	if p.Flow != "" {
		q.Set("flow", p.Flow)
	}

	frag := url.PathEscape(p.Name)
	return fmt.Sprintf("xray://%s@%s:%d?%s#%s",
		p.UUID, p.Host, p.Port, q.Encode(), frag,
	), nil
}
