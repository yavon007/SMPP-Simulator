import { defineStore } from 'pinia'
import { ref } from 'vue'
import { sessionApi, messageApi, statsApi, mockConfigApi, type Session, type Message, type Stats, type MockConfig } from '@/api'

export const useSessionStore = defineStore('session', () => {
  const sessions = ref<Session[]>([])
  const loading = ref(false)

  async function fetchSessions() {
    loading.value = true
    try {
      const res = await sessionApi.list()
      sessions.value = res.data.data
    } finally {
      loading.value = false
    }
  }

  async function deleteSession(id: string) {
    await sessionApi.delete(id)
    sessions.value = sessions.value.filter(s => s.id !== id)
  }

  function addSession(session: Session) {
    sessions.value.unshift(session)
  }

  function removeSession(id: string) {
    const session = sessions.value.find(s => s.id === id)
    if (session) {
      session.status = 'closed'
    }
  }

  return {
    sessions,
    loading,
    fetchSessions,
    deleteSession,
    addSession,
    removeSession
  }
})

export const useMessageStore = defineStore('message', () => {
  const messages = ref<Message[]>([])
  const total = ref(0)
  const page = ref(1)
  const pageSize = ref(20)
  const loading = ref(false)

  async function fetchMessages(params: {
    session_id?: string
    status?: string
    source_addr?: string
    dest_addr?: string
    start_time?: string
    end_time?: string
    page?: number
    page_size?: number
  } = {}) {
    loading.value = true
    try {
      const res = await messageApi.list({
        page: params.page || page.value,
        page_size: params.page_size || pageSize.value,
        session_id: params.session_id,
        status: params.status,
        source_addr: params.source_addr,
        dest_addr: params.dest_addr,
        start_time: params.start_time,
        end_time: params.end_time
      })
      messages.value = res.data.data
      total.value = res.data.total
      page.value = res.data.page
      pageSize.value = res.data.page_size
    } finally {
      loading.value = false
    }
  }

  function addMessage(message: Message) {
    messages.value.unshift(message)
    total.value++
  }

  function updateMessageStatus(id: string, status: string) {
    const message = messages.value.find(m => m.id === id)
    if (message) {
      message.status = status
    }
  }

  return {
    messages,
    total,
    page,
    pageSize,
    loading,
    fetchMessages,
    addMessage,
    updateMessageStatus
  }
})

export const useStatsStore = defineStore('stats', () => {
  const stats = ref<Stats>({
    active_connections: 0,
    total_messages: 0,
    pending_messages: 0,
    delivered_messages: 0,
    failed_messages: 0
  })

  async function fetchStats() {
    const res = await statsApi.get()
    stats.value = res.data
  }

  function updateStats(newStats: Partial<Stats>) {
    stats.value = { ...stats.value, ...newStats }
  }

  return {
    stats,
    fetchStats,
    updateStats
  }
})

export const useConfigStore = defineStore('config', () => {
  const config = ref<MockConfig>({
    auto_response: true,
    success_rate: 100,
    response_delay: 0,
    deliver_report: false,
    deliver_delay: 1000
  })
  const loading = ref(false)

  async function fetchConfig() {
    loading.value = true
    try {
      const res = await mockConfigApi.get()
      config.value = res.data
    } finally {
      loading.value = false
    }
  }

  async function updateConfig(newConfig: MockConfig) {
    loading.value = true
    try {
      const res = await mockConfigApi.update(newConfig)
      config.value = res.data
    } finally {
      loading.value = false
    }
  }

  return {
    config,
    loading,
    fetchConfig,
    updateConfig
  }
})
