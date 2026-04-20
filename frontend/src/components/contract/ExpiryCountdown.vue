<script setup lang="ts">
import { computed } from 'vue'

const props = defineProps<{
  days: number
}>()

const config = computed(() => {
  if (props.days > 30) return { color: 'var(--success)', label: `${props.days} 天后到期` }
  if (props.days >= 8) return { color: 'var(--warning)', label: `还有 ${props.days} 天到期` }
  if (props.days >= 1) return { color: '#EF4444', label: `即将到期：${props.days} 天` }
  if (props.days === 0) return { color: 'var(--danger)', label: '今天到期' }
  return { color: 'var(--danger)', label: `已过期 ${Math.abs(props.days)} 天` }
})
</script>

<template>
  <span class="expiry-countdown" :style="{ color: config.color }">
    {{ config.label }}
  </span>
</template>
