package xray

import (
	"api/database"
	"context"
	"database/sql"
	"errors"
	"time"
)

var (
	ErrorClientNotFound  = errors.New("xray client is not found")
	ErrorClientCreateBad = errors.New("incorrect data for create new vpn client")
)

type ClientPublicOut struct {
	Id        int64     `json:"id"`
	Alias     string    `json:"alias"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type ClientPrivateOut struct {
	ClientPublicOut
	AccessId string `json:"access_id"`
	UserId   int64  `json:"user_id"`
	InviteId int64  `json:"invite_id"`
}

type NewClientIn struct {
	InviteId int64  `json:"invite_id"`
	UserId   int64  `json:"user_id"`
	Alias    string `json:"alias"`
}

func (xraySrv *XrayService) GetClientById(clientId int64) (*ClientPrivateOut, error) {
	clientOut := new(ClientPrivateOut)
	err := xraySrv.DB.QueryRow(GetClientByIdQuery, clientId).Scan(
		&clientOut.Id, &clientOut.AccessId, &clientOut.UserId,
		&clientOut.InviteId, &clientOut.InviteId, &clientOut.Alias,
		clientOut.Status, &clientOut.CreatedAt, &clientOut.UpdatedAt)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrorClientNotFound
		default:
			xraySrv.logger.Printf("error: %#v", err)
			return nil, database.InternalDBError
		}
	}
	return clientOut, nil
}

func (xraySrv *XrayService) GetClientsByUserId(userId int64) (*[]ClientPublicOut, error) {
	clientList := make([]ClientPublicOut, 0, 3)

	queryResult, err := xraySrv.DB.Query(GetClientsByUserIdQuery, userId)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrorClientNotFound
		default:
			xraySrv.logger.Printf("error %#v", err)
			return nil, database.InternalDBError
		}
	}
	for queryResult.Next() {
		client := ClientPublicOut{}
		err = queryResult.Scan(
			&client.Id, &client.Alias, client.Status,
			&client.CreatedAt, &client.UpdatedAt)
		if err != nil {
			xraySrv.logger.Printf("error %#v", err)
			return nil, database.InternalDBError
		}
		clientList = append(clientList, client)
	}
	return &clientList, nil
}

func (xraySrv *XrayService) NewClient(newClientIn *NewClientIn, externalTx *sql.Tx) (*ClientPublicOut, error) {
	if externalTx == nil {
		var err error
		externalTx, err = xraySrv.DB.BeginTx(context.Background(), nil)
		defer func() { _ = externalTx.Commit() }()
		if err != nil {
			func() { _ = externalTx.Rollback() }()
			return nil, database.InternalDBError
		}
	}
	clientOut := ClientPublicOut{}
	err := externalTx.QueryRow(
		NewClientQuery, &newClientIn.InviteId, &newClientIn.UserId, &newClientIn.Alias).Scan(
		&clientOut.Id, &clientOut.Alias, &clientOut.Status,
		&clientOut.CreatedAt, &clientOut.UpdatedAt)
	if err != nil {
		xraySrv.logger.Printf("%#v", err)
		return nil, ErrorClientCreateBad
	}
	return &clientOut, nil
}
