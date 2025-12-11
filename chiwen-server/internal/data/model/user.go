package model

import (
	"database/sql"
	"time"
)

type User struct {
	ID           uint           `gorm:"primarykey" db:"id" json:"id"`
	Username     string         `gorm:"unique;size:64;not null" db:"username" json:"username"`
	PasswordHash string         `gorm:"column:password_hash;size:255;not null" db:"password_hash" json:"-"` // 隐藏返回
	Name         sql.NullString `gorm:"size:64" db:"name" json:"name"`
	Email        sql.NullString `gorm:"size:128" db:"email" json:"email"`
	Phone        sql.NullString `gorm:"size:32" db:"phone" json:"phone"`
	IsActive     bool           `gorm:"default:true" db:"is_active" json:"is_active"`
	IsAdmin      bool           `gorm:"default:false" db:"is_admin" json:"is_admin"`
	LdapDN       sql.NullString `gorm:"column:ldap_dn;size:255" db:"ldap_dn" json:"ldap_dn"`
	CreatedAt    time.Time      `db:"created_at" json:"created_at"`
	UpdatedAt    time.Time      `db:"updated_at" json:"updated_at"`
	LastLoginAt  *time.Time     `db:"last_login_at" json:"last_login_at"`
	LastLoginIP  sql.NullString `gorm:"size:45" db:"last_login_ip" json:"last_login_ip"`
}

func (User) TableName() string {
	return "users"
}
