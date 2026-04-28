package security

import (
	"api/database"
	"api/middleware/utils"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
)

type Role struct {
	Id          int64  `json:"id"`
	RoleName    string `json:"role_name"`
	AccessLevel int64  `json:"access_level"`
}

var RoleNotFound = errors.New("role not found")

func (SecSrv *SecurityService) AllowForRole(roleId int64, handler http.Handler) http.Handler {
	return http.HandlerFunc(func(next http.ResponseWriter, request *http.Request) {
		ctx := request.Context()
		userId := GetUserIdFromContext(&ctx)
		var hasAccess bool
		err := SecSrv.DB.QueryRow(
			`SELECT user_role.access_level >= role_req.access_level AS has_access
					FROM public.users AS usr
					JOIN public.roles AS user_role ON usr.role_id = user_role.id
					JOIN public.roles AS role_req ON role_req.id = $1
					WHERE usr.id = $2;`, roleId, userId).Scan(&hasAccess)
		if err != nil {
			switch {
			case errors.Is(err, sql.ErrNoRows):
				utils.WriteJSON(next, http.StatusNotFound, utils.ErrorCallback{ErrorText: err.Error()})
			default:
				utils.WriteJSON(next, http.StatusInternalServerError, utils.ErrorCallback{ErrorText: err.Error()})
			}
			return
		}
		if !hasAccess {
			utils.WriteJSON(next, http.StatusForbidden, utils.ErrorCallback{ErrorText: fmt.Sprintf("endpoint is able only for role_id: %d and higher", roleId)})
			return
		}
		handler.ServeHTTP(next, request)
	})
}

func (SecSrv *SecurityService) SetRoleForUser(userId int64, roleId int64) (*UserPrivateOut, error) {
	userOut := UserPrivateOut{}
	err := SecSrv.DB.QueryRow(
		"UPDATE public.users SET role_id=$1, updated_at=now() WHERE id = $2 "+
			"RETURNING id, created_at, updated_at, username, email, status, role_id;", roleId, userId).Scan(
		&userOut.Id, &userOut.CreatedAt, &userOut.UpdatedAt, &userOut.Username,
		&userOut.Email, &userOut.Status, &userOut.RoleId)
	if err != nil {
		fmt.Println(err)
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, UserNotFound
		default:
			return nil, database.InternalDBError
		}
	}
	return &userOut, nil
}

func (SecSrv *SecurityService) GetRole(roleId int64) (*Role, error) {
	roleOut := Role{}
	err := SecSrv.DB.QueryRow(
		"SELECT id, role_name, access_level FROM public.roles WHERE id = $1", roleId,
	).Scan(roleOut.Id, roleOut.RoleName, roleOut.AccessLevel)
	if err != nil {
		fmt.Println(err)
		switch {
		case errors.Is(err, RoleNotFound):
			return nil, err
		case errors.Is(err, database.InternalDBError):
			return nil, err

		}
	}
	return &roleOut, nil
}
