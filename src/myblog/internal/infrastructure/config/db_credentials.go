// src/myblog/internal/infrastructure/config/db_credentials.go
package config

import "errors"

type DBCredentials struct {
	DBName     string
	DBUser     string
	DBPassword string
}

func (db *DBCredentials) Validate() error {
	switch {
	case db.DBName == "":
		return errors.New("dBName required")
	case db.DBUser == "":
		return errors.New("dBUser required")
	case db.DBPassword == "":
		return errors.New("dBPassword required")
	}
	return nil
}
