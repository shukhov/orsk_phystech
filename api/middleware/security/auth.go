package security

import (
	"api/database"
	"api/middleware/utils"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"net/http"
	"os"
	"strings"
	"time"
)

var secretKey = []byte(os.Getenv("SECRET_KEY"))

var IncorrectLoginOrPassword = errors.New("incorrect login or password")

type ctxKey int

const userIDKey ctxKey = iota

type JWTAuthData struct {
	UserId int64 `json:"user_id"`
	jwt.RegisteredClaims
}

type UserLoginIn struct {
	Email    string `json:"email,omitempty"`
	Password string `json:"password"`
}

type AuthToken struct {
	Token string `json:"token"`
}

func generateToken(userId int64) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userId,
		"exp":     time.Now().Add(time.Hour * 24 * 30).Unix(), // Срок действия — месяц
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(secretKey)
}

func parseToken(token string) (*JWTAuthData, error) {
	claims := &JWTAuthData{}
	_, err := jwt.ParseWithClaims(token, claims, func(token *jwt.Token) (interface{}, error) {
		return secretKey, nil
	}, jwt.WithoutClaimsValidation())
	if err != nil {
		return nil, err
	}
	return claims, nil
}

func GetUserIdFromContext(ctx *context.Context) int64 {
	v := (*ctx).Value(userIDKey)
	id, _ := v.(int64)
	return id
}

func (SecSrv *SecurityService) RequireAuth(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(next http.ResponseWriter, request *http.Request) {
		auth := request.Header.Get("Authorization")
		tokens := strings.SplitN(auth, " ", 2)
		if len(tokens) != 2 {
			utils.WriteJSON(next, http.StatusUnauthorized, utils.ErrorCallback{ErrorText: "missing Bearer Token"})
			return
		}
		jwtAuth, err := parseToken(tokens[1])
		if err != nil {
			utils.WriteJSON(next, http.StatusBadRequest, utils.ErrorCallback{ErrorText: fmt.Sprintf("failed to parse token: %#v", err)})
			return
		}
		if jwtAuth.ExpiresAt.Unix() <= time.Now().Unix() {
			utils.WriteJSON(next, http.StatusUnauthorized, utils.ErrorCallback{ErrorText: "authorization token is expired"})
			return
		}
		ctx := context.WithValue(request.Context(), userIDKey, jwtAuth.UserId)
		handler.ServeHTTP(next, request.WithContext(ctx))
	})
}

func (SecSrv *SecurityService) Login(userIn *UserLoginIn) (string, error) {
	var hash string
	var id int64
	err := SecSrv.DB.QueryRow(
		"SELECT id, password_hash FROM public.users WHERE email = $1", userIn.Email,
	).Scan(&id, &hash)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", IncorrectLoginOrPassword
		}
		fmt.Println(err)
		return "", database.InternalDBError
	}
	if !(bcrypt.CompareHashAndPassword([]byte(hash), []byte(userIn.Password)) == nil) {
		return "", IncorrectLoginOrPassword
	}
	token, err := generateToken(id)
	if err != nil {
		fmt.Println(err)
		return "", database.InternalDBError
	}
	return token, nil
}
