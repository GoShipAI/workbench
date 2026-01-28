package main

import (
	"context"
	"log"
)

// App struct
type App struct {
	ctx context.Context
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx

	// 初始化数据库
	if err := InitDB(); err != nil {
		log.Printf("初始化数据库失败: %v", err)
	}
}

// shutdown is called when the app is closing
func (a *App) shutdown(ctx context.Context) {
	// 关闭数据库连接
	if err := CloseDB(); err != nil {
		log.Printf("关闭数据库失败: %v", err)
	}
}
