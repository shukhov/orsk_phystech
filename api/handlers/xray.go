package handlers

import (
	"api/database"
	"api/middleware/utils"
	"api/middleware/xray"
	"errors"
	"net/http"
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
