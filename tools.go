package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// ToolRegistry 工具注册表
type ToolRegistry struct {
	tools map[string]AgentTool
}

// NewToolRegistry 创建工具注册表
func NewToolRegistry() *ToolRegistry {
	registry := &ToolRegistry{
		tools: make(map[string]AgentTool),
	}
	registry.registerBuiltinTools()
	return registry
}

// registerBuiltinTools 注册内置工具
func (r *ToolRegistry) registerBuiltinTools() {
	r.tools[ToolClaudeCode] = AgentTool{
		Name:        ToolClaudeCode,
		Description: "调用 Claude Code CLI 执行复杂的代码任务。适用于需要阅读、修改代码或执行多步骤开发任务。",
		Type:        "cli",
		Schema: `{
			"type": "object",
			"properties": {
				"task": {"type": "string", "description": "要完成的任务描述"},
				"working_dir": {"type": "string", "description": "工作目录路径"}
			},
			"required": ["task"]
		}`,
	}

	r.tools[ToolShell] = AgentTool{
		Name:        ToolShell,
		Description: "执行 shell 命令。用于运行构建、测试、安装依赖等操作。",
		Type:        "builtin",
		Schema: `{
			"type": "object",
			"properties": {
				"command": {"type": "string", "description": "要执行的命令"},
				"working_dir": {"type": "string", "description": "工作目录路径"}
			},
			"required": ["command"]
		}`,
	}

	r.tools[ToolReadFile] = AgentTool{
		Name:        ToolReadFile,
		Description: "读取文件内容。",
		Type:        "builtin",
		Schema: `{
			"type": "object",
			"properties": {
				"path": {"type": "string", "description": "文件路径"}
			},
			"required": ["path"]
		}`,
	}

	r.tools[ToolWriteFile] = AgentTool{
		Name:        ToolWriteFile,
		Description: "写入内容到文件。",
		Type:        "builtin",
		Schema: `{
			"type": "object",
			"properties": {
				"path": {"type": "string", "description": "文件路径"},
				"content": {"type": "string", "description": "文件内容"}
			},
			"required": ["path", "content"]
		}`,
	}

	r.tools[ToolListFiles] = AgentTool{
		Name:        ToolListFiles,
		Description: "列出目录下的文件。",
		Type:        "builtin",
		Schema: `{
			"type": "object",
			"properties": {
				"path": {"type": "string", "description": "目录路径"},
				"pattern": {"type": "string", "description": "文件匹配模式，如 *.go"}
			},
			"required": ["path"]
		}`,
	}

	r.tools[ToolAskUser] = AgentTool{
		Name:        ToolAskUser,
		Description: "向用户提问，获取额外信息或确认。当需要澄清需求或做重要决定时使用。",
		Type:        "builtin",
		Schema: `{
			"type": "object",
			"properties": {
				"question": {"type": "string", "description": "问题内容"},
				"options": {"type": "array", "items": {"type": "string"}, "description": "可选的回答选项"}
			},
			"required": ["question"]
		}`,
	}

	r.tools[ToolComplete] = AgentTool{
		Name:        ToolComplete,
		Description: "标记任务完成。当任务已经完成时调用此工具。",
		Type:        "builtin",
		Schema: `{
			"type": "object",
			"properties": {
				"summary": {"type": "string", "description": "完成总结"}
			},
			"required": ["summary"]
		}`,
	}
}

// GetTool 获取工具
func (r *ToolRegistry) GetTool(name string) (AgentTool, bool) {
	tool, ok := r.tools[name]
	return tool, ok
}

// GetTools 获取指定名称的工具列表
func (r *ToolRegistry) GetTools(names []string) []AgentTool {
	var result []AgentTool
	for _, name := range names {
		if tool, ok := r.tools[name]; ok {
			result = append(result, tool)
		}
	}
	return result
}

// GetAllTools 获取所有工具
func (r *ToolRegistry) GetAllTools() []AgentTool {
	var result []AgentTool
	for _, tool := range r.tools {
		result = append(result, tool)
	}
	return result
}

// ToolInput 工具调用输入
type ToolInput struct {
	Task       string   `json:"task,omitempty"`
	WorkingDir string   `json:"working_dir,omitempty"`
	Command    string   `json:"command,omitempty"`
	Path       string   `json:"path,omitempty"`
	Content    string   `json:"content,omitempty"`
	Pattern    string   `json:"pattern,omitempty"`
	Question   string   `json:"question,omitempty"`
	Options    []string `json:"options,omitempty"`
	Summary    string   `json:"summary,omitempty"`
}

// ToolResult 工具执行结果
type ToolResult struct {
	Success     bool   `json:"success"`
	Output      string `json:"output"`
	Error       string `json:"error,omitempty"`
	NeedsUser   bool   `json:"needs_user,omitempty"`   // 是否需要用户输入
	IsCompleted bool   `json:"is_completed,omitempty"` // 任务是否完成
}

// ToolExecutor 工具执行器
type ToolExecutor struct {
	registry   *ToolRegistry
	workingDir string // 默认工作目录
}

// NewToolExecutor 创建工具执行器
func NewToolExecutor(workingDir string) *ToolExecutor {
	return &ToolExecutor{
		registry:   NewToolRegistry(),
		workingDir: workingDir,
	}
}

// Execute 执行工具
func (e *ToolExecutor) Execute(toolName string, inputJSON string) ToolResult {
	log.Printf("执行工具: %s, 输入: %s", toolName, inputJSON)

	var input ToolInput
	if err := json.Unmarshal([]byte(inputJSON), &input); err != nil {
		return ToolResult{Success: false, Error: fmt.Sprintf("解析输入失败: %v", err)}
	}

	// 设置默认工作目录
	if input.WorkingDir == "" {
		input.WorkingDir = e.workingDir
	}

	switch toolName {
	case ToolClaudeCode:
		return e.executeClaudeCode(input)
	case ToolShell:
		return e.executeShell(input)
	case ToolReadFile:
		return e.executeReadFile(input)
	case ToolWriteFile:
		return e.executeWriteFile(input)
	case ToolListFiles:
		return e.executeListFiles(input)
	case ToolAskUser:
		return e.executeAskUser(input)
	case ToolComplete:
		return e.executeComplete(input)
	default:
		return ToolResult{Success: false, Error: fmt.Sprintf("未知工具: %s", toolName)}
	}
}

// executeClaudeCode 执行 Claude Code CLI
func (e *ToolExecutor) executeClaudeCode(input ToolInput) ToolResult {
	// 检查 claude 命令是否存在
	claudePath, err := exec.LookPath("claude")
	if err != nil {
		return ToolResult{
			Success: false,
			Error:   "未找到 claude 命令，请确保已安装 Claude Code CLI",
		}
	}

	// 构建命令
	args := []string{"-p", input.Task, "--output-format", "text"}

	cmd := exec.Command(claudePath, args...)
	if input.WorkingDir != "" {
		cmd.Dir = input.WorkingDir
	}

	// 设置超时
	done := make(chan error, 1)
	var output strings.Builder

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return ToolResult{Success: false, Error: fmt.Sprintf("创建输出管道失败: %v", err)}
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return ToolResult{Success: false, Error: fmt.Sprintf("创建错误管道失败: %v", err)}
	}

	if err := cmd.Start(); err != nil {
		return ToolResult{Success: false, Error: fmt.Sprintf("启动命令失败: %v", err)}
	}

	// 读取输出
	go func() {
		scanner := bufio.NewScanner(io.MultiReader(stdout, stderr))
		for scanner.Scan() {
			output.WriteString(scanner.Text() + "\n")
		}
		done <- cmd.Wait()
	}()

	// 等待完成或超时 (10分钟)
	select {
	case err := <-done:
		if err != nil {
			return ToolResult{
				Success: false,
				Output:  output.String(),
				Error:   fmt.Sprintf("命令执行失败: %v", err),
			}
		}
		return ToolResult{Success: true, Output: output.String()}
	case <-time.After(10 * time.Minute):
		cmd.Process.Kill()
		return ToolResult{
			Success: false,
			Output:  output.String(),
			Error:   "命令执行超时",
		}
	}
}

// executeShell 执行 shell 命令
func (e *ToolExecutor) executeShell(input ToolInput) ToolResult {
	cmd := exec.Command("sh", "-c", input.Command)
	if input.WorkingDir != "" {
		cmd.Dir = input.WorkingDir
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		return ToolResult{
			Success: false,
			Output:  string(output),
			Error:   fmt.Sprintf("命令执行失败: %v", err),
		}
	}

	return ToolResult{Success: true, Output: string(output)}
}

// executeReadFile 读取文件
func (e *ToolExecutor) executeReadFile(input ToolInput) ToolResult {
	path := input.Path
	if !filepath.IsAbs(path) && e.workingDir != "" {
		path = filepath.Join(e.workingDir, path)
	}

	content, err := os.ReadFile(path)
	if err != nil {
		return ToolResult{Success: false, Error: fmt.Sprintf("读取文件失败: %v", err)}
	}

	return ToolResult{Success: true, Output: string(content)}
}

// executeWriteFile 写入文件
func (e *ToolExecutor) executeWriteFile(input ToolInput) ToolResult {
	path := input.Path
	if !filepath.IsAbs(path) && e.workingDir != "" {
		path = filepath.Join(e.workingDir, path)
	}

	// 确保目录存在
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return ToolResult{Success: false, Error: fmt.Sprintf("创建目录失败: %v", err)}
	}

	if err := os.WriteFile(path, []byte(input.Content), 0644); err != nil {
		return ToolResult{Success: false, Error: fmt.Sprintf("写入文件失败: %v", err)}
	}

	return ToolResult{Success: true, Output: fmt.Sprintf("文件已写入: %s", path)}
}

// executeListFiles 列出文件
func (e *ToolExecutor) executeListFiles(input ToolInput) ToolResult {
	path := input.Path
	if !filepath.IsAbs(path) && e.workingDir != "" {
		path = filepath.Join(e.workingDir, path)
	}

	var files []string

	if input.Pattern != "" {
		// 使用 glob 模式
		pattern := filepath.Join(path, input.Pattern)
		matches, err := filepath.Glob(pattern)
		if err != nil {
			return ToolResult{Success: false, Error: fmt.Sprintf("匹配模式失败: %v", err)}
		}
		files = matches
	} else {
		// 列出目录
		entries, err := os.ReadDir(path)
		if err != nil {
			return ToolResult{Success: false, Error: fmt.Sprintf("读取目录失败: %v", err)}
		}
		for _, entry := range entries {
			info, _ := entry.Info()
			if info != nil {
				files = append(files, fmt.Sprintf("%s\t%d\t%s",
					entry.Name(), info.Size(), info.ModTime().Format("2006-01-02 15:04")))
			} else {
				files = append(files, entry.Name())
			}
		}
	}

	return ToolResult{Success: true, Output: strings.Join(files, "\n")}
}

// executeAskUser 询问用户
func (e *ToolExecutor) executeAskUser(input ToolInput) ToolResult {
	// 返回需要用户输入的标记
	output := input.Question
	if len(input.Options) > 0 {
		output += "\n选项: " + strings.Join(input.Options, " / ")
	}

	return ToolResult{
		Success:   true,
		Output:    output,
		NeedsUser: true,
	}
}

// executeComplete 完成任务
func (e *ToolExecutor) executeComplete(input ToolInput) ToolResult {
	return ToolResult{
		Success:     true,
		Output:      input.Summary,
		IsCompleted: true,
	}
}

// BuildToolsPrompt 构建工具描述的 prompt
func BuildToolsPrompt(tools []AgentTool) string {
	var sb strings.Builder
	sb.WriteString("你可以使用以下工具:\n\n")

	for _, tool := range tools {
		sb.WriteString(fmt.Sprintf("## %s\n%s\n\n", tool.Name, tool.Description))
	}

	sb.WriteString(`
当你需要使用工具时，请严格按照以下 JSON 格式输出:

{"thought": "你的思考过程", "action": "工具名称", "action_input": {工具参数}}

示例:
{"thought": "我需要先看看项目结构", "action": "list_files", "action_input": {"path": "."}}
{"thought": "需要读取配置文件", "action": "read_file", "action_input": {"path": "config.json"}}
{"thought": "需要运行测试", "action": "shell", "action_input": {"command": "npm test"}}
{"thought": "任务已完成", "action": "complete", "action_input": {"summary": "已成功完成xxx"}}

重要:
1. 每次只执行一个工具
2. 根据工具执行结果决定下一步
3. 如果不确定，使用 ask_user 询问用户
4. 完成任务后，必须调用 complete 工具
`)

	return sb.String()
}
