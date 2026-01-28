package main

import (
	"fmt"
	"log"
)

// GetProjectTimeStats 获取项目时间占比统计
func (a *App) GetProjectTimeStats(startDate, endDate string) ([]ProjectTimeStats, error) {
	if db == nil {
		return nil, fmt.Errorf("数据库未初始化")
	}

	rows, err := db.Query(`
		SELECT
			COALESCE(t.project_id, 0) as project_id,
			COALESCE(p.name, '未分类') as project_name,
			COALESCE(p.color, '#86909c') as color,
			SUM(CASE WHEN t.status = 'completed' AND t.actual_hours > 0
				THEN t.actual_hours ELSE t.hours END) as total_hours,
			COUNT(*) as task_count
		FROM tasks t
		LEFT JOIN projects p ON t.project_id = p.id
		WHERE t.date >= ? AND t.date <= ?
		GROUP BY COALESCE(t.project_id, 0)
		ORDER BY total_hours DESC
	`, startDate, endDate)
	if err != nil {
		log.Printf("查询项目时间统计失败: %v", err)
		return nil, fmt.Errorf("查询项目时间统计失败: %v", err)
	}
	defer rows.Close()

	var stats []ProjectTimeStats
	var totalHours float64

	for rows.Next() {
		var s ProjectTimeStats
		if err := rows.Scan(&s.ProjectID, &s.ProjectName, &s.Color,
			&s.TotalHours, &s.TaskCount); err != nil {
			return nil, fmt.Errorf("扫描统计数据失败: %v", err)
		}
		totalHours += s.TotalHours
		stats = append(stats, s)
	}

	// 计算百分比
	for i := range stats {
		if totalHours > 0 {
			stats[i].Percentage = (stats[i].TotalHours / totalHours) * 100
		}
	}

	return stats, nil
}

// GetDailyTaskStats 获取每日任务统计
func (a *App) GetDailyTaskStats(startDate, endDate string) ([]DailyTaskStats, error) {
	if db == nil {
		return nil, fmt.Errorf("数据库未初始化")
	}

	rows, err := db.Query(`
		SELECT
			date,
			COUNT(*) as total_count,
			SUM(CASE WHEN status = 'completed' THEN 1 ELSE 0 END) as completed_count
		FROM tasks
		WHERE date >= ? AND date <= ? AND date IS NOT NULL
		GROUP BY date
		ORDER BY date ASC
	`, startDate, endDate)
	if err != nil {
		log.Printf("查询每日任务统计失败: %v", err)
		return nil, fmt.Errorf("查询每日任务统计失败: %v", err)
	}
	defer rows.Close()

	var stats []DailyTaskStats
	for rows.Next() {
		var s DailyTaskStats
		if err := rows.Scan(&s.Date, &s.TotalCount, &s.CompletedCount); err != nil {
			return nil, fmt.Errorf("扫描统计数据失败: %v", err)
		}
		// 计算完成率
		if s.TotalCount > 0 {
			s.CompletionRate = float64(s.CompletedCount) / float64(s.TotalCount) * 100
		}
		stats = append(stats, s)
	}

	return stats, nil
}

// GetReportData 获取完整报表数据
func (a *App) GetReportData(startDate, endDate string) (*ReportData, error) {
	projectStats, err := a.GetProjectTimeStats(startDate, endDate)
	if err != nil {
		return nil, err
	}

	dailyStats, err := a.GetDailyTaskStats(startDate, endDate)
	if err != nil {
		return nil, err
	}

	// 计算汇总数据
	var summary ReportSummary
	var totalRate float64

	for _, d := range dailyStats {
		summary.TotalTasks += d.TotalCount
		summary.CompletedTasks += d.CompletedCount
		totalRate += d.CompletionRate
	}

	for _, p := range projectStats {
		summary.TotalHours += p.TotalHours
	}

	// 查询已完成工时
	if db != nil {
		db.QueryRow(`
			SELECT COALESCE(SUM(CASE WHEN actual_hours > 0 THEN actual_hours ELSE hours END), 0)
			FROM tasks
			WHERE date >= ? AND date <= ? AND status = 'completed'
		`, startDate, endDate).Scan(&summary.CompletedHours)
	}

	if len(dailyStats) > 0 {
		summary.AverageRate = totalRate / float64(len(dailyStats))
	}

	return &ReportData{
		ProjectStats: projectStats,
		DailyStats:   dailyStats,
		Summary:      summary,
	}, nil
}
