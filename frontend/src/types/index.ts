// Message status enum
export enum MessageStatus {
  Pending = 'pending',
  Delivered = 'delivered',
  Failed = 'failed'
}

// Bind type enum
export enum BindType {
  Transmitter = 'TX',
  Receiver = 'RX',
  Transceiver = 'TR'
}

// Encoding type enum
export enum EncodingType {
  GSM7 = 'GSM7',
  UCS2 = 'UCS2'
}

// Session interface
export interface Session {
  id: string
  system_id: string
  bind_type: string
  remote_addr: string
  connected_at: string
  status: string
}

// Message interface
export interface Message {
  id: string
  session_id: string
  message_id: string
  sequence_num: number
  source_addr: string
  dest_addr: string
  content: string
  encoding: string
  status: MessageStatus | string
  created_at: string
  delivered_at?: string
}

// Stats interface
export interface Stats {
  active_connections: number
  total_messages: number
  pending_messages: number
  delivered_messages: number
  failed_messages: number
}

// SessionStats interface (for session detail)
export interface SessionStats {
  total: number
  delivered: number
  failed: number
  pending: number
  success_rate: number
}

// SessionDetail interface
export interface SessionDetail {
  session: Session
  stats: SessionStats
  recent_messages: Message[]
}

// MockConfig interface
export interface MockConfig {
  auto_response: boolean
  success_rate: number
  response_delay: number
  deliver_report: boolean
  deliver_delay: number
}

// Receiver interface (for sending messages)
export interface Receiver {
  id: string
  system_id: string
  bind_type: string
  remote_addr: string
}

// WebSocket event types
export interface WsSessionConnectEvent {
  type: 'session_connect'
  session: Session
}

export interface WsSessionDisconnectEvent {
  type: 'session_disconnect'
  session_id: string
}

export interface WsMessageReceivedEvent {
  type: 'message_received'
  message: Message
}

export interface WsMessageDeliveredEvent {
  type: 'message_delivered'
  message_id: string
}

export type WsEvent =
  | WsSessionConnectEvent
  | WsSessionDisconnectEvent
  | WsMessageReceivedEvent
  | WsMessageDeliveredEvent

// SystemConfig interface
export interface SystemConfig {
  smpp_port: string
  http_port: string
  db_type: string
  redis_enabled: boolean
  redis_status: string
  admin_password: string
  jwt_expiry: number
  cors_origins: string
  login_rate_limit: number
}

// UpdateSystemConfigRequest interface
export interface UpdateSystemConfigRequest {
  old_password?: string
  new_password?: string
  confirm_password?: string
  jwt_expiry?: number
  cors_origins?: string
  login_rate_limit?: number
}

// RateLimitStatus interface
export interface RateLimitStatus {
  remaining: number
  reset_at: string | null
  total: number
  is_limited: boolean
  window_seconds: number
}

// MessageTemplate interface
export interface MessageTemplate {
  id: string
  name: string
  content: string
  encoding: string
  created_at: string
}

// CreateTemplateRequest interface
export interface CreateTemplateRequest {
  name: string
  content: string
  encoding?: string
}

// UpdateTemplateRequest interface
export interface UpdateTemplateRequest {
  name: string
  content: string
  encoding?: string
}
