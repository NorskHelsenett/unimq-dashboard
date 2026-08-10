import type { VhostNotification } from '@/types/notifications'
import type { LogEntry, TestResult, VhostObj } from '@/types/notifications'
import { apiFetch } from '@/lib/apiClient'

export interface ApiResponse<T> {
  code: number
  message: string
  body: T
}

export function getSelectedVhost(vhosts: string[]): string {
  const params = new URLSearchParams(window.location.search)
  const vhost = params.get('vhost')
  return (vhost && vhosts.includes(vhost)) ? vhost : (vhosts[0] ?? '')
}

function jsonPost(body: unknown): RequestInit {
  return {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(body),
  }
}

export async function getVhosts(): Promise<string[]> {
  const res = await apiFetch('/api/v1/vhosts')
  const data: ApiResponse<VhostObj[]> = await res.json()
  return (data.body ?? []).map(v => v.name)
}

export async function getVhostNotification(vhost: string): Promise<VhostNotification | null> {
  const res = await apiFetch(`/api/v1/notifications/${encodeURIComponent(vhost)}`)
  if (!res.ok) return null
  const data: ApiResponse<VhostNotification> = await res.json()
  return data.body ?? null
}

export async function addRule(
  vhost: string,
  rule: { name: string; type: string; queue_name?: string; threshold?: number; message?: string }
): Promise<void> {
  await apiFetch(
    `/api/v1/notifications/${encodeURIComponent(vhost)}/rules`,
    jsonPost({ ...rule, enabled: true })
  )
}

export async function updateRule(
  vhost: string,
  ruleId: string,
  threshold: number,
  message: string
): Promise<globalThis.Response> {
  return apiFetch(
    `/api/v1/notifications/${encodeURIComponent(vhost)}/rules/${encodeURIComponent(ruleId)}`,
    jsonPost({ threshold: Number(threshold), message })
  )
}

export async function toggleRule(vhost: string, ruleId: string): Promise<void> {
  await apiFetch(
    `/api/v1/notifications/${encodeURIComponent(vhost)}/rules/${encodeURIComponent(ruleId)}/toggle`,
    { method: 'POST' }
  )
}

export async function testRule(vhost: string, ruleId: string): Promise<TestResult> {
  const res = await apiFetch(
    `/api/v1/notifications/${encodeURIComponent(vhost)}/rules/${encodeURIComponent(ruleId)}/test`,
    { method: 'POST' }
  )
  const data: ApiResponse<TestResult> = await res.json()
  return data.body
}

export async function deleteRule(vhost: string, ruleId: string): Promise<void> {
  await apiFetch(
    `/api/v1/notifications/${encodeURIComponent(vhost)}/rules/${encodeURIComponent(ruleId)}`,
    { method: 'DELETE' }
  )
}

export async function addRecipient(
  vhost: string,
  recipient: { name: string; type: string; url?: string; email?: string }
): Promise<void> {
  await apiFetch(
    `/api/v1/notifications/${encodeURIComponent(vhost)}/recipients`,
    jsonPost(recipient)
  )
}

export async function deleteRecipient(vhost: string, recipientId: string): Promise<void> {
  await apiFetch(
    `/api/v1/notifications/${encodeURIComponent(vhost)}/recipients/${encodeURIComponent(recipientId)}`,
    { method: 'DELETE' }
  )
}

export async function getAlarmLogs(vHost_name: string, alarmType?: string): Promise<LogEntry[]> {
  const url = alarmType
    ? `/api/v1/alarms/${encodeURIComponent(vHost_name)}?type=${encodeURIComponent(alarmType)}`
    : `/api/v1/alarms/${encodeURIComponent(vHost_name)}`
  const res = await apiFetch(url)
  const data: ApiResponse<{ entries?: LogEntry[] }> | null = await res.json().catch(() => null)
  return data?.body?.entries ?? []
}
