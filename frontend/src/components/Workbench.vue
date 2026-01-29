<script lang="ts" setup>
import { ref, computed, onMounted, watch, defineProps } from 'vue'

const props = defineProps<{
  active: boolean
}>()
import { GetWorkbenchData, GetProjects, CreateTask, CompleteTask, GetPendingTasks, AssignTaskToDate, GetTasksByDate } from '../../wailsjs/go/main/App'
import { main } from '../../wailsjs/go/models'
import { Message } from '@arco-design/web-vue'
import dayjs from 'dayjs'
import TaskAIChat from './TaskAIChat.vue'

interface TaskForm {
  project_id?: number
  name: string
  description: string
  hours: number
}

interface CompleteForm {
  actual_start: string
  actual_hours: number
}

const loading = ref(false)
const data = ref<main.WorkbenchData>(new main.WorkbenchData({
  today_tasks: [],
  total_count: 0,
  completed_count: 0,
  planned_hours: 0,
  completed_hours: 0,
  pending_count: 0
}))

// 视图切换: today=今日任务, tomorrow=明日待办
const viewMode = ref<'today' | 'tomorrow'>('today')
const tomorrowTasks = ref<main.Task[]>([])

const projects = ref<main.Project[]>([])
const taskModalVisible = ref(false)
const todoModalVisible = ref(false)

const defaultForm = (): TaskForm => ({
  project_id: undefined,
  name: '',
  description: '',
  hours: 0
})

const taskForm = ref<TaskForm>(defaultForm())
const todoForm = ref<TaskForm>(defaultForm())

// 完成任务相关
const completeModalVisible = ref(false)
const selectedTask = ref<main.Task | null>(null)
const completeForm = ref<CompleteForm>({
  actual_start: '',
  actual_hours: 0
})

// 任务详情相关
const detailModalVisible = ref(false)
const viewTask = ref<main.Task | null>(null)

// AI会话相关
const aiChatVisible = ref(false)
const aiChatTask = ref<main.Task | null>(null)

const openAIChat = (task: main.Task) => {
  aiChatTask.value = task
  aiChatVisible.value = true
}

// 日程规划相关
const planModalVisible = ref(false)
const planDate = ref(dayjs().format('YYYY-MM-DD'))
const pendingTasks = ref<main.Task[]>([])
const selectedTaskIds = ref<number[]>([])
const planLoading = ref(false)
const newTaskName = ref('')
const newTaskHours = ref(0)
const planDateTasks = ref<main.Task[]>([])
const loadingPlanDateTasks = ref(false)

const openDetailModal = (task: main.Task) => {
  viewTask.value = task
  detailModalVisible.value = true
}

const getPriorityText = (priority: string) => {
  switch (priority) {
    case 'high': return '高'
    case 'medium': return '中'
    case 'low': return '低'
    default: return '中'
  }
}

const getPriorityColor = (priority: string) => {
  switch (priority) {
    case 'high': return '#F53F3F'
    case 'medium': return '#FF7D00'
    case 'low': return '#86909c'
    default: return '#FF7D00'
  }
}

const getUrgencyText = (urgency: string) => {
  switch (urgency) {
    case 'high': return '高'
    case 'medium': return '中'
    case 'low': return '低'
    default: return '中'
  }
}

const getUrgencyColor = (urgency: string) => {
  switch (urgency) {
    case 'high': return '#F53F3F'
    case 'medium': return '#FF7D00'
    case 'low': return '#86909c'
    default: return '#FF7D00'
  }
}

const completionRate = computed(() => {
  const total = data.value.total_count ?? 0
  const completed = data.value.completed_count ?? 0
  if (total === 0) return 0
  return Math.min(Math.round((completed / total) * 100), 100)
})

const hoursRate = computed(() => {
  const planned = data.value.planned_hours ?? 0
  const completed = data.value.completed_hours ?? 0
  if (planned === 0) return 0
  return Math.min(Math.round((completed / planned) * 100), 100)
})

const formatPercent = (percent: number) => `${percent}%`

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

const loadTomorrowTasks = async () => {
  loading.value = true
  try {
    const tomorrow = dayjs().add(1, 'day').format('YYYY-MM-DD')
    const result = await GetTasksByDate(tomorrow)
    tomorrowTasks.value = result || []
  } catch (err) {
    console.error('加载明日任务失败:', err)
    Message.error('加载明日任务失败')
  } finally {
    loading.value = false
  }
}

const switchView = (mode: 'today' | 'tomorrow') => {
  viewMode.value = mode
  if (mode === 'today') {
    loadData()
  } else {
    loadTomorrowTasks()
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

const openCompleteModal = (task: main.Task) => {
  selectedTask.value = task
  completeForm.value = {
    actual_start: task.start_time || dayjs().format('HH:mm'),
    actual_hours: task.hours || 0
  }
  completeModalVisible.value = true
}

const handleComplete = async () => {
  if (!selectedTask.value) return

  try {
    const input: main.CompleteTaskInput = {
      id: selectedTask.value.id,
      actual_start: completeForm.value.actual_start || undefined,
      actual_hours: completeForm.value.actual_hours
    }
    await CompleteTask(input)
    Message.success('任务已完成')
    completeModalVisible.value = false
    await loadData()
  } catch (err) {
    console.error('完成任务失败:', err)
    Message.error('操作失败')
  }
}

const openTaskModal = async () => {
  taskForm.value = defaultForm()
  await loadProjects()
  taskModalVisible.value = true
}

const openTodoModal = async () => {
  todoForm.value = defaultForm()
  await loadProjects()
  todoModalVisible.value = true
}

const handleTaskSubmit = async () => {
  if (!taskForm.value.name.trim()) {
    Message.warning('请输入任务名称')
    return
  }

  try {
    const input: main.TaskInput = {
      id: 0,
      project_id: taskForm.value.project_id,
      name: taskForm.value.name,
      description: taskForm.value.description,
      date: dayjs().format('YYYY-MM-DD'),
      start_time: undefined,
      end_time: undefined,
      hours: taskForm.value.hours,
      deadline: undefined,
      priority: 'medium',
      urgency: 'medium',
      status: 'scheduled'
    }

    await CreateTask(input)
    Message.success('任务已创建')
    taskModalVisible.value = false
    await loadData()
  } catch (err) {
    console.error('创建任务失败:', err)
    Message.error('创建失败')
  }
}

const handleTodoSubmit = async () => {
  if (!todoForm.value.name.trim()) {
    Message.warning('请输入任务名称')
    return
  }

  try {
    const input: main.TaskInput = {
      id: 0,
      project_id: todoForm.value.project_id,
      name: todoForm.value.name,
      description: todoForm.value.description,
      date: undefined,
      start_time: undefined,
      end_time: undefined,
      hours: todoForm.value.hours,
      deadline: undefined,
      priority: 'medium',
      urgency: 'medium',
      status: 'pending'
    }

    await CreateTask(input)
    Message.success('待办已创建')
    todoModalVisible.value = false
    await loadData()
  } catch (err) {
    console.error('创建待办失败:', err)
    Message.error('创建失败')
  }
}

// 加载指定日期已有任务
const loadPlanDateTasks = async (date: string) => {
  loadingPlanDateTasks.value = true
  try {
    const result = await GetTasksByDate(date)
    planDateTasks.value = result || []
  } catch (err) {
    console.error('加载日期任务失败:', err)
    planDateTasks.value = []
  } finally {
    loadingPlanDateTasks.value = false
  }
}

// 选定日期的工时统计
const planDateTasksStats = computed(() => {
  const tasks = planDateTasks.value
  const totalHours = tasks.reduce((sum, t) => sum + (t.hours || 0), 0)
  const completedCount = tasks.filter(t => t.status === 'completed').length
  return {
    count: tasks.length,
    totalHours,
    completedCount
  }
})

// 打开规划面板
const openPlanModal = async () => {
  planDate.value = dayjs().format('YYYY-MM-DD')
  selectedTaskIds.value = []
  newTaskName.value = ''
  newTaskHours.value = 0
  planLoading.value = true
  planModalVisible.value = true

  try {
    const [pendingResult] = await Promise.all([
      GetPendingTasks(),
      loadPlanDateTasks(planDate.value)
    ])
    pendingTasks.value = pendingResult || []
  } catch (err) {
    console.error('加载待办任务失败:', err)
    Message.error('加载待办任务失败')
  } finally {
    planLoading.value = false
  }
}

// 监听规划日期变化
watch(planDate, async (newDate) => {
  if (planModalVisible.value && newDate) {
    await loadPlanDateTasks(newDate)
  }
})

// 切换任务选中状态
const toggleTaskSelection = (taskId: number) => {
  const index = selectedTaskIds.value.indexOf(taskId)
  if (index === -1) {
    selectedTaskIds.value.push(taskId)
  } else {
    selectedTaskIds.value.splice(index, 1)
  }
}

// 添加快速任务到规划
const addQuickTask = async () => {
  if (!newTaskName.value.trim()) {
    Message.warning('请输入任务名称')
    return
  }

  try {
    const input: main.TaskInput = {
      id: 0,
      name: newTaskName.value,
      description: '',
      date: planDate.value,
      hours: newTaskHours.value,
      priority: 'medium',
      urgency: 'medium',
      status: 'scheduled'
    }
    await CreateTask(input)
    Message.success('任务已添加')
    newTaskName.value = ''
    newTaskHours.value = 0
    // 刷新已安排任务列表
    await loadPlanDateTasks(planDate.value)
  } catch (err) {
    console.error('创建任务失败:', err)
    Message.error('创建失败')
  }
}

// 提交日程规划
const submitPlan = async () => {
  if (selectedTaskIds.value.length === 0) {
    Message.warning('请至少选择一个任务')
    return
  }

  planLoading.value = true
  try {
    for (const taskId of selectedTaskIds.value) {
      await AssignTaskToDate(taskId, planDate.value)
    }
    Message.success(`已将 ${selectedTaskIds.value.length} 个任务安排到 ${planDate.value}`)
    planModalVisible.value = false
    await loadData()
  } catch (err) {
    console.error('规划失败:', err)
    Message.error('规划失败')
  } finally {
    planLoading.value = false
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

// 当标签页激活时重新加载数据
watch(() => props.active, (isActive) => {
  if (isActive) {
    loadData()
  }
})

onMounted(() => {
  loadData()
})
</script>

<template>
  <div class="workbench">
    <a-spin :loading="loading">
      <a-row :gutter="16">
        <!-- 左侧：统计信息 -->
        <a-col :span="6">
          <!-- 快捷操作 -->
          <div class="quick-actions">
            <a-button type="primary" @click="openTaskModal">
              <template #icon><icon-plus /></template>
              新建今日任务
            </a-button>
            <a-button @click="openTodoModal">
              <template #icon><icon-plus /></template>
              新建待办
            </a-button>
          </div>

          <!-- 统计卡片 -->
          <a-row :gutter="12" class="stats-row">
            <a-col :span="12">
              <a-card class="stat-card">
                <a-statistic title="今日任务" :value="data.total_count">
                  <template #suffix>个</template>
                </a-statistic>
              </a-card>
            </a-col>
            <a-col :span="12">
              <a-card class="stat-card">
                <a-statistic title="已完成" :value="data.completed_count">
                  <template #suffix>个</template>
                </a-statistic>
              </a-card>
            </a-col>
            <a-col :span="12">
              <a-card class="stat-card">
                <a-statistic title="计划工时" :value="data.planned_hours" :precision="1">
                  <template #suffix>h</template>
                </a-statistic>
              </a-card>
            </a-col>
            <a-col :span="12">
              <a-card class="stat-card">
                <a-statistic title="完成工时" :value="data.completed_hours" :precision="1">
                  <template #suffix>h</template>
                </a-statistic>
              </a-card>
            </a-col>
          </a-row>

          <!-- 进度条 -->
          <a-card class="progress-card">
            <div class="progress-item">
              <div class="progress-title">任务完成进度</div>
              <div class="progress-row">
                <a-progress :percent="completionRate / 100" :stroke-width="10" :show-text="false" />
                <span class="progress-text">{{ completionRate }}%</span>
              </div>
            </div>
            <div class="progress-item">
              <div class="progress-title">工时完成进度</div>
              <div class="progress-row">
                <a-progress :percent="hoursRate / 100" :stroke-width="10" color="#00B42A" :show-text="false" />
                <span class="progress-text">{{ hoursRate }}%</span>
              </div>
            </div>
          </a-card>

          <!-- 待办提醒 -->
          <a-alert v-if="data.pending_count > 0" type="warning" class="pending-alert">
            您有 {{ data.pending_count }} 个待办任务尚未安排
          </a-alert>
        </a-col>

        <!-- 右侧：任务列表 -->
        <a-col :span="18">
          <a-card class="tasks-card">
            <template #title>
              <div class="card-title-tabs">
                <span
                  class="tab-item"
                  :class="{ active: viewMode === 'today' }"
                  @click="switchView('today')"
                >今日任务</span>
                <span
                  class="tab-item"
                  :class="{ active: viewMode === 'tomorrow' }"
                  @click="switchView('tomorrow')"
                >明日待办</span>
              </div>
            </template>
            <template #extra>
              <a-button type="text" @click="viewMode === 'today' ? loadData() : loadTomorrowTasks()">
                <template #icon><icon-refresh /></template>
                刷新
              </a-button>
            </template>

            <!-- 今日任务视图 -->
            <template v-if="viewMode === 'today'">
              <a-empty v-if="!data.today_tasks || data.today_tasks.length === 0" description="今日暂无任务" />

              <div v-else class="task-list">
                <div v-for="task in data.today_tasks" :key="task.id" class="task-card">
                  <div class="task-content" @click="openDetailModal(task)">
                    <div class="task-header">
                      <span class="task-name" :class="{ 'completed': task.status === 'completed' }">
                        {{ task.name }}
                      </span>
                      <span class="task-tags">
                        <a-tag v-if="task.project_name" size="small">{{ task.project_name }}</a-tag>
                        <a-tag :color="getStatusColor(task.status)" size="small">
                          {{ getStatusText(task.status) }}
                        </a-tag>
                      </span>
                    </div>
                    <div class="task-meta">
                      <span v-if="task.start_time" class="meta-item">
                        {{ task.start_time }}
                        <span v-if="task.end_time"> - {{ task.end_time }}</span>
                      </span>
                      <span v-if="task.hours" class="meta-item hours">{{ task.hours }}h</span>
                    </div>
                  </div>
                  <div class="task-actions">
                    <a-button
                      v-if="task.status !== 'completed'"
                      type="primary"
                      size="small"
                      @click.stop="openCompleteModal(task)"
                    >
                      完成
                    </a-button>
                    <a-button type="outline" size="small" @click.stop="openAIChat(task)">
                      <template #icon><icon-robot /></template>
                      AI
                    </a-button>
                  </div>
                </div>
              </div>
            </template>

            <!-- 明日待办视图（只读） -->
            <template v-else>
              <a-empty v-if="tomorrowTasks.length === 0" description="明日暂无安排" />

              <div v-else class="task-list">
                <div v-for="task in tomorrowTasks" :key="task.id" class="task-card readonly">
                  <div class="task-content">
                    <div class="task-header">
                      <span class="task-name" :class="{ 'completed': task.status === 'completed' }">
                        {{ task.name }}
                      </span>
                      <span class="task-tags">
                        <a-tag v-if="task.project_name" size="small">{{ task.project_name }}</a-tag>
                        <a-tag :color="getStatusColor(task.status)" size="small">
                          {{ getStatusText(task.status) }}
                        </a-tag>
                      </span>
                    </div>
                    <div class="task-meta">
                      <span v-if="task.start_time" class="meta-item">
                        {{ task.start_time }}
                        <span v-if="task.end_time"> - {{ task.end_time }}</span>
                      </span>
                      <span v-if="task.hours" class="meta-item hours">{{ task.hours }}h</span>
                    </div>
                  </div>
                </div>
              </div>
            </template>
          </a-card>
        </a-col>
      </a-row>
    </a-spin>

    <!-- 新建今日任务弹窗 -->
    <a-modal
      v-model:visible="taskModalVisible"
      title="新建今日任务"
      @ok="handleTaskSubmit"
      @cancel="taskModalVisible = false"
    >
      <a-form :model="taskForm" layout="vertical">
        <a-form-item label="任务名称" required>
          <a-input v-model="taskForm.name" placeholder="请输入任务名称" />
        </a-form-item>

        <a-form-item label="所属项目">
          <a-select v-model="taskForm.project_id" placeholder="选择项目（可选）" allow-clear>
            <a-option v-for="p in projects" :key="p.id" :value="p.id">
              {{ p.name }}
            </a-option>
          </a-select>
        </a-form-item>

        <a-form-item label="预计工时（小时）">
          <a-input-number
            v-model="taskForm.hours"
            :min="0"
            :max="24"
            :precision="1"
            :step="0.5"
            style="width: 100%"
          />
        </a-form-item>

        <a-form-item label="任务描述">
          <a-textarea
            v-model="taskForm.description"
            placeholder="请输入任务描述（可选）"
            :auto-size="{ minRows: 2, maxRows: 4 }"
          />
        </a-form-item>
      </a-form>
    </a-modal>

    <!-- 新建待办弹窗 -->
    <a-modal
      v-model:visible="todoModalVisible"
      title="新建待办"
      @ok="handleTodoSubmit"
      @cancel="todoModalVisible = false"
    >
      <a-form :model="todoForm" layout="vertical">
        <a-form-item label="任务名称" required>
          <a-input v-model="todoForm.name" placeholder="请输入任务名称" />
        </a-form-item>

        <a-form-item label="所属项目">
          <a-select v-model="todoForm.project_id" placeholder="选择项目（可选）" allow-clear>
            <a-option v-for="p in projects" :key="p.id" :value="p.id">
              {{ p.name }}
            </a-option>
          </a-select>
        </a-form-item>

        <a-form-item label="预计工时（小时）">
          <a-input-number
            v-model="todoForm.hours"
            :min="0"
            :max="100"
            :precision="1"
            :step="0.5"
            style="width: 100%"
          />
        </a-form-item>

        <a-form-item label="任务描述">
          <a-textarea
            v-model="todoForm.description"
            placeholder="请输入任务描述（可选）"
            :auto-size="{ minRows: 2, maxRows: 4 }"
          />
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
      <a-alert v-if="selectedTask" type="info" style="margin-bottom: 16px">
        任务：{{ selectedTask.name }}
      </a-alert>
      <a-form :model="completeForm" layout="vertical">
        <a-form-item label="实际开始时间">
          <a-time-picker
            v-model="completeForm.actual_start"
            format="HH:mm"
            style="width: 100%"
            placeholder="选择开始时间"
          />
        </a-form-item>

        <a-form-item label="实际工时（小时）">
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

    <!-- 任务详情弹窗 -->
    <a-modal
      v-model:visible="detailModalVisible"
      title="任务详情"
      :footer="false"
      @cancel="detailModalVisible = false"
      :width="520"
    >
      <div v-if="viewTask" class="task-detail">
        <div class="detail-header">
          <h3 class="detail-title">{{ viewTask.name }}</h3>
          <a-tag :color="getStatusColor(viewTask.status)">
            {{ getStatusText(viewTask.status) }}
          </a-tag>
        </div>

        <a-descriptions :column="2" bordered size="small" class="detail-desc">
          <a-descriptions-item label="所属项目">
            {{ viewTask.project_name || '无' }}
          </a-descriptions-item>
          <a-descriptions-item label="计划日期">
            {{ viewTask.date || '未安排' }}
          </a-descriptions-item>
          <a-descriptions-item label="开始时间">
            {{ viewTask.start_time || '-' }}
          </a-descriptions-item>
          <a-descriptions-item label="结束时间">
            {{ viewTask.end_time || '-' }}
          </a-descriptions-item>
          <a-descriptions-item label="预计工时">
            {{ viewTask.hours ? viewTask.hours + 'h' : '-' }}
          </a-descriptions-item>
          <a-descriptions-item label="实际工时">
            {{ viewTask.actual_hours ? viewTask.actual_hours + 'h' : '-' }}
          </a-descriptions-item>
          <a-descriptions-item label="重要程度">
            <a-tag :color="getPriorityColor(viewTask.priority || 'medium')" size="small">
              {{ getPriorityText(viewTask.priority || 'medium') }}
            </a-tag>
          </a-descriptions-item>
          <a-descriptions-item label="紧急程度">
            <a-tag :color="getUrgencyColor(viewTask.urgency || 'medium')" size="small">
              {{ getUrgencyText(viewTask.urgency || 'medium') }}
            </a-tag>
          </a-descriptions-item>
          <a-descriptions-item label="截止日期" :span="2">
            {{ viewTask.deadline || '无' }}
          </a-descriptions-item>
          <a-descriptions-item label="任务描述" :span="2">
            <div class="detail-description">{{ viewTask.description || '无' }}</div>
          </a-descriptions-item>
        </a-descriptions>

        <div v-if="viewTask.status !== 'completed'" class="detail-actions">
          <a-button type="primary" @click="detailModalVisible = false; openCompleteModal(viewTask)">
            完成任务
          </a-button>
        </div>
      </div>
    </a-modal>

    <!-- 日程规划浮动按钮 -->
    <div class="plan-fab" @click="openPlanModal">
      <div class="fab-inner">
        <icon-calendar class="fab-icon" />
      </div>
      <div class="fab-glow"></div>
    </div>

    <!-- 日程规划弹窗 -->
    <a-modal
      v-model:visible="planModalVisible"
      title="规划日程"
      :width="900"
      :footer="false"
      @cancel="planModalVisible = false"
    >
      <a-spin :loading="planLoading">
        <div class="plan-modal-content">
          <div class="plan-layout">
            <!-- 左侧：日期选择和待办 -->
            <div class="plan-left">
              <!-- 日期选择 -->
              <div class="plan-date-section">
                <div class="section-label">选择日期</div>
                <a-date-picker
                  v-model="planDate"
                  style="width: 100%"
                  :allow-clear="false"
                />
                <div class="quick-dates">
                  <a-tag
                    :color="planDate === dayjs().format('YYYY-MM-DD') ? 'arcoblue' : ''"
                    @click="planDate = dayjs().format('YYYY-MM-DD')"
                    class="quick-date-tag"
                  >今天</a-tag>
                  <a-tag
                    :color="planDate === dayjs().add(1, 'day').format('YYYY-MM-DD') ? 'arcoblue' : ''"
                    @click="planDate = dayjs().add(1, 'day').format('YYYY-MM-DD')"
                    class="quick-date-tag"
                  >明天</a-tag>
                  <a-tag
                    :color="planDate === dayjs().add(2, 'day').format('YYYY-MM-DD') ? 'arcoblue' : ''"
                    @click="planDate = dayjs().add(2, 'day').format('YYYY-MM-DD')"
                    class="quick-date-tag"
                  >后天</a-tag>
                </div>
              </div>

              <!-- 快速添加新任务 -->
              <div class="plan-quick-add">
                <div class="section-label">快速添加新任务</div>
                <div class="quick-add-row">
                  <a-input
                    v-model="newTaskName"
                    placeholder="输入任务名称"
                    style="flex: 1"
                    @keyup.enter="addQuickTask"
                  />
                  <a-input-number
                    v-model="newTaskHours"
                    :min="0"
                    :max="24"
                    :step="0.5"
                    :precision="1"
                    placeholder="工时"
                    style="width: 80px"
                  />
                  <a-button type="primary" @click="addQuickTask">
                    <icon-plus />
                  </a-button>
                </div>
              </div>

              <!-- 待办任务列表 -->
              <div class="plan-pending-section">
                <div class="section-label">
                  从待办中选择
                  <span class="selected-count" v-if="selectedTaskIds.length > 0">
                    已选 {{ selectedTaskIds.length }} 项
                  </span>
                </div>
                <a-empty v-if="pendingTasks.length === 0" description="暂无待办任务" />
                <div v-else class="pending-task-list">
                  <div
                    v-for="task in pendingTasks"
                    :key="task.id"
                    class="pending-task-item"
                    :class="{ selected: selectedTaskIds.includes(task.id) }"
                    @click="toggleTaskSelection(task.id)"
                  >
                    <a-checkbox :model-value="selectedTaskIds.includes(task.id)" />
                    <div class="pending-task-info">
                      <div class="pending-task-name">{{ task.name }}</div>
                      <div class="pending-task-meta">
                        <span v-if="task.project_name" class="task-project">{{ task.project_name }}</span>
                        <span v-if="task.hours" class="task-hours">{{ task.hours }}h</span>
                      </div>
                    </div>
                  </div>
                </div>
              </div>
            </div>

            <!-- 右侧：已安排任务 -->
            <div class="plan-right">
              <div class="plan-existing-section">
                <div class="section-header">
                  <div class="section-label">{{ planDate }} 已安排任务</div>
                  <div class="section-stats">
                    {{ planDateTasksStats.count }} 个任务，共 {{ planDateTasksStats.totalHours.toFixed(1) }}h
                  </div>
                </div>

                <a-spin :loading="loadingPlanDateTasks">
                  <div v-if="planDateTasks.length === 0" class="empty-date-tasks">
                    该日期暂无任务安排
                  </div>
                  <div v-else class="existing-task-list">
                    <div v-for="task in planDateTasks" :key="task.id" class="existing-task-item">
                      <div class="existing-task-main">
                        <a-tag v-if="task.status === 'completed'" size="small" color="green">已完成</a-tag>
                        <a-tag v-else-if="task.status === 'in_progress'" size="small" color="blue">进行中</a-tag>
                        <a-tag v-else size="small" color="orange">待开始</a-tag>
                        <span class="existing-task-name" :class="{ completed: task.status === 'completed' }">
                          {{ task.name }}
                        </span>
                      </div>
                      <div class="existing-task-meta">
                        <span v-if="task.start_time && task.end_time" class="time-range">
                          {{ task.start_time }} - {{ task.end_time }}
                        </span>
                        <span v-if="task.hours" class="hours-badge">{{ task.hours }}h</span>
                        <a-tag v-if="task.project_name" size="small">{{ task.project_name }}</a-tag>
                      </div>
                    </div>
                  </div>
                </a-spin>
              </div>
            </div>
          </div>

          <!-- 提交按钮 -->
          <div class="plan-actions">
            <a-button @click="planModalVisible = false">取消</a-button>
            <a-button
              type="primary"
              @click="submitPlan"
              :disabled="selectedTaskIds.length === 0"
            >
              确认安排 {{ selectedTaskIds.length > 0 ? `(${selectedTaskIds.length})` : '' }}
            </a-button>
          </div>
        </div>
      </a-spin>
    </a-modal>

    <!-- AI会话弹窗 -->
    <TaskAIChat
      v-model:visible="aiChatVisible"
      :task="aiChatTask"
    />
  </div>
</template>

<style scoped>
.workbench {
  padding: 0;
  width: 100%;
}

.workbench :deep(.arco-spin) {
  width: 100%;
  display: block;
}

.quick-actions {
  display: flex;
  gap: 8px;
  margin-bottom: 12px;
}

.quick-actions .arco-btn {
  flex: 1;
}

.stats-row {
  margin-bottom: 12px;
}

.stats-row .arco-col {
  margin-bottom: 12px;
}

.stat-card {
  background: #2a2a2b;
}

.stat-card :deep(.arco-statistic-title) {
  font-size: 12px;
}

.stat-card :deep(.arco-statistic-value) {
  font-size: 20px;
}

.progress-card {
  background: #2a2a2b;
  margin-bottom: 12px;
}

.progress-item {
  margin-bottom: 16px;
}

.progress-item:last-child {
  margin-bottom: 0;
}

.progress-title {
  margin-bottom: 8px;
  color: #86909c;
  font-size: 13px;
}

.progress-row {
  display: flex;
  align-items: center;
  gap: 12px;
}

.progress-row :deep(.arco-progress) {
  flex: 1;
}

.progress-text {
  font-size: 14px;
  font-weight: 500;
  color: #c9cdd4;
  min-width: 40px;
  text-align: right;
}

.pending-alert {
  margin-top: 0;
}

.tasks-card {
  background: #2a2a2b;
  height: 100%;
  min-height: 520px;
}

.card-title-tabs {
  display: flex;
  gap: 20px;
}

.tab-item {
  cursor: pointer;
  color: #86909c;
  font-weight: 500;
  padding-bottom: 4px;
  border-bottom: 2px solid transparent;
  transition: all 0.2s;
}

.tab-item:hover {
  color: #c9cdd4;
}

.tab-item.active {
  color: #165DFF;
  border-bottom-color: #165DFF;
}

.task-card.readonly {
  cursor: default;
}

.task-card.readonly .task-content {
  cursor: default;
}

.tasks-card :deep(.arco-card-body) {
  max-height: 460px;
  overflow-y: auto;
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
  background: #232324;
  border-radius: 6px;
  padding: 12px 16px;
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
  cursor: pointer;
}

.task-header {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 6px;
  flex-wrap: wrap;
}

.task-name {
  font-weight: 500;
  font-size: 15px;
}

.task-name.completed {
  text-decoration: line-through;
  color: #86909c;
}

.task-tags {
  display: flex;
  gap: 4px;
  flex-wrap: wrap;
}

.task-meta {
  display: flex;
  align-items: center;
  gap: 12px;
  color: #86909c;
  font-size: 13px;
}

.meta-item.hours {
  color: #165DFF;
}

.task-actions {
  display: flex;
  align-items: center;
  gap: 4px;
  flex-shrink: 0;
  margin-left: 12px;
}

/* 任务详情弹窗样式 */
.task-detail {
  padding: 0;
  font-size: 14px;
}

.detail-header {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 20px;
}

.detail-title {
  margin: 0;
  font-size: 20px;
  font-weight: 600;
}

.detail-desc {
  margin-bottom: 16px;
}

.detail-desc :deep(.arco-descriptions-item-label) {
  font-size: 14px;
  color: #86909c;
}

.detail-desc :deep(.arco-descriptions-item-value) {
  font-size: 14px;
}

.detail-description {
  white-space: pre-wrap;
  word-break: break-word;
  line-height: 1.6;
}

.detail-actions {
  margin-top: 20px;
  text-align: right;
}

/* 日程规划浮动按钮 */
.plan-fab {
  position: fixed;
  left: 50%;
  bottom: 32px;
  transform: translateX(-50%);
  cursor: pointer;
  z-index: 100;
}

.fab-inner {
  position: relative;
  width: 72px;
  height: 72px;
  border-radius: 50%;
  background: linear-gradient(135deg, #165DFF 0%, #0052D9 50%, #1890FF 100%);
  box-shadow: 0 6px 24px rgba(22, 93, 255, 0.5),
              0 0 48px rgba(22, 93, 255, 0.3),
              inset 0 1px 1px rgba(255, 255, 255, 0.2);
  color: white;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  z-index: 2;
}

.fab-icon {
  font-size: 30px;
}

.fab-inner:hover {
  transform: scale(1.08);
  box-shadow: 0 8px 32px rgba(22, 93, 255, 0.6),
              0 0 56px rgba(22, 93, 255, 0.4),
              inset 0 1px 1px rgba(255, 255, 255, 0.3);
}

.fab-inner:active {
  transform: scale(0.95);
}

.fab-glow {
  position: absolute;
  top: 0;
  left: 50%;
  transform: translateX(-50%);
  width: 72px;
  height: 72px;
  border-radius: 50%;
  background: linear-gradient(135deg, #165DFF 0%, #1890FF 100%);
  filter: blur(20px);
  opacity: 0.5;
  animation: glow 2.5s ease-in-out infinite;
  z-index: 1;
}

@keyframes glow {
  0%, 100% {
    opacity: 0.5;
    transform: translateX(-50%) scale(1);
  }
  50% {
    opacity: 0.7;
    transform: translateX(-50%) scale(1.15);
  }
}

/* 日程规划弹窗样式 */
.plan-modal-content {
  padding: 0;
}

.plan-layout {
  display: flex;
  gap: 24px;
  min-height: 400px;
}

.plan-left {
  flex: 1;
  min-width: 0;
}

.plan-right {
  width: 360px;
  flex-shrink: 0;
}

.section-label {
  font-size: 14px;
  font-weight: 500;
  color: #c9cdd4;
  margin-bottom: 10px;
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.selected-count {
  font-weight: normal;
  color: #165DFF;
  font-size: 13px;
}

.plan-date-section {
  margin-bottom: 20px;
}

.quick-dates {
  display: flex;
  gap: 8px;
  margin-top: 10px;
}

.quick-date-tag {
  cursor: pointer;
  transition: all 0.2s;
}

.quick-date-tag:hover {
  opacity: 0.8;
}

.plan-quick-add {
  margin-bottom: 20px;
  padding-bottom: 20px;
  border-bottom: 1px solid #333;
}

.quick-add-row {
  display: flex;
  gap: 8px;
  align-items: center;
}

.plan-pending-section {
  margin-bottom: 20px;
}

.pending-task-list {
  max-height: 240px;
  overflow-y: auto;
  border: 1px solid #333;
  border-radius: 6px;
}

.pending-task-item {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px;
  border-bottom: 1px solid #2a2a2b;
  cursor: pointer;
  transition: background 0.2s;
}

.pending-task-item:last-child {
  border-bottom: none;
}

.pending-task-item:hover {
  background: #2a2a2b;
}

.pending-task-item.selected {
  background: rgba(22, 93, 255, 0.1);
}

.pending-task-info {
  flex: 1;
  min-width: 0;
}

.pending-task-name {
  font-size: 14px;
  font-weight: 500;
  margin-bottom: 4px;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.pending-task-meta {
  display: flex;
  gap: 8px;
  font-size: 12px;
  color: #86909c;
}

.task-project {
  background: #333;
  padding: 1px 6px;
  border-radius: 3px;
}

.task-hours {
  color: #165DFF;
}

.plan-actions {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
  margin-top: 20px;
  padding-top: 16px;
  border-top: 1px solid #333;
  width: 100%;
}

/* 已安排任务区域样式 */
.plan-existing-section {
  background: #1e1e1f;
  border: 1px solid #3a3a3c;
  border-radius: 8px;
  padding: 16px;
  height: 100%;
  display: flex;
  flex-direction: column;
}

.plan-existing-section .section-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 12px;
  padding-bottom: 12px;
  border-bottom: 1px solid #3a3a3c;
}

.plan-existing-section .section-label {
  margin-bottom: 0;
}

.section-stats {
  color: #86909c;
  font-size: 13px;
}

.empty-date-tasks {
  text-align: center;
  color: #86909c;
  padding: 40px 0;
}

.existing-task-list {
  flex: 1;
  overflow-y: auto;
  display: flex;
  flex-direction: column;
  gap: 8px;
  max-height: 320px;
}

.existing-task-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 10px 12px;
  background: #2a2a2b;
  border-radius: 6px;
}

.existing-task-main {
  display: flex;
  align-items: center;
  gap: 8px;
  flex: 1;
  min-width: 0;
}

.existing-task-name {
  font-size: 14px;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.existing-task-name.completed {
  text-decoration: line-through;
  color: #86909c;
}

.existing-task-meta {
  display: flex;
  align-items: center;
  gap: 8px;
  color: #86909c;
  font-size: 12px;
  flex-shrink: 0;
}

.existing-task-meta .time-range {
  color: #4080ff;
}

.existing-task-meta .hours-badge {
  background: #3a3a3c;
  padding: 2px 6px;
  border-radius: 4px;
}
</style>
