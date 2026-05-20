package initialize

import (
	"context"
	"fmt"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm/logger"
	"k8soperation/pkg/app"
	"k8soperation/pkg/database"
)

func SetupDB(a *app.App) error {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=%s&parseTime=%t&loc=Local&timeout=1s&readTimeout=2s&writeTimeout=2s",
		a.DatabaseSetting.Username,
		a.DatabaseSetting.Password,
		a.DatabaseSetting.Host,
		a.DatabaseSetting.Port,
		a.DatabaseSetting.DBName,
		a.DatabaseSetting.Charset,
		a.DatabaseSetting.ParseTime,
	)

	dbConfig := mysql.New(mysql.Config{DSN: dsn})

	var err error
	a.DB, a.SQLDB, err = database.Connect(dbConfig, logger.Default.LogMode(logger.Info))
	if err != nil {
		return fmt.Errorf("connect db failed: %w", err)
	}

	a.SQLDB.SetMaxOpenConns(a.DatabaseSetting.MaxOpenConns)
	a.SQLDB.SetMaxIdleConns(a.DatabaseSetting.MaxIdleConns)
	a.SQLDB.SetConnMaxLifetime(time.Duration(a.DatabaseSetting.MaxLifeSeconds) * time.Second)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if err := a.SQLDB.PingContext(ctx); err != nil {
		return fmt.Errorf("db ping failed: %w", err)
	}

	return nil
}
