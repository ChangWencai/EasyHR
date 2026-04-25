<template>
  <div class="compliance-table-wrapper glass-card">
    <el-table
      :data="data"
      stripe
      class="compliance-table"
      :header-cell-style="{ background: '#F9FAFB', color: '#374151', fontWeight: 600 }"
      :row-class-name="rowClassName"
    >
      <slot />
    </el-table>
    <div class="pagination-wrapper">
      <el-pagination
        v-model:current-page="currentPage"
        :page-size="pageSize"
        :total="total"
        layout="total, prev, pager, next"
        @current-change="onPageChange"
      />
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'

const props = defineProps<{
  data: any[]
  total: number
  page: number
  pageSize: number
  rowClassName?: (row: any) => string
}>()
const emit = defineEmits<{ 'page-change': [page: number] }>()

const currentPage = computed({
  get: () => props.page,
  set: (v) => emit('page-change', v)
})

function onPageChange(p: number) {
  emit('page-change', p)
}
</script>

<style scoped lang="scss">
.compliance-table-wrapper { padding: 0; overflow: hidden; }
.pagination-wrapper { display: flex; justify-content: flex-end; padding: 16px 20px; border-top: 1px solid var(--border); }
</style>
