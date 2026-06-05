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

type ConnectionParams struct {
	XrayPort        int32
	XrayHost        string
	XrayListen      string
	XrayDest        string
	XrayPrivateKey  string
	XrayServerNames []string
	XrayShortIds    []string
}

type XrayService struct {
	DB         *sql.DB
	ConnParams *ConnectionParams
	logger     *log.Logger
}

func NewXrayService() *XrayService {
	vlsSvc := new(XrayService)
	vlsSvc.logger = log.New(os.Stdout, "XrayService: ", log.LstdFlags|log.Lshortfile)
	db, err := database.GetDB()
	if err != nil {
		panic(err)
	}
	vlsSvc.DB = db
	port, err := strconv.ParseInt(os.Getenv("XRAY_PORT"), 10, 32)
	if err != nil {
		panic(err)
	}
	vlsSvc.ConnParams = &ConnectionParams{
		XrayPort:        int32(port),
		XrayHost:        os.Getenv("XRAY_HOST"),
		XrayListen:      os.Getenv("XRAY_LISTEN"),
		XrayDest:        os.Getenv("XRAY_DEST"),
		XrayPrivateKey:  os.Getenv("XRAY_PRIVATE_KEY"),
		XrayServerNames: strings.Split(os.Getenv("XRAY_SERVER_NAMES"), ","),
		XrayShortIds:    strings.Split(os.Getenv("XRAY_SHORT_IDS"), ","),
	}

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
	Port           int32           `json:"port"`
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
	config := Config{
		Log: LogConfig{Loglevel: "warning"},
		Inbounds: []Inbound{
			{
				Listen:   xraySrv.ConnParams.XrayListen,
				Port:     xraySrv.ConnParams.XrayPort,
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
						Dest:        xraySrv.ConnParams.XrayDest,
						Xver:        0,
						ServerNames: xraySrv.ConnParams.XrayServerNames,
						PrivateKey:  xraySrv.ConnParams.XrayPrivateKey,
						ShortIds:    xraySrv.ConnParams.XrayShortIds,
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
