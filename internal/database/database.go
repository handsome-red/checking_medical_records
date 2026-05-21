package database

import (
	"fmt"
	"log"
	"med_book/internal/config"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type Database struct {
	db *gorm.DB
}

// NewDatabase создаёт новый экземпляр Database
func NewDatabase(cfg *config.Config) (*Database, error) {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=UTC",
		cfg.DBHost,
		cfg.DBUser,
		cfg.DBPassword,
		cfg.DBName,
		cfg.DBPort,
	)

	// Подключаемся к БД
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		// Logger: logger.Default.LogMode(logger.Info),
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Настройка пула соединений
	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	log.Println("Database connected successfully!")

	return &Database{db: db}, nil
}

// GetDB возвращает GORM DB инстанс
func (d *Database) GetDB() *gorm.DB {
	return d.db
}

// Close закрывает соединение с БД
func (d *Database) Close() error {
	if d.db == nil {
		return nil
	}

	sqlDB, err := d.db.DB()
	if err != nil {
		return err
	}

	return sqlDB.Close()
}

// Ping проверяет соединение с БД
func (d *Database) Ping() error {
	sqlDB, err := d.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Ping()
}
