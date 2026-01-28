<script lang="ts" setup>
import { ref, onMounted } from 'vue'
import {
  GetProjects,
  CreateProject,
  UpdateProject,
  DeleteProject
} from '../../wailsjs/go/main/App'
import { main } from '../../wailsjs/go/models'
import { Message } from '@arco-design/web-vue'

interface ProjectForm {
  id: number
  name: string
  description: string
  color: string
}

const loading = ref(false)
const projects = ref<main.Project[]>([])
const modalVisible = ref(false)
const isEditing = ref(false)

const presetColors = [
  '#165DFF', // 蓝色
  '#00B42A', // 绿色
  '#FF7D00', // 橙色
  '#F53F3F', // 红色
  '#722ED1', // 紫色
  '#0FC6C2', // 青色
  '#F7BA1E', // 黄色
  '#86909C'  // 灰色
]

const defaultForm = (): ProjectForm => ({
  id: 0,
  name: '',
  description: '',
  color: '#165DFF'
})

const form = ref<ProjectForm>(defaultForm())

const columns = [
  { title: '颜色', dataIndex: 'color', width: 80, slotName: 'color' },
  { title: '项目名称', dataIndex: 'name', ellipsis: true },
  { title: '描述', dataIndex: 'description', ellipsis: true },
  { title: '任务数', dataIndex: 'task_count', width: 100 },
  { title: '操作', slotName: 'actions', width: 160 }
]

const loadProjects = async () => {
  loading.value = true
  try {
    const result = await GetProjects()
    projects.value = result || []
  } catch (err) {
    console.error('加载项目失败:', err)
    Message.error('加载项目失败')
  } finally {
    loading.value = false
  }
}

const openCreateModal = () => {
  isEditing.value = false
  form.value = defaultForm()
  modalVisible.value = true
}

const openEditModal = (project: main.Project) => {
  isEditing.value = true
  form.value = {
    id: project.id,
    name: project.name,
    description: project.description,
    color: project.color
  }
  modalVisible.value = true
}

const handleSubmit = async () => {
  if (!form.value.name.trim()) {
    Message.warning('请输入项目名称')
    return
  }

  try {
    if (isEditing.value) {
      await UpdateProject(form.value.id, form.value.name, form.value.description, form.value.color)
      Message.success('项目已更新')
    } else {
      await CreateProject(form.value.name, form.value.description, form.value.color)
      Message.success('项目已创建')
    }

    modalVisible.value = false
    await loadProjects()
  } catch (err) {
    console.error('保存项目失败:', err)
    Message.error('保存失败')
  }
}

const handleDelete = async (project: main.Project) => {
  if (project.task_count > 0) {
    Message.warning(`该项目下还有 ${project.task_count} 个任务，删除项目后任务将变为无项目状态`)
  }

  try {
    await DeleteProject(project.id)
    Message.success('项目已删除')
    await loadProjects()
  } catch (err) {
    console.error('删除项目失败:', err)
    Message.error('删除失败')
  }
}

onMounted(() => {
  loadProjects()
})
</script>

<template>
  <div class="project-management">
    <!-- 工具栏 -->
    <div class="toolbar">
      <div class="title">项目管理</div>
      <a-button type="primary" @click="openCreateModal">
        <template #icon><icon-plus /></template>
        新建项目
      </a-button>
    </div>

    <!-- 项目表格 -->
    <a-table
      :loading="loading"
      :columns="columns"
      :data="projects"
      :pagination="false"
      row-key="id"
      class="projects-table"
    >
      <template #color="{ record }">
        <div
          class="color-dot"
          :style="{ backgroundColor: record.color }"
        />
      </template>
      <template #actions="{ record }">
        <a-space>
          <a-button type="text" size="small" @click="openEditModal(record)">
            编辑
          </a-button>
          <a-popconfirm content="确定删除此项目?" @ok="handleDelete(record)">
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
      :title="isEditing ? '编辑项目' : '新建项目'"
      @ok="handleSubmit"
      @cancel="modalVisible = false"
    >
      <a-form :model="form" layout="vertical">
        <a-form-item label="项目名称" required>
          <a-input v-model="form.name" placeholder="请输入项目名称" />
        </a-form-item>

        <a-form-item label="项目颜色">
          <div class="color-picker">
            <div
              v-for="color in presetColors"
              :key="color"
              class="color-option"
              :class="{ selected: form.color === color }"
              :style="{ backgroundColor: color }"
              @click="form.color = color"
            />
          </div>
        </a-form-item>

        <a-form-item label="项目描述">
          <a-textarea
            v-model="form.description"
            placeholder="请输入项目描述（可选）"
            :auto-size="{ minRows: 2, maxRows: 4 }"
          />
        </a-form-item>
      </a-form>
    </a-modal>
  </div>
</template>

<style scoped>
.project-management {
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
}

.projects-table {
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

.color-dot {
  width: 24px;
  height: 24px;
  border-radius: 50%;
}

.color-picker {
  display: flex;
  gap: 12px;
  flex-wrap: wrap;
}

.color-option {
  width: 32px;
  height: 32px;
  border-radius: 50%;
  cursor: pointer;
  border: 2px solid transparent;
  transition: all 0.2s;
}

.color-option:hover {
  transform: scale(1.1);
}

.color-option.selected {
  border-color: #fff;
  box-shadow: 0 0 0 2px rgba(255, 255, 255, 0.3);
}
</style>
