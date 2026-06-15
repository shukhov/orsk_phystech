package hysteria

import "math/rand"

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
	Insecure     int8
}

func (hystSrv *HysteriaService) makeXrayLinkParams(userpass *Userpass) *Params {
	// Guarantee: in `userpass` will come only 1 record
	var username, password string
	for k, v := range *userpass {
		username = k
		password = v
	}
	return &Params{
		Username:     username,
		Password:     password,
		Domain:       hystSrv.ConnParams.AcmeDomains[rand.Intn(len(hystSrv.ConnParams.AcmeDomains))],
		ObfsPassword: hystSrv.ConnParams.HysteriaObfsPassword,
		Obfs:         "salamander",
		Sni:          hystSrv.ConnParams.HysteriaMasqueradeProxyUrl,
		ProfileName:  "",
		Insecure:     0,
	}

}
