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

	// 自动创建必要的表
	if err := createTablesIfNotExist(); err != nil {
		zap.L().Error("create tables failed", zap.Error(err))
		return err
	}

	return nil
}

// createTablesIfNotExist 创建必要的表（如果不存在）
func createTablesIfNotExist() error {
	// 创建 users 表
	usersTableSQL := `
	CREATE TABLE IF NOT EXISTS users (
		id int unsigned NOT NULL AUTO_INCREMENT,
		username varchar(64) COLLATE utf8mb4_unicode_ci NOT NULL,
		password_hash varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL,
		name varchar(64) COLLATE utf8mb4_unicode_ci DEFAULT '',
		email varchar(128) COLLATE utf8mb4_unicode_ci DEFAULT '',
		phone varchar(32) COLLATE utf8mb4_unicode_ci DEFAULT '',
		is_active tinyint(1) DEFAULT '1',
		is_admin tinyint(1) DEFAULT '0',
		ldap_dn varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT '',
		created_at timestamp NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at timestamp NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
		last_login_at timestamp NULL DEFAULT NULL,
		last_login_ip varchar(45) COLLATE utf8mb4_unicode_ci DEFAULT '',
		PRIMARY KEY (id),
		UNIQUE KEY uk_username (username),
		KEY idx_is_active (is_active)
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户表';
	`

	_, err := db.Exec(usersTableSQL)
	if err != nil {
		return fmt.Errorf("create users table failed: %w", err)
	}

	// 检查并插入默认 admin 用户（如果不存在）
	insertAdminSQL := `
	INSERT IGNORE INTO users (username, password_hash, name, is_active, is_admin) 
	VALUES (?, ?, ?, ?, ?)
	`

	_, err = db.Exec(insertAdminSQL,
		"admin",
		"$2a$10$ZmRaa0ggeJmgxPQxn7d5vueiadwpb1.WRFFJwsMJQnqWBvpiHRdQK",
		"管理员",
		true,
		true,
	)
	if err != nil {
		return fmt.Errorf("insert admin user failed: %w", err)
	}

	zap.L().Info("Tables created/verified successfully")
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
