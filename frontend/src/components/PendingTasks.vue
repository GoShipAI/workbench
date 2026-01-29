<script lang="ts" setup>
import { ref, computed, onMounted, watch, defineProps } from 'vue'

const props = defineProps<{
  active: boolean
}>()
import {
  GetPendingTasks,
  GetProjects,
  CreateTask,
  UpdateTask,
  DeleteTask,
  AssignTaskToDate,
  GetTasksByDate
} from '../../wailsjs/go/main/App'
import { main } from '../../wailsjs/go/models'
import { Message } from '@arco-design/web-vue'
import dayjs from 'dayjs'

interface TaskForm {
  id: number
  project_id?: number
  name: string
  description: string
  hours: number
  deadline: string
  priority: string
  urgency: string
}

const loading = ref(false)
const tasks = ref<main.Task[]>([])
const projects = ref<main.Project[]>([])
const modalVisible = ref(false)
const assignModalVisible = ref(false)
const isEditing = ref(false)
const selectedTask = ref<main.Task | null>(null)
const assignDate = ref(dayjs().format('YYYY-MM-DD'))
const dateTasks = ref<main.Task[]>([])
const loadingDateTasks = ref(false)

// 筛选条件
const filterProjectId = ref<number | undefined>(undefined)

const defaultForm = (): TaskForm => ({
  id: 0,
  project_id: undefined,
  name: '',
  description: '',
  hours: 0,
  deadline: '',
  priority: 'medium',
  urgency: 'medium'
})

const form = ref<TaskForm>(defaultForm())

const priorityOptions = [
  { value: 'high', label: '高', color: '#F53F3F' },
  { value: 'medium', label: '中', color: '#FF7D00' },
  { value: 'low', label: '低', color: '#86909c' }
]

const urgencyOptions = [
  { value: 'high', label: '高', color: '#F53F3F' },
  { value: 'medium', label: '中', color: '#FF7D00' },
  { value: 'low', label: '低', color: '#86909c' }
]

// 筛选后的任务列表
const filteredTasks = computed(() => {
  let result = tasks.value
  if (filterProjectId.value) {
    result = result.filter(t => t.project_id === filterProjectId.value)
  }
  return result
})

// 选定日期的工时统计
const dateTasksStats = computed(() => {
  const tasks = dateTasks.value
  const totalHours = tasks.reduce((sum, t) => sum + (t.hours || 0), 0)
  const completedCount = tasks.filter(t => t.status === 'completed').length
  return {
    count: tasks.length,
    totalHours,
    completedCount
  }
})

const loadTasks = async () => {
  loading.value = true
  try {
    const result = await GetPendingTasks()
    tasks.value = result || []
  } catch (err) {
    console.error('加载待办任务失败:', err)
    Message.error('加载任务失败')
  } finally {
    loading.value = false
  }
}

const loadProjects = async () => {
  try {
    const result = await GetProjects()
    projects.value = result || []
  } catch (err) {
    console.error('加载项目失败:', err)
  }
}

const openCreateModal = async () => {
  isEditing.value = false
  form.value = defaultForm()
  await loadProjects()
  modalVisible.value = true
}

const openEditModal = async (task: main.Task) => {
  isEditing.value = true
  form.value = {
    id: task.id,
    project_id: task.project_id,
    name: task.name,
    description: task.description,
    hours: task.hours,
    deadline: task.deadline || '',
    priority: task.priority || 'medium',
    urgency: task.urgency || 'medium'
  }
  await loadProjects()
  modalVisible.value = true
}

const loadDateTasks = async (date: string) => {
  loadingDateTasks.value = true
  try {
    const result = await GetTasksByDate(date)
    dateTasks.value = result || []
  } catch (err) {
    console.error('加载日期任务失败:', err)
    dateTasks.value = []
  } finally {
    loadingDateTasks.value = false
  }
}

const openAssignModal = async (task: main.Task) => {
  selectedTask.value = task
  assignDate.value = dayjs().format('YYYY-MM-DD')
  assignModalVisible.value = true
  await loadDateTasks(assignDate.value)
}

// 监听分配日期变化，加载该日期的任务
watch(assignDate, async (newDate) => {
  if (assignModalVisible.value && newDate) {
    await loadDateTasks(newDate)
  }
})

const handleSubmit = async () => {
  if (!form.value.name.trim()) {
    Message.warning('请输入任务名称')
    return
  }

  try {
    const input: main.TaskInput = {
      id: form.value.id,
      project_id: form.value.project_id,
      name: form.value.name,
      description: form.value.description,
      date: undefined,
      start_time: undefined,
      end_time: undefined,
      hours: form.value.hours,
      deadline: form.value.deadline || undefined,
      priority: form.value.priority,
      urgency: form.value.urgency,
      status: 'pending'
    }

    if (isEditing.value) {
      await UpdateTask(input)
      Message.success('任务已更新')
    } else {
      await CreateTask(input)
      Message.success('任务已创建')
    }

    modalVisible.value = false
    await loadTasks()
  } catch (err) {
    console.error('保存任务失败:', err)
    Message.error('保存失败')
  }
}

const handleAssign = async () => {
  if (!selectedTask.value) return

  try {
    await AssignTaskToDate(selectedTask.value.id, assignDate.value)
    Message.success(`任务已分配到 ${assignDate.value}`)
    assignModalVisible.value = false
    await loadTasks()
  } catch (err) {
    console.error('分配任务失败:', err)
    Message.error('分配失败')
  }
}

const handleDelete = async (task: main.Task) => {
  try {
    await DeleteTask(task.id)
    Message.success('任务已删除')
    await loadTasks()
  } catch (err) {
    console.error('删除任务失败:', err)
    Message.error('删除失败')
  }
}

const getPriorityColor = (priority: string) => {
  return priorityOptions.find(p => p.value === priority)?.color || '#86909c'
}

const getUrgencyColor = (urgency: string) => {
  return urgencyOptions.find(u => u.value === urgency)?.color || '#86909c'
}

const isOverdue = (task: main.Task) => {
  if (!task.deadline) return false
  return dayjs(task.deadline).isBefore(dayjs(), 'day')
}

// 当标签页激活时重新加载数据
watch(() => props.active, (isActive) => {
  if (isActive) {
    loadTasks()
  }
})

onMounted(() => {
  loadTasks()
  loadProjects()
})
</script>

<template>
  <div class="pending-tasks">
    <!-- 工具栏 -->
    <div class="toolbar">
      <div class="toolbar-left">
        <div class="title">
          待办任务
          <a-badge :count="filteredTasks.length" :max-count="99" />
        </div>
        <a-select
          v-model="filterProjectId"
          placeholder="按项目筛选"
          allow-clear
          style="width: 160px"
        >
          <a-option v-for="p in projects" :key="p.id" :value="p.id">
            {{ p.name }}
          </a-option>
        </a-select>
      </div>
      <a-button type="primary" @click="openCreateModal">
        <template #icon><icon-plus /></template>
        新建任务
      </a-button>
    </div>

    <!-- 任务列表 -->
    <a-spin :loading="loading">
      <a-empty v-if="filteredTasks.length === 0" description="暂无待办任务" />

      <div v-else class="task-list">
        <div v-for="task in filteredTasks" :key="task.id" class="task-card">
          <div class="task-content">
            <div class="task-header">
              <span class="task-name">{{ task.name }}</span>
              <span class="task-tags">
                <a-tag v-if="task.project_name" size="small">{{ task.project_name }}</a-tag>
                <a-tag v-if="task.priority === 'high'" size="small" :color="getPriorityColor(task.priority)">重要</a-tag>
                <a-tag v-if="task.urgency === 'high'" size="small" :color="getUrgencyColor(task.urgency)">紧急</a-tag>
              </span>
            </div>
            <div class="task-meta">
              <span v-if="task.description" class="task-desc">{{ task.description }}</span>
              <span v-if="task.hours" class="meta-item">预计 {{ task.hours }}h</span>
              <span v-if="task.deadline" class="meta-item" :class="{ 'overdue': isOverdue(task) }">
                截止: {{ task.deadline }}
                <icon-exclamation-circle-fill v-if="isOverdue(task)" class="overdue-icon" />
              </span>
            </div>
          </div>
          <div class="task-actions">
            <a-button type="primary" size="small" @click="openAssignModal(task)">
              分配日期
            </a-button>
            <a-button type="text" size="small" @click="openEditModal(task)">
              编辑
            </a-button>
            <a-popconfirm content="确定删除此任务?" @ok="handleDelete(task)">
              <a-button type="text" size="small" status="danger">
                删除
              </a-button>
            </a-popconfirm>
          </div>
        </div>
      </div>
    </a-spin>

    <!-- 新建/编辑弹窗 -->
    <a-modal
      v-model:visible="modalVisible"
      :title="isEditing ? '编辑任务' : '新建待办任务'"
      @ok="handleSubmit"
      @cancel="modalVisible = false"
      :width="480"
    >
      <a-form :model="form" layout="vertical">
        <a-form-item label="任务名称" required>
          <a-input v-model="form.name" placeholder="请输入任务名称" />
        </a-form-item>

        <a-form-item label="所属项目">
          <a-select v-model="form.project_id" placeholder="选择项目（可选）" allow-clear>
            <a-option v-for="p in projects" :key="p.id" :value="p.id">
              {{ p.name }}
            </a-option>
          </a-select>
        </a-form-item>

        <a-row :gutter="16">
          <a-col :span="12">
            <a-form-item label="重要程度">
              <a-radio-group v-model="form.priority" type="button">
                <a-radio v-for="p in priorityOptions" :key="p.value" :value="p.value">
                  <span :style="{ color: form.priority === p.value ? '#fff' : p.color }">{{ p.label }}</span>
                </a-radio>
              </a-radio-group>
            </a-form-item>
          </a-col>
          <a-col :span="12">
            <a-form-item label="紧急程度">
              <a-radio-group v-model="form.urgency" type="button">
                <a-radio v-for="u in urgencyOptions" :key="u.value" :value="u.value">
                  <span :style="{ color: form.urgency === u.value ? '#fff' : u.color }">{{ u.label }}</span>
                </a-radio>
              </a-radio-group>
            </a-form-item>
          </a-col>
        </a-row>

        <a-row :gutter="16">
          <a-col :span="12">
            <a-form-item label="截止日期">
              <a-date-picker v-model="form.deadline" style="width: 100%" placeholder="可选" allow-clear />
            </a-form-item>
          </a-col>
          <a-col :span="12">
            <a-form-item label="预计工时（小时）">
              <a-input-number
                v-model="form.hours"
                :min="0"
                :max="100"
                :precision="1"
                :step="0.5"
                style="width: 100%"
              />
            </a-form-item>
          </a-col>
        </a-row>

        <a-form-item label="任务描述">
          <a-textarea
            v-model="form.description"
            placeholder="请输入任务描述（可选）"
            :auto-size="{ minRows: 2, maxRows: 4 }"
          />
        </a-form-item>
      </a-form>
    </a-modal>

    <!-- 分配日期弹窗 -->
    <a-modal
      v-model:visible="assignModalVisible"
      title="规划日程"
      @ok="handleAssign"
      @cancel="assignModalVisible = false"
      :width="640"
    >
      <div class="assign-modal-content">
        <!-- 当前任务信息 -->
        <a-alert v-if="selectedTask" type="info" class="current-task-alert">
          <template #title>待规划任务</template>
          <div class="current-task-info">
            <span class="task-name">{{ selectedTask.name }}</span>
            <span v-if="selectedTask.hours" class="task-hours">预计 {{ selectedTask.hours }}h</span>
          </div>
        </a-alert>

        <!-- 日期选择 -->
        <div class="date-picker-row">
          <span class="date-label">选择日期</span>
          <a-date-picker
            v-model="assignDate"
            style="width: 200px"
            :allow-clear="false"
          />
        </div>

        <!-- 当天已安排任务 -->
        <div class="date-tasks-section">
          <div class="section-header">
            <span class="section-title">{{ assignDate }} 已安排任务</span>
            <span class="section-stats">
              {{ dateTasksStats.count }} 个任务，共 {{ dateTasksStats.totalHours.toFixed(1) }}h
            </span>
          </div>

          <a-spin :loading="loadingDateTasks">
            <div v-if="dateTasks.length === 0" class="empty-date-tasks">
              该日期暂无任务安排
            </div>
            <div v-else class="date-tasks-list">
              <div v-for="task in dateTasks" :key="task.id" class="date-task-item">
                <div class="date-task-main">
                  <a-tag v-if="task.status === 'completed'" size="small" color="green">已完成</a-tag>
                  <a-tag v-else-if="task.status === 'in_progress'" size="small" color="blue">进行中</a-tag>
                  <span class="date-task-name" :class="{ completed: task.status === 'completed' }">
                    {{ task.name }}
                  </span>
                </div>
                <div class="date-task-meta">
                  <span v-if="task.start_time && task.end_time" class="time-range">
                    {{ task.start_time }} - {{ task.end_time }}
                  </span>
                  <span v-if="task.hours" class="hours">{{ task.hours }}h</span>
                  <a-tag v-if="task.project_name" size="small">{{ task.project_name }}</a-tag>
                </div>
              </div>
            </div>
          </a-spin>
        </div>
      </div>
    </a-modal>
  </div>
</template>

<style scoped>
.pending-tasks {
  padding: 0;
  width: 100%;
}

.pending-tasks :deep(.arco-spin) {
  width: 100%;
  display: block;
}

.toolbar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
}

.toolbar-left {
  display: flex;
  align-items: center;
  gap: 16px;
}

.title {
  font-size: 16px;
  font-weight: 500;
  display: flex;
  align-items: center;
  gap: 8px;
}

.task-list {
  display: flex;
  flex-direction: column;
  gap: 8px;
  width: 100%;
}

.task-card {
  display: flex;
  justify-content: space-between;
  align-items: center;
  background: #2a2a2b;
  border-radius: 8px;
  padding: 16px 20px;
  transition: background 0.2s;
  width: 100%;
  box-sizing: border-box;
}

.task-card:hover {
  background: #333;
}

.task-content {
  flex: 1;
  min-width: 0;
}

.task-header {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 8px;
  flex-wrap: wrap;
}

.task-name {
  font-weight: 500;
  font-size: 15px;
}

.task-tags {
  display: flex;
  gap: 4px;
  flex-wrap: wrap;
}

.task-meta {
  display: flex;
  align-items: center;
  gap: 16px;
  color: #86909c;
  font-size: 13px;
  flex-wrap: wrap;
}

.task-desc {
  max-width: 400px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.meta-item {
  white-space: nowrap;
}

.task-actions {
  display: flex;
  align-items: center;
  gap: 4px;
  flex-shrink: 0;
  margin-left: 16px;
}

.overdue {
  color: #F53F3F;
}

.overdue-icon {
  margin-left: 4px;
  color: #F53F3F;
}

/* 分配日期弹窗样式 */
.assign-modal-content {
  min-height: 300px;
  width: 100%;
}

.current-task-alert {
  margin-bottom: 16px;
}

.date-picker-row {
  display: flex;
  align-items: center;
  gap: 16px;
  margin-bottom: 16px;
}

.date-label {
  font-size: 14px;
  color: #c9cdd4;
}

.current-task-info {
  display: flex;
  align-items: center;
  gap: 12px;
}

.current-task-info .task-name {
  font-weight: 500;
}

.current-task-info .task-hours {
  color: #86909c;
  font-size: 13px;
}

.date-tasks-section {
  margin-top: 0;
  border: 1px solid #3a3a3c;
  border-radius: 8px;
  padding: 16px;
  background: #1e1e1f;
}

.date-tasks-section :deep(.arco-spin),
.date-tasks-section :deep(.arco-spin-children) {
  width: 100% !important;
  display: block !important;
}

.section-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 12px;
  padding-bottom: 8px;
  border-bottom: 1px solid #3a3a3c;
}

.section-title {
  font-weight: 500;
  font-size: 14px;
}

.section-stats {
  color: #86909c;
  font-size: 13px;
}

.empty-date-tasks {
  text-align: center;
  color: #86909c;
  padding: 24px 0;
}

.date-tasks-list {
  display: flex;
  flex-direction: column;
  gap: 8px;
  max-height: 280px;
  overflow-y: auto;
  width: 100%;
}

.date-task-item {
  display: flex;
  align-items: center;
  padding: 10px 14px;
  background: #2a2a2b;
  border-radius: 6px;
  width: 100%;
  box-sizing: border-box;
  gap: 12px;
}

.date-task-main {
  display: flex;
  align-items: center;
  gap: 10px;
  flex: 1;
  min-width: 0;
}

.date-task-name {
  font-size: 14px;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.date-task-name.completed {
  text-decoration: line-through;
  color: #86909c;
}

.date-task-meta {
  display: flex;
  align-items: center;
  gap: 8px;
  color: #86909c;
  font-size: 12px;
  margin-left: auto;
}

.date-task-meta .time-range {
  color: #4080ff;
}

.date-task-meta .hours {
  background: #3a3a3c;
  padding: 2px 6px;
  border-radius: 4px;
}
</style>
