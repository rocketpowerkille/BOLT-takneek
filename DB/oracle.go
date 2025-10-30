package DB

import (
    "database/sql"
    "fmt"
    "log"

    "github.com/sirupsen/logrus"
    _ "github.com/godror/godror"
)

var OracleDB *sql.DB

func TestOracleConnection(host, port, service, user, password string) error {
    logrus.WithFields(logrus.Fields{
        "host":    host,
        "port":    port,
        "service": service,
        "user":    user,
    }).Info("Testing Oracle connection")

    dsn := fmt.Sprintf("%s/%s@%s:%s/%s", user, password, host, port, service)

    db, err := sql.Open("godror", dsn)
    if err != nil {
        logrus.WithError(err).Error("Failed to connect to Oracle")
        return fmt.Errorf("failed to connect to Oracle: %v", err)
    }
    defer db.Close()

    if err := db.Ping(); err != nil {
        logrus.WithError(err).Error("Oracle ping failed")
        return fmt.Errorf("Oracle ping failed: %v", err)
    }

    logrus.Info("Oracle connection test successful")
    return nil
}

func ConnectOracle(host, port, service, user, password string) {
    dsn := fmt.Sprintf("%s/%s@%s:%s/%s", user, password, host, port, service)

    db, err := sql.Open("godror", dsn)
    if err != nil {
        log.Fatalf("Failed to connect to Oracle: %v", err)
    }

    if err := db.Ping(); err != nil {
        log.Fatalf("Oracle ping failed: %v", err)
    }

    OracleDB = db
    log.Println("✅ Connected to Oracle database successfully")
}
