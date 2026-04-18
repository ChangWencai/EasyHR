<template>
  <div v-if="carousels.length > 0" class="home-carousel">
    <el-carousel
      :interval="4000"
      trigger="click"
      indicator-position="outside"
      type="card"
      height="160px"
      arrow="never"
      @change="onChange"
    >
      <el-carousel-item v-for="item in carousels" :key="item.id">
        <a :href="item.link_url || 'javascript:void(0)'" class="carousel-link" target="_blank">
          <img :src="item.image_url" :alt="'carousel-' + item.id" class="carousel-image" />
        </a>
      </el-carousel-item>
    </el-carousel>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { fetchCarousels, type CarouselItem } from '@/api/carousel'

const carousels = ref<CarouselItem[]>([])

async function load() {
  try {
    const res = await fetchCarousels()
    const now = new Date()
    carousels.value = (res.data || []).filter((item: CarouselItem) => {
      if (!item.active) return false
      const start = item.start_at ? new Date(item.start_at) : null
      const end = item.end_at ? new Date(item.end_at) : null
      if (start && now < start) return false
      if (end && now > end) return false
      return true
    })
  } catch {
    carousels.value = []
  }
}

function onChange(_index: number) {
  // No-op: carousel auto-advances
}

onMounted(load)
</script>

<style scoped lang="scss">
.home-carousel {
  margin-bottom: 16px;
  border-radius: 8px;
  overflow: hidden;
}

.carousel-link {
  display: block;
  width: 100%;
  height: 100%;
  text-decoration: none;
}

.carousel-image {
  width: 100%;
  height: 100%;
  object-fit: cover;
  border-radius: 8px;
}

:deep(.el-carousel__item--card) {
  border-radius: 8px;
}

:deep(.el-carousel__item) {
  background-color: #f5f7fa;
}

:deep(.el-carousel__indicator--horizontal) {
  padding: 4px 4px;
}

:deep(.el-carousel__button) {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  background-color: rgba(79, 110, 247, 0.3);
}

:deep(.el-carousel__indicator.is-active .el-carousel__button) {
  background-color: #4F6EF7;
}
</style>
