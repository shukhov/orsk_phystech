package xray

import (
	"api/database"
	"encoding/base64"
	"fmt"
	"golang.org/x/crypto/curve25519"
	"math/rand"
	"net/url"
	"os"
)

type ConnectionLinkOut struct {
	ConnectionLink string `json:"connection_link"`
}

type ConnectionAccessOut struct {
	AccessKey string
	Alias     string
}

type Params struct {
	UUID string
	Host string
	Port int32
	PBK  string
	SNI  string
	SID  string
	FP   string
	Flow string
	Name string
}

func (xraySrv *XrayService) xrayPrivateToPublicBase64(privB64 string) (string, error) {
	priv, err := base64.RawURLEncoding.DecodeString(privB64)
	if err != nil {
		xraySrv.logger.Printf("base64url decode private key: %v", err)
		return "", database.InternalDBError
	}
	if len(priv) != 32 {
		xraySrv.logger.Printf("invalid private key length: got %d, want 32", len(priv))
		return "", database.InternalDBError
	}

	pub, err := curve25519.X25519(priv, curve25519.Basepoint)
	if err != nil {
		xraySrv.logger.Printf("x25519 derive public key: %v", err)
		return "", database.InternalDBError
	}

	return base64.RawURLEncoding.EncodeToString(pub), nil
}

func (xraySrv *XrayService) makeXrayLinkParams(connAccess *ConnectionAccessOut) (*Params, error) {
	pubKey, err := xraySrv.xrayPrivateToPublicBase64(os.Getenv("XRAY_PRIVATE_KEY"))
	if err != nil {
		return nil, err
	}
	params := Params{
		UUID: connAccess.AccessKey,
		Host: xraySrv.ConnParams.XrayHost,
		Port: xraySrv.ConnParams.XrayPort,
		PBK:  pubKey,
		SNI:  xraySrv.ConnParams.XrayServerNames[rand.Intn(len(xraySrv.ConnParams.XrayServerNames))],
		SID:  xraySrv.ConnParams.XrayShortIds[rand.Intn(len(xraySrv.ConnParams.XrayShortIds))],
		FP:   "chrome",
		Flow: "xtls-rprx-vision",
		Name: connAccess.Alias,
	}
	return &params, nil
}

func (xraySrv *XrayService) link(p *Params) string {
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
	return fmt.Sprintf("vless://%s@%s:%d?%s#%s",
		p.UUID, p.Host, p.Port, q.Encode(), frag,
	)
}

func (xraySrv *XrayService) GetXrayLinkById(clientId int64, userId int64) (*ConnectionLinkOut, error) {
	connAccess := ConnectionAccessOut{}
	err := xraySrv.DB.QueryRow(GetXrayLinkByIdQuery, clientId, userId).Scan(&connAccess.AccessKey, &connAccess.Alias)
	if err != nil {
		return nil, ErrorClientNotFound
	}
	params, err := xraySrv.makeXrayLinkParams(&connAccess)
	if err != nil {
		return nil, err
	}
	link := xraySrv.link(params)
	return &ConnectionLinkOut{ConnectionLink: link}, nil
}
