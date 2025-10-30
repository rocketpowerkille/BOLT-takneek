package DB

import (
	"fmt"

	"github.com/sirupsen/logrus"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var MySQLDB *gorm.DB

func TestMySQLConnection(host, port, password, dbName, user string) error {
	logrus.WithFields(logrus.Fields{
		"host":   host,
		"port":   port,
		"dbName": dbName,
		"user":   user,
	}).Info("Testing MySQL connection")

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		user, password, host, port, dbName)

	database, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		logrus.WithError(err).Error("Failed to connect to MySQL database")
		return fmt.Errorf("failed to connect to MySQL database: %v", err)
	}

	// Test the connection
	sqlDB, err := database.DB()
	if err != nil {
		logrus.WithError(err).Error("Failed to get database instance")
		return fmt.Errorf("failed to get database instance: %v", err)
	}

	if err := sqlDB.Ping(); err != nil {
		logrus.WithError(err).Error("Failed to ping MySQL database")
		return fmt.Errorf("failed to ping MySQL database: %v", err)
	}

	// Close test connection
	sqlDB.Close()
	logrus.Info("MySQL connection test successful")
	return nil
}

func ConnectMySQL(host, port, password, dbName, user string) {
	// Data Source Name (DSN) format for MySQL
	// Example: user:password@tcp(localhost:3306)/dbname?charset=utf8mb4&parseTime=True&loc=Local
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		user, password, host, port, dbName)

	database, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		logrus.Fatal("Failed to connect to MySQL database: ", err)
	}

	MySQLDB = database

	logrus.Info("✅ Connected to MySQL database successfully")
}
