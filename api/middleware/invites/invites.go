package invites

import (
	"api/database"
	"api/middleware/xray"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"log"
	"os"
	"time"
)

var (
	IncorrectInviteError         = errors.New("incorrect invite")
	InviteNotFoundError          = errors.New("invite not found")
	TypeOfVPNIsUnknownOrNotFound = errors.New("type of vpn is unknown or not found")
)

var secret = []byte(os.Getenv("SECRET_KEY"))

type InviteService struct {
	DB     *sql.DB
	logger *log.Logger
}

func NewInviteService() *InviteService {
	invSvc := new(InviteService)
	invSvc.logger = log.New(os.Stdout, "XrayService: ", log.LstdFlags|log.Lshortfile)
	db, err := database.GetDB()
	if err != nil {
		panic(err)
	}
	invSvc.DB = db
	return invSvc
}

type InviteOut struct {
	Id int64 `json:"id"`

	InviteHash string `json:"invite_hash"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	ExpiresAt time.Time `json:"expires_at"`

	Status  string `json:"status"`
	VPNType string `json:"vpn_type"`
}

type InviteIn struct {
	InviteWord string    `json:"invite_word"`
	VPNType    string    `json:"vpn_type"`
	ExpiresAt  time.Time `json:"expires_at,omitempty"`
}

type InviteActivateIn struct {
	UserId     int64  `json:"user_id"`
	InviteWord string `json:"invite_word"`
	Alias      string `json:"alias"`
}

type InviteCheckOut struct {
	Id      int64  `json:"id"`
	Alias   string `json:"alias"`
	VPNType string `json:"vpn_type"`
}

func KeyLookupHash(key string, pepper []byte) string {
	mac := hmac.New(sha256.New, pepper)
	mac.Write([]byte(key))
	return hex.EncodeToString(mac.Sum(nil))
}

func (invSvc *InviteService) NewInvite(inviteIn *InviteIn) (*InviteOut, error) {
	hash := KeyLookupHash(inviteIn.InviteWord, secret)
	inviteOut := new(InviteOut)
	err := invSvc.DB.QueryRow(
		NewInviteQuery, hash, inviteIn.VPNType, inviteIn.ExpiresAt).Scan(
		&inviteOut.Id, &inviteOut.InviteHash, &inviteOut.CreatedAt, &inviteOut.UpdatedAt,
		&inviteOut.ExpiresAt, &inviteOut.Status, &inviteOut.VPNType)
	if err != nil {
		invSvc.logger.Printf("%#v", err)
		return nil, IncorrectInviteError
	}
	return inviteOut, nil
}

func (invSvc *InviteService) ActivateInvite(inviteCheckIn *InviteActivateIn) (*InviteCheckOut, error) {
	ctx := context.Background()
	tx, err := invSvc.DB.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelRepeatableRead})
	if err != nil {
		invSvc.logger.Println(err)
		return nil, database.InternalDBError
	}
	var vpnType string
	var inviteId int64
	defer func() { _ = tx.Commit() }()

	err = tx.QueryRow(GetInviteInfoQuery, KeyLookupHash(inviteCheckIn.InviteWord, secret)).Scan(
		&inviteId, &vpnType)

	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, InviteNotFoundError
		default:
			invSvc.logger.Println(err)
			return nil, database.InternalDBError
		}

	}
	var inviteCheckOut InviteCheckOut
	switch vpnType {
	case "vless":
		newVLESSClient := xray.NewClientIn{
			InviteId: inviteId,
			UserId:   inviteCheckIn.UserId,
			Alias:    inviteCheckIn.Alias,
		}
		newClient, err := xray.XraySrv.NewClient(&newVLESSClient, tx)
		if err != nil {
			return nil, IncorrectInviteError
		}
		inviteCheckOut.Id = newClient.Id
		inviteCheckOut.Alias = newClient.Alias
		inviteCheckOut.VPNType = "vless"
	default:
		return nil, TypeOfVPNIsUnknownOrNotFound
	}
	_, err = tx.Exec(ActivateInviteQuery, inviteId)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, InviteNotFoundError
		default:
			return nil, IncorrectInviteError
		}
	}
	return &inviteCheckOut, nil
}
