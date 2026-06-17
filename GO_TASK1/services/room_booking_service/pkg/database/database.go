package database

import (
    "database/sql"   
    "fmt"
    "strings"

    _ "github.com/lib/pq"
    _ "github.com/microsoft/go-mssqldb"
    "gorm.io/driver/postgres"
    "gorm.io/driver/sqlserver"
    "gorm.io/gorm"
)


type Db struct {
    Gorm  *gorm.DB
    SqlDb *sql.DB
}

type DBConnector interface {
    EstablishConnection(dbURL string) (*Db, error)
}

type DBService struct{}

func NewDBService() *DBService {
    return &DBService{}
}

func (s *DBService) EstablishConnection(dbURL string) (*Db, error) {
    var dialector gorm.Dialector

    if strings.HasPrefix(dbURL, "sqlite://") {
        return nil, fmt.Errorf("sqlite not supported on this build")  
    } else if strings.HasPrefix(dbURL, "sqlserver://") {
        dialector = sqlserver.Open(dbURL)
    } else if strings.HasPrefix(dbURL, "postgres://") {
        dialector = postgres.Open(dbURL)
    } else {
        return nil, fmt.Errorf("unsupported database URL: %s", dbURL)
    }

    db, err := gorm.Open(dialector, &gorm.Config{})
    if err != nil {
        return nil, err
    }

    sqlDB, err := db.DB()
    if err != nil {
        return nil, err
    }

    if err := sqlDB.Ping(); err != nil {
        return nil, err
    }

    return &Db{
        Gorm:  db,
        SqlDb: sqlDB,
    }, nil
}