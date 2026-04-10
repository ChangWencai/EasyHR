import { defineStore } from 'pinia'
import { ref } from 'vue'

export interface OrgInfo {
  id: number
  name: string
  credit_code: string
  city: string
  city_id?: number
  contact_name: string
  contact_phone: string
}

export interface UserInfo {
  id: number
  name: string
  phone: string
  role: string
}

export const useUserStore = defineStore('user', () => {
  const user = ref<UserInfo | null>(null)
  const org = ref<OrgInfo | null>(null)

  function setUser(u: UserInfo) {
    user.value = u
  }

  function setOrg(o: OrgInfo) {
    org.value = o
  }

  function clear() {
    user.value = null
    org.value = null
  }

  return { user, org, setUser, setOrg, clear }
})
