<script lang="ts" setup>
import { ref, onMounted } from 'vue'
import Workbench from './components/Workbench.vue'
import TaskManagement from './components/TaskManagement.vue'
import PendingTasks from './components/PendingTasks.vue'
import Settings from './components/Settings.vue'

const activeTab = ref('workbench')

onMounted(() => {
  // 设置 Arco Design 暗色主题
  document.body.setAttribute('arco-theme', 'dark')
})
</script>

<template>
  <a-layout class="layout">
    <a-layout-header class="header">
      <div class="logo">
        <span class="logo-icon">📋</span>
        <span class="logo-text">工作台</span>
      </div>
    </a-layout-header>

    <a-layout-content class="content">
      <a-tabs v-model:active-key="activeTab" class="main-tabs">
        <a-tab-pane key="workbench" title="工作台">
          <Workbench :active="activeTab === 'workbench'" />
        </a-tab-pane>
        <a-tab-pane key="tasks" title="任务管理">
          <TaskManagement :active="activeTab === 'tasks'" />
        </a-tab-pane>
        <a-tab-pane key="pending" title="待办">
          <PendingTasks :active="activeTab === 'pending'" />
        </a-tab-pane>
        <a-tab-pane key="settings" title="设置">
          <Settings />
        </a-tab-pane>
      </a-tabs>
    </a-layout-content>
  </a-layout>
</template>

<style scoped>
.layout {
  min-height: 100vh;
  background: #17171a;
}

.header {
  display: flex;
  align-items: center;
  padding: 0 24px;
  background: #232324;
  border-bottom: 1px solid #333;
  height: 60px;
}

.logo {
  display: flex;
  align-items: center;
  gap: 8px;
}

.logo-icon {
  font-size: 24px;
  color: #165DFF;
}

.logo-text {
  font-size: 20px;
  font-weight: 600;
  color: #fff;
}

.subtitle {
  margin-left: 16px;
  color: #86909c;
  font-size: 14px;
}

.content {
  padding: 24px;
  background: #17171a;
}

.main-tabs {
  background: #232324;
  border-radius: 8px;
  padding: 16px;
}

:deep(.arco-tabs-nav) {
  background: transparent;
}

:deep(.arco-tabs-content) {
  padding-top: 16px;
}
</style>
