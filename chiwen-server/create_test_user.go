package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/go-sql-driver/mysql"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	// 数据库连接信息
	dsn := "myuser:MyUserP@ss123@tcp(192.168.19.100:3306)/myapp?charset=utf8mb4&parseTime=True&loc=Local"

	// 打开数据库连接
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// 测试连接
	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	fmt.Println("Connected to database successfully!")

	// 创建测试用户
	username := "testuser"
	password := "testpassword123"

	// 生成bcrypt哈希密码
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Fatalf("Failed to hash password: %v", err)
	}

	// 检查用户是否已存在
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM users WHERE username = ?", username).Scan(&count)
	if err != nil {
		log.Fatalf("Failed to check if user exists: %v", err)
	}

	if count > 0 {
		fmt.Printf("User '%s' already exists. Updating password...\n", username)

		// 更新现有用户的密码
		_, err = db.Exec("UPDATE users SET password_hash = ?, is_active = 1 WHERE username = ?",
			string(hashedPassword), username)
		if err != nil {
			log.Fatalf("Failed to update user: %v", err)
		}

		fmt.Printf("Updated password for user '%s'\n", username)
	} else {
		// 插入新用户
		query := `INSERT INTO users (username, password_hash, name, email, is_active, is_admin) 
		          VALUES (?, ?, ?, ?, ?, ?)`

		_, err = db.Exec(query,
			username,
			string(hashedPassword),
			"Test User",
			"test@example.com",
			true,  // is_active
			false) // is_admin

		if err != nil {
			log.Fatalf("Failed to insert user: %v", err)
		}

		fmt.Printf("Created test user '%s' with password '%s'\n", username, password)
	}

	fmt.Println("Test user created/updated successfully!")
	fmt.Printf("Username: %s\n", username)
	fmt.Printf("Password: %s\n", password)
}
