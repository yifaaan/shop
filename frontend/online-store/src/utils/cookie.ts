// Cookie 工具（用户 token / 昵称的持久化）
export const cookie = {
  set(name: string, value: string, days = 7): void {
    const exdate = new Date()
    exdate.setDate(exdate.getDate() + days)
    document.cookie = `${name}=${encodeURIComponent(value)};expires=${exdate.toUTCString()};path=/`
  },
  get(name: string): string | null {
    const arr = document.cookie.match(new RegExp('(^| )' + name + '=([^;]*)(;|$)'))
    return arr ? decodeURIComponent(arr[2]) : null
  },
  del(name: string): void {
    const value = cookie.get(name)
    if (value !== null) {
      document.cookie = `${name}=;expires=Thu, 01 Jan 1970 00:00:01 GMT;path=/`
    }
  },
}
