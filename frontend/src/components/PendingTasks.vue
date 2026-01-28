<script lang="ts" setup>
import { ref, onMounted } from 'vue'
import {
  GetPendingTasks,
  GetProjects,
  CreateTask,
  UpdateTask,
  DeleteTask,
  AssignTaskToDate
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
}

const loading = ref(false)
const tasks = ref<main.Task[]>([])
const projects = ref<main.Project[]>([])
const modalVisible = ref(false)
const assignModalVisible = ref(false)
const isEditing = ref(false)
const selectedTask = ref<main.Task | null>(null)
const assignDate = ref(dayjs().format('YYYY-MM-DD'))

const defaultForm = (): TaskForm => ({
  id: 0,
  project_id: undefined,
  name: '',
  description: '',
  hours: 0
})

const form = ref<TaskForm>(defaultForm())

const loadTasks = async () => {
  loading.value = true
  try {
    const result = await GetPendingTasks()
    tasks.value = result || []
  } catch (err) {
    console.error('加载待处理任务失败:', err)
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
  modalVisible.value = true
}

const openEditModal = (task: main.Task) => {
  isEditing.value = true
  form.value = {
    id: task.id,
    project_id: task.project_id,
    name: task.name,
    description: task.description,
    hours: task.hours
  }
  modalVisible.value = true
}

const openAssignModal = (task: main.Task) => {
  selectedTask.value = task
  assignDate.value = dayjs().format('YYYY-MM-DD')
  assignModalVisible.value = true
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
      date: undefined,
      start_time: undefined,
      end_time: undefined,
      hours: form.value.hours,
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

onMounted(() => {
  loadTasks()
  loadProjects()
})
</script>

<template>
  <div class="pending-tasks">
    <!-- 工具栏 -->
    <div class="toolbar">
      <div class="title">
        待处理任务
        <a-badge :count="tasks.length" :max-count="99" />
      </div>
      <a-button type="primary" @click="openCreateModal">
        <template #icon><icon-plus /></template>
        新建任务
      </a-button>
    </div>

    <!-- 任务列表 -->
    <a-spin :loading="loading">
      <a-empty v-if="tasks.length === 0" description="暂无待处理任务" />

      <a-list v-else :bordered="false">
        <a-list-item v-for="task in tasks" :key="task.id" class="task-item">
          <a-list-item-meta>
            <template #title>
              <span class="task-name">{{ task.name }}</span>
              <a-tag v-if="task.project_name" size="small" class="project-tag">
                {{ task.project_name }}
              </a-tag>
            </template>
            <template #description>
              <span v-if="task.description" class="task-desc">{{ task.description }}</span>
              <span v-if="task.hours" class="hours-text">预计 {{ task.hours }} 小时</span>
            </template>
          </a-list-item-meta>
          <template #actions>
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
          </template>
        </a-list-item>
      </a-list>
    </a-spin>

    <!-- 新建/编辑弹窗 -->
    <a-modal
      v-model:visible="modalVisible"
      :title="isEditing ? '编辑任务' : '新建待处理任务'"
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
      title="分配日期"
      @ok="handleAssign"
      @cancel="assignModalVisible = false"
    >
      <a-form layout="vertical">
        <a-form-item label="选择日期">
          <a-date-picker
            v-model="assignDate"
            style="width: 100%"
            :allow-clear="false"
          />
        </a-form-item>
        <a-alert v-if="selectedTask" type="info">
          任务「{{ selectedTask.name }}」将被分配到选定日期
        </a-alert>
      </a-form>
    </a-modal>
  </div>
</template>

<style scoped>
.pending-tasks {
  padding: 0;
}

.toolbar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
}

.title {
  font-size: 16px;
  font-weight: 500;
  display: flex;
  align-items: center;
  gap: 8px;
}

.task-item {
  background: #2a2a2b;
  border-radius: 8px;
  margin-bottom: 8px;
  padding: 12px 16px;
}

.task-name {
  font-weight: 500;
}

.project-tag {
  margin-left: 8px;
}

.task-desc {
  color: #86909c;
}

.hours-text {
  margin-left: 12px;
  color: #165DFF;
}
</style>
