# Workbench

**Personal productivity app with Agentic AI assistant**

[English](#english) | [中文](#中文)

---

## English

A personal productivity app featuring an **Agentic AI assistant** for task management, time tracking, and autonomous workflow execution. Built with Wails (Go + Vue).

### Features

#### 🤖 Agentic AI Assistant
- **Autonomous Execution** - ReAct-based AI agent that thinks, plans, and executes tasks
- **Tool Integration** - Built-in tools: Shell commands, file read/write, directory listing
- **Multi-turn Conversations** - Maintains context across interactions
- **Customizable Agents** - Create multiple agents with different prompts and tool configurations
- **OpenAI-compatible API** - Works with DeepSeek, OpenAI, and other providers

#### 📋 Task Management
- **Dashboard** - Today's tasks overview with progress tracking and statistics
- **Task Management** - Organize tasks by date, project, status with time tracking
- **Inbox** - Quick capture ideas, assign to dates later
- **Projects** - Categorize tasks with color-coded projects

### Download

Download the latest release for your platform:

- [macOS (Universal)](https://github.com/user/Workbench/releases) - Apple Silicon & Intel
- [Windows](https://github.com/user/Workbench/releases) - x64

### Tech Stack

- **Framework**: Wails v2 (Go + Vue)
- **Frontend**: Vue 3 + TypeScript + Arco Design
- **Backend**: Go + SQLite
- **AI**: OpenAI-compatible API (DeepSeek, OpenAI, etc.)

### Development

Prerequisites:
- Go 1.21+
- Node.js 18+
- [Wails CLI](https://wails.io/docs/gettingstarted/installation)

```bash
# Install dependencies
make frontend-install

# Development mode (hot reload)
make dev

# Build
make build           # Current platform
make build-all       # macOS + Windows

# See all commands
make help
```

### Data Location

- macOS: `~/Library/Application Support/Workbench/`
- Windows: `%APPDATA%/Workbench/`

### License

MIT License

---

## 中文

一款集成 **Agentic AI 智能助手**的个人效率管理软件，帮助你管理日常任务、跟踪工时，并通过 AI 自主执行工作流。

### 功能介绍

#### 🤖 Agentic AI 智能助手
- **自主执行** - 基于 ReAct 模式的 AI Agent，能够思考、规划并自主执行任务
- **工具集成** - 内置工具：Shell 命令、文件读写、目录浏览等
- **多轮对话** - 支持上下文连续对话，追踪执行步骤
- **自定义 Agent** - 可创建多个 Agent，配置不同的提示词和工具
- **兼容 OpenAI API** - 支持 DeepSeek、OpenAI 等多种 AI 服务

#### 📋 工作台
- 查看今日任务列表和完成进度
- 统计卡片：任务数、已完成数、计划工时、完成工时
- 快速创建今日任务或待办
- 日程规划：从待办池安排任务到指定日期

#### ✅ 任务管理
- 按日期范围、项目、状态筛选任务
- 任务状态流转：已安排 → 进行中 → 已完成
- 支持设置截止日期、优先级、紧急程度
- 工时录入：直接填写或通过开始/结束时间自动计算

#### 📝 待办
- 任务收集箱，快速记录想法
- 先捕获，后整理 - 无需立即决定执行时间
- 随时可将待办分配到具体日期，转为正式任务

#### 📁 项目管理
- 创建项目并设置颜色标识
- 任务可关联项目，便于分类和统计
- 支持项目归档

### 工作流示例

**任务管理流程：**
```
1. 想到要做的事 → 快速添加到「待办」
2. 规划时间时 → 从「待办」分配到具体日期 → 进入「任务管理」
3. 执行当天 → 在「工作台」查看今日任务并完成
```

**AI Agent 使用场景：**
```
• 让 AI 帮你分析项目代码结构
• 自动执行 Shell 命令完成批量操作
• 读取文件内容并生成报告
• 多步骤任务的自主规划与执行
```

### 下载安装

前往 [Releases](https://github.com/user/Workbench/releases) 页面下载：

- **macOS**: `Workbench.app` (通用版本，支持 Intel 和 Apple Silicon)
- **Windows**: `Workbench.exe`

### 从源码构建

环境要求：
- Go 1.21+
- Node.js 18+
- [Wails CLI](https://wails.io/docs/gettingstarted/installation)

```bash
# 安装前端依赖
make frontend-install

# 开发模式（热重载）
make dev

# 构建
make build           # 当前平台
make build-all       # macOS + Windows

# 查看所有命令
make help
```

### 数据存储位置

- macOS: `~/Library/Application Support/Workbench/`
- Windows: `%APPDATA%/Workbench/`

### 开源协议

MIT License
