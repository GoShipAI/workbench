package main

import (
	"fmt"
	"log"
	"math"
	"time"
)

// GetTasksByDate 根据日期获取任务
func (a *App) GetTasksByDate(date string) ([]Task, error) {
	if db == nil {
		return nil, fmt.Errorf("数据库未初始化")
	}

	rows, err := db.Query(`
		SELECT t.id, t.project_id, COALESCE(p.name, '') as project_name,
			   t.name, t.description, t.date, t.start_time, t.end_time,
			   t.hours, t.status, t.created_at
		FROM tasks t
		LEFT JOIN projects p ON t.project_id = p.id
		WHERE t.date = ?
		ORDER BY t.start_time, t.created_at
	`, date)
	if err != nil {
		log.Printf("查询任务失败: %v", err)
		return nil, fmt.Errorf("查询任务失败: %v", err)
	}
	defer rows.Close()

	return scanTasks(rows)
}

// GetPendingTasks 获取待处理任务（无日期）
func (a *App) GetPendingTasks() ([]Task, error) {
	if db == nil {
		return nil, fmt.Errorf("数据库未初始化")
	}

	rows, err := db.Query(`
		SELECT t.id, t.project_id, COALESCE(p.name, '') as project_name,
			   t.name, t.description, t.date, t.start_time, t.end_time,
			   t.hours, t.status, t.created_at
		FROM tasks t
		LEFT JOIN projects p ON t.project_id = p.id
		WHERE t.date IS NULL
		ORDER BY t.created_at DESC
	`)
	if err != nil {
		log.Printf("查询待处理任务失败: %v", err)
		return nil, fmt.Errorf("查询待处理任务失败: %v", err)
	}
	defer rows.Close()

	return scanTasks(rows)
}

// scanTasks 扫描任务结果集
func scanTasks(rows interface{ Next() bool; Scan(...interface{}) error }) ([]Task, error) {
	var tasks []Task
	for rows.Next() {
		var t Task
		if err := rows.Scan(&t.ID, &t.ProjectID, &t.ProjectName, &t.Name, &t.Description,
			&t.Date, &t.StartTime, &t.EndTime, &t.Hours, &t.Status, &t.CreatedAt); err != nil {
			log.Printf("扫描任务失败: %v", err)
			return nil, fmt.Errorf("扫描任务失败: %v", err)
		}
		tasks = append(tasks, t)
	}
	return tasks, nil
}

// CreateTask 创建任务
func (a *App) CreateTask(input TaskInput) (*Task, error) {
	if db == nil {
		return nil, fmt.Errorf("数据库未初始化")
	}

	if input.Name == "" {
		return nil, fmt.Errorf("任务名称不能为空")
	}

	// 确定状态
	status := input.Status
	if status == "" {
		if input.Date == nil || *input.Date == "" {
			status = TaskStatusPending
		} else {
			status = TaskStatusScheduled
		}
	}

	result, err := db.Exec(`
		INSERT INTO tasks (project_id, name, description, date, start_time, end_time, hours, status)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`, input.ProjectID, input.Name, input.Description, input.Date, input.StartTime, input.EndTime, input.Hours, status)
	if err != nil {
		log.Printf("创建任务失败: %v", err)
		return nil, fmt.Errorf("创建任务失败: %v", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("获取任务ID失败: %v", err)
	}

	// 查询创建的任务
	var t Task
	err = db.QueryRow(`
		SELECT t.id, t.project_id, COALESCE(p.name, '') as project_name,
			   t.name, t.description, t.date, t.start_time, t.end_time,
			   t.hours, t.status, t.created_at
		FROM tasks t
		LEFT JOIN projects p ON t.project_id = p.id
		WHERE t.id = ?
	`, id).Scan(&t.ID, &t.ProjectID, &t.ProjectName, &t.Name, &t.Description,
		&t.Date, &t.StartTime, &t.EndTime, &t.Hours, &t.Status, &t.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("查询任务失败: %v", err)
	}

	log.Printf("创建任务成功: %s (ID: %d)", input.Name, id)
	return &t, nil
}

// UpdateTask 更新任务
func (a *App) UpdateTask(input TaskInput) error {
	if db == nil {
		return fmt.Errorf("数据库未初始化")
	}

	if input.Name == "" {
		return fmt.Errorf("任务名称不能为空")
	}

	_, err := db.Exec(`
		UPDATE tasks
		SET project_id = ?, name = ?, description = ?, date = ?,
			start_time = ?, end_time = ?, hours = ?, status = ?
		WHERE id = ?
	`, input.ProjectID, input.Name, input.Description, input.Date,
		input.StartTime, input.EndTime, input.Hours, input.Status, input.ID)
	if err != nil {
		log.Printf("更新任务失败: %v", err)
		return fmt.Errorf("更新任务失败: %v", err)
	}

	log.Printf("更新任务成功: ID=%d", input.ID)
	return nil
}

// DeleteTask 删除任务
func (a *App) DeleteTask(id int64) error {
	if db == nil {
		return fmt.Errorf("数据库未初始化")
	}

	_, err := db.Exec(`DELETE FROM tasks WHERE id = ?`, id)
	if err != nil {
		log.Printf("删除任务失败: %v", err)
		return fmt.Errorf("删除任务失败: %v", err)
	}

	log.Printf("删除任务成功: ID=%d", id)
	return nil
}

// AssignTaskToDate 将任务分配到指定日期
func (a *App) AssignTaskToDate(taskID int64, date string) error {
	if db == nil {
		return fmt.Errorf("数据库未初始化")
	}

	status := TaskStatusScheduled
	_, err := db.Exec(`
		UPDATE tasks SET date = ?, status = ? WHERE id = ?
	`, date, status, taskID)
	if err != nil {
		log.Printf("分配任务日期失败: %v", err)
		return fmt.Errorf("分配任务日期失败: %v", err)
	}

	log.Printf("任务 %d 已分配到 %s", taskID, date)
	return nil
}

// UpdateTaskStatus 更新任务状态
func (a *App) UpdateTaskStatus(id int64, status string) error {
	if db == nil {
		return fmt.Errorf("数据库未初始化")
	}

	_, err := db.Exec(`UPDATE tasks SET status = ? WHERE id = ?`, status, id)
	if err != nil {
		log.Printf("更新任务状态失败: %v", err)
		return fmt.Errorf("更新任务状态失败: %v", err)
	}

	log.Printf("任务 %d 状态已更新为 %s", id, status)
	return nil
}

// CalculateHours 根据开始和结束时间计算工时
func (a *App) CalculateHours(startTime, endTime string) float64 {
	start, err1 := time.Parse("15:04", startTime)
	end, err2 := time.Parse("15:04", endTime)
	if err1 != nil || err2 != nil {
		return 0
	}

	duration := end.Sub(start)
	if duration < 0 {
		duration += 24 * time.Hour // 跨天处理
	}

	// 四舍五入到一位小数
	return math.Round(duration.Hours()*10) / 10
}

// GetWorkbenchData 获取工作台数据
func (a *App) GetWorkbenchData() (*WorkbenchData, error) {
	if db == nil {
		return nil, fmt.Errorf("数据库未初始化")
	}

	today := time.Now().Format("2006-01-02")

	// 获取今日任务
	todayTasks, err := a.GetTasksByDate(today)
	if err != nil {
		return nil, err
	}

	// 计算统计数据
	totalCount := len(todayTasks)
	completedCount := 0
	var plannedHours, completedHours float64

	for _, t := range todayTasks {
		plannedHours += t.Hours
		if t.Status == TaskStatusCompleted {
			completedCount++
			completedHours += t.Hours
		}
	}

	// 获取待处理任务数
	var pendingCount int
	err = db.QueryRow(`SELECT COUNT(*) FROM tasks WHERE date IS NULL`).Scan(&pendingCount)
	if err != nil {
		log.Printf("查询待处理任务数失败: %v", err)
	}

	return &WorkbenchData{
		TodayTasks:     todayTasks,
		TotalCount:     totalCount,
		CompletedCount: completedCount,
		PlannedHours:   plannedHours,
		CompletedHours: completedHours,
		PendingCount:   pendingCount,
	}, nil
}
