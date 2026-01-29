package main

import (
	"fmt"
	"log"
	"time"
)

// StartConversation 开始一个AI会话
func (a *App) StartConversation(input StartConversationInput) (*ConversationDetail, error) {
	if db == nil {
		return nil, fmt.Errorf("数据库未初始化")
	}

	// 验证任务存在
	task, err := a.GetTask(input.TaskID)
	if err != nil {
		return nil, fmt.Errorf("任务不存在: %v", err)
	}

	// 验证Agent存在
	agent, err := a.GetAgent(input.AgentID)
	if err != nil {
		return nil, fmt.Errorf("Agent不存在: %v", err)
	}

	// 创建会话
	result, err := db.Exec(`
		INSERT INTO task_conversations (task_id, agent_id, status)
		VALUES (?, ?, ?)
	`, input.TaskID, input.AgentID, ConversationStatusActive)
	if err != nil {
		log.Printf("创建会话失败: %v", err)
		return nil, fmt.Errorf("创建会话失败: %v", err)
	}

	convID, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("获取会话ID失败: %v", err)
	}

	// 构建初始上下文消息
	contextMsg := fmt.Sprintf(`# 任务信息
- 任务名称: %s
- 任务描述: %s
- 所属项目: %s
- 状态: %s
- 计划日期: %s
- 截止日期: %s`,
		task.Name,
		task.Description,
		task.ProjectName,
		task.Status,
		safeString(task.Date),
		safeString(task.Deadline),
	)

	if input.ExtraContext != "" {
		contextMsg += "\n\n# 补充说明\n" + input.ExtraContext
	}

	// 保存系统上下文消息
	_, err = a.saveMessage(convID, "system", contextMsg, MessageTypeText, "{}")
	if err != nil {
		log.Printf("保存上下文消息失败: %v", err)
	}

	// 触发AI处理（异步）
	go a.runAIConversation(convID, agent)

	// 返回会话详情
	return a.GetConversationDetail(convID)
}

// GetConversationDetail 获取会话详情
func (a *App) GetConversationDetail(conversationID int64) (*ConversationDetail, error) {
	if db == nil {
		return nil, fmt.Errorf("数据库未初始化")
	}

	// 获取会话
	var conv TaskConversation
	err := db.QueryRow(`
		SELECT c.id, c.task_id, c.agent_id, COALESCE(a.name, '') as agent_name, c.status, c.created_at, c.updated_at
		FROM task_conversations c
		LEFT JOIN agents a ON c.agent_id = a.id
		WHERE c.id = ?
	`, conversationID).Scan(&conv.ID, &conv.TaskID, &conv.AgentID, &conv.AgentName, &conv.Status, &conv.CreatedAt, &conv.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("会话不存在: %v", err)
	}

	// 获取消息列表
	messages, err := a.getConversationMessages(conversationID)
	if err != nil {
		return nil, err
	}

	// 获取任务信息
	task, err := a.GetTask(conv.TaskID)
	if err != nil {
		return nil, err
	}

	return &ConversationDetail{
		Conversation: conv,
		Messages:     messages,
		Task:         *task,
	}, nil
}

// GetTaskConversations 获取任务的所有会话
func (a *App) GetTaskConversations(taskID int64) ([]TaskConversation, error) {
	if db == nil {
		return nil, fmt.Errorf("数据库未初始化")
	}

	rows, err := db.Query(`
		SELECT c.id, c.task_id, c.agent_id, COALESCE(a.name, '') as agent_name, c.status, c.created_at, c.updated_at
		FROM task_conversations c
		LEFT JOIN agents a ON c.agent_id = a.id
		WHERE c.task_id = ?
		ORDER BY c.created_at DESC
	`, taskID)
	if err != nil {
		return nil, fmt.Errorf("查询会话失败: %v", err)
	}
	defer rows.Close()

	var conversations []TaskConversation
	for rows.Next() {
		var conv TaskConversation
		if err := rows.Scan(&conv.ID, &conv.TaskID, &conv.AgentID, &conv.AgentName, &conv.Status, &conv.CreatedAt, &conv.UpdatedAt); err != nil {
			return nil, fmt.Errorf("扫描会话失败: %v", err)
		}
		conversations = append(conversations, conv)
	}

	return conversations, nil
}

// SendMessage 用户发送消息
func (a *App) SendMessage(input SendMessageInput) (*ConversationDetail, error) {
	if db == nil {
		return nil, fmt.Errorf("数据库未初始化")
	}

	// 获取会话
	var conv TaskConversation
	var agentID int64
	err := db.QueryRow(`SELECT id, agent_id, status FROM task_conversations WHERE id = ?`, input.ConversationID).Scan(&conv.ID, &agentID, &conv.Status)
	if err != nil {
		return nil, fmt.Errorf("会话不存在: %v", err)
	}

	// 保存用户消息
	_, err = a.saveMessage(input.ConversationID, "user", input.Content, MessageTypeText, "{}")
	if err != nil {
		return nil, fmt.Errorf("保存消息失败: %v", err)
	}

	// 更新会话状态为活跃
	_, err = db.Exec(`UPDATE task_conversations SET status = ?, updated_at = ? WHERE id = ?`,
		ConversationStatusActive, time.Now(), input.ConversationID)
	if err != nil {
		log.Printf("更新会话状态失败: %v", err)
	}

	// 获取Agent并继续执行
	agent, err := a.GetAgent(agentID)
	if err != nil {
		return nil, fmt.Errorf("获取Agent失败: %v", err)
	}

	// 异步继续AI处理
	go a.runAIConversation(input.ConversationID, agent)

	return a.GetConversationDetail(input.ConversationID)
}

// StopConversation 停止会话
func (a *App) StopConversation(conversationID int64) error {
	if db == nil {
		return fmt.Errorf("数据库未初始化")
	}

	_, err := db.Exec(`UPDATE task_conversations SET status = ?, updated_at = ? WHERE id = ?`,
		ConversationStatusFailed, time.Now(), conversationID)
	if err != nil {
		return fmt.Errorf("停止会话失败: %v", err)
	}

	return nil
}

// getConversationMessages 获取会话消息
func (a *App) getConversationMessages(conversationID int64) ([]ConversationMessage, error) {
	rows, err := db.Query(`
		SELECT id, conversation_id, role, content, message_type, metadata, created_at
		FROM conversation_messages
		WHERE conversation_id = ?
		ORDER BY created_at ASC
	`, conversationID)
	if err != nil {
		return nil, fmt.Errorf("查询消息失败: %v", err)
	}
	defer rows.Close()

	var messages []ConversationMessage
	for rows.Next() {
		var msg ConversationMessage
		if err := rows.Scan(&msg.ID, &msg.ConversationID, &msg.Role, &msg.Content, &msg.MessageType, &msg.Metadata, &msg.CreatedAt); err != nil {
			return nil, fmt.Errorf("扫描消息失败: %v", err)
		}
		messages = append(messages, msg)
	}

	return messages, nil
}

// saveMessage 保存消息
func (a *App) saveMessage(conversationID int64, role, content, msgType, metadata string) (int64, error) {
	result, err := db.Exec(`
		INSERT INTO conversation_messages (conversation_id, role, content, message_type, metadata)
		VALUES (?, ?, ?, ?, ?)
	`, conversationID, role, content, msgType, metadata)
	if err != nil {
		return 0, fmt.Errorf("保存消息失败: %v", err)
	}

	// 更新会话时间
	db.Exec(`UPDATE task_conversations SET updated_at = ? WHERE id = ?`, time.Now(), conversationID)

	return result.LastInsertId()
}

// updateConversationStatus 更新会话状态
func (a *App) updateConversationStatus(conversationID int64, status string) error {
	_, err := db.Exec(`UPDATE task_conversations SET status = ?, updated_at = ? WHERE id = ?`,
		status, time.Now(), conversationID)
	return err
}

// safeString 安全获取字符串指针的值
func safeString(s *string) string {
	if s == nil {
		return "未设置"
	}
	return *s
}
