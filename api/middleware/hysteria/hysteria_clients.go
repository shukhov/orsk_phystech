package hysteria

import (
	"api/database"
	"api/middleware/security"
	"context"
	"database/sql"
	"errors"
	"time"
)

var (
	ErrorClientNotFound  = errors.New("hysteria client is not found")
	ErrorClientCreateBad = errors.New("incorrect data for create new hysteria client")
	ErrorClientUpdateBad = errors.New("incorrect data for update hysteria client alias")
	ErrorClientDeleteBad = errors.New("incorrect data for delete hysteria client")
	ErrorClientForbidden = errors.New("hysteria client does not belong to this user")
)

type Passwords = []string

type ClientPublicOut struct {
	Id        int64     `json:"id"`
	Alias     string    `json:"alias"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type ClientPrivateOut struct {
	ClientPublicOut
	Password string `json:"access_key"`
	UserId   int64  `json:"user_id"`
	InviteId int64  `json:"invite_id"`
}

type NewClientIn struct {
	InviteId int64  `json:"invite_id"`
	UserId   int64  `json:"user_id"`
	Alias    string `json:"alias"`
}

type UpdateClientAliasIn struct {
	NewAlias string `json:"new_alias"`
}

type LastUpdateOut struct {
	UpdatedAt time.Time `json:"updated_at"`
}

func (hystSrv *HysteriaService) GetClientById(clientId int64) (*ClientPrivateOut, error) {
	clientOut := new(ClientPrivateOut)
	err := hystSrv.DB.QueryRow(GetClientByIdQuery, clientId).Scan(
		&clientOut.Id, &clientOut.Password, &clientOut.UserId,
		&clientOut.InviteId, &clientOut.Alias, &clientOut.Status,
		&clientOut.CreatedAt, &clientOut.UpdatedAt)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrorClientNotFound
		default:
			hystSrv.logger.Printf("error: %#v", err)
			return nil, database.InternalDBError
		}
	}
	return clientOut, nil
}

func (hystSrv *HysteriaService) GetClientsByUserId(userId int64) (*[]ClientPublicOut, error) {
	_, err := security.SecSrv.GetUserById(userId)
	if err != nil {
		return nil, security.UserNotFound
	}

	clientList := make([]ClientPublicOut, 0, 3)

	queryResult, err := hystSrv.DB.Query(GetClientsByUserIdQuery, userId)
	if err != nil {
		hystSrv.logger.Printf("error %#v", err)
		return nil, database.InternalDBError
	}
	for queryResult.Next() {
		client := ClientPublicOut{}
		err = queryResult.Scan(
			&client.Id, &client.Alias, &client.Status,
			&client.CreatedAt, &client.UpdatedAt)
		if err != nil {
			hystSrv.logger.Printf("error %#v", err)
			return nil, database.InternalDBError
		}
		clientList = append(clientList, client)
	}
	return &clientList, nil
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

func (hystSrv *HysteriaService) GetAllInConfigClients() (*Passwords, error) {
	inConfigClient := make(Passwords, 0, 3)
	rows, err := hystSrv.DB.Query(GetAllInConfigClientsQuery)
	if err != nil {
		return nil, database.InternalDBError
	}
	for rows.Next() {
		var password string
		err = rows.Scan(&password)
		if err != nil {
			return nil, database.InternalDBError
		}
		inConfigClient = append(inConfigClient, password)
	}
	return &inConfigClient, nil
}

func (hystSrv *HysteriaService) UpdateClientAlias(clientId int64, userId int64, updateIn *UpdateClientAliasIn) (*ClientPublicOut, error) {
	client, err := hystSrv.GetClientById(clientId)
	if err != nil {
		return nil, err
	}
	if client.UserId != userId {
		return nil, ErrorClientForbidden
	}
	clientOut := new(ClientPublicOut)
	err = hystSrv.DB.QueryRow(UpdateClientAliasQuery, updateIn.NewAlias, clientId).Scan(
		&clientOut.Id, &clientOut.Alias, &clientOut.Status,
		&clientOut.CreatedAt, &clientOut.UpdatedAt)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrorClientNotFound
		default:
			hystSrv.logger.Printf("error updating client alias: %#v", err)
			return nil, ErrorClientUpdateBad
		}
	}
	return clientOut, nil
}

func (hystSrv *HysteriaService) GetLastUpdate() (*LastUpdateOut, error) {
	lastUpdateOut := LastUpdateOut{}
	err := hystSrv.DB.QueryRow(LastUpdateQuery).Scan(&lastUpdateOut.UpdatedAt)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrorClientNotFound
		default:
			hystSrv.logger.Printf("last update error: %#v", err)
			return nil, database.InternalDBError
		}
	}
	return &lastUpdateOut, nil
}

func (hystSrv *HysteriaService) DeleteClientById(clientId int64, userId int64) (*ClientPublicOut, error) {
	client, err := hystSrv.GetClientById(clientId)
	if err != nil {
		return nil, err
	}
	if client.UserId != userId {
		return nil, ErrorClientForbidden
	}
	clientOut := new(ClientPublicOut)
	err = hystSrv.DB.QueryRow(DeleteClientByIdQuery, clientId).Scan(
		&clientOut.Id, &clientOut.Alias, &clientOut.Status,
		&clientOut.CreatedAt, &clientOut.UpdatedAt)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrorClientNotFound
		default:
			hystSrv.logger.Printf("error deleting client: %#v", err)
			return nil, ErrorClientDeleteBad
		}
	}
	return clientOut, nil
}
