package handlers

import (
	"api/database"
	"api/middleware/invites"
	"api/middleware/utils"
	"errors"
	"net/http"
)

func NewInvite(writer http.ResponseWriter, request *http.Request) {
	inviteIn := invites.InviteIn{}
	err := utils.ReadJSON(request, &inviteIn)
	if err != nil {
		utils.WriteJSON(writer, http.StatusBadRequest, utils.ErrorCallback{ErrorText: err.Error()})
		return
	}
	inviteOut, err := invites.InvSrv.NewInvite(&inviteIn)
	if err != nil {
		utils.WriteJSON(writer, http.StatusBadRequest, utils.ErrorCallback{ErrorText: err.Error()})
		return
	}
	utils.WriteJSON(writer, http.StatusCreated, *inviteOut)
	return
}

func ActivateInvite(writer http.ResponseWriter, request *http.Request) {
	inviteActivateIn := invites.InviteActivateIn{}
	err := utils.ReadJSON(request, &inviteActivateIn)
	if err != nil {
		utils.WriteJSON(writer, http.StatusBadRequest, utils.ErrorCallback{ErrorText: err.Error()})
		return
	}
	inviteCheckOut, err := invites.InvSrv.ActivateInvite(&inviteActivateIn)
	if err != nil {
		switch {
		case errors.Is(err, database.InternalDBError):
			utils.WriteJSON(writer, http.StatusInternalServerError, utils.ErrorCallback{ErrorText: err.Error()})
		case errors.Is(err, invites.InviteNotFoundError):
			utils.WriteJSON(writer, http.StatusNotFound, utils.ErrorCallback{ErrorText: err.Error()})
		case errors.Is(err, invites.TypeOfVPNIsUnknownOrNotFound):
			utils.WriteJSON(writer, http.StatusNotFound, utils.ErrorCallback{ErrorText: err.Error()})
		case errors.Is(err, invites.IncorrectInviteError):
			utils.WriteJSON(writer, http.StatusBadRequest, utils.ErrorCallback{ErrorText: err.Error()})
		default:
			utils.WriteJSON(writer, http.StatusBadRequest, utils.ErrorCallback{ErrorText: err.Error()})
		}
		return
	}
	utils.WriteJSON(writer, http.StatusOK, *inviteCheckOut)
}
