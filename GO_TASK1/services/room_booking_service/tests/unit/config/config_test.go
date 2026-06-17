package config_test
 
import (
	"os"
	"testing"
 
	config "room_booking_management/room_booking_service/internal/config"
 
	"github.com/stretchr/testify/assert"
)
 
func TestMain(m *testing.M) {
	os.Exit(m.Run())
}
 
func TestLoadConfig(t *testing.T) {
	t.Run("Successful Load", func(t *testing.T) {
		os.Setenv("DATABASE_URL", "sqlserver://user:pass@localhost:1433?database=testdb")
		os.Setenv("PORT", "8080")
		os.Setenv("PROJECT_ROOT", "/app/root")
		os.Setenv("BASE_URL", "http://localhost:8080")
 
		cfg, err := config.LoadConfig("room-booking-service", "nonexistent-env")
 
		assert.NoError(t, err)
		assert.NotNil(t, cfg)
		assert.Equal(t, "8080", cfg.Port)
		assert.Equal(t, "sqlserver://user:pass@localhost:1433?database=testdb", cfg.DatabaseURL)
		assert.Equal(t, "/app/root", cfg.ProjectRoot)
		assert.Equal(t, "http://localhost:8080", cfg.BaseUrl)
	})
 
	t.Run("Missing DATABASE_URL Returns Error", func(t *testing.T) {
		os.Unsetenv("DATABASE_URL")
 
		cfg, err := config.LoadConfig("room-booking-service", "nonexistent-env")
 
		assert.Nil(t, cfg)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "DATABASE_URL")
	})
 
	t.Run("Default Values Applied", func(t *testing.T) {
		os.Setenv("DATABASE_URL", "sqlserver://user:pass@localhost:1433?database=testdb")
		os.Unsetenv("LOG_LEVEL")
		os.Unsetenv("LOG_DIR")
		os.Unsetenv("LOG_FILE_NAME")
 
		cfg, err := config.LoadConfig("room-booking-service", "nonexistent-env")
     
		assert.NoError(t, err)
		assert.Equal(t, "info", cfg.LogLevel)
		assert.Equal(t, "./logs", cfg.LogDir)
  	})
}
             