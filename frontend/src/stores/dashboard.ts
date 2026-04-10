import { defineStore } from 'pinia'
import { ref } from 'vue'
import { fetchDashboard, type TodoItem, type DashboardOverview } from '@/api/dashboard'

export const useDashboardStore = defineStore('dashboard', () => {
  const todos = ref<TodoItem[]>([])
  const overview = ref<DashboardOverview | null>(null)
  const loading = ref(false)
  const overviewExpanded = ref(true)

  async function load() {
    loading.value = true
    try {
      const data = await fetchDashboard()
      todos.value = data.todos
      overview.value = data.overview
    } finally {
      loading.value = false
    }
  }

  function toggleOverview() {
    overviewExpanded.value = !overviewExpanded.value
  }

  function removeTodo(type: string) {
    todos.value = todos.value.filter((t) => t.type !== type)
  }

  return { todos, overview, loading, overviewExpanded, load, toggleOverview, removeTodo }
})
