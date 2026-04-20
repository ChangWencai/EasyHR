import { defineStore } from 'pinia'
import { ref, computed } from 'vue'

export const useAuthStore = defineStore('auth', () => {
  const token = ref<string | null>(localStorage.getItem('token'))

  const isLoggedIn = computed(() => !!token.value)

  /** Decode org_id from JWT payload (base64url → JSON) */
  const orgId = computed<number | null>(() => {
    if (!token.value) return null
    try {
      const payload = token.value.split('.')[1]
      const base64 = payload.replace(/-/g, '+').replace(/_/g, '/')
      const json = atob(base64)
      const claims = JSON.parse(json) as { org_id?: number }
      return claims.org_id ?? null
    } catch {
      return null
    }
  })

  function setToken(newToken: string) {
    token.value = newToken
    localStorage.setItem('token', newToken)
  }

  function logout() {
    token.value = null
    localStorage.removeItem('token')
  }

  return { token, isLoggedIn, orgId, setToken, logout }
})
