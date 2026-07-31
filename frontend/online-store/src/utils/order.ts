export const ORDER_PAYMENT_TIMEOUT_MS = 15 * 60 * 1000

export type OrderStatusTagType = 'success' | 'warning' | 'info'

function toTimestamp(value?: number | string): number | null {
  if (value === undefined || value === null || value === '') return null

  const raw = typeof value === 'string' ? value.trim() : value
  const numeric = Number(raw)
  if (raw !== '' && Number.isFinite(numeric) && numeric > 0) {
    return numeric < 1_000_000_000_000 ? numeric * 1000 : numeric
  }

  if (typeof raw === 'string') {
    const parsed = Date.parse(raw)
    if (Number.isFinite(parsed)) return parsed
  }

  return null
}

export function orderCreatedAt(addTime?: number | string): number | null {
  return toTimestamp(addTime)
}

export function orderDeadline(addTime?: number | string): number | null {
  const createdAt = orderCreatedAt(addTime)
  return createdAt === null ? null : createdAt + ORDER_PAYMENT_TIMEOUT_MS
}

export function orderRemainingMs(addTime?: number | string, now = Date.now()): number | null {
  const deadline = orderDeadline(addTime)
  return deadline === null ? null : Math.max(0, deadline - now)
}

export function formatCountdown(milliseconds: number | null): string {
  if (milliseconds === null) return '--:--'

  const seconds = Math.ceil(Math.max(0, milliseconds) / 1000)
  const minutesPart = Math.floor(seconds / 60)
  const secondsPart = seconds % 60
  return `${String(minutesPart).padStart(2, '0')}:${String(secondsPart).padStart(2, '0')}`
}

export function formatOrderTime(addTime?: number | string): string {
  const timestamp = orderCreatedAt(addTime)
  if (timestamp === null) return '--'

  const date = new Date(timestamp)
  if (Number.isNaN(date.getTime())) return '--'

  const pad = (value: number) => String(value).padStart(2, '0')
  return [
    `${date.getFullYear()}-${pad(date.getMonth() + 1)}-${pad(date.getDate())}`,
    `${pad(date.getHours())}:${pad(date.getMinutes())}:${pad(date.getSeconds())}`,
  ].join(' ')
}

export function formatMoney(value?: number | string): string {
  const amount = Number(value ?? 0)
  return `￥${(Number.isFinite(amount) ? amount : 0).toFixed(2)}`
}

export function orderStatusText(status?: number | string): string {
  switch (Number(status)) {
    case 1:
      return '待支付'
    case 2:
      return '已支付'
    case 3:
      return '已取消'
    default:
      return '未知状态'
  }
}

export function orderStatusType(status?: number | string): OrderStatusTagType {
  switch (Number(status)) {
    case 1:
      return 'warning'
    case 2:
      return 'success'
    default:
      return 'info'
  }
}

export function isPendingOrder(status?: number | string): boolean {
  return Number(status) === 1
}
