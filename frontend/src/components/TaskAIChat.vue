<script lang="ts" setup>
import { ref, computed, watch, nextTick, onMounted, onUnmounted } from 'vue'
import {
  StartConversation,
  SendMessage,
  GetConversationDetail,
  GetTaskConversations,
  StopConversation,
  GetEnabledAgents,
  GetConversationSteps
} from '../../wailsjs/go/main/App'
import { main } from '../../wailsjs/go/models'
import { Message } from '@arco-design/web-vue'

const props = defineProps<{
  visible: boolean
  task: main.Task | null
}>()

const emit = defineEmits<{
  (e: 'update:visible', value: boolean): void
}>()

const modalVisible = computed({
  get: () => props.visible,
  set: (val) => emit('update:visible', val)
})

// 数据状态
const agents = ref<main.Agent[]>([])
const conversations = ref<main.TaskConversation[]>([])
const currentConversation = ref<main.ConversationDetail | null>(null)
const currentSteps = ref<main.AgentStep[]>([])
const selectedAgentId = ref<number | null>(null)
const userInput = ref('')
const extraContext = ref('')
const loading = ref(false)
const sending = ref(false)
const messagesRef = ref<HTMLElement | null>(null)
const showSteps = ref(true) // 是否显示执行步骤

// 轮询定时器
let pollTimer: number | null = null

// 加载可用的Agent
const loadAgents = async () => {
  try {
    const result = await GetEnabledAgents()
    agents.value = result || []
    if (agents.value.length > 0 && !selectedAgentId.value) {
      selectedAgentId.value = agents.value[0].id
    }
  } catch (err) {
    console.error('加载Agent失败:', err)
  }
}

// 加载任务的会话历史
const loadConversations = async () => {
  if (!props.task) return
  try {
    const result = await GetTaskConversations(props.task.id)
    conversations.value = result || []
  } catch (err) {
    console.error('加载会话历史失败:', err)
  }
}

// 加载会话详情
const loadConversationDetail = async (convId: number) => {
  loading.value = true
  try {
    const result = await GetConversationDetail(convId)
    currentConversation.value = result
    // 同时加载执行步骤
    await loadConversationSteps(convId)
    scrollToBottom()
  } catch (err) {
    console.error('加载会话详情失败:', err)
    Message.error('加载会话失败')
  } finally {
    loading.value = false
  }
}

// 加载会话的执行步骤
const loadConversationSteps = async (convId: number) => {
  try {
    const result = await GetConversationSteps(convId)
    currentSteps.value = result || []
  } catch (err) {
    console.error('加载执行步骤失败:', err)
    currentSteps.value = []
  }
}

// 开始新会话
const startNewConversation = async () => {
  if (!props.task || !selectedAgentId.value) {
    Message.warning('请选择一个Agent')
    return
  }

  loading.value = true
  try {
    const input: main.StartConversationInput = {
      task_id: props.task.id,
      agent_id: selectedAgentId.value,
      extra_context: extraContext.value
    }
    const result = await StartConversation(input)
    currentConversation.value = result
    extraContext.value = ''
    await loadConversations()
    startPolling()
    scrollToBottom()
  } catch (err) {
    console.error('开始会话失败:', err)
    Message.error('开始会话失败')
  } finally {
    loading.value = false
  }
}

// 发送消息
const sendMessage = async () => {
  if (!currentConversation.value || !userInput.value.trim()) return

  sending.value = true
  const message = userInput.value.trim()
  userInput.value = ''

  try {
    const input: main.SendMessageInput = {
      conversation_id: currentConversation.value.conversation.id,
      content: message
    }
    const result = await SendMessage(input)
    currentConversation.value = result
    startPolling()
    scrollToBottom()
  } catch (err) {
    console.error('发送消息失败:', err)
    Message.error('发送失败')
    userInput.value = message // 恢复输入
  } finally {
    sending.value = false
  }
}

// 停止会话
const stopCurrentConversation = async () => {
  if (!currentConversation.value) return

  try {
    await StopConversation(currentConversation.value.conversation.id)
    await loadConversationDetail(currentConversation.value.conversation.id)
    stopPolling()
    Message.success('会话已停止')
  } catch (err) {
    console.error('停止会话失败:', err)
  }
}

// 选择快捷回复
const selectOption = (option: string) => {
  userInput.value = option
  sendMessage()
}

// 开始轮询
const startPolling = () => {
  stopPolling()
  pollTimer = window.setInterval(async () => {
    if (!currentConversation.value) return

    const status = currentConversation.value.conversation.status
    if (status === 'active') {
      await loadConversationDetail(currentConversation.value.conversation.id)
    } else {
      stopPolling()
    }
  }, 2000)
}

// 停止轮询
const stopPolling = () => {
  if (pollTimer) {
    clearInterval(pollTimer)
    pollTimer = null
  }
}

// 滚动到底部
const scrollToBottom = () => {
  nextTick(() => {
    if (messagesRef.value) {
      messagesRef.value.scrollTop = messagesRef.value.scrollHeight
    }
  })
}

// 返回会话列表
const backToList = () => {
  stopPolling()
  currentConversation.value = null
}

// 格式化时间
const formatTime = (date: string) => {
  return new Date(date).toLocaleString('zh-CN', {
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit'
  })
}

// 获取状态文本
const getStatusText = (status: string) => {
  switch (status) {
    case 'active': return '处理中'
    case 'waiting_user': return '等待回复'
    case 'completed': return '已完成'
    case 'failed': return '失败'
    default: return status
  }
}

// 获取状态颜色
const getStatusColor = (status: string) => {
  switch (status) {
    case 'active': return '#165DFF'
    case 'waiting_user': return '#FF7D00'
    case 'completed': return '#00B42A'
    case 'failed': return '#F53F3F'
    default: return '#86909c'
  }
}

// 解析消息选项
const parseOptions = (metadata: string): string[] => {
  try {
    const data = JSON.parse(metadata)
    return data.options || []
  } catch {
    return []
  }
}

// 解析消息元数据
const parseMetadata = (metadata: string): { step_num?: number; action?: string; tool?: string; success?: boolean } => {
  try {
    return JSON.parse(metadata)
  } catch {
    return {}
  }
}

// 获取步骤状态图标
const getStepStatusIcon = (status: string) => {
  switch (status) {
    case 'success': return 'icon-check-circle'
    case 'failed': return 'icon-close-circle'
    case 'running': return 'icon-loading'
    default: return 'icon-clock-circle'
  }
}

// 获取步骤状态颜色
const getStepStatusColor = (status: string) => {
  switch (status) {
    case 'success': return '#00B42A'
    case 'failed': return '#F53F3F'
    case 'running': return '#165DFF'
    default: return '#86909c'
  }
}

// 获取工具名称的显示文本
const getToolDisplayName = (tool: string) => {
  const names: Record<string, string> = {
    'claude_code': 'Claude Code',
    'shell': 'Shell',
    'read_file': '读取文件',
    'write_file': '写入文件',
    'list_files': '列出文件',
    'ask_user': '询问用户',
    'complete': '完成'
  }
  return names[tool] || tool
}

// 监听可见性变化
watch(() => props.visible, async (visible) => {
  if (visible && props.task) {
    await loadAgents()
    await loadConversations()
    currentConversation.value = null
  } else {
    stopPolling()
  }
})

// 监听当前会话状态
watch(() => currentConversation.value?.conversation.status, (status) => {
  if (status === 'active') {
    startPolling()
  } else {
    stopPolling()
  }
})

onUnmounted(() => {
  stopPolling()
})
</script>

<template>
  <a-modal
    v-model:visible="modalVisible"
    :title="task ? `AI 跟进: ${task.name}` : 'AI 跟进'"
    :width="700"
    :footer="false"
    class="ai-chat-modal"
  >
    <div class="ai-chat-container">
      <!-- 会话列表视图 -->
      <template v-if="!currentConversation">
        <!-- 新建会话 -->
        <div class="new-conversation">
          <a-form layout="vertical">
            <a-form-item label="选择 Agent">
              <a-select v-model="selectedAgentId" placeholder="选择一个AI助手">
                <a-option v-for="agent in agents" :key="agent.id" :value="agent.id">
                  {{ agent.name }}
                  <span class="agent-desc">{{ agent.description }}</span>
                </a-option>
              </a-select>
            </a-form-item>
            <a-form-item label="补充说明（可选）">
              <a-textarea
                v-model="extraContext"
                placeholder="提供额外的上下文信息，帮助AI更好地理解任务..."
                :auto-size="{ minRows: 2, maxRows: 4 }"
              />
            </a-form-item>
            <a-button type="primary" :loading="loading" @click="startNewConversation" long>
              <template #icon><icon-robot /></template>
              开始 AI 会话
            </a-button>
          </a-form>
        </div>

        <!-- 历史会话 -->
        <div v-if="conversations.length > 0" class="conversation-history">
          <div class="section-title">历史会话</div>
          <div class="conversation-list">
            <div
              v-for="conv in conversations"
              :key="conv.id"
              class="conversation-item"
              @click="loadConversationDetail(conv.id)"
            >
              <div class="conv-info">
                <span class="conv-agent">{{ conv.agent_name }}</span>
                <span class="conv-time">{{ formatTime(conv.created_at) }}</span>
              </div>
              <a-tag :color="getStatusColor(conv.status)" size="small">
                {{ getStatusText(conv.status) }}
              </a-tag>
            </div>
          </div>
        </div>
      </template>

      <!-- 会话详情视图 -->
      <template v-else>
        <!-- 头部 -->
        <div class="chat-header">
          <a-button type="text" size="small" @click="backToList">
            <template #icon><icon-left /></template>
            返回
          </a-button>
          <div class="header-info">
            <span class="agent-name">{{ currentConversation.conversation.agent_name }}</span>
            <a-tag :color="getStatusColor(currentConversation.conversation.status)" size="small">
              {{ getStatusText(currentConversation.conversation.status) }}
            </a-tag>
          </div>
          <div class="header-actions">
            <a-button
              v-if="currentSteps.length > 0"
              type="text"
              size="small"
              @click="showSteps = !showSteps"
            >
              <template #icon><icon-list /></template>
              {{ showSteps ? '隐藏步骤' : '显示步骤' }}
            </a-button>
            <a-button
              v-if="currentConversation.conversation.status === 'active'"
              type="text"
              status="danger"
              size="small"
              @click="stopCurrentConversation"
            >
              停止
            </a-button>
          </div>
        </div>

        <!-- 执行步骤时间线 -->
        <div v-if="showSteps && currentSteps.length > 0" class="steps-timeline">
          <div class="steps-title">
            <icon-thunderbolt />
            执行步骤 ({{ currentSteps.length }})
          </div>
          <div class="steps-list">
            <div
              v-for="step in currentSteps"
              :key="step.id"
              class="step-item"
              :class="step.status"
            >
              <div class="step-indicator">
                <component
                  :is="getStepStatusIcon(step.status)"
                  :style="{ color: getStepStatusColor(step.status) }"
                />
              </div>
              <div class="step-content">
                <div class="step-header">
                  <span class="step-num">步骤 {{ step.step_num }}</span>
                  <a-tag size="small" :color="getStepStatusColor(step.status)">
                    {{ getToolDisplayName(step.action) }}
                  </a-tag>
                </div>
                <div class="step-thought">{{ step.thought }}</div>
                <div v-if="step.observation" class="step-observation">
                  <div class="observation-label">执行结果:</div>
                  <pre class="observation-content">{{ step.observation.slice(0, 200) }}{{ step.observation.length > 200 ? '...' : '' }}</pre>
                </div>
                <div v-if="step.error" class="step-error">
                  <icon-exclamation-circle />
                  {{ step.error }}
                </div>
              </div>
            </div>
          </div>
        </div>

        <!-- 消息列表 -->
        <div ref="messagesRef" class="messages-container" :class="{ loading, 'with-steps': showSteps && currentSteps.length > 0 }">
          <div
            v-for="msg in currentConversation.messages"
            :key="msg.id"
            class="message"
            :class="[msg.role, msg.type]"
          >
            <!-- 系统消息（上下文/工具执行结果） -->
            <template v-if="msg.role === 'system'">
              <div class="system-message" :class="{ 'tool-result': msg.type === 'result' }">
                <template v-if="msg.type === 'result'">
                  <icon-code />
                  <span class="tool-info">
                    <template v-if="parseMetadata(msg.metadata).tool">
                      工具执行: {{ getToolDisplayName(parseMetadata(msg.metadata).tool || '') }}
                      <a-tag
                        size="small"
                        :color="parseMetadata(msg.metadata).success ? '#00B42A' : '#F53F3F'"
                      >
                        {{ parseMetadata(msg.metadata).success ? '成功' : '失败' }}
                      </a-tag>
                    </template>
                    <template v-else>
                      工具执行结果
                    </template>
                  </span>
                </template>
                <template v-else>
                  <icon-info-circle />
                  <span>任务上下文已发送</span>
                </template>
              </div>
            </template>

            <!-- 用户消息 -->
            <template v-else-if="msg.role === 'user'">
              <div class="user-message">
                <div class="message-content">{{ msg.content }}</div>
              </div>
            </template>

            <!-- AI消息 -->
            <template v-else>
              <div class="assistant-message">
                <div class="message-avatar">
                  <icon-robot />
                </div>
                <div class="message-body">
                  <!-- 错误消息 -->
                  <div v-if="msg.type === 'error'" class="error-content">
                    {{ msg.content }}
                  </div>
                  <!-- 问题消息（带选项） -->
                  <template v-else-if="msg.type === 'question'">
                    <div class="message-content">{{ msg.content }}</div>
                    <div v-if="parseOptions(msg.metadata).length > 0" class="options">
                      <a-button
                        v-for="opt in parseOptions(msg.metadata)"
                        :key="opt"
                        size="small"
                        @click="selectOption(opt)"
                      >
                        {{ opt }}
                      </a-button>
                    </div>
                  </template>
                  <!-- 结果消息 -->
                  <div v-else-if="msg.type === 'result'" class="result-content">
                    <icon-check-circle class="result-icon" />
                    {{ msg.content }}
                  </div>
                  <!-- 普通消息 -->
                  <div v-else class="message-content">{{ msg.content }}</div>
                </div>
              </div>
            </template>
          </div>

          <!-- 加载中指示器 -->
          <div v-if="currentConversation.conversation.status === 'active'" class="typing-indicator">
            <span></span><span></span><span></span>
          </div>
        </div>

        <!-- 输入区域 -->
        <div
          v-if="currentConversation.conversation.status === 'waiting_user'"
          class="input-area"
        >
          <a-input
            v-model="userInput"
            placeholder="输入回复..."
            @press-enter="sendMessage"
            :disabled="sending"
          >
            <template #suffix>
              <a-button type="primary" size="small" :loading="sending" @click="sendMessage">
                发送
              </a-button>
            </template>
          </a-input>
        </div>

        <!-- 已完成提示 -->
        <div v-else-if="currentConversation.conversation.status === 'completed'" class="completed-tip">
          <icon-check-circle-fill class="completed-icon" />
          会话已完成
        </div>

        <!-- 失败提示 -->
        <div v-else-if="currentConversation.conversation.status === 'failed'" class="failed-tip">
          <icon-close-circle-fill class="failed-icon" />
          会话已终止
        </div>
      </template>
    </div>
  </a-modal>
</template>

<style scoped>
.ai-chat-container {
  min-height: 400px;
  display: flex;
  flex-direction: column;
}

.new-conversation {
  padding: 16px 0;
}

.agent-desc {
  color: #86909c;
  font-size: 12px;
  margin-left: 8px;
}

.section-title {
  font-size: 14px;
  color: #86909c;
  margin: 24px 0 12px;
  padding-bottom: 8px;
  border-bottom: 1px solid #333;
}

.conversation-list {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.conversation-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 12px;
  background: #2a2a2b;
  border-radius: 6px;
  cursor: pointer;
  transition: background 0.2s;
}

.conversation-item:hover {
  background: #333;
}

.conv-info {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.conv-agent {
  font-weight: 500;
}

.conv-time {
  font-size: 12px;
  color: #86909c;
}

/* 聊天界面 */
.chat-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding-bottom: 12px;
  border-bottom: 1px solid #333;
  margin-bottom: 12px;
}

.header-info {
  display: flex;
  align-items: center;
  gap: 8px;
}

.agent-name {
  font-weight: 500;
}

.header-actions {
  display: flex;
  gap: 8px;
}

/* 执行步骤时间线 */
.steps-timeline {
  margin-bottom: 12px;
  padding: 12px;
  background: #232324;
  border-radius: 8px;
  max-height: 200px;
  overflow-y: auto;
}

.steps-title {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 13px;
  font-weight: 500;
  color: #86909c;
  margin-bottom: 12px;
}

.steps-list {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.step-item {
  display: flex;
  gap: 10px;
  padding: 8px;
  background: #2a2a2b;
  border-radius: 6px;
  font-size: 12px;
}

.step-item.running {
  border-left: 2px solid #165DFF;
}

.step-item.success {
  border-left: 2px solid #00B42A;
}

.step-item.failed {
  border-left: 2px solid #F53F3F;
}

.step-indicator {
  flex-shrink: 0;
  width: 20px;
  display: flex;
  justify-content: center;
  padding-top: 2px;
}

.step-content {
  flex: 1;
  min-width: 0;
}

.step-header {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 4px;
}

.step-num {
  font-weight: 500;
  color: #e5e5e5;
}

.step-thought {
  color: #a0a0a0;
  line-height: 1.4;
  margin-bottom: 6px;
}

.step-observation {
  margin-top: 6px;
  padding: 6px 8px;
  background: #1e1e1f;
  border-radius: 4px;
}

.observation-label {
  font-size: 11px;
  color: #86909c;
  margin-bottom: 4px;
}

.observation-content {
  font-family: monospace;
  font-size: 11px;
  color: #a0a0a0;
  white-space: pre-wrap;
  word-break: break-all;
  margin: 0;
}

.step-error {
  display: flex;
  align-items: center;
  gap: 4px;
  color: #F53F3F;
  font-size: 11px;
  margin-top: 4px;
}

.messages-container {
  flex: 1;
  overflow-y: auto;
  padding: 12px 0;
  max-height: 350px;
  min-height: 200px;
}

.messages-container.with-steps {
  max-height: 200px;
  min-height: 150px;
}

.message {
  margin-bottom: 16px;
}

.system-message {
  display: flex;
  align-items: center;
  gap: 8px;
  justify-content: center;
  color: #86909c;
  font-size: 12px;
  padding: 8px;
  background: #2a2a2b;
  border-radius: 6px;
}

.system-message.tool-result {
  background: #1e2632;
  border: 1px solid #2a3a4a;
}

.tool-info {
  display: flex;
  align-items: center;
  gap: 8px;
}

.user-message {
  display: flex;
  justify-content: flex-end;
}

.user-message .message-content {
  background: #165DFF;
  color: #fff;
  padding: 10px 14px;
  border-radius: 12px 12px 2px 12px;
  max-width: 80%;
}

.assistant-message {
  display: flex;
  gap: 10px;
}

.message-avatar {
  width: 32px;
  height: 32px;
  background: #165DFF;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.message-body {
  flex: 1;
}

.message-body .message-content {
  background: #2a2a2b;
  padding: 10px 14px;
  border-radius: 2px 12px 12px 12px;
  white-space: pre-wrap;
  line-height: 1.6;
}

.error-content {
  background: rgba(245, 63, 63, 0.1);
  color: #F53F3F;
  padding: 10px 14px;
  border-radius: 6px;
}

.result-content {
  background: rgba(0, 180, 42, 0.1);
  color: #00B42A;
  padding: 10px 14px;
  border-radius: 6px;
  display: flex;
  align-items: flex-start;
  gap: 8px;
}

.result-icon {
  flex-shrink: 0;
  margin-top: 2px;
}

.options {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  margin-top: 12px;
}

/* 输入加载动画 */
.typing-indicator {
  display: flex;
  gap: 4px;
  padding: 12px 16px;
  background: #2a2a2b;
  border-radius: 12px;
  width: fit-content;
  margin-left: 42px;
}

.typing-indicator span {
  width: 8px;
  height: 8px;
  background: #86909c;
  border-radius: 50%;
  animation: typing 1.4s infinite ease-in-out;
}

.typing-indicator span:nth-child(1) { animation-delay: 0s; }
.typing-indicator span:nth-child(2) { animation-delay: 0.2s; }
.typing-indicator span:nth-child(3) { animation-delay: 0.4s; }

@keyframes typing {
  0%, 60%, 100% { transform: translateY(0); }
  30% { transform: translateY(-4px); }
}

/* 输入区域 */
.input-area {
  padding-top: 12px;
  border-top: 1px solid #333;
  margin-top: auto;
}

.completed-tip,
.failed-tip {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  padding: 16px;
  border-top: 1px solid #333;
  margin-top: auto;
}

.completed-icon {
  color: #00B42A;
  font-size: 18px;
}

.failed-icon {
  color: #F53F3F;
  font-size: 18px;
}

:deep(.arco-modal-body) {
  padding: 16px 20px;
}
</style>
