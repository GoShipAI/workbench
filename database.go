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
			archived INTEGER DEFAULT 0,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		return fmt.Errorf("创建 projects 表失败: %v", err)
	}

	// 模型提供商表
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS model_providers (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL UNIQUE,
			label TEXT NOT NULL,
			api_key TEXT DEFAULT '',
			base_url TEXT DEFAULT '',
			enabled INTEGER DEFAULT 1,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		return fmt.Errorf("创建 model_providers 表失败: %v", err)
	}

	// Agent表
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS agents (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			description TEXT DEFAULT '',
			type TEXT DEFAULT 'executor',
			prompt TEXT DEFAULT '',
			provider_id INTEGER,
			model TEXT DEFAULT '',
			tools TEXT DEFAULT '[]',
			working_dir TEXT DEFAULT '',
			max_retries INTEGER DEFAULT 3,
			enabled INTEGER DEFAULT 1,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (provider_id) REFERENCES model_providers(id) ON DELETE SET NULL
		)
	`)
	if err != nil {
		return fmt.Errorf("创建 agents 表失败: %v", err)
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
			deadline TEXT,
			priority TEXT DEFAULT 'medium',
			urgency TEXT DEFAULT 'medium',
			status TEXT DEFAULT 'pending',
			actual_start TEXT,
			actual_hours REAL DEFAULT 0,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE SET NULL
		)
	`)
	if err != nil {
		return fmt.Errorf("创建 tasks 表失败: %v", err)
	}

	// 迁移：为旧表添加新字段（如果不存在）
	migrationColumns := []string{
		"ALTER TABLE tasks ADD COLUMN deadline TEXT",
		"ALTER TABLE tasks ADD COLUMN priority TEXT DEFAULT 'medium'",
		"ALTER TABLE tasks ADD COLUMN urgency TEXT DEFAULT 'medium'",
		"ALTER TABLE tasks ADD COLUMN actual_start TEXT",
		"ALTER TABLE tasks ADD COLUMN actual_hours REAL DEFAULT 0",
		"ALTER TABLE projects ADD COLUMN archived INTEGER DEFAULT 0",
		// Agent 表新增字段
		"ALTER TABLE agents ADD COLUMN type TEXT DEFAULT 'executor'",
		"ALTER TABLE agents ADD COLUMN tools TEXT DEFAULT '[]'",
		"ALTER TABLE agents ADD COLUMN working_dir TEXT DEFAULT ''",
		"ALTER TABLE agents ADD COLUMN max_retries INTEGER DEFAULT 3",
	}
	for _, sql := range migrationColumns {
		db.Exec(sql) // 忽略错误，因为列可能已存在
	}

	// Agent执行步骤表
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS agent_steps (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			conversation_id INTEGER NOT NULL,
			step_num INTEGER NOT NULL,
			thought TEXT DEFAULT '',
			action TEXT DEFAULT '',
			action_input TEXT DEFAULT '{}',
			observation TEXT DEFAULT '',
			status TEXT DEFAULT 'pending',
			error TEXT DEFAULT '',
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (conversation_id) REFERENCES task_conversations(id) ON DELETE CASCADE
		)
	`)
	if err != nil {
		return fmt.Errorf("创建 agent_steps 表失败: %v", err)
	}

	_, err = db.Exec(`CREATE INDEX IF NOT EXISTS idx_steps_conversation ON agent_steps(conversation_id)`)
	if err != nil {
		return fmt.Errorf("创建 steps conversation_id 索引失败: %v", err)
	}

	// 初始化默认模型提供商
	defaultProviders := []struct {
		name    string
		label   string
		baseURL string
	}{
		{"deepseek", "DeepSeek", "https://api.deepseek.com"},
		{"tongyi", "通义千问", "https://dashscope.aliyuncs.com/compatible-mode/v1"},
		{"volcengine", "火山引擎", "https://ark.cn-beijing.volces.com/api/v3"},
	}
	for _, p := range defaultProviders {
		db.Exec(`INSERT OR IGNORE INTO model_providers (name, label, base_url) VALUES (?, ?, ?)`,
			p.name, p.label, p.baseURL)
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

	_, err = db.Exec(`CREATE INDEX IF NOT EXISTS idx_tasks_deadline ON tasks(deadline)`)
	if err != nil {
		return fmt.Errorf("创建 deadline 索引失败: %v", err)
	}

	// AI会话表
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS task_conversations (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			task_id INTEGER NOT NULL,
			agent_id INTEGER NOT NULL,
			status TEXT DEFAULT 'active',
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (task_id) REFERENCES tasks(id) ON DELETE CASCADE,
			FOREIGN KEY (agent_id) REFERENCES agents(id) ON DELETE CASCADE
		)
	`)
	if err != nil {
		return fmt.Errorf("创建 task_conversations 表失败: %v", err)
	}

	// 会话消息表
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS conversation_messages (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			conversation_id INTEGER NOT NULL,
			role TEXT NOT NULL,
			content TEXT DEFAULT '',
			message_type TEXT DEFAULT 'text',
			metadata TEXT DEFAULT '{}',
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (conversation_id) REFERENCES task_conversations(id) ON DELETE CASCADE
		)
	`)
	if err != nil {
		return fmt.Errorf("创建 conversation_messages 表失败: %v", err)
	}

	// 会话索引
	_, err = db.Exec(`CREATE INDEX IF NOT EXISTS idx_conversations_task ON task_conversations(task_id)`)
	if err != nil {
		return fmt.Errorf("创建 task_id 索引失败: %v", err)
	}

	_, err = db.Exec(`CREATE INDEX IF NOT EXISTS idx_messages_conversation ON conversation_messages(conversation_id)`)
	if err != nil {
		return fmt.Errorf("创建 conversation_id 索引失败: %v", err)
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
