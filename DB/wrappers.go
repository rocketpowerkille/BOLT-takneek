package DB

import (
    "database/sql"
    "time"
    "gorm.io/gorm"
    "github.com/pclubiitk/dbcli/config"
    "github.com/sirupsen/logrus"
)

// DBInterface is a common interface for any database type
type DBInterface interface {
    RawQuery(query string, args ...interface{}) (*sql.Rows, error)
    ExecQuery(query string, args ...interface{}) error
    BeginTx() (TxInterface, error)
    Close() error
}

// TxInterface represents a database transaction
type TxInterface interface {
    RawQuery(query string, args ...interface{}) (*sql.Rows, error)
    ExecQuery(query string, args ...interface{}) error
    Commit() error
    Rollback() error
}

// --------------------------------------------------------

type SQLWrapper struct {
    DB *sql.DB
}

func (s *SQLWrapper) RawQuery(query string, args ...interface{}) (*sql.Rows, error) {
    return s.DB.Query(query, args...)
}

func (s *SQLWrapper) ExecQuery(query string, args ...interface{}) error {
    _, err := s.DB.Exec(query, args...)
    return err
}

func (s *SQLWrapper) BeginTx() (TxInterface, error) {
    tx, err := s.DB.Begin()
    if err != nil {
        return nil, err
    }
    return &SQLTx{tx: tx}, nil
}

func (s *SQLWrapper) Close() error {
    return s.DB.Close()
}

// ConfigureConnectionPool configures connection pool settings for SQL database
func (s *SQLWrapper) ConfigureConnectionPool(cfg *config.Config) {
    if cfg == nil {
        return
    }

    pool := cfg.Database.ConnectionPool
    s.DB.SetMaxOpenConns(pool.MaxConnections)
    s.DB.SetMaxIdleConns(pool.MaxConnections / 2) // Half of max connections as idle
    s.DB.SetConnMaxLifetime(time.Duration(pool.ConnMaxLifetime) * time.Second)
    s.DB.SetConnMaxIdleTime(time.Duration(pool.MaxIdleTime) * time.Second)

    logrus.WithFields(logrus.Fields{
        "maxConnections":   pool.MaxConnections,
        "maxIdleTime":      pool.MaxIdleTime,
        "connMaxLifetime":  pool.ConnMaxLifetime,
    }).Info("SQL connection pool configured")
}

// --------------------------------------------------------

type SQLTx struct {
    tx *sql.Tx
}

func (t *SQLTx) RawQuery(query string, args ...interface{}) (*sql.Rows, error) {
    return t.tx.Query(query, args...)
}

func (t *SQLTx) ExecQuery(query string, args ...interface{}) error {
    _, err := t.tx.Exec(query, args...)
    return err
}

func (t *SQLTx) Commit() error {
    return t.tx.Commit()
}

func (t *SQLTx) Rollback() error {
    return t.tx.Rollback()
}

// --------------------------------------------------------

type GormWrapper struct {
    DB *gorm.DB
}

func (g *GormWrapper) RawQuery(query string, args ...interface{}) (*sql.Rows, error) {
    tx := g.DB.Raw(query, args...)
    return tx.Rows()
}

func (g *GormWrapper) ExecQuery(query string, args ...interface{}) error {
    tx := g.DB.Exec(query, args...)
    return tx.Error
}

func (g *GormWrapper) BeginTx() (TxInterface, error) {
    tx := g.DB.Begin()
    if tx.Error != nil {
        return nil, tx.Error
    }
    return &GormTx{tx: tx}, nil
}

func (g *GormWrapper) Close() error {
    sqlDB, err := g.DB.DB()
    if err != nil {
        return err
    }
    return sqlDB.Close()
}

// ConfigureConnectionPool configures connection pool settings for GORM database
func (g *GormWrapper) ConfigureConnectionPool(cfg *config.Config) error {
    if cfg == nil {
        return nil
    }

    sqlDB, err := g.DB.DB()
    if err != nil {
        return err
    }

    pool := cfg.Database.ConnectionPool
    sqlDB.SetMaxOpenConns(pool.MaxConnections)
    sqlDB.SetMaxIdleConns(pool.MaxConnections / 2) // Half of max connections as idle
    sqlDB.SetConnMaxLifetime(time.Duration(pool.ConnMaxLifetime) * time.Second)
    sqlDB.SetConnMaxIdleTime(time.Duration(pool.MaxIdleTime) * time.Second)

    logrus.WithFields(logrus.Fields{
        "maxConnections":   pool.MaxConnections,
        "maxIdleTime":      pool.MaxIdleTime,
        "connMaxLifetime":  pool.ConnMaxLifetime,
    }).Info("GORM connection pool configured")
    
    return nil
}

// --------------------------------------------------------

type GormTx struct {
    tx *gorm.DB
}

func (t *GormTx) RawQuery(query string, args ...interface{}) (*sql.Rows, error) {
    tx := t.tx.Raw(query, args...)
    return tx.Rows()
}

func (t *GormTx) ExecQuery(query string, args ...interface{}) error {
    tx := t.tx.Exec(query, args...)
    return tx.Error
}

func (t *GormTx) Commit() error {
    tx := t.tx.Commit()
    return tx.Error
}

func (t *GormTx) Rollback() error {
    tx := t.tx.Rollback()
    return tx.Error
}
