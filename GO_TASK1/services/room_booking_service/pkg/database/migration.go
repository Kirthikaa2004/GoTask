package database

import (
    "room_booking_management/room_booking_service/internal/loggers"
    "room_booking_management/room_booking_service/internal/models"

    "go.uber.org/zap"
)

type Migration struct{}


func NewMigration() *Migration {
    return &Migration{}
}

func (m *Migration) AutoMigrate(db *Db, logger *loggers.Logger) error {
    logger.Info("Running database migration...")

    err := db.Gorm.AutoMigrate(
        &models.Customer{},
        &models.Room{},
        &models.Booking{},
    )
    if err != nil {
        logger.Error("Migration failed", zap.Error(err))
        return err  
    }

    logger.Info("Database migration completed successfully")
    return nil
}