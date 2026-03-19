import axios from 'axios'
import { useAuthStore } from '@/stores/auth'
import router from '@/router'
import type { Session, Message, Stats, MockConfig, Receiver } from '@/types'

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

// Response interceptor - handle 401 errors
api.interceptors.response.use(
  (response) => response,
  (error) => {
    if (error.response?.status === 401) {
      const authStore = useAuthStore()
      authStore.logout()
      router.push('/login')
    }
    return Promise.reject(error)
  }
)

// Session API
export const sessionApi = {
  list: () => api.get<{ data: Session[], total: number }>('/sessions'),
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
  fail: (id: string) => api.post(`/messages/${id}/fail`)
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
