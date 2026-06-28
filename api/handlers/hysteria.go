package handlers

import (
	"api/database"
	"api/middleware/hysteria"
	"api/middleware/security"
	"api/middleware/utils"
	"errors"
	"net/http"
	"strconv"
)

func GetHysteriaConfig(writer http.ResponseWriter, _ *http.Request) {
	config, err := hysteria.HystSrv.GetConfig()
	if err != nil {
		switch {
		case errors.Is(err, hysteria.ErrorClientNotFound):
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

func GetHysteriaClientsByUserId(writer http.ResponseWriter, request *http.Request) {
	userIdString := request.PathValue("user_id")
	userId, err := strconv.ParseInt(userIdString, 10, 64)
	if err != nil {
		utils.WriteJSON(writer, http.StatusBadRequest, utils.ErrorCallback{ErrorText: err.Error()})
		return
	}
	clients, err := hysteria.HystSrv.GetClientsByUserId(userId)
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
}

func GetHysteriaClientById(writer http.ResponseWriter, request *http.Request) {
	clientIdString := request.PathValue("client_id")
	clientId, err := strconv.ParseInt(clientIdString, 10, 64)
	if err != nil {
		utils.WriteJSON(writer, http.StatusBadRequest, utils.ErrorCallback{ErrorText: err.Error()})
		return
	}
	client, err := hysteria.HystSrv.GetClientById(clientId)
	if err != nil {
		switch {
		case errors.Is(err, hysteria.ErrorClientNotFound):
			utils.WriteJSON(writer, http.StatusNotFound, utils.ErrorCallback{ErrorText: err.Error()})
		default:
			utils.WriteJSON(writer, http.StatusInternalServerError, utils.ErrorCallback{ErrorText: err.Error()})
		}
		return
	}
	utils.WriteJSON(writer, http.StatusOK, *client)
}

func GetHysteriaLinkById(writer http.ResponseWriter, request *http.Request) {
	ctx := request.Context()
	userId := security.GetUserIdFromContext(&ctx)
	clientId, err := strconv.ParseInt(request.PathValue("client_id"), 10, 64)
	if err != nil {
		utils.WriteJSON(writer, http.StatusBadRequest, utils.ErrorCallback{ErrorText: err.Error()})
		return
	}
	link, err := hysteria.HystSrv.GetXrayLinkById(clientId, userId)
	if err != nil {
		switch {
		case errors.Is(err, database.InternalDBError):
			utils.WriteJSON(writer, http.StatusInternalServerError, utils.ErrorCallback{ErrorText: err.Error()})
		case errors.Is(err, hysteria.ErrorClientNotFound):
			utils.WriteJSON(writer, http.StatusNotFound, utils.ErrorCallback{ErrorText: err.Error()})
		default:
			utils.WriteJSON(writer, http.StatusBadRequest, utils.ErrorCallback{ErrorText: err.Error()})
		}
		return
	}
	utils.WriteJSON(writer, http.StatusOK, *link)
}

func GetHysteriaLastUpdate(writer http.ResponseWriter, request *http.Request) {
	glu, err := hysteria.HystSrv.GetLastUpdate()
	if err != nil {
		switch {
		case errors.Is(err, hysteria.ErrorClientNotFound):
			utils.WriteJSON(writer, http.StatusNotFound, utils.ErrorCallback{ErrorText: err.Error()})
		default:
			utils.WriteJSON(writer, http.StatusInternalServerError, utils.ErrorCallback{ErrorText: err.Error()})
		}
		return
	}
	utils.WriteJSON(writer, http.StatusOK, *glu)
}

func UpdateHysteriaClientAlias(writer http.ResponseWriter, request *http.Request) {
	clientIdString := request.PathValue("client_id")
	clientId, err := strconv.ParseInt(clientIdString, 10, 64)
	if err != nil {
		utils.WriteJSON(writer, http.StatusBadRequest, utils.ErrorCallback{ErrorText: err.Error()})
		return
	}
	ctx := request.Context()
	userId := security.GetUserIdFromContext(&ctx)
	updateIn := hysteria.UpdateClientAliasIn{}
	err = utils.ReadJSON(request, &updateIn)
	if err != nil {
		utils.WriteJSON(writer, http.StatusBadRequest, utils.ErrorCallback{ErrorText: err.Error()})
		return
	}
	client, err := hysteria.HystSrv.UpdateClientAlias(clientId, userId, &updateIn)
	if err != nil {
		switch {
		case errors.Is(err, hysteria.ErrorClientNotFound):
			utils.WriteJSON(writer, http.StatusNotFound, utils.ErrorCallback{ErrorText: err.Error()})
		case errors.Is(err, hysteria.ErrorClientForbidden):
			utils.WriteJSON(writer, http.StatusForbidden, utils.ErrorCallback{ErrorText: err.Error()})
		case errors.Is(err, hysteria.ErrorClientUpdateBad):
			utils.WriteJSON(writer, http.StatusBadRequest, utils.ErrorCallback{ErrorText: err.Error()})
		default:
			utils.WriteJSON(writer, http.StatusInternalServerError, utils.ErrorCallback{ErrorText: err.Error()})
		}
		return
	}
	utils.WriteJSON(writer, http.StatusOK, *client)
}

func DeleteHysteriaClientById(writer http.ResponseWriter, request *http.Request) {
	ctx := request.Context()
	userId := security.GetUserIdFromContext(&ctx)
	clientIdString := request.PathValue("client_id")
	clientId, err := strconv.ParseInt(clientIdString, 10, 64)
	if err != nil {
		utils.WriteJSON(writer, http.StatusBadRequest, utils.ErrorCallback{ErrorText: err.Error()})
		return
	}
	client, err := hysteria.HystSrv.DeleteClientById(clientId, userId)
	if err != nil {
		switch {
		case errors.Is(err, hysteria.ErrorClientNotFound):
			utils.WriteJSON(writer, http.StatusNotFound, utils.ErrorCallback{ErrorText: err.Error()})
		case errors.Is(err, hysteria.ErrorClientForbidden):
			utils.WriteJSON(writer, http.StatusForbidden, utils.ErrorCallback{ErrorText: err.Error()})
		case errors.Is(err, hysteria.ErrorClientDeleteBad):
			utils.WriteJSON(writer, http.StatusBadRequest, utils.ErrorCallback{ErrorText: err.Error()})
		default:
			utils.WriteJSON(writer, http.StatusInternalServerError, utils.ErrorCallback{ErrorText: err.Error()})
		}
		return
	}
	utils.WriteJSON(writer, http.StatusOK, *client)
}
