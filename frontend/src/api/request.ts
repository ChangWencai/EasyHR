import axios, { type AxiosError } from 'axios'
import router from '@/router'
import { useMessage } from '@/composables/useMessage'

const request = axios.create({
  baseURL: import.meta.env.VITE_API_BASE_URL || '/api/v1',
  timeout: 10000,
})

const $msg = useMessage()

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

// Error code → user-friendly message mapping (D-10-12)
const ERROR_MESSAGES: Record<number, string> = {
  400: '请求参数错误，请检查输入',
  401: '登录已过期，请重新登录',  // handled separately with redirect
  403: '您没有权限进行此操作',
  404: '请求的数据不存在',
  409: '数据冲突，请刷新后重试',
  422: '数据验证失败，请检查输入',
  500: '服务器异常，请稍后重试',
  502: '服务暂时不可用，请稍后重试',
  503: '服务暂时不可用，请稍后重试',
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
  async (error: AxiosError) => {
    const err = error as unknown as ApiError
    const status = err.response?.status

    // 401: redirect to login (per existing pattern)
    if (status === 401) {
      localStorage.removeItem('token')
      $msg.error('登录已过期，请重新登录')
      router.push('/login')
      return Promise.reject(error)
    }

    // Determine user message
    let userMessage = ERROR_MESSAGES[status ?? 0] ?? '操作失败，请稍后重试'

    // Network errors (no response)
    if (!status) {
      if (error.code === 'ECONNABORTED' || error.message?.includes('timeout')) {
        userMessage = '请求超时，请稍后重试'
      } else {
        userMessage = '网络连接失败，请检查网络后重试'
      }
    }

    // Show error with action guidance for retryable errors (D-10-13)
    const retryable = [500, 502, 503, 'ECONNABORTED', 'ERR_NETWORK'].includes(
      status ?? (error.code as string),
    )

    if (retryable) {
      $msg.error(userMessage, { showActions: true })
    } else {
      $msg.error(userMessage)
    }

    return Promise.reject(error)
  },
)

export default request
