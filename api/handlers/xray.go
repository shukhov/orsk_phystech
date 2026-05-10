package handlers

import (
	"api/database"
	"api/middleware/security"
	"api/middleware/utils"
	"api/middleware/xray"
	"errors"
	"net/http"
	"strconv"
)

func GetConfig(writer http.ResponseWriter, _ *http.Request) {
	config, err := xray.XraySrv.GetConfig()
	if err != nil {
		switch {
		case errors.Is(err, xray.ErrorClientNotFound):
			utils.WriteJSON(writer, http.StatusNotFound, utils.ErrorCallback{ErrorText: err.Error()})
		case errors.Is(err, database.InternalDBError):
			utils.WriteJSON(writer, http.StatusInternalServerError, utils.ErrorCallback{ErrorText: err.Error()})
		default:
			utils.WriteJSON(writer, http.StatusBadRequest, utils.ErrorCallback{ErrorText: err.Error()})
		}
		return
	}
	utils.WriteJSON(writer, http.StatusOK, *config)
}

func GetClientsByUserId(writer http.ResponseWriter, request *http.Request) {
	userIdString := request.PathValue("user_id")
	userId, err := strconv.ParseInt(userIdString, 10, 64)
	if err != nil {
		utils.WriteJSON(writer, http.StatusBadRequest, utils.ErrorCallback{ErrorText: err.Error()})
		return
	}
	clients, err := xray.XraySrv.GetClientsByUserId(userId)
	if err != nil {
		switch {
		case errors.Is(err, security.UserNotFound):
			utils.WriteJSON(writer, http.StatusNotFound, utils.ErrorCallback{ErrorText: err.Error()})
		default:
			utils.WriteJSON(writer, http.StatusInternalServerError, utils.ErrorCallback{ErrorText: err.Error()})
		}
		return
	}
	utils.WriteJSON(writer, http.StatusOK, *clients)
	return
}

func GetClientById(writer http.ResponseWriter, request *http.Request) {
	clientIdString := request.PathValue("client_id")
	clientId, err := strconv.ParseInt(clientIdString, 10, 64)
	if err != nil {
		utils.WriteJSON(writer, http.StatusBadRequest, utils.ErrorCallback{ErrorText: err.Error()})
		return
	}
	client, err := xray.XraySrv.GetClientById(clientId)
	if err != nil {
		switch {
		case errors.Is(err, xray.ErrorClientNotFound):
			utils.WriteJSON(writer, http.StatusNotFound, utils.ErrorCallback{ErrorText: err.Error()})
		default:
			utils.WriteJSON(writer, http.StatusInternalServerError, utils.ErrorCallback{ErrorText: err.Error()})
		}
		return
	}
	utils.WriteJSON(writer, http.StatusOK, *client)
	return
}

func GetXrayLinkById(writer http.ResponseWriter, request *http.Request) {
	ctx := request.Context()
	userId := security.GetUserIdFromContext(&ctx)
	clientId, err := strconv.ParseInt(request.PathValue("client_id"), 10, 64)
	if err != nil {
		utils.WriteJSON(writer, http.StatusBadRequest, utils.ErrorCallback{ErrorText: err.Error()})
		return
	}
	link, err := xray.XraySrv.GetXrayLinkById(clientId, userId)
	if err != nil {
		switch {
		case errors.Is(err, database.InternalDBError):
			utils.WriteJSON(writer, http.StatusInternalServerError, utils.ErrorCallback{ErrorText: err.Error()})
		case errors.Is(err, xray.ErrorClientNotFound):
			utils.WriteJSON(writer, http.StatusNotFound, utils.ErrorCallback{ErrorText: err.Error()})
		default:
			utils.WriteJSON(writer, http.StatusBadRequest, utils.ErrorCallback{ErrorText: err.Error()})
		}
		return
	}
	utils.WriteJSON(writer, http.StatusOK, *link)
	return

}
