<script lang="ts" setup>
import { ref, computed, watch, onMounted } from 'vue'
import { use } from 'echarts/core'
import { CanvasRenderer } from 'echarts/renderers'
import { PieChart, BarChart, LineChart } from 'echarts/charts'
import {
  TitleComponent,
  TooltipComponent,
  LegendComponent,
  GridComponent
} from 'echarts/components'
import VChart from 'vue-echarts'
import { GetReportData } from '../../wailsjs/go/main/App'
import { main } from '../../wailsjs/go/models'
import { Message } from '@arco-design/web-vue'
import dayjs from 'dayjs'

// 注册 ECharts 组件
use([
  CanvasRenderer,
  PieChart,
  BarChart,
  LineChart,
  TitleComponent,
  TooltipComponent,
  LegendComponent,
  GridComponent
])

const props = defineProps<{
  active: boolean
}>()

// 状态
const loading = ref(false)
const dateRange = ref<string[]>([
  dayjs().subtract(6, 'day').format('YYYY-MM-DD'),
  dayjs().format('YYYY-MM-DD')
])
const reportData = ref<main.ReportData | null>(null)

// 快捷日期范围
const quickRanges = [
  { label: '7天', days: 7 },
  { label: '14天', days: 14 },
  { label: '30天', days: 30 },
  { label: '本月', type: 'month' }
]

// 饼图配置
const pieChartOption = computed(() => {
  if (!reportData.value?.project_stats?.length) {
    return null
  }

  const data = reportData.value.project_stats.map(item => ({
    name: item.project_name,
    value: Math.round(item.total_hours * 10) / 10,
    itemStyle: { color: item.color }
  }))

  return {
    backgroundColor: 'transparent',
    tooltip: {
      trigger: 'item',
      formatter: '{b}: {c}h ({d}%)'
    },
    legend: {
      orient: 'vertical',
      right: 20,
      top: 'center',
      textStyle: { color: '#c9cdd4' }
    },
    series: [
      {
        type: 'pie',
        radius: ['40%', '70%'],
        center: ['35%', '50%'],
        avoidLabelOverlap: true,
        itemStyle: {
          borderRadius: 4,
          borderColor: '#232324',
          borderWidth: 2
        },
        label: {
          show: true,
          color: '#c9cdd4',
          formatter: '{b}\n{d}%'
        },
        emphasis: {
          label: {
            show: true,
            fontSize: 14,
            fontWeight: 'bold'
          }
        },
        data
      }
    ]
  }
})

// 柱状图配置（每日任务统计）
const barChartOption = computed(() => {
  if (!reportData.value?.daily_stats?.length) {
    return null
  }

  const dates = reportData.value.daily_stats.map(d => d.date.slice(5)) // MM-DD
  const totalCounts = reportData.value.daily_stats.map(d => d.total_count)
  const completedCounts = reportData.value.daily_stats.map(d => d.completed_count)
  const completionRates = reportData.value.daily_stats.map(d =>
    Math.round(d.completion_rate)
  )

  return {
    backgroundColor: 'transparent',
    tooltip: {
      trigger: 'axis',
      axisPointer: { type: 'shadow' }
    },
    legend: {
      data: ['总任务', '已完成', '完成率'],
      textStyle: { color: '#c9cdd4' },
      top: 0
    },
    grid: {
      left: '3%',
      right: '4%',
      bottom: '3%',
      containLabel: true
    },
    xAxis: {
      type: 'category',
      data: dates,
      axisLine: { lineStyle: { color: '#4e5969' } },
      axisLabel: { color: '#86909c' }
    },
    yAxis: [
      {
        type: 'value',
        name: '任务数',
        axisLine: { lineStyle: { color: '#4e5969' } },
        axisLabel: { color: '#86909c' },
        splitLine: { lineStyle: { color: '#333' } }
      },
      {
        type: 'value',
        name: '完成率',
        max: 100,
        axisLine: { lineStyle: { color: '#4e5969' } },
        axisLabel: { color: '#86909c', formatter: '{value}%' },
        splitLine: { show: false }
      }
    ],
    series: [
      {
        name: '总任务',
        type: 'bar',
        data: totalCounts,
        itemStyle: { color: '#165DFF' }
      },
      {
        name: '已完成',
        type: 'bar',
        data: completedCounts,
        itemStyle: { color: '#00B42A' }
      },
      {
        name: '完成率',
        type: 'line',
        yAxisIndex: 1,
        data: completionRates,
        smooth: true,
        itemStyle: { color: '#FF7D00' },
        lineStyle: { width: 2 }
      }
    ]
  }
})

// 加载数据
const loadData = async () => {
  loading.value = true
  try {
    const result = await GetReportData(dateRange.value[0], dateRange.value[1])
    reportData.value = result
  } catch (err) {
    console.error('加载报表数据失败:', err)
    Message.error('加载报表数据失败')
  } finally {
    loading.value = false
  }
}

// 设置快捷日期范围
const setQuickRange = (range: typeof quickRanges[0]) => {
  if (range.type === 'month') {
    dateRange.value = [
      dayjs().startOf('month').format('YYYY-MM-DD'),
      dayjs().format('YYYY-MM-DD')
    ]
  } else if (range.days) {
    dateRange.value = [
      dayjs().subtract(range.days - 1, 'day').format('YYYY-MM-DD'),
      dayjs().format('YYYY-MM-DD')
    ]
  }
}

// 监听日期变化
watch(dateRange, () => {
  loadData()
})

// 监听 tab 激活
watch(() => props.active, (isActive) => {
  if (isActive) {
    loadData()
  }
})

onMounted(() => {
  if (props.active) {
    loadData()
  }
})
</script>

<template>
  <div class="report">
    <!-- 工具栏 -->
    <div class="toolbar">
      <div class="toolbar-left">
        <a-range-picker
          v-model="dateRange"
          style="width: 260px"
          :allow-clear="false"
        />
        <div class="quick-ranges">
          <a-tag
            v-for="range in quickRanges"
            :key="range.label"
            class="quick-tag"
            @click="setQuickRange(range)"
          >
            {{ range.label }}
          </a-tag>
        </div>
      </div>
      <a-button type="text" @click="loadData">
        <template #icon><icon-refresh /></template>
        刷新
      </a-button>
    </div>

    <a-spin :loading="loading" style="display: block; width: 100%;">
      <!-- 统计概览 -->
      <a-row :gutter="16" class="summary-row" v-if="reportData?.summary">
        <a-col :span="6">
          <a-card class="stat-card">
            <a-statistic title="总任务数" :value="reportData.summary.total_tasks">
              <template #suffix>个</template>
            </a-statistic>
          </a-card>
        </a-col>
        <a-col :span="6">
          <a-card class="stat-card">
            <a-statistic title="已完成" :value="reportData.summary.completed_tasks">
              <template #suffix>个</template>
            </a-statistic>
          </a-card>
        </a-col>
        <a-col :span="6">
          <a-card class="stat-card">
            <a-statistic title="总工时" :value="reportData.summary.total_hours" :precision="1">
              <template #suffix>h</template>
            </a-statistic>
          </a-card>
        </a-col>
        <a-col :span="6">
          <a-card class="stat-card">
            <a-statistic title="平均完成率" :value="reportData.summary.average_rate" :precision="1">
              <template #suffix>%</template>
            </a-statistic>
          </a-card>
        </a-col>
      </a-row>

      <!-- 图表区域 -->
      <a-row :gutter="16" class="charts-row">
        <!-- 项目时间占比饼图 -->
        <a-col :span="12">
          <a-card title="项目时间占比" class="chart-card">
            <div v-if="pieChartOption" class="chart-container">
              <v-chart :option="pieChartOption" autoresize />
            </div>
            <a-empty v-else description="暂无数据" />
          </a-card>
        </a-col>

        <!-- 每日任务统计柱状图 -->
        <a-col :span="12">
          <a-card title="每日任务统计" class="chart-card">
            <div v-if="barChartOption" class="chart-container">
              <v-chart :option="barChartOption" autoresize />
            </div>
            <a-empty v-else description="暂无数据" />
          </a-card>
        </a-col>
      </a-row>

      <!-- 项目明细表格 -->
      <a-card title="项目工时明细" class="detail-card" v-if="reportData?.project_stats?.length">
        <a-table
          :data="reportData.project_stats"
          :pagination="false"
          row-key="project_id"
        >
          <template #columns>
            <a-table-column title="项目" data-index="project_name">
              <template #cell="{ record }">
                <span :style="{ color: record.color }">{{ record.project_name }}</span>
              </template>
            </a-table-column>
            <a-table-column title="任务数" data-index="task_count" />
            <a-table-column title="总工时" data-index="total_hours">
              <template #cell="{ record }">
                {{ record.total_hours.toFixed(1) }}h
              </template>
            </a-table-column>
            <a-table-column title="占比">
              <template #cell="{ record }">
                <a-progress
                  :percent="record.percentage / 100"
                  :stroke-width="8"
                  :show-text="false"
                  :color="record.color"
                  style="width: 100px; display: inline-flex; margin-right: 8px;"
                />
                {{ record.percentage.toFixed(1) }}%
              </template>
            </a-table-column>
          </template>
        </a-table>
      </a-card>
    </a-spin>
  </div>
</template>

<style scoped>
.report {
  padding: 0;
  width: 100%;
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

.quick-ranges {
  display: flex;
  gap: 8px;
}

.quick-tag {
  cursor: pointer;
  transition: all 0.2s;
}

.quick-tag:hover {
  color: #165DFF;
  border-color: #165DFF;
}

.summary-row {
  margin-bottom: 16px;
}

.stat-card {
  background: #2a2a2b;
}

.charts-row {
  margin-bottom: 16px;
}

.chart-card {
  background: #2a2a2b;
  height: 100%;
}

.chart-container {
  height: 320px;
}

.detail-card {
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
