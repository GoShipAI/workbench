package main

import "time"

// Project 项目
type Project struct {
	ID          int64     `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Color       string    `json:"color"`
	Archived    bool      `json:"archived"`   // 是否归档
	CreatedAt   time.Time `json:"created_at"`
	TaskCount   int       `json:"task_count"` // 任务数量（查询时填充）
}

// ModelProvider 模型提供商
type ModelProvider struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`      // 提供商名称: deepseek/tongyi/volcengine
	Label     string    `json:"label"`     // 显示名称
	APIKey    string    `json:"api_key"`   // API Key
	BaseURL   string    `json:"base_url"`  // API Base URL (可选)
	Enabled   bool      `json:"enabled"`   // 是否启用
	CreatedAt time.Time `json:"created_at"`
}

// Agent AI助手
type Agent struct {
	ID          int64     `json:"id"`
	Name        string    `json:"name"`        // Agent名称
	Description string    `json:"description"` // 描述
	Prompt      string    `json:"prompt"`      // 系统提示词
	ProviderID  *int64    `json:"provider_id"` // 关联的模型提供商
	Model       string    `json:"model"`       // 模型名称
	Enabled     bool      `json:"enabled"`     // 是否启用
	CreatedAt   time.Time `json:"created_at"`
}

// ModelProviderInput 创建/更新模型提供商的输入
type ModelProviderInput struct {
	ID      int64  `json:"id"`
	Name    string `json:"name"`
	Label   string `json:"label"`
	APIKey  string `json:"api_key"`
	BaseURL string `json:"base_url"`
	Enabled bool   `json:"enabled"`
}

// AgentInput 创建/更新Agent的输入
type AgentInput struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Prompt      string `json:"prompt"`
	ProviderID  *int64 `json:"provider_id"`
	Model       string `json:"model"`
	Enabled     bool   `json:"enabled"`
}

// 模型提供商常量
const (
	ProviderDeepSeek    = "deepseek"
	ProviderTongyi      = "tongyi"
	ProviderVolcEngine  = "volcengine"
)

// Task 任务
type Task struct {
	ID          int64     `json:"id"`
	ProjectID   *int64    `json:"project_id"`
	ProjectName string    `json:"project_name"` // 项目名称（查询时填充）
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Date        *string   `json:"date"`       // 计划日期 YYYY-MM-DD, nil=待办
	StartTime   *string   `json:"start_time"` // 计划开始时间 HH:MM
	EndTime     *string   `json:"end_time"`   // 计划结束时间 HH:MM
	Hours       float64   `json:"hours"`      // 预计工时
	Deadline    *string   `json:"deadline"`   // 截止日期 YYYY-MM-DD
	Priority    string    `json:"priority"`   // 重要程度: high/medium/low
	Urgency     string    `json:"urgency"`    // 紧急程度: high/medium/low
	Status      string    `json:"status"`     // pending/scheduled/in_progress/completed
	ActualStart *string   `json:"actual_start"` // 实际开始时间 HH:MM (完成时填写)
	ActualHours float64   `json:"actual_hours"` // 实际工时 (完成时填写)
	CreatedAt   time.Time `json:"created_at"`
}

// TaskInput 创建/更新任务的输入
type TaskInput struct {
	ID          int64   `json:"id"`
	ProjectID   *int64  `json:"project_id"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Date        *string `json:"date"`
	StartTime   *string `json:"start_time"`
	EndTime     *string `json:"end_time"`
	Hours       float64 `json:"hours"`
	Deadline    *string `json:"deadline"`
	Priority    string  `json:"priority"`
	Urgency     string  `json:"urgency"`
	Status      string  `json:"status"`
}

// CompleteTaskInput 完成任务时的输入
type CompleteTaskInput struct {
	ID          int64   `json:"id"`
	ActualStart *string `json:"actual_start"` // 实际开始时间 HH:MM
	ActualHours float64 `json:"actual_hours"` // 实际工时
}

// WorkbenchData 工作台数据
type WorkbenchData struct {
	TodayTasks      []Task  `json:"today_tasks"`
	TotalCount      int     `json:"total_count"`
	CompletedCount  int     `json:"completed_count"`
	PlannedHours    float64 `json:"planned_hours"`
	CompletedHours  float64 `json:"completed_hours"`
	PendingCount    int     `json:"pending_count"` // 待处理任务数
}

// 任务状态常量
const (
	TaskStatusPending    = "pending"     // 待办（无日期）
	TaskStatusScheduled  = "scheduled"   // 已安排（有日期但未开始）
	TaskStatusInProgress = "in_progress" // 进行中
	TaskStatusCompleted  = "completed"   // 已完成
)

// 优先级常量（重要程度）
const (
	PriorityHigh   = "high"   // 高
	PriorityMedium = "medium" // 中
	PriorityLow    = "low"    // 低
)

// 紧急程度常量
const (
	UrgencyHigh   = "high"   // 高
	UrgencyMedium = "medium" // 中
	UrgencyLow    = "low"    // 低
)
