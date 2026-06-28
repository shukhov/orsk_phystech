package hysteria

import (
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"
)

type ConnectionLinkOut struct {
	ConnectionLink string `json:"connection_link"`
}

type ConnectionAccessOut struct {
	AccessKey string
	Alias     string
}

type Params struct {
	Username     string
	Password     string
	Domain       string
	ObfsPassword string
	Obfs         string
	Sni          string
	ProfileName  string
	PinSHA256    string
	Port         int16
	Insecure     int8
}

type InConfigClient struct {
	Userpass Userpass
	Alias    string
}

func (hystSrv *HysteriaService) getPin() string {
	pin, err := os.ReadFile("/etc/hysteria/pin.txt")
	if err != nil {
		return ""
	}

	return strings.TrimSpace(string(pin))
}

func (hystSrv *HysteriaService) makeXrayLinkParams(username, password, alias string) *Params {
	// Guarantee: in `inConfigClient` will come only 1 record
	port, err := strconv.ParseInt(strings.Split(hystSrv.ConnParams.HysteriaListen, ":")[1], 10, 16)
	if err != nil {
		panic(err)
	}
	return &Params{
		Username:     username,
		Password:     password,
		Domain:       hystSrv.ConnParams.HysteriaHost,
		ObfsPassword: hystSrv.ConnParams.HysteriaObfsPassword,
		Obfs:         "salamander",
		Sni:          hystSrv.ConnParams.HysteriaSni,
		ProfileName:  alias,
		PinSHA256:    hystSrv.getPin(),
		Port:         int16(port),
		Insecure:     0,
	}
}

func (hystSrv *HysteriaService) link(p *Params) string {
	q := url.Values{}
	q.Set("obfs", p.Obfs)
	q.Set("obfs-password", p.ObfsPassword)
	q.Set("sni", p.Sni)
	q.Set("insecure", strconv.Itoa(int(p.Insecure)))
	q.Set("pinSHA256", p.PinSHA256)

	frag := url.PathEscape(p.ProfileName)
	return fmt.Sprintf("hysteria2://%s:%s@%s:%d/?%s#%s",
		p.Username, p.Password, p.Domain, p.Port, q.Encode(), frag,
	)
}

func (hystSrv *HysteriaService) GetXrayLinkById(clientId int64, userId int64) (*ConnectionLinkOut, error) {
	var username, password, alias string
	err := hystSrv.DB.QueryRow(GetHysteriaLinkByIdQuery, clientId, userId).Scan(&username, &password, &alias)
	if err != nil {
		return nil, ErrorClientNotFound
	}
	params := hystSrv.makeXrayLinkParams(username, password, alias)

	link := hystSrv.link(params)
	return &ConnectionLinkOut{ConnectionLink: link}, nil

}
