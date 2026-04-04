import axios from 'axios'
import { ElMessage } from 'element-plus'
import { useAuthStore } from '@/stores/auth'
import router from '@/router'
import type { Session, Message, Stats, MockConfig, Receiver, SessionDetail } from '@/types'

const api = axios.create({
  baseURL: '/api',
  timeout: 10000
})

// Request interceptor - attach auth token
api.interceptors.request.use((config) => {
  const token = localStorage.getItem('token')
  if (token) {
    config.headers.Authorization = `Bearer ${token}`
  }
  return config
})

// Response interceptor - handle errors globally
api.interceptors.response.use(
  (response) => response,
  (error) => {
    const status = error.response?.status
    const message = error.response?.data?.error || error.message || 'Unknown error'

    if (status === 401) {
      const authStore = useAuthStore()
      authStore.logout()
      router.push('/login')
      ElMessage.error('登录已过期，请重新登录')
    } else if (status === 403) {
      ElMessage.error('没有权限执行此操作')
    } else if (status === 404) {
      ElMessage.error('请求的资源不存在')
    } else if (status === 429) {
      ElMessage.error('请求过于频繁，请稍后再试')
    } else if (status >= 500) {
      ElMessage.error('服务器错误，请稍后再试')
    } else if (error.code === 'ECONNABORTED') {
      ElMessage.error('请求超时，请检查网络连接')
    } else if (!status) {
      ElMessage.error('网络连接失败，请检查网络')
    } else {
      ElMessage.error(message)
    }

    return Promise.reject(error)
  }
)

// Session API
export const sessionApi = {
  list: () => api.get<{ data: Session[], total: number }>('/sessions'),
  getStats: (id: string) => api.get<SessionDetail>(`/sessions/${id}/stats`),
  delete: (id: string) => api.delete(`/sessions/${id}`)
}

// Message API
export const messageApi = {
  list: (params: {
    session_id?: string
    status?: string
    source_addr?: string
    dest_addr?: string
    start_time?: string
    end_time?: string
    page?: number
    page_size?: number
  }) => api.get<{ data: Message[], total: number, page: number, page_size: number }>('/messages', { params }),
  get: (id: string) => api.get<Message>(`/messages/${id}`),
  deliver: (id: string) => api.post(`/messages/${id}/deliver`),
  fail: (id: string) => api.post(`/messages/${id}/fail`),
  batchDelete: (ids: string[]) => api.delete<{ message: string; deleted_count: number }>('/messages/batch', { data: { ids } }),
  export: (params: {
    session_id?: string
    status?: string
    source_addr?: string
    dest_addr?: string
    start_time?: string
    end_time?: string
    format?: 'csv' | 'json'
  }) => api.get('/messages/export', { params, responseType: 'blob' })
}

// Stats API
export const statsApi = {
  get: () => api.get<Stats>('/stats')
}

// Mock Config API
export const mockConfigApi = {
  get: () => api.get<MockConfig>('/mock/config'),
  update: (config: MockConfig) => api.put<MockConfig>('/mock/config', config)
}

// Auth API
export const authApi = {
  login: (username: string, password: string) =>
    api.post<{ token: string }>('/auth/login', { username, password }),
  status: () => api.get<{ authenticated: boolean; username?: string }>('/auth/status')
}

// Data API
export const dataApi = {
  deleteAllMessages: () => api.delete<{ message: string }>('/data/messages'),
  deleteAllSessions: () => api.delete<{ message: string }>('/data/sessions'),
  clearAllData: () => api.delete<{ message: string }>('/data/all')
}

// Send Message API
export const sendApi = {
  listReceivers: () => api.get<{ data: Receiver[] }>('/send/receivers'),
  sendMessage: (params: {
    session_id: string
    source_addr: string
    dest_addr: string
    content: string
    encoding?: 'GSM7' | 'UCS2'
  }) => api.post<{ message: string; session_id: string }>('/send', params)
}

export default api
