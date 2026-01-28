package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"

	_ "modernc.org/sqlite"
)

var (
	db     *sql.DB
	dbOnce sync.Once
	dbErr  error
)

// getConfigDir 获取配置目录
func getConfigDir() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("获取用户目录失败: %v", err)
	}

	configDir := filepath.Join(homeDir, "Library", "Application Support", "Workbench")

	// 确保目录存在
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return "", fmt.Errorf("创建配置目录失败: %v", err)
	}

	return configDir, nil
}

// InitDB 初始化数据库连接
func InitDB() error {
	dbOnce.Do(func() {
		log.Println("初始化数据库...")
		configDir, err := getConfigDir()
		if err != nil {
			dbErr = fmt.Errorf("获取配置目录失败: %v", err)
			log.Printf("数据库初始化失败: %v", dbErr)
			return
		}

		dbPath := filepath.Join(configDir, "workbench.db")
		log.Printf("数据库路径: %s", dbPath)

		// 使用 WAL 模式和超时设置
		dsn := fmt.Sprintf("%s?_busy_timeout=5000&_journal_mode=WAL", dbPath)
		db, err = sql.Open("sqlite", dsn)
		if err != nil {
			dbErr = fmt.Errorf("打开数据库失败: %v", err)
			log.Printf("数据库初始化失败: %v", dbErr)
			return
		}

		// 设置连接池参数
		db.SetMaxOpenConns(1)
		db.SetMaxIdleConns(1)
		db.SetConnMaxLifetime(time.Hour)

		// 验证连接
		if err := db.Ping(); err != nil {
			dbErr = fmt.Errorf("数据库连接失败: %v", err)
			log.Printf("数据库初始化失败: %v", dbErr)
			return
		}

		// 创建表
		if err := createTables(); err != nil {
			dbErr = fmt.Errorf("创建表失败: %v", err)
			log.Printf("数据库初始化失败: %v", dbErr)
			return
		}

		log.Println("数据库初始化成功")
	})

	return dbErr
}

// createTables 创建数据库表
func createTables() error {
	// 项目表
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS projects (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL UNIQUE,
			description TEXT DEFAULT '',
			color TEXT DEFAULT '#165DFF',
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		return fmt.Errorf("创建 projects 表失败: %v", err)
	}

	// 任务表
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS tasks (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			project_id INTEGER,
			name TEXT NOT NULL,
			description TEXT DEFAULT '',
			date TEXT,
			start_time TEXT,
			end_time TEXT,
			hours REAL DEFAULT 0,
			status TEXT DEFAULT 'pending',
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE SET NULL
		)
	`)
	if err != nil {
		return fmt.Errorf("创建 tasks 表失败: %v", err)
	}

	// 创建索引
	_, err = db.Exec(`CREATE INDEX IF NOT EXISTS idx_tasks_date ON tasks(date)`)
	if err != nil {
		return fmt.Errorf("创建 date 索引失败: %v", err)
	}

	_, err = db.Exec(`CREATE INDEX IF NOT EXISTS idx_tasks_status ON tasks(status)`)
	if err != nil {
		return fmt.Errorf("创建 status 索引失败: %v", err)
	}

	_, err = db.Exec(`CREATE INDEX IF NOT EXISTS idx_tasks_project ON tasks(project_id)`)
	if err != nil {
		return fmt.Errorf("创建 project_id 索引失败: %v", err)
	}

	return nil
}

// GetDB 获取数据库连接
func GetDB() *sql.DB {
	return db
}

// CloseDB 关闭数据库连接
func CloseDB() error {
	if db != nil {
		return db.Close()
	}
	return nil
}
