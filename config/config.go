package config

import (
	"fmt"
	"os"
)

type DB struct {
	Host string
	Port string
	Pass string
	User string
	Name string
}

func (cfg *DB) DSN() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
		cfg.User, cfg.Pass, cfg.Host, cfg.Port, cfg.Name, "disable")
}

func Database() DB {
	return DB{
		Host: os.Getenv("POSTGRES_HOST"),
		Port: os.Getenv("POSTGRES_PORT"),
		Pass: os.Getenv("POSTGRES_PASSWORD"),
		User: os.Getenv("POSTGRES_USER"),
		Name: os.Getenv("POSTGRES_DB"),
	}
}
