import axios from 'axios'
import { useAuthStore } from '@/stores/auth'
import router from '@/router'

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

// Types
export interface Session {
  id: string
  system_id: string
  bind_type: string
  remote_addr: string
  connected_at: string
  status: string
}

export interface Message {
  id: string
  session_id: string
  message_id: string
  sequence_num: number
  source_addr: string
  dest_addr: string
  content: string
  encoding: string
  status: string
  created_at: string
  delivered_at?: string
}

export interface Stats {
  active_connections: number
  total_messages: number
  pending_messages: number
  delivered_messages: number
  failed_messages: number
}

export interface MockConfig {
  auto_response: boolean
  success_rate: number
  response_delay: number
  deliver_report: boolean
  deliver_delay: number
}

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
    page?: number
    page_size?: number
  }) => api.get<{ data: Message[], total: number, page: number, page_size: number }>('/messages', { params }),
  get: (id: string) => api.get<Message>(`/messages/${id}`),
  deliver: (id: string) => api.post(`/messages/${id}/deliver`)
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

export default api
