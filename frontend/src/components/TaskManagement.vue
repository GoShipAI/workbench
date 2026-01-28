<script lang="ts" setup>
import { ref, watch, onMounted } from 'vue'
import {
  GetTasksByDate,
  GetProjects,
  CreateTask,
  UpdateTask,
  DeleteTask,
  UpdateTaskStatus,
  CalculateHours
} from '../../wailsjs/go/main/App'
import { main } from '../../wailsjs/go/models'
import { Message } from '@arco-design/web-vue'
import dayjs from 'dayjs'

interface TaskForm {
  id: number
  project_id?: number
  name: string
  description: string
  date: string
  start_time: string
  end_time: string
  hours: number
  status: string
  hours_mode: 'direct' | 'calculate'
}

const loading = ref(false)
const selectedDate = ref(dayjs().format('YYYY-MM-DD'))
const tasks = ref<main.Task[]>([])
const projects = ref<main.Project[]>([])
const modalVisible = ref(false)
const isEditing = ref(false)

const defaultForm = (): TaskForm => ({
  id: 0,
  project_id: undefined,
  name: '',
  description: '',
  date: selectedDate.value,
  start_time: '',
  end_time: '',
  hours: 0,
  status: 'scheduled',
  hours_mode: 'direct'
})

const form = ref<TaskForm>(defaultForm())

const columns = [
  { title: '项目', dataIndex: 'project_name', width: 120 },
  { title: '任务名称', dataIndex: 'name', ellipsis: true },
  { title: '开始时间', dataIndex: 'start_time', width: 100 },
  { title: '结束时间', dataIndex: 'end_time', width: 100 },
  { title: '工时', dataIndex: 'hours', width: 80 },
  { title: '状态', dataIndex: 'status', width: 100, slotName: 'status' },
  { title: '操作', slotName: 'actions', width: 200 }
]

const loadTasks = async () => {
  loading.value = true
  try {
    const result = await GetTasksByDate(selectedDate.value)
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

const openCreateModal = () => {
  isEditing.value = false
  form.value = defaultForm()
  form.value.date = selectedDate.value
  modalVisible.value = true
}

const openEditModal = (task: main.Task) => {
  isEditing.value = true
  form.value = {
    id: task.id,
    project_id: task.project_id,
    name: task.name,
    description: task.description,
    date: task.date || selectedDate.value,
    start_time: task.start_time || '',
    end_time: task.end_time || '',
    hours: task.hours,
    status: task.status,
    hours_mode: 'direct'
  }
  modalVisible.value = true
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
  try {
    await UpdateTaskStatus(task.id, status)
    Message.success('状态已更新')
    await loadTasks()
  } catch (err) {
    console.error('更新状态失败:', err)
    Message.error('更新失败')
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

watch(selectedDate, () => {
  loadTasks()
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
      <a-date-picker
        v-model="selectedDate"
        style="width: 200px"
        :allow-clear="false"
      />
      <a-button type="primary" @click="openCreateModal">
        <template #icon><icon-plus /></template>
        新建任务
      </a-button>
    </div>

    <!-- 任务表格 -->
    <a-table
      :loading="loading"
      :columns="columns"
      :data="tasks"
      :pagination="false"
      row-key="id"
      class="tasks-table"
    >
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

        <a-form-item label="日期">
          <a-date-picker v-model="form.date" style="width: 100%" />
        </a-form-item>

        <a-form-item label="工时录入方式">
          <a-radio-group v-model="form.hours_mode">
            <a-radio value="direct">直接填写工时</a-radio>
            <a-radio value="calculate">通过时间计算</a-radio>
          </a-radio-group>
        </a-form-item>

        <a-row :gutter="16">
          <a-col :span="12">
            <a-form-item label="开始时间">
              <a-time-picker
                v-model="form.start_time"
                format="HH:mm"
                style="width: 100%"
              />
            </a-form-item>
          </a-col>
          <a-col :span="12">
            <a-form-item v-if="form.hours_mode === 'calculate'" label="结束时间">
              <a-time-picker
                v-model="form.end_time"
                format="HH:mm"
                style="width: 100%"
              />
            </a-form-item>
            <a-form-item v-else label="工时（小时）">
              <a-input-number
                v-model="form.hours"
                :min="0"
                :max="24"
                :precision="1"
                :step="0.5"
                style="width: 100%"
              />
            </a-form-item>
          </a-col>
        </a-row>

        <a-form-item v-if="form.hours_mode === 'calculate'" label="计算工时">
          <a-input-number
            v-model="form.hours"
            :min="0"
            :max="24"
            :precision="1"
            disabled
            style="width: 100%"
          />
        </a-form-item>

        <a-form-item label="任务描述">
          <a-textarea
            v-model="form.description"
            placeholder="请输入任务描述（可选）"
            :auto-size="{ minRows: 2, maxRows: 4 }"
          />
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
}

.tasks-table {
  background: #2a2a2b;
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
