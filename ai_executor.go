package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"regexp"
	"strings"
	"time"
)

// OpenAI兼容的API请求/响应结构
type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ChatRequest struct {
	Model       string        `json:"model"`
	Messages    []ChatMessage `json:"messages"`
	Temperature float64       `json:"temperature,omitempty"`
	MaxTokens   int           `json:"max_tokens,omitempty"`
}

type ChatChoice struct {
	Message      ChatMessage `json:"message"`
	FinishReason string      `json:"finish_reason"`
}

type ChatResponse struct {
	Choices []ChatChoice `json:"choices"`
	Error   *struct {
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

// AgentAction AI响应的动作结构
type AgentAction struct {
	Thought     string          `json:"thought"`
	Action      string          `json:"action"`
	ActionInput json.RawMessage `json:"action_input"`
}

// ReActExecutor ReAct模式执行器
type ReActExecutor struct {
	app            *App
	conversationID int64
	agent          *Agent
	provider       *ModelProvider
	toolExecutor   *ToolExecutor
	maxSteps       int // 最大步骤数，防止无限循环
}

// NewReActExecutor 创建ReAct执行器
func NewReActExecutor(app *App, conversationID int64, agent *Agent, provider *ModelProvider) *ReActExecutor {
	workingDir := agent.WorkingDir
	if workingDir == "" {
		workingDir = "."
	}

	return &ReActExecutor{
		app:            app,
		conversationID: conversationID,
		agent:          agent,
		provider:       provider,
		toolExecutor:   NewToolExecutor(workingDir),
		maxSteps:       20, // 默认最多20步
	}
}

// Run 运行ReAct执行循环
func (r *ReActExecutor) Run() {
	log.Printf("开始ReAct执行: conversationID=%d, agent=%s", r.conversationID, r.agent.Name)

	stepNum := 0
	for stepNum < r.maxSteps {
		stepNum++
		log.Printf("执行步骤 %d", stepNum)

		// 1. 构建Prompt
		prompt, err := r.buildPrompt()
		if err != nil {
			r.handleError(fmt.Sprintf("构建Prompt失败: %v", err))
			return
		}

		// 2. 调用LLM
		response, err := r.callLLM(prompt)
		if err != nil {
			r.handleError(fmt.Sprintf("调用LLM失败: %v", err))
			return
		}

		// 3. 解析响应
		action, err := r.parseResponse(response)
		if err != nil {
			// 解析失败，可能是格式问题，保存原始响应并等待用户
			log.Printf("解析响应失败: %v, 原始响应: %s", err, response)
			r.app.saveMessage(r.conversationID, "assistant", response, MessageTypeText, "{}")
			r.app.updateConversationStatus(r.conversationID, ConversationStatusWaitingUser)
			return
		}

		// 4. 保存步骤记录
		step := &AgentStep{
			ConversationID: r.conversationID,
			StepNum:        stepNum,
			Thought:        action.Thought,
			Action:         action.Action,
			ActionInput:    string(action.ActionInput),
			Status:         StepStatusRunning,
		}
		stepID, err := r.saveStep(step)
		if err != nil {
			log.Printf("保存步骤失败: %v", err)
		}
		step.ID = stepID

		// 发送思考过程给前端
		r.app.saveMessage(r.conversationID, "assistant", action.Thought, MessageTypeText,
			fmt.Sprintf(`{"step_num":%d,"action":"%s"}`, stepNum, action.Action))

		// 5. 检查是否完成
		if action.Action == ToolComplete {
			var input ToolInput
			json.Unmarshal(action.ActionInput, &input)
			r.updateStepStatus(step.ID, StepStatusSuccess, input.Summary, "")
			r.app.saveMessage(r.conversationID, "assistant", input.Summary, MessageTypeResult, "{}")
			r.app.updateConversationStatus(r.conversationID, ConversationStatusCompleted)
			log.Printf("任务完成: %s", input.Summary)
			return
		}

		// 6. 检查是否需要询问用户
		if action.Action == ToolAskUser {
			var input ToolInput
			json.Unmarshal(action.ActionInput, &input)
			metadata := "{}"
			if len(input.Options) > 0 {
				optionsJSON, _ := json.Marshal(map[string]interface{}{
					"options": input.Options,
				})
				metadata = string(optionsJSON)
			}
			r.updateStepStatus(step.ID, StepStatusSuccess, input.Question, "")
			r.app.saveMessage(r.conversationID, "assistant", input.Question, MessageTypeQuestion, metadata)
			r.app.updateConversationStatus(r.conversationID, ConversationStatusWaitingUser)
			log.Printf("等待用户输入: %s", input.Question)
			return
		}

		// 7. 执行工具
		result := r.toolExecutor.Execute(action.Action, string(action.ActionInput))

		// 8. 更新步骤状态
		if result.Success {
			r.updateStepStatus(step.ID, StepStatusSuccess, result.Output, "")
		} else {
			r.updateStepStatus(step.ID, StepStatusFailed, result.Output, result.Error)
		}

		// 9. 将观察结果作为消息保存（用于下次LLM调用）
		observationMsg := fmt.Sprintf("[工具执行结果]\n工具: %s\n状态: %s\n输出:\n%s",
			action.Action,
			map[bool]string{true: "成功", false: "失败"}[result.Success],
			result.Output)
		if result.Error != "" {
			observationMsg += fmt.Sprintf("\n错误: %s", result.Error)
		}
		r.app.saveMessage(r.conversationID, "system", observationMsg, MessageTypeResult,
			fmt.Sprintf(`{"step_num":%d,"tool":"%s","success":%t}`, stepNum, action.Action, result.Success))

		// 10. 如果需要用户输入（ask_user工具返回），暂停循环
		if result.NeedsUser {
			r.app.updateConversationStatus(r.conversationID, ConversationStatusWaitingUser)
			return
		}

		// 继续下一步
	}

	// 达到最大步骤数
	r.handleError(fmt.Sprintf("达到最大步骤数限制 (%d)，任务中止", r.maxSteps))
}

// buildPrompt 构建完整的Prompt
func (r *ReActExecutor) buildPrompt() ([]ChatMessage, error) {
	var messages []ChatMessage

	// 1. 系统提示词
	systemPrompt := r.buildSystemPrompt()
	messages = append(messages, ChatMessage{
		Role:    "system",
		Content: systemPrompt,
	})

	// 2. 获取历史消息
	historyMsgs, err := r.app.getConversationMessages(r.conversationID)
	if err != nil {
		return nil, fmt.Errorf("获取历史消息失败: %v", err)
	}

	// 3. 转换历史消息
	for _, msg := range historyMsgs {
		role := msg.Role
		if role == "system" {
			// 工具执行结果作为用户消息发送给LLM
			role = "user"
		}
		messages = append(messages, ChatMessage{
			Role:    role,
			Content: msg.Content,
		})
	}

	return messages, nil
}

// buildSystemPrompt 构建系统提示词
func (r *ReActExecutor) buildSystemPrompt() string {
	var sb strings.Builder

	// 基本角色
	sb.WriteString("你是一个任务执行Agent，能够自主完成软件工程任务。\n\n")

	// Agent自定义提示词
	if r.agent.Prompt != "" {
		sb.WriteString("## Agent设定\n")
		sb.WriteString(r.agent.Prompt)
		sb.WriteString("\n\n")
	}

	// 可用工具
	var toolNames []string
	if r.agent.Tools != "" && r.agent.Tools != "[]" {
		json.Unmarshal([]byte(r.agent.Tools), &toolNames)
	}
	if len(toolNames) == 0 {
		// 默认工具
		toolNames = []string{ToolShell, ToolReadFile, ToolWriteFile, ToolListFiles, ToolAskUser, ToolComplete}
	}
	tools := r.toolExecutor.registry.GetTools(toolNames)
	sb.WriteString(BuildToolsPrompt(tools))

	return sb.String()
}

// parseResponse 解析LLM响应，提取动作
func (r *ReActExecutor) parseResponse(response string) (*AgentAction, error) {
	response = strings.TrimSpace(response)

	// 尝试直接解析JSON
	var action AgentAction
	if err := json.Unmarshal([]byte(response), &action); err == nil {
		if action.Action != "" {
			return &action, nil
		}
	}

	// 尝试从Markdown代码块中提取JSON
	jsonPattern := regexp.MustCompile("(?s)```(?:json)?\\s*\\n?(.+?)\\n?```")
	if matches := jsonPattern.FindStringSubmatch(response); len(matches) > 1 {
		jsonStr := strings.TrimSpace(matches[1])
		if err := json.Unmarshal([]byte(jsonStr), &action); err == nil {
			if action.Action != "" {
				return &action, nil
			}
		}
	}

	// 尝试找到JSON对象（{...}）
	jsonObjPattern := regexp.MustCompile(`(?s)\{[^{}]*"action"[^{}]*\}`)
	if matches := jsonObjPattern.FindString(response); matches != "" {
		if err := json.Unmarshal([]byte(matches), &action); err == nil {
			if action.Action != "" {
				return &action, nil
			}
		}
	}

	return nil, fmt.Errorf("无法从响应中解析动作JSON")
}

// callLLM 调用LLM API
func (r *ReActExecutor) callLLM(messages []ChatMessage) (string, error) {
	model := r.agent.Model
	if model == "" {
		model = "deepseek-chat"
	}

	reqBody := ChatRequest{
		Model:       model,
		Messages:    messages,
		Temperature: 0.3, // 降低温度，使输出更确定
		MaxTokens:   2000,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("序列化请求失败: %v", err)
	}

	apiURL := r.provider.BaseURL
	if !strings.HasSuffix(apiURL, "/") {
		apiURL += "/"
	}
	apiURL += "chat/completions"

	log.Printf("调用LLM API: %s, model=%s", apiURL, model)

	req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("创建请求失败: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+r.provider.APIKey)

	client := &http.Client{Timeout: 120 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("请求失败: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("读取响应失败: %v", err)
	}

	log.Printf("LLM响应状态: %d", resp.StatusCode)

	var chatResp ChatResponse
	if err := json.Unmarshal(body, &chatResp); err != nil {
		return "", fmt.Errorf("解析响应失败: %v, body: %s", err, string(body))
	}

	if chatResp.Error != nil {
		return "", fmt.Errorf("API错误: %s", chatResp.Error.Message)
	}

	if len(chatResp.Choices) == 0 {
		return "", fmt.Errorf("LLM未返回响应")
	}

	return chatResp.Choices[0].Message.Content, nil
}

// saveStep 保存执行步骤
func (r *ReActExecutor) saveStep(step *AgentStep) (int64, error) {
	if db == nil {
		return 0, fmt.Errorf("数据库未初始化")
	}

	result, err := db.Exec(`
		INSERT INTO agent_steps (conversation_id, step_num, thought, action, action_input, observation, status, error)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`, step.ConversationID, step.StepNum, step.Thought, step.Action, step.ActionInput, step.Observation, step.Status, step.Error)

	if err != nil {
		return 0, fmt.Errorf("插入步骤失败: %v", err)
	}

	return result.LastInsertId()
}

// updateStepStatus 更新步骤状态
func (r *ReActExecutor) updateStepStatus(stepID int64, status string, observation string, errMsg string) {
	if db == nil {
		return
	}

	_, err := db.Exec(`
		UPDATE agent_steps SET status = ?, observation = ?, error = ? WHERE id = ?
	`, status, observation, errMsg, stepID)

	if err != nil {
		log.Printf("更新步骤状态失败: %v", err)
	}
}

// handleError 处理错误
func (r *ReActExecutor) handleError(errMsg string) {
	log.Printf("ReAct执行错误: %s", errMsg)
	r.app.saveMessage(r.conversationID, "assistant", errMsg, MessageTypeError, "{}")
	r.app.updateConversationStatus(r.conversationID, ConversationStatusFailed)
}

// ============ 保留原有的入口函数，但改为使用ReActExecutor ============

// runAIConversation 运行AI会话（入口函数）
func (a *App) runAIConversation(conversationID int64, agent *Agent) {
	log.Printf("开始AI会话: conversationID=%d, agent=%s", conversationID, agent.Name)

	// 获取Provider
	if agent.ProviderID == nil {
		a.saveMessage(conversationID, "assistant", "错误：Agent未配置模型提供商", MessageTypeError, "{}")
		a.updateConversationStatus(conversationID, ConversationStatusFailed)
		return
	}

	provider, err := a.GetModelProvider(*agent.ProviderID)
	if err != nil {
		a.saveMessage(conversationID, "assistant", fmt.Sprintf("错误：获取模型提供商失败: %v", err), MessageTypeError, "{}")
		a.updateConversationStatus(conversationID, ConversationStatusFailed)
		return
	}

	if provider.APIKey == "" {
		a.saveMessage(conversationID, "assistant", "错误：模型提供商未配置API Key", MessageTypeError, "{}")
		a.updateConversationStatus(conversationID, ConversationStatusFailed)
		return
	}

	// 使用ReAct执行器
	executor := NewReActExecutor(a, conversationID, agent, provider)
	executor.Run()
}

// GetConversationSteps 获取会话的执行步骤
func (a *App) GetConversationSteps(conversationID int64) ([]AgentStep, error) {
	if db == nil {
		return nil, fmt.Errorf("数据库未初始化")
	}

	rows, err := db.Query(`
		SELECT id, conversation_id, step_num, thought, action, action_input, observation, status, error, created_at
		FROM agent_steps
		WHERE conversation_id = ?
		ORDER BY step_num ASC
	`, conversationID)
	if err != nil {
		return nil, fmt.Errorf("查询步骤失败: %v", err)
	}
	defer rows.Close()

	var steps []AgentStep
	for rows.Next() {
		var step AgentStep
		if err := rows.Scan(&step.ID, &step.ConversationID, &step.StepNum, &step.Thought,
			&step.Action, &step.ActionInput, &step.Observation, &step.Status, &step.Error, &step.CreatedAt); err != nil {
			return nil, fmt.Errorf("扫描步骤失败: %v", err)
		}
		steps = append(steps, step)
	}

	return steps, nil
}
