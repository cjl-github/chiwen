package mysql

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/go-sql-driver/mysql" // 不要忘了导入数据库驱动
	"github.com/jmoiron/sqlx"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

var db *sqlx.DB

func InitDB() (err error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		viper.GetString("mysql.user"),
		viper.GetString("mysql.password"),
		viper.GetString("mysql.host"),
		viper.GetInt("mysql.port"),
		viper.GetString("mysql.dbname"),
	)

	db, err = sqlx.Connect("mysql", dsn)
	if err != nil {
		zap.L().Error("connect DB failed", zap.Error(err))
		return
	}

	// 设置连接池参数
	db.SetMaxOpenConns(viper.GetInt("mysql.max_open_connes"))
	db.SetMaxIdleConns(viper.GetInt("mysql.max_idle_connes"))
	db.SetConnMaxLifetime(time.Hour)
	db.SetConnMaxIdleTime(time.Minute * 30)

	// 测试连接
	err = db.Ping()
	if err != nil {
		zap.L().Error("ping DB failed", zap.Error(err))
		return err
	}

	zap.L().Info("Database connected successfully")
	return nil
}

// DB 返回数据库连接实例
func DB() *sqlx.DB {
	return db
}

// GetDB 返回原始的 *sql.DB 用于某些操作
func GetDB() *sql.DB {
	return db.DB
}

func Close() {
	if db != nil {
		_ = db.Close()
		zap.L().Info("Database connection closed")
	}
}
