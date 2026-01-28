<script lang="ts" setup>
import { ref, onMounted } from 'vue'
import {
  GetAllProjects,
  CreateProject,
  UpdateProject,
  DeleteProject,
  ArchiveProject,
  GetModelProviders,
  UpdateModelProvider,
  GetAgents,
  CreateAgent,
  UpdateAgent,
  DeleteAgent
} from '../../wailsjs/go/main/App'
import { main } from '../../wailsjs/go/models'
import { Message } from '@arco-design/web-vue'

// 当前选中的设置项
const activeKey = ref('projects')

// ========== 项目管理 ==========
const projects = ref<main.Project[]>([])
const projectModalVisible = ref(false)
const isEditingProject = ref(false)
const projectForm = ref({
  id: 0,
  name: '',
  description: '',
  color: '#165DFF'
})

const loadProjects = async () => {
  try {
    const result = await GetAllProjects()
    projects.value = result || []
  } catch (err) {
    console.error('加载项目失败:', err)
    Message.error('加载项目失败')
  }
}

const openCreateProject = () => {
  isEditingProject.value = false
  projectForm.value = { id: 0, name: '', description: '', color: '#165DFF' }
  projectModalVisible.value = true
}

const openEditProject = (p: main.Project) => {
  isEditingProject.value = true
  projectForm.value = {
    id: p.id,
    name: p.name,
    description: p.description,
    color: p.color
  }
  projectModalVisible.value = true
}

const handleProjectSubmit = async () => {
  if (!projectForm.value.name.trim()) {
    Message.warning('请输入项目名称')
    return
  }

  try {
    if (isEditingProject.value) {
      await UpdateProject(
        projectForm.value.id,
        projectForm.value.name,
        projectForm.value.description,
        projectForm.value.color
      )
      Message.success('项目已更新')
    } else {
      await CreateProject(
        projectForm.value.name,
        projectForm.value.description,
        projectForm.value.color
      )
      Message.success('项目已创建')
    }
    projectModalVisible.value = false
    await loadProjects()
  } catch (err) {
    console.error('保存项目失败:', err)
    Message.error('保存失败')
  }
}

const handleArchiveProject = async (p: main.Project) => {
  try {
    await ArchiveProject(p.id, !p.archived)
    Message.success(p.archived ? '项目已取消归档' : '项目已归档')
    await loadProjects()
  } catch (err) {
    console.error('归档项目失败:', err)
    Message.error('操作失败')
  }
}

const handleDeleteProject = async (p: main.Project) => {
  try {
    await DeleteProject(p.id)
    Message.success('项目已删除')
    await loadProjects()
  } catch (err) {
    console.error('删除项目失败:', err)
    Message.error('删除失败')
  }
}

// ========== 模型提供商 ==========
const providers = ref<main.ModelProvider[]>([])
const providerModalVisible = ref(false)
const providerForm = ref({
  id: 0,
  name: '',
  label: '',
  api_key: '',
  base_url: '',
  enabled: true
})

const loadProviders = async () => {
  try {
    const result = await GetModelProviders()
    providers.value = result || []
  } catch (err) {
    console.error('加载模型提供商失败:', err)
    Message.error('加载模型提供商失败')
  }
}

const openEditProvider = (p: main.ModelProvider) => {
  providerForm.value = {
    id: p.id,
    name: p.name,
    label: p.label,
    api_key: p.api_key,
    base_url: p.base_url,
    enabled: p.enabled
  }
  providerModalVisible.value = true
}

const handleProviderSubmit = async () => {
  try {
    const input: main.ModelProviderInput = {
      id: providerForm.value.id,
      name: providerForm.value.name,
      label: providerForm.value.label,
      api_key: providerForm.value.api_key,
      base_url: providerForm.value.base_url,
      enabled: providerForm.value.enabled
    }
    await UpdateModelProvider(input)
    Message.success('配置已保存')
    providerModalVisible.value = false
    await loadProviders()
  } catch (err) {
    console.error('保存配置失败:', err)
    Message.error('保存失败')
  }
}

// ========== Agent管理 ==========
const agents = ref<main.Agent[]>([])
const agentModalVisible = ref(false)
const isEditingAgent = ref(false)
const agentForm = ref({
  id: 0,
  name: '',
  description: '',
  prompt: '',
  provider_id: undefined as number | undefined,
  model: '',
  enabled: true
})

const loadAgents = async () => {
  try {
    const result = await GetAgents()
    agents.value = result || []
  } catch (err) {
    console.error('加载Agent失败:', err)
    Message.error('加载Agent失败')
  }
}

const openCreateAgent = () => {
  isEditingAgent.value = false
  agentForm.value = {
    id: 0,
    name: '',
    description: '',
    prompt: '',
    provider_id: undefined,
    model: '',
    enabled: true
  }
  agentModalVisible.value = true
}

const openEditAgent = (agent: main.Agent) => {
  isEditingAgent.value = true
  agentForm.value = {
    id: agent.id,
    name: agent.name,
    description: agent.description,
    prompt: agent.prompt,
    provider_id: agent.provider_id || undefined,
    model: agent.model,
    enabled: agent.enabled
  }
  agentModalVisible.value = true
}

const handleAgentSubmit = async () => {
  if (!agentForm.value.name.trim()) {
    Message.warning('请输入Agent名称')
    return
  }

  try {
    const input: main.AgentInput = {
      id: agentForm.value.id,
      name: agentForm.value.name,
      description: agentForm.value.description,
      prompt: agentForm.value.prompt,
      provider_id: agentForm.value.provider_id,
      model: agentForm.value.model,
      enabled: agentForm.value.enabled
    }

    if (isEditingAgent.value) {
      await UpdateAgent(input)
      Message.success('Agent已更新')
    } else {
      await CreateAgent(input)
      Message.success('Agent已创建')
    }
    agentModalVisible.value = false
    await loadAgents()
  } catch (err) {
    console.error('保存Agent失败:', err)
    Message.error('保存失败')
  }
}

const handleDeleteAgent = async (agent: main.Agent) => {
  try {
    await DeleteAgent(agent.id)
    Message.success('Agent已删除')
    await loadAgents()
  } catch (err) {
    console.error('删除Agent失败:', err)
    Message.error('删除失败')
  }
}

const getProviderName = (providerId: number | undefined) => {
  if (!providerId) return '-'
  const provider = providers.value.find(p => p.id === providerId)
  return provider?.label || '-'
}

onMounted(() => {
  loadProjects()
  loadProviders()
  loadAgents()
})
</script>

<template>
  <div class="settings">
    <a-tabs v-model:active-key="activeKey" type="rounded" class="settings-tabs">
      <!-- 项目管理 -->
      <a-tab-pane key="projects" title="项目管理">
        <div class="section-header">
          <span class="section-title">项目列表</span>
          <a-button type="primary" size="small" @click="openCreateProject">
            <template #icon><icon-plus /></template>
            新建项目
          </a-button>
        </div>

        <a-table :data="projects" :pagination="false" row-key="id" class="settings-table">
          <template #columns>
            <a-table-column title="项目名称" data-index="name">
              <template #cell="{ record }">
                <div class="project-name-cell">
                  <span class="color-dot" :style="{ background: record.color }"></span>
                  <span :class="{ 'archived-text': record.archived }">{{ record.name }}</span>
                  <a-tag v-if="record.archived" size="small" color="gray">已归档</a-tag>
                </div>
              </template>
            </a-table-column>
            <a-table-column title="描述" data-index="description" ellipsis />
            <a-table-column title="任务数" data-index="task_count" :width="80" />
            <a-table-column title="操作" :width="200">
              <template #cell="{ record }">
                <a-button type="text" size="small" @click="openEditProject(record)">编辑</a-button>
                <a-button type="text" size="small" @click="handleArchiveProject(record)">
                  {{ record.archived ? '取消归档' : '归档' }}
                </a-button>
                <a-popconfirm content="确定删除此项目?" @ok="handleDeleteProject(record)">
                  <a-button type="text" size="small" status="danger">删除</a-button>
                </a-popconfirm>
              </template>
            </a-table-column>
          </template>
        </a-table>
      </a-tab-pane>

      <!-- 模型提供商 -->
      <a-tab-pane key="providers" title="模型提供商">
        <div class="section-header">
          <span class="section-title">API配置</span>
        </div>

        <a-list :bordered="false" class="provider-list">
          <a-list-item v-for="p in providers" :key="p.id" class="provider-item">
            <a-list-item-meta>
              <template #title>
                <div class="provider-title">
                  <span>{{ p.label }}</span>
                  <a-tag v-if="p.api_key" size="small" color="green">已配置</a-tag>
                  <a-tag v-else size="small" color="gray">未配置</a-tag>
                </div>
              </template>
              <template #description>
                <div class="provider-desc">
                  <span>{{ p.base_url }}</span>
                </div>
              </template>
            </a-list-item-meta>
            <template #actions>
              <a-switch v-model="p.enabled" size="small" @change="() => { providerForm.id = p.id; providerForm.api_key = p.api_key; providerForm.base_url = p.base_url; providerForm.enabled = p.enabled; handleProviderSubmit() }" />
              <a-button type="text" size="small" @click="openEditProvider(p)">配置</a-button>
            </template>
          </a-list-item>
        </a-list>
      </a-tab-pane>

      <!-- Agent管理 -->
      <a-tab-pane key="agents" title="Agent管理">
        <div class="section-header">
          <span class="section-title">Agent列表</span>
          <a-button type="primary" size="small" @click="openCreateAgent">
            <template #icon><icon-plus /></template>
            新建Agent
          </a-button>
        </div>

        <a-table :data="agents" :pagination="false" row-key="id" class="settings-table">
          <template #columns>
            <a-table-column title="名称" data-index="name" />
            <a-table-column title="描述" data-index="description" ellipsis />
            <a-table-column title="模型提供商" :width="120">
              <template #cell="{ record }">
                {{ getProviderName(record.provider_id) }}
              </template>
            </a-table-column>
            <a-table-column title="模型" data-index="model" :width="150" />
            <a-table-column title="状态" :width="80">
              <template #cell="{ record }">
                <a-tag v-if="record.enabled" size="small" color="green">启用</a-tag>
                <a-tag v-else size="small" color="gray">禁用</a-tag>
              </template>
            </a-table-column>
            <a-table-column title="操作" :width="120">
              <template #cell="{ record }">
                <a-button type="text" size="small" @click="openEditAgent(record)">编辑</a-button>
                <a-popconfirm content="确定删除此Agent?" @ok="handleDeleteAgent(record)">
                  <a-button type="text" size="small" status="danger">删除</a-button>
                </a-popconfirm>
              </template>
            </a-table-column>
          </template>
        </a-table>
      </a-tab-pane>
    </a-tabs>

    <!-- 项目编辑弹窗 -->
    <a-modal
      v-model:visible="projectModalVisible"
      :title="isEditingProject ? '编辑项目' : '新建项目'"
      @ok="handleProjectSubmit"
      @cancel="projectModalVisible = false"
    >
      <a-form :model="projectForm" layout="vertical">
        <a-form-item label="项目名称" required>
          <a-input v-model="projectForm.name" placeholder="请输入项目名称" />
        </a-form-item>
        <a-form-item label="项目描述">
          <a-textarea
            v-model="projectForm.description"
            placeholder="请输入项目描述（可选）"
            :auto-size="{ minRows: 2, maxRows: 4 }"
          />
        </a-form-item>
        <a-form-item label="项目颜色">
          <a-color-picker v-model="projectForm.color" show-text />
        </a-form-item>
      </a-form>
    </a-modal>

    <!-- 模型提供商配置弹窗 -->
    <a-modal
      v-model:visible="providerModalVisible"
      title="配置模型提供商"
      @ok="handleProviderSubmit"
      @cancel="providerModalVisible = false"
    >
      <a-form :model="providerForm" layout="vertical">
        <a-form-item label="API Key">
          <a-input-password v-model="providerForm.api_key" placeholder="请输入API Key" />
        </a-form-item>
        <a-form-item label="Base URL">
          <a-input v-model="providerForm.base_url" placeholder="API Base URL" />
        </a-form-item>
        <a-form-item label="启用">
          <a-switch v-model="providerForm.enabled" />
        </a-form-item>
      </a-form>
    </a-modal>

    <!-- Agent编辑弹窗 -->
    <a-modal
      v-model:visible="agentModalVisible"
      :title="isEditingAgent ? '编辑Agent' : '新建Agent'"
      @ok="handleAgentSubmit"
      @cancel="agentModalVisible = false"
      :width="560"
    >
      <a-form :model="agentForm" layout="vertical">
        <a-form-item label="Agent名称" required>
          <a-input v-model="agentForm.name" placeholder="请输入Agent名称" />
        </a-form-item>
        <a-form-item label="描述">
          <a-input v-model="agentForm.description" placeholder="请输入描述（可选）" />
        </a-form-item>
        <a-row :gutter="16">
          <a-col :span="12">
            <a-form-item label="模型提供商">
              <a-select v-model="agentForm.provider_id" placeholder="选择提供商" allow-clear>
                <a-option v-for="p in providers" :key="p.id" :value="p.id" :disabled="!p.api_key">
                  {{ p.label }}
                  <span v-if="!p.api_key" style="color: #86909c"> (未配置)</span>
                </a-option>
              </a-select>
            </a-form-item>
          </a-col>
          <a-col :span="12">
            <a-form-item label="模型名称">
              <a-input v-model="agentForm.model" placeholder="如: deepseek-chat" />
            </a-form-item>
          </a-col>
        </a-row>
        <a-form-item label="系统提示词">
          <a-textarea
            v-model="agentForm.prompt"
            placeholder="请输入系统提示词"
            :auto-size="{ minRows: 4, maxRows: 10 }"
          />
        </a-form-item>
        <a-form-item label="启用">
          <a-switch v-model="agentForm.enabled" />
        </a-form-item>
      </a-form>
    </a-modal>
  </div>
</template>

<style scoped>
.settings {
  padding: 0;
}

.settings-tabs {
  background: transparent;
}

.section-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
}

.section-title {
  font-size: 14px;
  font-weight: 500;
  color: #86909c;
}

.settings-table {
  background: #2a2a2b;
  border-radius: 8px;
}

.project-name-cell {
  display: flex;
  align-items: center;
  gap: 8px;
}

.color-dot {
  width: 12px;
  height: 12px;
  border-radius: 50%;
  flex-shrink: 0;
}

.archived-text {
  color: #86909c;
}

.provider-list {
  background: transparent;
}

.provider-item {
  background: #2a2a2b;
  border-radius: 8px;
  margin-bottom: 8px;
  padding: 12px 16px;
}

.provider-title {
  display: flex;
  align-items: center;
  gap: 8px;
}

.provider-desc {
  color: #86909c;
  font-size: 12px;
}
</style>
