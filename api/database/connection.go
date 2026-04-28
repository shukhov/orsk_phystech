package database

import (
	"database/sql"
	"errors"
	"fmt"
	_ "github.com/lib/pq"
	"os"
	"strconv"
)

type DBConnectConfig struct {
	Host     string
	Port     int
	Username string
	Password string
	DBName   string
}

var InternalDBError = errors.New("internal database error")

func NewConn(conn *DBConnectConfig) (*sql.DB, error) {
	url := fmt.Sprintf("user=%s password=%s host=%s port=%v dbname=%s sslmode=disable",
		conn.Username, conn.Password, conn.Host, conn.Port, conn.DBName)
	db, err := sql.Open("postgres", url)
	if err != nil {
		return nil, err
	}
	return db, nil
}

func GetDBConnectConfig() *DBConnectConfig {
	port, _ := strconv.Atoi(os.Getenv("POSTGRES_PORT"))
	return &DBConnectConfig{
		Host:     os.Getenv("POSTGRES_HOST"),
		Port:     port,
		Username: os.Getenv("POSTGRES_USER"),
		Password: os.Getenv("POSTGRES_PASSWORD"),
		DBName:   os.Getenv("POSTGRES_DB"),
	}
}

func GetDB() (*sql.DB, error) {
	config := GetDBConnectConfig()
	conn, err := NewConn(config)
	if err != nil {
		return nil, err
	}
	return conn, nil
}
