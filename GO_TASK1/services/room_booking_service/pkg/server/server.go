package server

import (
    "net/http"

    "room_booking_management/room_booking_service/internal/config"
    "room_booking_management/room_booking_service/internal/handlers"
    "room_booking_management/room_booking_service/internal/loggers"
    "room_booking_management/room_booking_service/pkg/database"

    "github.com/gorilla/mux"
    "github.com/rs/cors"
    "go.uber.org/zap"
)

type Application struct {
    Logger    *loggers.Logger    
    Router    *mux.Router
    Config    *config.Config
    DBService database.DBConnector
}

func InitializeApp(env string) (*Application, error) {
    cfg, err := config.LoadConfig("room-booking-service", env)
    if err != nil {
        return nil, err
    }
    logger := loggers.NewLogger(loggers.LogConfig{
        Level:       cfg.LogLevel,
        LogDir:      cfg.LogDir,
        FileName:    cfg.LogFileName,
        ServiceName: "room-booking-service",
    })
    logger.Info("Logger initialized")

    dbService := database.NewDBService()
    logger.Info("Initializing database connection")

    db, err := dbService.EstablishConnection(cfg.DatabaseURL)
    if err != nil {
        logger.Error("Failed to connect to database", zap.Error(err))
        return nil, err
    }
    logger.Info("Database connection established successfully")

    migration := database.NewMigration()
    if err := migration.AutoMigrate(db, logger); err != nil {
        logger.Error("Failed to run database migrations", zap.Error(err))
        return nil, err
    }

    router := mux.NewRouter()
    handlers.SetupRoutes(router, db, logger, cfg)

    return &Application{
        Logger:    logger,
        Router:    router,
        Config:    cfg,
        DBService: dbService,
    }, nil
}

func GetCorsOptions() cors.Options {
    return cors.Options{
        AllowedOrigins: []string{"*"},
        AllowedMethods: []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
        AllowedHeaders: []string{"Content-Type", "Authorization", "accessToken", "deviceIdentifier"},
    }
}

func RunServer(app *Application) error {
    app.Logger.Info("Starting server on port", zap.String("port", app.Config.Port)) 

    corsOpts := GetCorsOptions()
    handler := cors.New(corsOpts).Handler(app.Router)
    return http.ListenAndServe(":"+app.Config.Port, handler)
}