<script lang="ts" setup>
import { ref, computed, onMounted } from 'vue'
import { GetWorkbenchData, UpdateTaskStatus } from '../../wailsjs/go/main/App'
import { main } from '../../wailsjs/go/models'
import { Message } from '@arco-design/web-vue'

const loading = ref(false)
const data = ref<main.WorkbenchData>(new main.WorkbenchData({
  today_tasks: [],
  total_count: 0,
  completed_count: 0,
  planned_hours: 0,
  completed_hours: 0,
  pending_count: 0
}))

const completionRate = computed(() => {
  if (data.value.total_count === 0) return 0
  return Math.round((data.value.completed_count / data.value.total_count) * 100)
})

const hoursRate = computed(() => {
  if (data.value.planned_hours === 0) return 0
  return Math.round((data.value.completed_hours / data.value.planned_hours) * 100)
})

const loadData = async () => {
  loading.value = true
  try {
    const result = await GetWorkbenchData()
    if (result) {
      data.value = result
    }
  } catch (err) {
    console.error('加载工作台数据失败:', err)
    Message.error('加载数据失败')
  } finally {
    loading.value = false
  }
}

const completeTask = async (task: main.Task) => {
  try {
    await UpdateTaskStatus(task.id, 'completed')
    Message.success('任务已完成')
    await loadData()
  } catch (err) {
    console.error('更新任务状态失败:', err)
    Message.error('操作失败')
  }
}

const getStatusColor = (status: string) => {
  switch (status) {
    case 'completed': return '#00B42A'
    case 'in_progress': return '#165DFF'
    case 'scheduled': return '#FF7D00'
    default: return '#86909c'
  }
}

const getStatusText = (status: string) => {
  switch (status) {
    case 'completed': return '已完成'
    case 'in_progress': return '进行中'
    case 'scheduled': return '已安排'
    default: return '待处理'
  }
}

onMounted(() => {
  loadData()
})
</script>

<template>
  <div class="workbench">
    <a-spin :loading="loading">
      <!-- 统计卡片 -->
      <a-row :gutter="16" class="stats-row">
        <a-col :span="6">
          <a-card class="stat-card">
            <a-statistic title="今日任务" :value="data.total_count">
              <template #suffix>个</template>
            </a-statistic>
          </a-card>
        </a-col>
        <a-col :span="6">
          <a-card class="stat-card">
            <a-statistic title="已完成" :value="data.completed_count">
              <template #suffix>个</template>
            </a-statistic>
          </a-card>
        </a-col>
        <a-col :span="6">
          <a-card class="stat-card">
            <a-statistic title="计划工时" :value="data.planned_hours" :precision="1">
              <template #suffix>小时</template>
            </a-statistic>
          </a-card>
        </a-col>
        <a-col :span="6">
          <a-card class="stat-card">
            <a-statistic title="完成工时" :value="data.completed_hours" :precision="1">
              <template #suffix>小时</template>
            </a-statistic>
          </a-card>
        </a-col>
      </a-row>

      <!-- 进度条 -->
      <a-row :gutter="16" class="progress-row">
        <a-col :span="12">
          <a-card class="progress-card">
            <div class="progress-title">任务完成进度</div>
            <a-progress :percent="completionRate" :stroke-width="12" />
          </a-card>
        </a-col>
        <a-col :span="12">
          <a-card class="progress-card">
            <div class="progress-title">工时完成进度</div>
            <a-progress :percent="hoursRate" :stroke-width="12" color="#00B42A" />
          </a-card>
        </a-col>
      </a-row>

      <!-- 待处理提醒 -->
      <a-alert v-if="data.pending_count > 0" type="warning" class="pending-alert">
        您有 {{ data.pending_count }} 个待处理任务尚未安排日期
      </a-alert>

      <!-- 今日任务列表 -->
      <a-card title="今日任务" class="tasks-card">
        <template #extra>
          <a-button type="text" @click="loadData">
            <template #icon><icon-refresh /></template>
            刷新
          </a-button>
        </template>

        <a-empty v-if="!data.today_tasks || data.today_tasks.length === 0" description="今日暂无任务" />

        <a-list v-else :bordered="false">
          <a-list-item v-for="task in data.today_tasks" :key="task.id">
            <a-list-item-meta>
              <template #title>
                <span :class="{ 'completed-task': task.status === 'completed' }">
                  {{ task.name }}
                </span>
                <a-tag v-if="task.project_name" size="small" class="project-tag">
                  {{ task.project_name }}
                </a-tag>
              </template>
              <template #description>
                <span v-if="task.start_time">{{ task.start_time }}</span>
                <span v-if="task.start_time && task.end_time"> - {{ task.end_time }}</span>
                <span v-if="task.hours" class="hours-text">
                  {{ task.hours }}小时
                </span>
              </template>
            </a-list-item-meta>
            <template #actions>
              <a-tag :color="getStatusColor(task.status)" size="small">
                {{ getStatusText(task.status) }}
              </a-tag>
              <a-button
                v-if="task.status !== 'completed'"
                type="primary"
                size="small"
                @click="completeTask(task)"
              >
                完成
              </a-button>
            </template>
          </a-list-item>
        </a-list>
      </a-card>
    </a-spin>
  </div>
</template>

<style scoped>
.workbench {
  padding: 0;
}

.stats-row {
  margin-bottom: 16px;
}

.stat-card {
  background: #2a2a2b;
}

.progress-row {
  margin-bottom: 16px;
}

.progress-card {
  background: #2a2a2b;
}

.progress-title {
  margin-bottom: 12px;
  color: #86909c;
  font-size: 14px;
}

.pending-alert {
  margin-bottom: 16px;
}

.tasks-card {
  background: #2a2a2b;
}

.completed-task {
  text-decoration: line-through;
  color: #86909c;
}

.project-tag {
  margin-left: 8px;
}

.hours-text {
  margin-left: 12px;
  color: #165DFF;
}
</style>
