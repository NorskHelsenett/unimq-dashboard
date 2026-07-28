import type { RecipientsProps } from '@/components/notifications/RecipientCard'
import type { AlarmProps } from '@/components/notifications/AlarmCard'

export interface VhostNotification {
  Name: string
  Recipients: RecipientsProps[]
  Rules: AlarmProps[]
  Notified: boolean
}
