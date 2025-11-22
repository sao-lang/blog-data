package bootstrap

import (
	"blog/internal/config"
	"blog/internal/models"
	"fmt"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func SetupPgSql(cfg *config.Config) (*gorm.DB, error) {
	dsn := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable", cfg.PgSql.User, cfg.PgSql.Password, cfg.PgSql.Host, cfg.PgSql.Port, cfg.PgSql.DatabaseName)
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	err = db.AutoMigrate(&models.User{})

	if err != nil {
		return nil, err
	}
	return db, nil
}
