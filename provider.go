package main

import (
	"fmt"
	"log"
)

// GetModelProviders 获取所有模型提供商
func (a *App) GetModelProviders() ([]ModelProvider, error) {
	if db == nil {
		return nil, fmt.Errorf("数据库未初始化")
	}

	rows, err := db.Query(`
		SELECT id, name, label, api_key, base_url, enabled, created_at
		FROM model_providers
		ORDER BY id
	`)
	if err != nil {
		log.Printf("查询模型提供商失败: %v", err)
		return nil, fmt.Errorf("查询模型提供商失败: %v", err)
	}
	defer rows.Close()

	var providers []ModelProvider
	for rows.Next() {
		var p ModelProvider
		if err := rows.Scan(&p.ID, &p.Name, &p.Label, &p.APIKey, &p.BaseURL, &p.Enabled, &p.CreatedAt); err != nil {
			log.Printf("扫描模型提供商失败: %v", err)
			return nil, fmt.Errorf("扫描模型提供商失败: %v", err)
		}
		providers = append(providers, p)
	}

	return providers, nil
}

// GetModelProvider 获取单个模型提供商
func (a *App) GetModelProvider(id int64) (*ModelProvider, error) {
	if db == nil {
		return nil, fmt.Errorf("数据库未初始化")
	}

	var p ModelProvider
	err := db.QueryRow(`
		SELECT id, name, label, api_key, base_url, enabled, created_at
		FROM model_providers WHERE id = ?
	`, id).Scan(&p.ID, &p.Name, &p.Label, &p.APIKey, &p.BaseURL, &p.Enabled, &p.CreatedAt)
	if err != nil {
		log.Printf("查询模型提供商失败: %v", err)
		return nil, fmt.Errorf("查询模型提供商失败: %v", err)
	}

	return &p, nil
}

// UpdateModelProvider 更新模型提供商（主要用于更新API Key）
func (a *App) UpdateModelProvider(input ModelProviderInput) error {
	if db == nil {
		return fmt.Errorf("数据库未初始化")
	}

	_, err := db.Exec(`
		UPDATE model_providers
		SET api_key = ?, base_url = ?, enabled = ?
		WHERE id = ?
	`, input.APIKey, input.BaseURL, input.Enabled, input.ID)
	if err != nil {
		log.Printf("更新模型提供商失败: %v", err)
		return fmt.Errorf("更新模型提供商失败: %v", err)
	}

	log.Printf("更新模型提供商成功: ID=%d", input.ID)
	return nil
}

// GetEnabledProviders 获取已启用的模型提供商
func (a *App) GetEnabledProviders() ([]ModelProvider, error) {
	if db == nil {
		return nil, fmt.Errorf("数据库未初始化")
	}

	rows, err := db.Query(`
		SELECT id, name, label, api_key, base_url, enabled, created_at
		FROM model_providers
		WHERE enabled = 1 AND api_key != ''
		ORDER BY id
	`)
	if err != nil {
		log.Printf("查询模型提供商失败: %v", err)
		return nil, fmt.Errorf("查询模型提供商失败: %v", err)
	}
	defer rows.Close()

	var providers []ModelProvider
	for rows.Next() {
		var p ModelProvider
		if err := rows.Scan(&p.ID, &p.Name, &p.Label, &p.APIKey, &p.BaseURL, &p.Enabled, &p.CreatedAt); err != nil {
			log.Printf("扫描模型提供商失败: %v", err)
			return nil, fmt.Errorf("扫描模型提供商失败: %v", err)
		}
		providers = append(providers, p)
	}

	return providers, nil
}
