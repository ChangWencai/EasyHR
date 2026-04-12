import axios, { type AxiosError } from 'axios'
import { ElMessage } from 'element-plus'
import router from '@/router'

const request = axios.create({
  baseURL: import.meta.env.VITE_API_BASE_URL || '/api/v1',
  timeout: 10000,
})

const PUBLIC_AUTH_PATHS = ['/auth/send-code', '/auth/login', '/auth/register', '/auth/login/password', '/auth/refresh']

interface ApiError {
  response?: {
    status?: number
    data?: {
      code?: number
      message?: string
    }
  }
}

request.interceptors.request.use((config) => {
  const isPublic = PUBLIC_AUTH_PATHS.some((p) => config.url?.startsWith(p))
  if (!isPublic) {
    const token = localStorage.getItem('token')
    if (token) {
      config.headers.Authorization = `Bearer ${token}`
    }
  }
  return config
})

request.interceptors.response.use(
  (response) => response.data,
  (error: AxiosError) => {
    const err = error as unknown as ApiError
    if (err.response?.status === 401) {
      localStorage.removeItem('token')
      ElMessage.error('登录已过期，请重新登录')
      router.push('/login')
    }
    return Promise.reject(error)
  },
)

export default request
