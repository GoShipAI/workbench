package main

import (
	"fmt"
	"log"
)

// GetProjects 获取所有项目
func (a *App) GetProjects() ([]Project, error) {
	if db == nil {
		return nil, fmt.Errorf("数据库未初始化")
	}

	rows, err := db.Query(`
		SELECT p.id, p.name, p.description, p.color, p.created_at,
			   (SELECT COUNT(*) FROM tasks WHERE project_id = p.id) as task_count
		FROM projects p
		ORDER BY p.name
	`)
	if err != nil {
		log.Printf("查询项目失败: %v", err)
		return nil, fmt.Errorf("查询项目失败: %v", err)
	}
	defer rows.Close()

	var projects []Project
	for rows.Next() {
		var p Project
		if err := rows.Scan(&p.ID, &p.Name, &p.Description, &p.Color, &p.CreatedAt, &p.TaskCount); err != nil {
			log.Printf("扫描项目失败: %v", err)
			return nil, fmt.Errorf("扫描项目失败: %v", err)
		}
		projects = append(projects, p)
	}

	return projects, nil
}

// CreateProject 创建项目
func (a *App) CreateProject(name, description, color string) (*Project, error) {
	if db == nil {
		return nil, fmt.Errorf("数据库未初始化")
	}

	if name == "" {
		return nil, fmt.Errorf("项目名称不能为空")
	}

	if color == "" {
		color = "#165DFF"
	}

	result, err := db.Exec(`
		INSERT INTO projects (name, description, color)
		VALUES (?, ?, ?)
	`, name, description, color)
	if err != nil {
		log.Printf("创建项目失败: %v", err)
		return nil, fmt.Errorf("创建项目失败: %v", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("获取项目ID失败: %v", err)
	}

	// 查询创建的项目
	var p Project
	err = db.QueryRow(`
		SELECT id, name, description, color, created_at
		FROM projects WHERE id = ?
	`, id).Scan(&p.ID, &p.Name, &p.Description, &p.Color, &p.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("查询项目失败: %v", err)
	}

	log.Printf("创建项目成功: %s (ID: %d)", name, id)
	return &p, nil
}

// UpdateProject 更新项目
func (a *App) UpdateProject(id int64, name, description, color string) error {
	if db == nil {
		return fmt.Errorf("数据库未初始化")
	}

	if name == "" {
		return fmt.Errorf("项目名称不能为空")
	}

	_, err := db.Exec(`
		UPDATE projects
		SET name = ?, description = ?, color = ?
		WHERE id = ?
	`, name, description, color, id)
	if err != nil {
		log.Printf("更新项目失败: %v", err)
		return fmt.Errorf("更新项目失败: %v", err)
	}

	log.Printf("更新项目成功: ID=%d", id)
	return nil
}

// DeleteProject 删除项目
func (a *App) DeleteProject(id int64) error {
	if db == nil {
		return fmt.Errorf("数据库未初始化")
	}

	_, err := db.Exec(`DELETE FROM projects WHERE id = ?`, id)
	if err != nil {
		log.Printf("删除项目失败: %v", err)
		return fmt.Errorf("删除项目失败: %v", err)
	}

	log.Printf("删除项目成功: ID=%d", id)
	return nil
}
