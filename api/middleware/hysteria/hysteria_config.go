package hysteria

import (
	"api/database"
	"database/sql"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
)

type ConnectionParams struct {
	AcmeDomains                []string
	HysteriaMasqueradeProxyUrl string
	HysteriaListen             string
	HysteriaAcmeEmail          string
	HysteriaObfsPassword       string
	HysteriaBandwidthUp        int64
	HysteriaBandwidthDown      int64
}

type HysteriaService struct {
	DB         *sql.DB
	ConnParams *ConnectionParams
	logger     *log.Logger
}

var HystSrv = NewHysteriaService()

func NewHysteriaService() *HysteriaService {
	HystSrv := new(HysteriaService)
	HystSrv.logger = log.New(os.Stdout, "HysteriaService: ", log.LstdFlags|log.Lshortfile)
	db, err := database.GetDB()
	if err != nil {
		panic(err)
	}
	HystSrv.DB = db
	bandwidthUp, err := strconv.ParseInt(os.Getenv("HYSTERIA_BANDWIDTH_UP"), 10, 64)
	if err != nil {
		panic(err)
	}
	bandwidthDown, err := strconv.ParseInt(os.Getenv("HYSTERIA_BANDWIDTH_DOWN"), 10, 64)
	if err != nil {
		panic(err)
	}
	HystSrv.ConnParams = &ConnectionParams{
		HysteriaListen:             os.Getenv("HYSTERIA_LISTEN"),
		HysteriaAcmeEmail:          os.Getenv("HYSTERIA_ACME_EMAIL"),
		HysteriaObfsPassword:       os.Getenv("HYSTERIA_OBFS_PASSWORD"),
		HysteriaMasqueradeProxyUrl: os.Getenv("HYSTERIA_MASQUERADE_PROXY_URL"),
		AcmeDomains:                strings.Split(os.Getenv("ACME_DOMAINS"), ","),
		HysteriaBandwidthUp:        bandwidthUp,
		HysteriaBandwidthDown:      bandwidthDown,
	}
	return HystSrv
}

type Config struct {
	Listen string `json:"listen"`

	TLS *TLSConfig `json:"tls,omitempty"`

	Auth AuthConfig `json:"auth"`

	Bandwidth Bandwidth `json:"bandwidth"`

	Obfs *ObfsConfig `json:"obfs,omitempty"`

	Masquerade *MasqueradeConfig `json:"masquerade,omitempty"`
}

type TLSConfig struct {
	Type string     `json:"type"`
	ACME ACMEConfig `json:"acme,omitempty"`
}

type ACMEConfig struct {
	Domains []string `json:"domains"`
	Email   string   `json:"email"`
}

type AuthConfig struct {
	Type     string   `json:"type"`
	Userpass Userpass `json:"userpass,omitempty"`
}

type ObfsConfig struct {
	Type       string         `json:"type"`
	Salamander SalamanderObfs `json:"salamander,omitempty"`
}

type SalamanderObfs struct {
	Password string `json:"password"`
}

type MasqueradeConfig struct {
	Type  string          `json:"type"`
	Proxy MasqueradeProxy `json:"proxy,omitempty"`
}

type MasqueradeProxy struct {
	URL         string `json:"url"`
	RewriteHost bool   `json:"rewriteHost"`
}

type Bandwidth struct {
	Up   string `json:"up"`
	Down string `json:"down"`
}

func (hystSrv *HysteriaService) GetConfig() (*Config, error) {
	inConfigClinent, err := hystSrv.GetAllInConfigClients()
	if err != nil {
		return nil, err
	}
	return &Config{
		Listen: hystSrv.ConnParams.HysteriaListen,
		TLS: &TLSConfig{
			Type: "tls",
			ACME: ACMEConfig{
				Domains: hystSrv.ConnParams.AcmeDomains,
				Email:   hystSrv.ConnParams.HysteriaAcmeEmail,
			},
		},
		Auth: AuthConfig{
			Type:     "userpass",
			Userpass: *inConfigClinent,
		},
		Bandwidth: Bandwidth{
			Up:   fmt.Sprintf("%d mbps", hystSrv.ConnParams.HysteriaBandwidthUp),
			Down: fmt.Sprintf("%d mbps", hystSrv.ConnParams.HysteriaBandwidthDown),
		},
		Obfs: &ObfsConfig{
			Type: "salamander",
			Salamander: SalamanderObfs{
				Password: hystSrv.ConnParams.HysteriaObfsPassword,
			},
		},
		Masquerade: &MasqueradeConfig{
			Type: "proxy",
			Proxy: MasqueradeProxy{
				URL:         hystSrv.ConnParams.HysteriaMasqueradeProxyUrl,
				RewriteHost: false,
			},
		},
	}, nil

}

/*

listen: :443
tls:
    type: acme
    acme:
        domains:
            - vpn.example.com
        email: admin@example.com
auth:
    type: userpass
    userpass:
        42-100: 550e8400-e29b-41d4-a716-446655440000
        42-101: 6c2a2e34-f53b-4d17-b34c-f7dd8c1b0a5e
obfs:
    type: salamander
    salamander:
        password: some-obfs-password
masquerade:
    type: proxy
    proxy:
        url: https://www.bing.com
        rewriteHost: true
*/
