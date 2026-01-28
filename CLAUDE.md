# 工作台 项目指南

## 项目概述

工作台是一个个人效率管理软件，帮助用户管理日常任务、跟踪工时和组织项目。

## 技术栈

- **框架**: Wails v2 (Go 后端 + Vue 前端)
- **后端**: Go 1.24+ SQLite (WAL 模式)
- **前端**: Vue 3 + TypeScript + Vite
- **UI 组件**: Arco Design Vue

## 项目结构

```
TaskFlow/
├── main.go                 # Wails 应用入口
├── app.go                  # App 结构体和生命周期
├── model.go                # Go 数据模型
├── database.go             # SQLite 数据库操作
├── task.go                 # 任务业务逻辑
├── project.go              # 项目业务逻辑
├── go.mod / go.sum
├── wails.json
├── Makefile
├── build/
│   ├── appicon.png
│   └── darwin/Info.plist
└── frontend/
    ├── index.html
    ├── package.json
    ├── tsconfig.json
    ├── vite.config.ts
    └── src/
        ├── main.ts
        ├── App.vue
        ├── style.css
        └── components/
            ├── Workbench.vue       # 工作台
            ├── TaskManagement.vue  # 任务管理
            ├── PendingTasks.vue    # 待处理任务
            └── ProjectManagement.vue # 项目管理
```

## 常用命令

```bash
# 开发模式（支持热重载）
make dev

# 构建当前平台
make build

# 构建指定平台
make build-universal  # macOS 通用版本
make build-windows    # Windows
make build-linux      # Linux

# 前端依赖安装
make frontend-install

# 查看所有命令
make help
```

## 开发规范

### Go 后端
- 后端逻辑分布在多个文件：database.go, task.go, project.go
- 使用 Wails 的绑定机制暴露方法给前端调用
- 数据存储在用户数据目录的 SQLite 数据库

### Vue 前端
- 组件使用 Vue 3 Composition API + `<script setup>` 语法
- 使用 TypeScript 进行类型检查
- UI 组件统一使用 Arco Design Vue
- 暗色主题

### 数据存储路径
- macOS: `~/Library/Application Support/Workbench/`
- Windows: `%APPDATA%/Workbench/`

## 核心功能

### 1. 工作台 (Workbench)
- 今日任务概览
- 统计卡片：任务数、已完成数、计划工时、完成工时
- 进度条显示完成百分比
- 快速操作：完成/编辑任务

### 2. 任务管理 (TaskManagement)
- 按日期查看和管理任务
- 任务属性：项目、名称、描述、时间、工时、状态
- 两种工时录入方式：
  - 直接填写工时
  - 通过开始/结束时间计算

### 3. 待处理任务 (PendingTasks)
- 无日期的任务池
- 可随时分配到具体日期
- 任务规划和整理

### 4. 项目管理 (ProjectManagement)
- 创建和管理项目
- 项目属性：名称、描述、颜色
- 查看项目下的任务统计

## 数据模型

### Project (项目)
```go
type Project struct {
    ID          int64     `json:"id"`
    Name        string    `json:"name"`
    Description string    `json:"description"`
    Color       string    `json:"color"`
    CreatedAt   time.Time `json:"created_at"`
}
```

### Task (任务)
```go
type Task struct {
    ID          int64   `json:"id"`
    ProjectID   *int64  `json:"project_id"`
    ProjectName string  `json:"project_name"`
    Name        string  `json:"name"`
    Description string  `json:"description"`
    Date        *string `json:"date"`       // YYYY-MM-DD, nil=待处理
    StartTime   *string `json:"start_time"` // HH:MM
    EndTime     *string `json:"end_time"`   // HH:MM
    Hours       float64 `json:"hours"`
    Status      string  `json:"status"`     // pending/scheduled/in_progress/completed
    CreatedAt   time.Time `json:"created_at"`
}
```

## 任务状态流转

- `pending`: 待处理（无日期）
- `scheduled`: 已安排（有日期但未开始）
- `in_progress`: 进行中
- `completed`: 已完成
