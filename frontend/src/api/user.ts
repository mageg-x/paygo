import request, { type ApiResponse } from './request'

// 用户信息接口
export interface User {
  uid: number
  username: string
  email: string
  phone: string
  money: number
  status: number
}

// 商户登录
export function userLogin(data: { uid: number; password: string } | { uid: number; key: string }): Promise<ApiResponse<User>> {
  return request.post('/user/login', data)
}

// 商户注册
export function userRegister(data: {
  email: string
  phone?: string
  password: string
  invite_code?: string
}): Promise<ApiResponse> {
  return request.post('/user/reg', data)
}

// 登出
export function userLogout(): Promise<ApiResponse> {
  return request.post('/user/logout')
}

// 获取用户信息
export function getUserInfo(): Promise<ApiResponse> {
  return request.get('/user/index')
}

// 获取订单列表
export function getUserOrders(params: { page?: number; limit?: number; status?: number }): Promise<ApiResponse> {
  return request.get('/user/orders', { params })
}

// 获取结算列表
export function getUserSettles(params: { page?: number; limit?: number }): Promise<ApiResponse> {
  return request.get('/user/settles', { params })
}

// 申请结算
export function applySettle(data: {
  account: string
  username: string
  money: number
  type: number
}): Promise<ApiResponse> {
  return request.post('/user/settle/apply', data)
}

// 获取资金记录
export function getUserRecords(params: { page?: number; limit?: number; action?: number }): Promise<ApiResponse> {
  return request.get('/user/records', { params })
}

// 更新资料
export function updateProfile(data: { username?: string; phone?: string; qq?: string }): Promise<ApiResponse> {
  return request.post('/user/editinfo', data)
}

// 实名认证
export function submitCertificate(data: { certname: string; certno: string; certtype: number }): Promise<ApiResponse> {
  return request.post('/user/certificate', data)
}
