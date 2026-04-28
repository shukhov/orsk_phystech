package handlers

import (
	"api/database"
	"api/middleware/security"
	"api/middleware/utils"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"strconv"
)

var SecSrv = security.SecSrv

func Register(writer http.ResponseWriter, request *http.Request) {
	regData := security.UserRegisterIn{}
	err := utils.ReadJSON(request, &regData)
	if err != nil {
		utils.WriteJSON(writer, http.StatusBadRequest, utils.ErrorCallback{ErrorText: err.Error()})
		return
	}
	result, err := security.SecSrv.Create(&regData)
	if err != nil {
		utils.WriteJSON(writer, http.StatusBadRequest, utils.ErrorCallback{ErrorText: err.Error()})
		return
	}
	utils.WriteJSON(writer, 201, result)
	return
}

func Login(writer http.ResponseWriter, request *http.Request) {
	loginData := security.UserLoginIn{}
	err := utils.ReadJSON(request, &loginData)
	fmt.Println(err, loginData)
	if err != nil {
		utils.WriteJSON(writer, http.StatusBadRequest, utils.ErrorCallback{ErrorText: err})
		return
	}
	token, err := security.SecSrv.Login(&loginData)
	fmt.Println(err)
	if err != nil {
		switch {
		case errors.Is(err, security.IncorrectLoginOrPassword):
			utils.WriteJSON(writer, http.StatusUnauthorized, utils.ErrorCallback{ErrorText: err.Error()})
		case errors.Is(err, database.InternalDBError):
			utils.WriteJSON(writer, http.StatusInternalServerError, utils.ErrorCallback{ErrorText: err.Error()})
		default:
			utils.WriteJSON(writer, http.StatusBadRequest, utils.ErrorCallback{ErrorText: err.Error()})
		}
		return
	}
	utils.WriteJSON(writer, http.StatusCreated, security.AuthToken{Token: token})
	return
}

func Me(writer http.ResponseWriter, request *http.Request) {
	ctx := request.Context()
	userId := security.GetUserIdFromContext(&ctx)
	userOut, err := security.SecSrv.Me(userId)
	if err != nil {
		switch {
		case errors.Is(err, security.UserNotFound):
			utils.WriteJSON(writer, http.StatusNotFound, utils.ErrorCallback{ErrorText: err.Error()})
		case errors.Is(err, database.InternalDBError):
			utils.WriteJSON(writer, http.StatusInternalServerError, utils.ErrorCallback{ErrorText: err.Error()})
		}
		return
	}
	utils.WriteJSON(writer, http.StatusOK, userOut)
	return
}

func GetUserById(writer http.ResponseWriter, request *http.Request) {
	useIdStr := request.PathValue("user_id")
	userId, err := strconv.ParseInt(useIdStr, 10, 64)
	if err != nil {
		utils.WriteJSON(writer, http.StatusBadRequest, utils.ErrorCallback{ErrorText: err.Error()})
		return
	}
	userOut, err := security.SecSrv.GetUserById(userId)
	if err != nil {
		switch {
		case errors.Is(err, security.UserNotFound):
			utils.WriteJSON(writer, http.StatusNotFound, utils.ErrorCallback{ErrorText: err.Error()})
		case errors.Is(err, database.InternalDBError):
			utils.WriteJSON(writer, http.StatusInternalServerError, utils.ErrorCallback{ErrorText: err.Error()})
		default:
			utils.WriteJSON(writer, http.StatusBadRequest, utils.ErrorCallback{ErrorText: err.Error()})
		}
		return
	}
	utils.WriteJSON(writer, http.StatusOK, userOut)
	return
}

func SetRoleForUser(writer http.ResponseWriter, request *http.Request) {
	strUserId := request.PathValue("user_id")
	userId, err := strconv.ParseInt(strUserId, 10, 64)
	if err != nil {
		utils.WriteJSON(writer, http.StatusBadRequest, utils.ErrorCallback{ErrorText: err.Error()})
		return
	}
	strRoleId := request.PathValue("role_id")
	roleId, err := strconv.ParseInt(strRoleId, 10, 64)
	if err != nil {
		utils.WriteJSON(writer, http.StatusBadRequest, utils.ErrorCallback{ErrorText: err.Error()})
		return
	}
	userOut, err := security.SecSrv.SetRoleForUser(userId, roleId)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			utils.WriteJSON(writer, http.StatusNotFound, utils.ErrorCallback{ErrorText: err.Error()})
		case errors.Is(err, database.InternalDBError):
			utils.WriteJSON(writer, http.StatusInternalServerError, utils.ErrorCallback{ErrorText: err.Error()})
		default:
			utils.WriteJSON(writer, http.StatusBadRequest, utils.ErrorCallback{ErrorText: err.Error()})
		}
		return
	}
	utils.WriteJSON(writer, http.StatusOK, userOut)
}

func GetRole(writer http.ResponseWriter, request *http.Request) {
	roleIdStr := request.PathValue("role_id")
	roleId, err := strconv.ParseInt(roleIdStr, 10, 64)
	if err != nil {
		utils.WriteJSON(writer, http.StatusBadRequest, utils.ErrorCallback{ErrorText: err.Error()})
		return
	}
	roleOut, err := security.SecSrv.GetRole(roleId)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			utils.WriteJSON(writer, http.StatusNotFound, utils.ErrorCallback{ErrorText: err.Error()})
		case errors.Is(err, database.InternalDBError):
			utils.WriteJSON(writer, http.StatusInternalServerError, utils.ErrorCallback{ErrorText: err.Error()})
		default:
			utils.WriteJSON(writer, http.StatusBadRequest, utils.ErrorCallback{ErrorText: err.Error()})
		}
		return
	}
	utils.WriteJSON(writer, http.StatusOK, roleOut)
}
