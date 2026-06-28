package hysteria

import (
	"api/database"
	"database/sql"
	"fmt"
	"log"
	"os"
	"strconv"
)

type ConnectionParams struct {
	HysteriaHost               string
	HysteriaListen             string
	HysteriaObfsPassword       string
	HysteriaMasqueradeProxyUrl string
	HysteriaSni                string
	HysteriaPassword           string
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
		HysteriaHost:               os.Getenv("HYSTERIA_HOST"),
		HysteriaListen:             os.Getenv("HYSTERIA_LISTEN"),
		HysteriaObfsPassword:       os.Getenv("HYSTERIA_OBFS_PASSWORD"),
		HysteriaMasqueradeProxyUrl: os.Getenv("HYSTERIA_MASQUERADE_PROXY_URL"),
		HysteriaSni:                os.Getenv("HYSTERIA_SNI"),
		HysteriaPassword:           os.Getenv("HYSTERIA_PASSWORD"),
		HysteriaBandwidthUp:        bandwidthUp,
		HysteriaBandwidthDown:      bandwidthDown,
	}
	return HystSrv
}

type Config struct {
	Listen     string            `json:"listen"`
	TLS        TLSConfig         `json:"tls"`
	Auth       AuthConfig        `json:"auth"`
	Bandwidth  Bandwidth         `json:"bandwidth"`
	Obfs       *ObfsConfig       `json:"obfs,omitempty"`
	Masquerade *MasqueradeConfig `json:"masquerade,omitempty"`
}

type TLSConfig struct {
	Cert     string `json:"cert"`
	Key      string `json:"key"`
	SNIGuard string `json:"sniGuard,omitempty"`
}

type AuthConfig struct {
	Type     string   `json:"type"`
	Password []string `json:"password,omitempty"`
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
		TLS: TLSConfig{
			Cert:     "/etc/hysteria/server.crt",
			Key:      "/etc/hysteria/server.key",
			SNIGuard: "disable",
		},
		Auth: AuthConfig{
			Type:     "password",
			Password: *inConfigClinent,
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
    type: password
    password: some-password
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
