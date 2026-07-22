import { defineStore } from 'pinia'
import { ref } from 'vue'
import { cookie } from '@/utils/cookie'
import type { UserInfo } from '@/types'

export const useUserStore = defineStore('user', () => {
  const name = ref<string>(cookie.get('name') ?? '')
  const token = ref<string>(cookie.get('token') ?? '')

  function setInfo(n?: string, t?: string) {
    name.value = n ?? cookie.get('name') ?? ''
    token.value = t ?? cookie.get('token') ?? ''
  }

  function login(nickname: string, t: string) {
    cookie.set('name', nickname, 7)
    cookie.set('token', t, 7)
    setInfo(nickname, t)
  }

  function logout() {
    cookie.del('name')
    cookie.del('token')
    name.value = ''
    token.value = ''
  }

  function isLoggedIn(): boolean {
    return !!token.value
  }

  const userInfo = ref<UserInfo>({ name: '', token: '' })

  return { name, token, userInfo, setInfo, login, logout, isLoggedIn }
})
