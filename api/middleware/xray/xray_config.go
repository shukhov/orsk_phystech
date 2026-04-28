package xray

import (
	"api/database"
	"database/sql"
	"errors"
	"log"
	"os"
	"strconv"
	"strings"
)

type XrayService struct {
	DB     *sql.DB
	logger *log.Logger
}

func NewXrayService() *XrayService {
	vlsSvc := new(XrayService)
	vlsSvc.logger = log.New(os.Stdout, "XrayService: ", log.LstdFlags|log.Lshortfile)
	db, err := database.GetDB()
	if err != nil {
		panic(err)
	}
	vlsSvc.DB = db
	return vlsSvc
}

type Config struct {
	Log       LogConfig  `json:"log"`
	Inbounds  []Inbound  `json:"inbounds"`
	Outbounds []Outbound `json:"outbounds"`
}

type LogConfig struct {
	Loglevel string `json:"loglevel"`
}

type Inbound struct {
	Listen         string          `json:"listen"`
	Port           int64           `json:"port"`
	Protocol       string          `json:"protocol"`
	Settings       InboundSettings `json:"settings"`
	StreamSettings StreamSettings  `json:"streamSettings"`
	Sniffing       Sniffing        `json:"sniffing"`
}

type InboundSettings struct {
	Clients    []InConfigClient `json:"clients"`
	Decryption string           `json:"decryption"`
}

type InConfigClient struct {
	Id   string `json:"id"`
	Flow string `json:"flow"`
}

type StreamSettings struct {
	Network         string          `json:"network"`
	Security        string          `json:"security"`
	RealitySettings RealitySettings `json:"realitySettings"`
}

type RealitySettings struct {
	Show        bool     `json:"show"`
	Dest        string   `json:"dest"`
	Xver        int64    `json:"xver"`
	ServerNames []string `json:"serverNames"`
	PrivateKey  string   `json:"privateKey"`
	ShortIds    []string `json:"shortIds"`
}

type Sniffing struct {
	Enabled      bool     `json:"enabled"`
	DestOverride []string `json:"destOverride"`
}

type Outbound struct {
	Protocol string `json:"protocol"`
}

var XraySrv = NewXrayService()

func (xraySrv *XrayService) GetAllInConfigClients() (*[]InConfigClient, error) {
	clients := make([]InConfigClient, 0, 10)

	idList, err := xraySrv.DB.Query(GetAllInConfigClientsQuery)

	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrorClientNotFound
		default:
			return nil, database.InternalDBError
		}
	}
	for idList.Next() {
		client := InConfigClient{}
		err = idList.Scan(&client.Id, &client.Flow)
		if err != nil {
			return nil, database.InternalDBError
		}
		clients = append(clients, client)
	}
	return &clients, nil
}

func (xraySrv *XrayService) GetConfig() (*Config, error) {
	inConfigClients, err := xraySrv.GetAllInConfigClients()
	if err != nil {
		return nil, err
	}
	xrayPort, err := strconv.ParseInt(os.Getenv("XRAY_PORT"), 10, 64)
	config := Config{
		Log: LogConfig{Loglevel: "warning"},
		Inbounds: []Inbound{
			{
				Listen:   os.Getenv("XRAY_LISTEN"),
				Port:     xrayPort,
				Protocol: "vless",
				Settings: InboundSettings{
					Clients:    *inConfigClients,
					Decryption: "none",
				},
				StreamSettings: StreamSettings{
					Network:  "tcp",
					Security: "reality",
					RealitySettings: RealitySettings{
						Show:        false,
						Dest:        os.Getenv("XRAY_DEST"),
						Xver:        0,
						ServerNames: strings.Split(os.Getenv("XRAY_SERVER_NAMES"), ","),
						PrivateKey:  os.Getenv("XRAY_PRIVATE_KEY"),
						ShortIds:    strings.Split(os.Getenv("XRAY_SHORT_IDS"), ","),
					},
				},
				Sniffing: Sniffing{
					Enabled:      false,
					DestOverride: []string{"http", "tls", "quic"},
				},
			},
		},
		Outbounds: []Outbound{
			{Protocol: "freedom"},
		},
	}
	return &config, nil
}
