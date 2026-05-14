package security

import (
	"api/database"
	"bytes"
	"database/sql"
	"errors"
	"fmt"
	"golang.org/x/crypto/bcrypt"
	"log"
	"regexp"
	"time"
	"unicode/utf8"
)

type SecurityService struct {
	DB     *sql.DB
	logger log.Logger
}

type UserPublicOut struct {
	Id        int64     `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	Username string `json:"username"`
}

type UserPrivateOut struct {
	UserPublicOut
	Email  string `json:"email,omitempty"`
	Status string `json:"status"`
	RoleId int64  `json:"role_id"`
}

type UserRegisterIn struct {
	UserLoginIn
	Username string `json:"username"`
}

type UserId struct {
	UserId int64 `json:"user_id"`
}

var (
	InvalidUsername = errors.New("invalid username")
	InvalidPassword = errors.New("invalid password")
	InvalidEmail    = errors.New("invalid email")
	UserNotFound    = errors.New("user not found")
)

var SecSrv = NewSecurityService()

func NewSecurityService() *SecurityService {
	db, err := database.GetDB()
	if err != nil {
		log.Fatal(err)
	}
	return &SecurityService{DB: db}
}

func PasswordHash(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	strHash := bytes.NewBuffer(hash).String()
	return strHash, nil
}

func validateUsername(username string) bool {
	usernameRe := regexp.MustCompile(`^[A-Za-z0-9]{1,25}$`)
	return usernameRe.MatchString(username)
}

func validatePassword(password string) bool {
	return utf8.RuneCountInString(password) >= 8
}

func validateEmail(email string) bool {
	fmt.Println(email)
	var emailRe = regexp.MustCompile(`^[A-Za-z0-9._%+\-]+@[A-Za-z0-9.\-]+\.[A-Za-z]{2,}$`)
	return emailRe.MatchString(email)
}

func (SecSrv *SecurityService) CreateUser(userIn *UserRegisterIn) (*UserPublicOut, error) {
	if !validateUsername(userIn.Username) {
		return nil, InvalidUsername
	}
	if !validatePassword(userIn.Password) {
		return nil, InvalidPassword
	}
	if !validateEmail(userIn.Email) {
		return nil, InvalidEmail
	}
	hash, err := PasswordHash(userIn.Password)
	if err != nil {
		return nil, err
	}
	out := UserPublicOut{}
	err = SecSrv.DB.QueryRow(CreateUserQuery, userIn.Username, hash, userIn.Email).Scan(&out.Id, &out.Username, &out.CreatedAt, &out.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %#v", err)
	}
	return &out, nil
}

func (SecSrv *SecurityService) GetUserById(userId int64) (*UserPublicOut, error) {
	userOut := UserPublicOut{}
	err := SecSrv.DB.QueryRow(GetUserByIdQuery, userId).Scan(
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
	err := SecSrv.DB.QueryRow(MeQuery, userId).Scan(
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
