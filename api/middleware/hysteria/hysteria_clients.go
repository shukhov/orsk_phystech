package hysteria

import (
	"api/database"
	"context"
	"database/sql"
	"errors"
	"time"
)

var (
	ErrorClientNotFound  = errors.New("hysteria client is not found")
	ErrorClientCreateBad = errors.New("incorrect data for create new hysteria client")
)

type Userpass = map[string]string

type ClientPublicOut struct {
	Id        int64     `json:"id"`
	Alias     string    `json:"alias"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type NewClientIn struct {
	InviteId int64  `json:"invite_id"`
	UserId   int64  `json:"user_id"`
	Alias    string `json:"alias"`
}

func (hystSrv *HysteriaService) NewClient(newClientIn *NewClientIn, externalTx *sql.Tx) (*ClientPublicOut, error) {
	var isExternalTx = true
	if externalTx == nil {
		var err error
		externalTx, err = hystSrv.DB.BeginTx(context.Background(), nil)
		if err != nil {
			func() { _ = externalTx.Rollback() }()
			return nil, database.InternalDBError
		}
		isExternalTx = false
	}
	clientOut := ClientPublicOut{}
	err := externalTx.QueryRow(
		NewClientQuery, &newClientIn.InviteId, &newClientIn.UserId, &newClientIn.Alias).Scan(
		&clientOut.Id, &clientOut.Alias, &clientOut.Status,
		&clientOut.CreatedAt, &clientOut.UpdatedAt)
	if err != nil {
		hystSrv.logger.Printf("%#v", err)
		return nil, ErrorClientCreateBad
	}
	if !isExternalTx {
		err = externalTx.Commit()
		if err != nil {
			return nil, database.InternalDBError
		}
	}
	return &clientOut, nil
}

func (hystSrv *HysteriaService) GetAllInConfigClients() (*Userpass, error) {
	inConfigClient := make(Userpass)
	rows, err := hystSrv.DB.Query(GetAllInConfigClientsQuery)
	if err != nil {
		return nil, database.InternalDBError
	}
	for rows.Next() {
		var user, password string
		err = rows.Scan(&user, &password)
		if err != nil {
			return nil, database.InternalDBError
		}
		inConfigClient[user] = password
	}
	return &inConfigClient, nil
}
