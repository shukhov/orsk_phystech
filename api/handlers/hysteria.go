package handlers

import (
	"api/middleware/hysteria"
	"api/middleware/utils"
	"errors"
	"net/http"
	"strconv"
)

func UpdateHysteriaClientAlias(writer http.ResponseWriter, request *http.Request) {
	clientIdString := request.PathValue("client_id")
	clientId, err := strconv.ParseInt(clientIdString, 10, 64)
	if err != nil {
		utils.WriteJSON(writer, http.StatusBadRequest, utils.ErrorCallback{ErrorText: err.Error()})
		return
	}
	updateIn := hysteria.UpdateClientAliasIn{}
	err = utils.ReadJSON(request, &updateIn)
	if err != nil {
		utils.WriteJSON(writer, http.StatusBadRequest, utils.ErrorCallback{ErrorText: err.Error()})
		return
	}
	client, err := hysteria.HystSrv.UpdateClientAlias(clientId, &updateIn)
	if err != nil {
		switch {
		case errors.Is(err, hysteria.ErrorClientNotFound):
			utils.WriteJSON(writer, http.StatusNotFound, utils.ErrorCallback{ErrorText: err.Error()})
		case errors.Is(err, hysteria.ErrorClientUpdateBad):
			utils.WriteJSON(writer, http.StatusBadRequest, utils.ErrorCallback{ErrorText: err.Error()})
		default:
			utils.WriteJSON(writer, http.StatusInternalServerError, utils.ErrorCallback{ErrorText: err.Error()})
		}
		return
	}
	utils.WriteJSON(writer, http.StatusOK, *client)
}
