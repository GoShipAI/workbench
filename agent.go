package main

import (
	"fmt"
	"log"
)

// GetAgents 获取所有Agent
func (a *App) GetAgents() ([]Agent, error) {
	if db == nil {
		return nil, fmt.Errorf("数据库未初始化")
	}

	rows, err := db.Query(`
		SELECT id, name, description, prompt, provider_id, model, enabled, created_at
		FROM agents
		ORDER BY created_at DESC
	`)
	if err != nil {
		log.Printf("查询Agent失败: %v", err)
		return nil, fmt.Errorf("查询Agent失败: %v", err)
	}
	defer rows.Close()

	var agents []Agent
	for rows.Next() {
		var agent Agent
		if err := rows.Scan(&agent.ID, &agent.Name, &agent.Description, &agent.Prompt,
			&agent.ProviderID, &agent.Model, &agent.Enabled, &agent.CreatedAt); err != nil {
			log.Printf("扫描Agent失败: %v", err)
			return nil, fmt.Errorf("扫描Agent失败: %v", err)
		}
		agents = append(agents, agent)
	}

	return agents, nil
}

// GetAgent 获取单个Agent
func (a *App) GetAgent(id int64) (*Agent, error) {
	if db == nil {
		return nil, fmt.Errorf("数据库未初始化")
	}

	var agent Agent
	err := db.QueryRow(`
		SELECT id, name, description, prompt, provider_id, model, enabled, created_at
		FROM agents WHERE id = ?
	`, id).Scan(&agent.ID, &agent.Name, &agent.Description, &agent.Prompt,
		&agent.ProviderID, &agent.Model, &agent.Enabled, &agent.CreatedAt)
	if err != nil {
		log.Printf("查询Agent失败: %v", err)
		return nil, fmt.Errorf("查询Agent失败: %v", err)
	}

	return &agent, nil
}

// CreateAgent 创建Agent
func (a *App) CreateAgent(input AgentInput) (*Agent, error) {
	if db == nil {
		return nil, fmt.Errorf("数据库未初始化")
	}

	if input.Name == "" {
		return nil, fmt.Errorf("Agent名称不能为空")
	}

	result, err := db.Exec(`
		INSERT INTO agents (name, description, prompt, provider_id, model, enabled)
		VALUES (?, ?, ?, ?, ?, ?)
	`, input.Name, input.Description, input.Prompt, input.ProviderID, input.Model, input.Enabled)
	if err != nil {
		log.Printf("创建Agent失败: %v", err)
		return nil, fmt.Errorf("创建Agent失败: %v", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("获取Agent ID失败: %v", err)
	}

	log.Printf("创建Agent成功: %s (ID: %d)", input.Name, id)
	return a.GetAgent(id)
}

// UpdateAgent 更新Agent
func (a *App) UpdateAgent(input AgentInput) error {
	if db == nil {
		return fmt.Errorf("数据库未初始化")
	}

	if input.Name == "" {
		return fmt.Errorf("Agent名称不能为空")
	}

	_, err := db.Exec(`
		UPDATE agents
		SET name = ?, description = ?, prompt = ?, provider_id = ?, model = ?, enabled = ?
		WHERE id = ?
	`, input.Name, input.Description, input.Prompt, input.ProviderID, input.Model, input.Enabled, input.ID)
	if err != nil {
		log.Printf("更新Agent失败: %v", err)
		return fmt.Errorf("更新Agent失败: %v", err)
	}

	log.Printf("更新Agent成功: ID=%d", input.ID)
	return nil
}

// DeleteAgent 删除Agent
func (a *App) DeleteAgent(id int64) error {
	if db == nil {
		return fmt.Errorf("数据库未初始化")
	}

	_, err := db.Exec(`DELETE FROM agents WHERE id = ?`, id)
	if err != nil {
		log.Printf("删除Agent失败: %v", err)
		return fmt.Errorf("删除Agent失败: %v", err)
	}

	log.Printf("删除Agent成功: ID=%d", id)
	return nil
}

// GetEnabledAgents 获取已启用的Agent
func (a *App) GetEnabledAgents() ([]Agent, error) {
	if db == nil {
		return nil, fmt.Errorf("数据库未初始化")
	}

	rows, err := db.Query(`
		SELECT id, name, description, prompt, provider_id, model, enabled, created_at
		FROM agents
		WHERE enabled = 1
		ORDER BY created_at DESC
	`)
	if err != nil {
		log.Printf("查询Agent失败: %v", err)
		return nil, fmt.Errorf("查询Agent失败: %v", err)
	}
	defer rows.Close()

	var agents []Agent
	for rows.Next() {
		var agent Agent
		if err := rows.Scan(&agent.ID, &agent.Name, &agent.Description, &agent.Prompt,
			&agent.ProviderID, &agent.Model, &agent.Enabled, &agent.CreatedAt); err != nil {
			log.Printf("扫描Agent失败: %v", err)
			return nil, fmt.Errorf("扫描Agent失败: %v", err)
		}
		agents = append(agents, agent)
	}

	return agents, nil
}
