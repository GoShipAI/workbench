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

// Agent AI助手/执行器
type Agent struct {
	ID          int64     `json:"id"`
	Name        string    `json:"name"`        // Agent名称
	Description string    `json:"description"` // 描述
	Type        string    `json:"type"`        // 类型: planner/executor
	Prompt      string    `json:"prompt"`      // 系统提示词
	ProviderID  *int64    `json:"provider_id"` // 关联的模型提供商
	Model       string    `json:"model"`       // 模型名称
	Tools       string    `json:"tools"`       // 可用工具列表 JSON ["claude_code", "shell"]
	WorkingDir  string    `json:"working_dir"` // 默认工作目录
	MaxRetries  int       `json:"max_retries"` // 最大重试次数
	Enabled     bool      `json:"enabled"`     // 是否启用
	CreatedAt   time.Time `json:"created_at"`
}

// AgentTool 工具定义
type AgentTool struct {
	Name        string `json:"name"`        // 工具名称
	Description string `json:"description"` // 描述（给LLM看）
	Type        string `json:"type"`        // 类型: cli/builtin
	Schema      string `json:"schema"`      // 参数 JSON Schema
}

// AgentStep 执行步骤
type AgentStep struct {
	ID             int64     `json:"id"`
	ConversationID int64     `json:"conversation_id"`
	StepNum        int       `json:"step_num"`
	Thought        string    `json:"thought"`      // AI的思考
	Action         string    `json:"action"`       // 选择的工具
	ActionInput    string    `json:"action_input"` // 工具输入 JSON
	Observation    string    `json:"observation"`  // 执行结果
	Status         string    `json:"status"`       // pending/running/success/failed
	Error          string    `json:"error"`        // 错误信息
	CreatedAt      time.Time `json:"created_at"`
}

// 步骤状态常量
const (
	StepStatusPending = "pending"
	StepStatusRunning = "running"
	StepStatusSuccess = "success"
	StepStatusFailed  = "failed"
)

// 工具名称常量
const (
	ToolClaudeCode = "claude_code" // 调用 Claude Code CLI
	ToolShell      = "shell"       // 执行 shell 命令
	ToolReadFile   = "read_file"   // 读取文件
	ToolWriteFile  = "write_file"  // 写入文件
	ToolListFiles  = "list_files"  // 列出文件
	ToolAskUser    = "ask_user"    // 询问用户
	ToolComplete   = "complete"    // 完成任务
)

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
	Type        string `json:"type"`
	Prompt      string `json:"prompt"`
	ProviderID  *int64 `json:"provider_id"`
	Model       string `json:"model"`
	Tools       string `json:"tools"`       // JSON数组
	WorkingDir  string `json:"working_dir"`
	MaxRetries  int    `json:"max_retries"`
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

// ProjectTimeStats 项目时间统计
type ProjectTimeStats struct {
	ProjectID   int64   `json:"project_id"`
	ProjectName string  `json:"project_name"`
	Color       string  `json:"color"`
	TotalHours  float64 `json:"total_hours"` // 总工时
	TaskCount   int     `json:"task_count"`  // 任务数量
	Percentage  float64 `json:"percentage"`  // 占比百分比
}

// DailyTaskStats 每日任务统计
type DailyTaskStats struct {
	Date           string  `json:"date"`            // 日期 YYYY-MM-DD
	TotalCount     int     `json:"total_count"`     // 总任务数
	CompletedCount int     `json:"completed_count"` // 完成数量
	CompletionRate float64 `json:"completion_rate"` // 完成率 (0-100)
}

// ReportSummary 报表汇总
type ReportSummary struct {
	TotalTasks     int     `json:"total_tasks"`     // 总任务数
	CompletedTasks int     `json:"completed_tasks"` // 已完成任务数
	TotalHours     float64 `json:"total_hours"`     // 总工时
	CompletedHours float64 `json:"completed_hours"` // 已完成工时
	AverageRate    float64 `json:"average_rate"`    // 平均完成率
}

// ReportData 报表数据汇总
type ReportData struct {
	ProjectStats []ProjectTimeStats `json:"project_stats"` // 项目时间统计
	DailyStats   []DailyTaskStats   `json:"daily_stats"`   // 每日任务统计
	Summary      ReportSummary      `json:"summary"`       // 汇总数据
}

// ========== AI会话相关模型 ==========

// TaskConversation 任务AI会话
type TaskConversation struct {
	ID        int64     `json:"id"`
	TaskID    int64     `json:"task_id"`
	AgentID   int64     `json:"agent_id"`
	AgentName string    `json:"agent_name"` // Agent名称（查询时填充）
	Status    string    `json:"status"`     // active/waiting_user/completed/failed
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// ConversationMessage 会话消息
type ConversationMessage struct {
	ID             int64     `json:"id"`
	ConversationID int64     `json:"conversation_id"`
	Role           string    `json:"role"`     // user/assistant/system
	Content        string    `json:"content"`  // 消息内容
	MessageType    string    `json:"type"`     // text/action/question/result/error
	Metadata       string    `json:"metadata"` // JSON: 执行的动作、选项等
	CreatedAt      time.Time `json:"created_at"`
}

// 会话状态常量
const (
	ConversationStatusActive      = "active"       // AI正在处理
	ConversationStatusWaitingUser = "waiting_user" // 等待用户回复
	ConversationStatusCompleted   = "completed"    // 已完成
	ConversationStatusFailed      = "failed"       // 失败
)

// 消息类型常量
const (
	MessageTypeText     = "text"     // 普通文本
	MessageTypeAction   = "action"   // 执行动作
	MessageTypeQuestion = "question" // 询问用户
	MessageTypeResult   = "result"   // 执行结果
	MessageTypeError    = "error"    // 错误信息
)

// StartConversationInput 开始AI会话的输入
type StartConversationInput struct {
	TaskID       int64  `json:"task_id"`
	AgentID      int64  `json:"agent_id"`
	ExtraContext string `json:"extra_context"` // 额外上下文（可选）
}

// SendMessageInput 发送消息的输入
type SendMessageInput struct {
	ConversationID int64  `json:"conversation_id"`
	Content        string `json:"content"`
}

// ConversationDetail 会话详情（包含消息列表）
type ConversationDetail struct {
	Conversation TaskConversation      `json:"conversation"`
	Messages     []ConversationMessage `json:"messages"`
	Task         Task                  `json:"task"`
}
