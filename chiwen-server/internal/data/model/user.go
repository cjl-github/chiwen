package model

import "time"

type User struct {
	ID           uint       `gorm:"primarykey" json:"id"`
	Username     string     `gorm:"unique;size:64;not null" json:"username"`
	PasswordHash string     `gorm:"column:password_hash;size:255;not null" json:"-"` // 隐藏返回
	Name         string     `gorm:"size:64" json:"name"`
	Email        string     `gorm:"size:128" json:"email"`
	Phone        string     `gorm:"size:32" json:"phone"`
	IsActive     bool       `gorm:"default:true" json:"is_active"`
	IsAdmin      bool       `gorm:"default:false" json:"is_admin"`
	LdapDN       string     `gorm:"column:ldap_dn;size:255" json:"ldap_dn"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
	LastLoginAt  *time.Time `json:"last_login_at"`
	LastLoginIP  string     `gorm:"size:45" json:"last_login_ip"`
}

func (User) TableName() string {
	return "users"
}
