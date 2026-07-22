import axios, { type AxiosInstance, type AxiosResponse, type InternalAxiosRequestConfig } from 'axios'
import { ElMessage } from 'element-plus'
import { cookie } from './cookie'
import router from '@/router'

// 统一响应数据结构（后端 envelope: { total, data } 或裸对象）
const service: AxiosInstance = axios.create({
  timeout: 15000,
})

// 请求拦截器：带上 token（后端读取 x-token 头）
service.interceptors.request.use(
  (config: InternalAxiosRequestConfig) => {
    const token = cookie.get('token')
    if (token) {
      config.headers.set('x-token', token)
      config.headers.set('Authorization', `JWT ${token}`)
    }
    return config
  },
  (error) => Promise.reject(error),
)

// 响应拦截器：统一处理鉴权与错误提示
service.interceptors.response.use(
  (response: AxiosResponse) => response.data,
  (error) => {
    const res = error?.response
    if (res) {
      switch (res.status) {
        case 401: {
          const msg = res.data?.msg || '登录已过期，请重新登录'
          ElMessage.error(typeof msg === 'string' ? msg : '未登录')
          cookie.del('token')
          cookie.del('name')
          // 避免在登录页死循环跳转
          if (router.currentRoute.value.name !== 'login') {
            router.push({ name: 'login' })
          }
          break
        }
        case 403:
          ElMessage.error('您没有该操作权限')
          break
        case 500:
          ElMessage.error('服务器错误')
          break
      }
    } else {
      ElMessage.error('网络异常，请稍后重试')
    }
    return Promise.reject(res?.data ?? error)
  },
)

export default service
