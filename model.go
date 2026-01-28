package main

import "time"

// Project 项目
type Project struct {
	ID          int64     `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Color       string    `json:"color"`
	CreatedAt   time.Time `json:"created_at"`
	TaskCount   int       `json:"task_count"` // 任务数量（查询时填充）
}

// Task 任务
type Task struct {
	ID          int64   `json:"id"`
	ProjectID   *int64  `json:"project_id"`
	ProjectName string  `json:"project_name"` // 项目名称（查询时填充）
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Date        *string `json:"date"`       // YYYY-MM-DD, nil=待处理
	StartTime   *string `json:"start_time"` // HH:MM
	EndTime     *string `json:"end_time"`   // HH:MM
	Hours       float64 `json:"hours"`
	Status      string  `json:"status"` // pending/scheduled/in_progress/completed
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
	Status      string  `json:"status"`
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
	TaskStatusPending    = "pending"     // 待处理（无日期）
	TaskStatusScheduled  = "scheduled"   // 已安排（有日期但未开始）
	TaskStatusInProgress = "in_progress" // 进行中
	TaskStatusCompleted  = "completed"   // 已完成
)
