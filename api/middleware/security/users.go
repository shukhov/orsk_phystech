package security

import (
	"api/database"
	"database/sql"
	"errors"
	"log"
)

var UserNotFound = errors.New("user not found")

type UserId struct {
	UserId int64 `json:"user_id"`
}

func NewSecurityService() *SecurityService {
	db, err := database.GetDB()
	if err != nil {
		log.Fatal(err)
	}
	return &SecurityService{DB: db}
}

var SecSrv = NewSecurityService()

func (SecSrv *SecurityService) GetUserById(userId int64) (*UserPublicOut, error) {
	userOut := UserPublicOut{}
	err := SecSrv.DB.QueryRow("SELECT id, created_at, updated_at, status "+
		"FROM public.users WHERE id = $1", userId).Scan(
		&userOut.Id, &userOut.CreatedAt, &userOut.UpdatedAt, &userOut.Username)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, UserNotFound
		}
		return nil, database.InternalDBError
	}
	return &userOut, nil
}

func (SecSrv *SecurityService) Me(userId int64) (*UserPrivateOut, error) {
	userOut := UserPrivateOut{}
	err := SecSrv.DB.QueryRow(
		"SELECT id, created_at, updated_at, username, email, status, role_id "+
			"FROM public.users WHERE id = $1", userId).Scan(
		&userOut.Id, &userOut.CreatedAt, &userOut.UpdatedAt, &userOut.Username,
		&userOut.Email, &userOut.Status, &userOut.RoleId)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, UserNotFound
		}
		return nil, database.InternalDBError
	}
	return &userOut, nil

}
