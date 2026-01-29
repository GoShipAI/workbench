<script lang="ts" setup>
import { ref, computed, watch, onMounted, defineProps } from 'vue'

const props = defineProps<{
  active: boolean
}>()
import {
  GetTasksByDate,
  GetTasksByDateRange,
  GetProjects,
  CreateTask,
  UpdateTask,
  DeleteTask,
  UpdateTaskStatus,
  CompleteTask,
  CalculateHours
} from '../../wailsjs/go/main/App'
import { main } from '../../wailsjs/go/models'
import { Message } from '@arco-design/web-vue'
import dayjs from 'dayjs'
import TaskAIChat from './TaskAIChat.vue'

interface TaskForm {
  id: number
  project_id?: number
  name: string
  description: string
  date: string
  start_time: string
  end_time: string
  hours: number
  deadline: string
  priority: string
  urgency: string
  status: string
  hours_mode: 'direct' | 'calculate'
}

interface CompleteForm {
  id: number
  actual_start: string
  actual_hours: number
}

const loading = ref(false)
const dateRange = ref<string[]>([dayjs().format('YYYY-MM-DD'), dayjs().format('YYYY-MM-DD')])
const tasks = ref<main.Task[]>([])
const projects = ref<main.Project[]>([])
const modalVisible = ref(false)
const completeModalVisible = ref(false)
const isEditing = ref(false)

// 筛选条件
const filterProjectId = ref<number | undefined>(undefined)
const filterStatus = ref<string | undefined>(undefined)

const defaultForm = (): TaskForm => ({
  id: 0,
  project_id: undefined,
  name: '',
  description: '',
  date: dateRange.value[0],
  start_time: '',
  end_time: '',
  hours: 0,
  deadline: '',
  priority: 'medium',
  urgency: 'medium',
  status: 'scheduled',
  hours_mode: 'direct'
})

const form = ref<TaskForm>(defaultForm())

const completeForm = ref<CompleteForm>({
  id: 0,
  actual_start: '',
  actual_hours: 0
})

const currentTask = ref<main.Task | null>(null)

// AI会话相关
const aiChatVisible = ref(false)
const aiChatTask = ref<main.Task | null>(null)

const openAIChat = (task: main.Task) => {
  aiChatTask.value = task
  aiChatVisible.value = true
}

// 筛选后的任务列表
const filteredTasks = computed(() => {
  let result = tasks.value
  if (filterProjectId.value) {
    result = result.filter(t => t.project_id === filterProjectId.value)
  }
  if (filterStatus.value) {
    result = result.filter(t => t.status === filterStatus.value)
  }
  return result
})

// 判断是否为单日模式
const isSingleDay = computed(() => {
  return dateRange.value[0] === dateRange.value[1]
})

const columns = computed(() => {
  const baseColumns: any[] = [
    { title: '任务名称', dataIndex: 'name', ellipsis: true, slotName: 'name' },
    { title: '项目', dataIndex: 'project_name', width: 100 },
  ]

  // 日期范围模式显示日期列
  if (!isSingleDay.value) {
    baseColumns.push({ title: '日期', dataIndex: 'date', width: 100 })
  }

  baseColumns.push(
    { title: '开始', dataIndex: 'start_time', width: 70 },
    { title: '工时', dataIndex: 'hours', width: 60 },
    { title: '截止', dataIndex: 'deadline', width: 100, slotName: 'deadline' },
    { title: '状态', dataIndex: 'status', width: 90, slotName: 'status' },
    { title: '操作', slotName: 'actions', width: 180 }
  )

  return baseColumns
})

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

const statusOptions = [
  { value: 'scheduled', label: '已安排' },
  { value: 'in_progress', label: '进行中' },
  { value: 'completed', label: '已完成' }
]

const loadTasks = async () => {
  loading.value = true
  try {
    let result: main.Task[] | null
    if (isSingleDay.value) {
      result = await GetTasksByDate(dateRange.value[0])
    } else {
      result = await GetTasksByDateRange(dateRange.value[0], dateRange.value[1])
    }
    tasks.value = result || []
  } catch (err) {
    console.error('加载任务失败:', err)
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
  form.value.date = dateRange.value[0]
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
    date: task.date || dateRange.value[0],
    start_time: task.start_time || '',
    end_time: task.end_time || '',
    hours: task.hours,
    deadline: task.deadline || '',
    priority: task.priority || 'medium',
    urgency: task.urgency || 'medium',
    status: task.status,
    hours_mode: 'direct'
  }
  await loadProjects()
  modalVisible.value = true
}

const openCompleteModal = (task: main.Task) => {
  currentTask.value = task
  completeForm.value = {
    id: task.id,
    actual_start: task.start_time || dayjs().format('HH:mm'),
    actual_hours: task.hours || 0
  }
  completeModalVisible.value = true
}

const calculateTaskHours = async () => {
  if (form.value.start_time && form.value.end_time) {
    try {
      const hours = await CalculateHours(form.value.start_time, form.value.end_time)
      form.value.hours = hours
    } catch (err) {
      console.error('计算工时失败:', err)
    }
  }
}

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
      date: form.value.date || undefined,
      start_time: form.value.start_time || undefined,
      end_time: form.value.end_time || undefined,
      hours: form.value.hours,
      deadline: form.value.deadline || undefined,
      priority: form.value.priority,
      urgency: form.value.urgency,
      status: form.value.status
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

const handleComplete = async () => {
  try {
    const input: main.CompleteTaskInput = {
      id: completeForm.value.id,
      actual_start: completeForm.value.actual_start || undefined,
      actual_hours: completeForm.value.actual_hours
    }
    await CompleteTask(input)
    Message.success('任务已完成')
    completeModalVisible.value = false
    await loadTasks()
  } catch (err) {
    console.error('完成任务失败:', err)
    Message.error('操作失败')
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

const handleStatusChange = async (task: main.Task, status: string) => {
  if (status === 'completed') {
    openCompleteModal(task)
  } else {
    try {
      await UpdateTaskStatus(task.id, status)
      Message.success('状态已更新')
      await loadTasks()
    } catch (err) {
      console.error('更新状态失败:', err)
      Message.error('更新失败')
    }
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
    default: return '待办'
  }
}

const getPriorityColor = (priority: string) => {
  return priorityOptions.find(p => p.value === priority)?.color || '#86909c'
}

const getUrgencyColor = (urgency: string) => {
  return urgencyOptions.find(u => u.value === urgency)?.color || '#86909c'
}

const isOverdue = (task: main.Task) => {
  if (!task.deadline || task.status === 'completed') return false
  return dayjs(task.deadline).isBefore(dayjs(), 'day')
}

watch(dateRange, () => {
  loadTasks()
})

// 当标签页激活时重新加载数据
watch(() => props.active, (isActive) => {
  if (isActive) {
    loadTasks()
  }
})

watch(() => form.value.hours_mode, () => {
  if (form.value.hours_mode === 'calculate') {
    calculateTaskHours()
  }
})

watch([() => form.value.start_time, () => form.value.end_time], () => {
  if (form.value.hours_mode === 'calculate') {
    calculateTaskHours()
  }
})

onMounted(() => {
  loadTasks()
  loadProjects()
})
</script>

<template>
  <div class="task-management">
    <!-- 工具栏 -->
    <div class="toolbar">
      <div class="toolbar-left">
        <a-range-picker
          v-model="dateRange"
          style="width: 260px"
          :allow-clear="false"
        />
        <a-select
          v-model="filterProjectId"
          placeholder="按项目筛选"
          allow-clear
          style="width: 140px"
        >
          <a-option v-for="p in projects" :key="p.id" :value="p.id">
            {{ p.name }}
          </a-option>
        </a-select>
        <a-select
          v-model="filterStatus"
          placeholder="按状态筛选"
          allow-clear
          style="width: 120px"
        >
          <a-option v-for="s in statusOptions" :key="s.value" :value="s.value">
            {{ s.label }}
          </a-option>
        </a-select>
        <span class="task-count">共 {{ filteredTasks.length }} 条任务</span>
      </div>
      <a-button type="primary" @click="openCreateModal">
        <template #icon><icon-plus /></template>
        新建任务
      </a-button>
    </div>

    <!-- 任务表格 -->
    <a-table
      :loading="loading"
      :columns="columns"
      :data="filteredTasks"
      :pagination="false"
      row-key="id"
      class="tasks-table"
    >
      <template #name="{ record }">
        <div class="task-name-cell">
          <span class="task-name" :class="{ 'completed': record.status === 'completed' }">
            {{ record.name }}
          </span>
          <span class="task-tags">
            <a-tag v-if="record.priority === 'high'" size="small" :color="getPriorityColor(record.priority)">重要</a-tag>
            <a-tag v-if="record.urgency === 'high'" size="small" :color="getUrgencyColor(record.urgency)">紧急</a-tag>
          </span>
        </div>
      </template>
      <template #deadline="{ record }">
        <span v-if="record.deadline" :class="{ 'overdue': isOverdue(record) }">
          {{ record.deadline }}
          <icon-exclamation-circle-fill v-if="isOverdue(record)" class="overdue-icon" />
        </span>
      </template>
      <template #status="{ record }">
        <a-dropdown>
          <a-tag :color="getStatusColor(record.status)" style="cursor: pointer">
            {{ getStatusText(record.status) }}
          </a-tag>
          <template #content>
            <a-doption @click="handleStatusChange(record, 'scheduled')">已安排</a-doption>
            <a-doption @click="handleStatusChange(record, 'in_progress')">进行中</a-doption>
            <a-doption @click="handleStatusChange(record, 'completed')">已完成</a-doption>
          </template>
        </a-dropdown>
      </template>
      <template #actions="{ record }">
        <a-space>
          <a-button
            v-if="record.status !== 'completed'"
            type="primary"
            size="small"
            @click="openCompleteModal(record)"
          >
            完成
          </a-button>
          <a-button type="outline" size="small" @click="openAIChat(record)">
            <template #icon><icon-robot /></template>
            AI
          </a-button>
          <a-button type="text" size="small" @click="openEditModal(record)">
            编辑
          </a-button>
          <a-popconfirm content="确定删除此任务?" @ok="handleDelete(record)">
            <a-button type="text" size="small" status="danger">
              删除
            </a-button>
          </a-popconfirm>
        </a-space>
      </template>
    </a-table>

    <!-- 新建/编辑弹窗 -->
    <a-modal
      v-model:visible="modalVisible"
      :title="isEditing ? '编辑任务' : '新建任务'"
      @ok="handleSubmit"
      @cancel="modalVisible = false"
      :width="520"
    >
      <a-form :model="form" layout="vertical">
        <a-form-item label="任务名称" required>
          <a-input v-model="form.name" placeholder="请输入任务名称" />
        </a-form-item>

        <a-row :gutter="16">
          <a-col :span="12">
            <a-form-item label="所属项目">
              <a-select v-model="form.project_id" placeholder="选择项目" allow-clear>
                <a-option v-for="p in projects" :key="p.id" :value="p.id">
                  {{ p.name }}
                </a-option>
              </a-select>
            </a-form-item>
          </a-col>
          <a-col :span="12">
            <a-form-item label="计划日期">
              <a-date-picker v-model="form.date" style="width: 100%" />
            </a-form-item>
          </a-col>
        </a-row>

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

        <a-form-item label="截止日期">
          <a-date-picker v-model="form.deadline" style="width: 100%" placeholder="选择截止日期（可选）" allow-clear />
        </a-form-item>

        <a-form-item label="工时录入方式">
          <a-radio-group v-model="form.hours_mode">
            <a-radio value="direct">直接填写工时</a-radio>
            <a-radio value="calculate">通过时间计算</a-radio>
          </a-radio-group>
        </a-form-item>

        <a-row :gutter="16">
          <a-col :span="8">
            <a-form-item label="开始时间">
              <a-time-picker v-model="form.start_time" format="HH:mm" style="width: 100%" />
            </a-form-item>
          </a-col>
          <a-col :span="8">
            <a-form-item v-if="form.hours_mode === 'calculate'" label="结束时间">
              <a-time-picker v-model="form.end_time" format="HH:mm" style="width: 100%" />
            </a-form-item>
            <a-form-item v-else label="预计工时">
              <a-input-number v-model="form.hours" :min="0" :max="24" :precision="1" :step="0.5" style="width: 100%" />
            </a-form-item>
          </a-col>
          <a-col :span="8" v-if="form.hours_mode === 'calculate'">
            <a-form-item label="计算工时">
              <a-input-number v-model="form.hours" :min="0" :max="24" :precision="1" disabled style="width: 100%" />
            </a-form-item>
          </a-col>
        </a-row>

        <a-form-item label="任务描述">
          <a-textarea v-model="form.description" placeholder="请输入任务描述（可选）" :auto-size="{ minRows: 2, maxRows: 4 }" />
        </a-form-item>

        <a-form-item label="状态">
          <a-select v-model="form.status">
            <a-option value="scheduled">已安排</a-option>
            <a-option value="in_progress">进行中</a-option>
            <a-option value="completed">已完成</a-option>
          </a-select>
        </a-form-item>
      </a-form>
    </a-modal>

    <!-- 完成任务弹窗 -->
    <a-modal
      v-model:visible="completeModalVisible"
      title="完成任务"
      @ok="handleComplete"
      @cancel="completeModalVisible = false"
    >
      <a-alert v-if="currentTask" type="info" style="margin-bottom: 16px">
        确认完成任务「{{ currentTask.name }}」
      </a-alert>
      <a-form :model="completeForm" layout="vertical">
        <a-form-item label="实际开始时间">
          <a-time-picker v-model="completeForm.actual_start" format="HH:mm" style="width: 100%" />
        </a-form-item>
        <a-form-item label="实际工时（小时）" required>
          <a-input-number
            v-model="completeForm.actual_hours"
            :min="0"
            :max="24"
            :precision="1"
            :step="0.5"
            style="width: 100%"
          />
        </a-form-item>
      </a-form>
    </a-modal>

    <!-- AI会话弹窗 -->
    <TaskAIChat
      v-model:visible="aiChatVisible"
      :task="aiChatTask"
    />
  </div>
</template>

<style scoped>
.task-management {
  padding: 0;
}

.toolbar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
  flex-wrap: wrap;
  gap: 12px;
}

.toolbar-left {
  display: flex;
  align-items: center;
  gap: 12px;
  flex-wrap: wrap;
}

.task-count {
  color: #86909c;
  font-size: 14px;
}

.tasks-table {
  background: #2a2a2b;
}

.task-name-cell {
  display: flex;
  align-items: center;
  gap: 8px;
}

.task-name.completed {
  text-decoration: line-through;
  color: #86909c;
}

.task-tags {
  display: flex;
  gap: 4px;
}

.overdue {
  color: #F53F3F;
}

.overdue-icon {
  margin-left: 4px;
  color: #F53F3F;
}

:deep(.arco-table-th) {
  background: #232324;
}

:deep(.arco-table-tr) {
  background: #2a2a2b;
}

:deep(.arco-table-tr:hover) {
  background: #333;
}
</style>
