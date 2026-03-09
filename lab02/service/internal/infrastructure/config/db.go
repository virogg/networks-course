package config

import (
	"fmt"

	pkgconfig "github.com/virogg/networks-course/service/pkg/config"
)

type DatabaseConfig struct {
	Host     string
	Port     string
	Name     string
	User     string
	Password string
}

func (db *DatabaseConfig) GetDSN() string {
	return fmt.Sprintf("host=%s port=%s dbname=%s user=%s password=%s sslmode=disable",
		db.Host, db.Port, db.Name, db.User, db.Password)
}

func LoadDatabaseConfig() (*DatabaseConfig, error) {
	dbHost, err := pkgconfig.GetEnv("POSTGRES_HOST")
	if err != nil {
		return nil, err
	}

	dbPort, err := pkgconfig.GetEnv("POSTGRES_PORT")
	if err != nil {
		return nil, err
	}

	dbName, err := pkgconfig.GetEnv("POSTGRES_DB")
	if err != nil {
		return nil, err
	}

	dbUser, err := pkgconfig.GetEnv("POSTGRES_USER")
	if err != nil {
		return nil, err
	}

	dbPassword, err := pkgconfig.GetEnv("POSTGRES_PASSWORD")
	if err != nil {
		return nil, err
	}

	return &DatabaseConfig{
		Host:     dbHost,
		Port:     dbPort,
		Name:     dbName,
		User:     dbUser,
		Password: dbPassword,
	}, nil
}
