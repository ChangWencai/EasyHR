<template>
  <div class="ring-chart-wrapper">
    <div v-if="loading" class="ring-loading">
      <el-icon class="is-loading" size="20"><Loading /></el-icon>
    </div>
    <div v-else-if="!hasData" class="ring-empty">暂无数据</div>
    <div v-else class="ring-content">
      <v-chart :option="option" autoresize style="height: 200px" />
      <div class="ring-center-label">
        <div class="center-percent">{{ stats.percent }}%</div>
        <div class="center-count">{{ stats.completed }}/{{ stats.total }}</div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, ref, onMounted } from 'vue'
import VChart from 'vue-echarts'
import { use } from 'echarts/core'
import { PieChart } from 'echarts/charts'
import { TooltipComponent, LegendComponent } from 'echarts/components'
import { CanvasRenderer } from 'echarts/renderers'
import { Loading } from '@element-plus/icons-vue'
import { fetchTodoRingStats } from '@/api/dashboard'

use([PieChart, TooltipComponent, LegendComponent, CanvasRenderer])

interface Props {
  /** 'all': 全部事项, 'time-limited': 限时任务 */
  type: 'all' | 'time-limited'
}

const props = defineProps<Props>()
const loading = ref(false)
const stats = ref({ completed: 0, pending: 0, total: 0, percent: 0 })

const hasData = computed(() => stats.value.total > 0)

const option = computed(() => ({
  tooltip: { trigger: 'item', formatter: '{b}: {c} ({d}%)' },
  series: [{
    type: 'pie',
    radius: ['40%', '70%'],
    center: ['50%', '50%'],
    avoidLabelOverlap: false,
    label: { show: false },
    emphasis: {
      label: { show: false },
    },
    data: [
      {
        value: stats.value.completed,
        name: '已完成',
        itemStyle: { color: '#4F6EF7' },
      },
      {
        value: stats.value.pending,
        name: '待办',
        itemStyle: { color: '#E8ECF0' },
      },
    ],
  }],
}))

async function load() {
  loading.value = true
  try {
    const data = await fetchTodoRingStats(props.type)
    stats.value = data.stats
  } finally {
    loading.value = false
  }
}

onMounted(load)
</script>

<style scoped lang="scss">
.ring-chart-wrapper {
  display: flex;
  flex-direction: column;
  align-items: center;
  position: relative;
  min-height: 200px;
}

.ring-content {
  position: relative;
  width: 100%;
}

.ring-center-label {
  position: absolute;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%);
  text-align: center;
  pointer-events: none;
}

.center-percent {
  font-size: 22px;
  font-weight: 700;
  color: #172B4D;
  line-height: 1.2;
}

.center-count {
  font-size: 13px;
  color: #97A0AF;
  margin-top: 2px;
}

.ring-loading,
.ring-empty {
  height: 200px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #97A0AF;
  font-size: 13px;
}
</style>
