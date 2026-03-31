import axios, { type AxiosInstance, type AxiosResponse, type AxiosRequestConfig } from 'axios'
import { ElMessage } from 'element-plus'
import router from '@/router'

// API方法类型
export interface ApiResponse<T = any> {
  code: number
  msg: string
  data: T
  token?: string
  count?: number
  user?: any
}

// 分页参数
export interface PageParams {
  page?: number
  limit?: number
}

// 分页响应
export interface PageResponse<T> {
  list: T[]
  total: number
  page: number
  limit: number
}

// 自定义请求接口，返回 ApiResponse 而不是 AxiosResponse
interface RequestInstance {
  get<T = any>(url: string, config?: AxiosRequestConfig): Promise<ApiResponse<T>>
  post<T = any>(url: string, data?: any, config?: AxiosRequestConfig): Promise<ApiResponse<T>>
  put<T = any>(url: string, data?: any, config?: AxiosRequestConfig): Promise<ApiResponse<T>>
  delete<T = any>(url: string, config?: AxiosRequestConfig): Promise<ApiResponse<T>>
}

// 创建axios实例
const axiosInstance: AxiosInstance = axios.create({
  baseURL: '/api',
  timeout: 30000,
  headers: {
    'Content-Type': 'application/json'
  }
})

// 请求拦截器
axiosInstance.interceptors.request.use(
  (config) => {
    // 添加admin token
    const adminToken = localStorage.getItem('admin_token')
    if (adminToken) {
      config.headers['Admin-Token'] = adminToken
    }

    // 添加user token
    const userToken = localStorage.getItem('user_token')
    if (userToken) {
      config.headers['User-Token'] = userToken
    }

    return config
  },
  (error) => {
    return Promise.reject(error)
  }
)

// 响应拦截器
axiosInstance.interceptors.response.use(
  (response: AxiosResponse) => {
    const res = response.data

    if (res.code !== 0 && res.code !== undefined) {
      ElMessage.error(res.msg || '请求失败')
      return Promise.reject(new Error(res.msg || '请求失败'))
    }

    return res
  },
  (error) => {
    // 401 未授权，登录过期
    if (error.response?.status === 401) {
      const isAdminRoute = error.config?.url?.includes('/admin/')
      if (isAdminRoute) {
        localStorage.removeItem('admin_token')
        ElMessage.error('登录已过期，请重新登录')
        router.push('/admin/login')
      } else {
        localStorage.removeItem('user_token')
        ElMessage.error('登录已过期，请重新登录')
        router.push('/user/login')
      }
      return Promise.reject(new Error('登录已过期'))
    }

    const msg = error.response?.data?.msg || error.message || '网络错误'
    ElMessage.error(msg)
    return Promise.reject(new Error(msg))
  }
)

const request = axiosInstance as RequestInstance

export default request
