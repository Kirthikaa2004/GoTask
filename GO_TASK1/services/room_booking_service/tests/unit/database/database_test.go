package database_test

import (
	"log"
	"os"
	"testing"

	"room_booking_management/room_booking_service/pkg/database"

	"github.com/stretchr/testify/assert"
)

func setupDBEnv() {
	os.Setenv("DB_USERNAME", "username")
	os.Setenv("DB_PASSWORD", "password")
	os.Setenv("DB_HOST", "localhost")
	os.Setenv("DB_PORT", "1433")
	os.Setenv("DB_NAME", "testdb")
}

func teardownDBEnv() {
	os.Unsetenv("DB_USERNAME")
	os.Unsetenv("DB_PASSWORD")
	os.Unsetenv("DB_HOST")
	os.Unsetenv("DB_PORT")
	os.Unsetenv("DB_NAME")
}

func buildDSN() string {
	return "sqlserver://" +
		os.Getenv("DB_USERNAME") + ":" +
		os.Getenv("DB_PASSWORD") + "@" +
		os.Getenv("DB_HOST") + ":" +
		os.Getenv("DB_PORT") +
		"?database=" + os.Getenv("DB_NAME")
}

func TestEstablishConnection(t *testing.T) {
	setupDBEnv()
	defer teardownDBEnv()

	validDSN := buildDSN()

	tests := []struct {
		name        string
		dbURL       string
		wantErr     bool
		errContains string 
	}{
		{
			name:    "Valid SQL Server connection string format",
			dbURL:   validDSN,
			wantErr: false,
		},
		{
			name:        "Single slash - malformed sqlserver scheme",
			dbURL:       "sqlserver:/username:password@localhost:1433?database=testdb",
			wantErr:     true,
			errContains: "",
		},
		{
			name:        "Unsupported scheme - sqlite",
			dbURL:       "sqlite:///tmp/test.db",
			wantErr:     true,
			errContains: "sqlite not supported",
		},
		{
			name:        "Unsupported scheme - mysql",
			dbURL:       "mysql://username:password@localhost:3306/testdb",
			wantErr:     true,
			errContains: "unsupported database URL",
		},
		{
			name:        "Empty connection string",
			dbURL:       "",
			wantErr:     true,
			errContains: "unsupported database URL",
		},
		{
			name:        "Valid postgres scheme - no real server",
			dbURL:       "postgres://username:password@localhost:5432/testdb",
			wantErr:     true, 
			errContains: "",
		},
		{
			name:        "Malformed URL - missing host",
			dbURL:       "sqlserver://username:password@:1433?database=testdb",
			wantErr:     true,
			errContains: "",
		},
		{
			name:        "Malformed URL - missing database name",
			dbURL:       "sqlserver://username:password@localhost:1433?database=",
			wantErr:     true,
			errContains: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &database.DBService{}
			db, err := svc.EstablishConnection(tt.dbURL)

			if tt.wantErr {
				assert.Nil(t, db, "expected db to be nil on error")
				assert.NotNil(t, err, "expected an error but got none")
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
			} else {
				
				if err != nil {
					t.Logf("Skipping live DB assertions (no server available): %v", err)
					t.Skip("No database server available")
					return
				}
				assert.NotNil(t, db)
				assert.NotNil(t, db.Gorm, "expected Gorm instance to be set")
				assert.NotNil(t, db.SqlDb, "expected SqlDb instance to be set")

				if err := db.SqlDb.Ping(); err != nil {
					t.Logf("Post-connection ping failed: %v", err)
				}

				if cleanupErr := db.SqlDb.Close(); cleanupErr != nil {
					log.Printf("Failed to close the database: %v", cleanupErr)
				}
			}
		})
	}
}

func TestNewDBService(t *testing.T) {
	svc := database.NewDBService()
	assert.NotNil(t, svc)
}

func TestDBServiceImplementsInterface(t *testing.T) {
	var _ database.DBConnector = (*database.DBService)(nil)
}