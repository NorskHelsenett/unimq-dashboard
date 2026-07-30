import type { RecipientsProps } from '@/components/notifications/RecipientCard'

export interface VhostNotification {
  Name: string
  Recipients: RecipientsProps[]
  Rules: AlarmProps[]
  Notified: boolean
}


export interface AlarmProps {
    id?: string
    name?: string
    type?: string    
    queue_name?: string    
    threshold?: number
    message?: string
    enabled?: boolean
    status?: string //"firing", "ok", also based on enabled status
    last_fired?: string | null
    last_value?: number | null
}


export const alarmDropdownOptions = [
    { value: 'connections', label: 'Connections' },
    { value: 'channels', label: 'Channels' },
    { value: 'queues', label: 'Queues' },
    { value: 'unacked', label: 'Unacknowledged Messages' },
    { value: 'queue_messages', label: 'Messages in Queue' },
    { value: 'queue_size', label: 'Queue Size' },
    { value: 'no_consumer', label: 'No Consumers' },
    { value: 'maintenance', label: 'Maintenance Message' },
]


export interface NotifyRuleData {
  Vhost: string
  Rule: AlarmProps
  Msg: string
}

export interface VhostObj {
  name: string
}

export interface LogEntry {
  ts: string
  event: 'fired' | 'resolved'
  value?: number
  threshold: number
}

export interface TestResult {
  success: boolean
  message: string
}