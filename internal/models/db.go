package models

import (
	"context"
	"fmt"
	"ryzenlo/to2cloud/configs"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var DBClient *gorm.DB

func InitDBClient(ctx context.Context, conf *configs.Config) error {
	filename := fmt.Sprintf("%s/%s", conf.Sqlite.DirPath, conf.Sqlite.DBFileName)
	db, err := gorm.Open(sqlite.Open(filename), &gorm.Config{})
	if err != nil {
		return fmt.Errorf("failed to connect database,%w", err)
	}
	DBClient = db
	return nil
}
