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
  code?: string
}): Promise<ApiResponse> {
  return request.post('/user/reg', data)
}

// 商户注册 - 发送验证码
export function userRegisterSendCode(data: { email?: string; phone?: string }): Promise<ApiResponse> {
  return request.post('/user/reg/send', data)
}

// 登出
export function userLogout(): Promise<ApiResponse> {
  return request.post('/user/logout')
}

// 获取用户信息
export function getUserInfo(): Promise<ApiResponse> {
  return request.get('/user/info')
}

// 获取订单列表
export function getUserOrders(params: { page?: number; limit?: number; status?: number | string; trade_no?: string }): Promise<ApiResponse> {
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

// 找回密码 - 发送验证码
export function findPwdSendCode(email: string): Promise<ApiResponse> {
  return request.post('/user/findpwd/send', { email })
}

// 找回密码 - 重置密码
export function findPwdReset(data: { email: string; code: string; password: string }): Promise<ApiResponse> {
  return request.post('/user/findpwd/reset', data)
}

// 用户组转让记录
export function getUserGroupTransferList(): Promise<ApiResponse> {
  return request.get('/user/group/transfer/list')
}

// 获取用户组列表(商户端)
export function getUserGroupList(): Promise<ApiResponse> {
  return request.get('/user/group/list')
}

// 创建用户组转让
export function createUserGroupTransfer(data: { target_uid: number; group_id: number; price: number }): Promise<ApiResponse> {
  return request.post('/user/group/transfer/create', data)
}

// 商户订单操作
export function userOrderOp(data: { action: 'notify' | 'refund'; trade_no: string; money?: number }): Promise<ApiResponse> {
  return request.post('/user/order/op', data)
}

// 购买用户组
export function buyUserGroup(data: { group_id: number }): Promise<ApiResponse> {
  return request.post('/user/group/buy', data)
}
