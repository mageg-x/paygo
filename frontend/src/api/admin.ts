import request, { type ApiResponse } from './request'

// 管理员登录
export function adminLogin(data: { username: string; password: string }): Promise<ApiResponse<{ token: string }>> {
  return request.post('/admin/login', data)
}

// 管理员登出
export function adminLogout(): Promise<ApiResponse> {
  return request.post('/admin/logout')
}

// 获取统计数据
export function getAdminStats(): Promise<ApiResponse> {
  return request.get('/admin/stats')
}

// 获取商户列表
export function getUserList(params: { page?: number; limit?: number }): Promise<ApiResponse> {
  return request.get('/admin/users', { params })
}

// 添加商户
export function addUser(data: {
  gid?: number
  phone?: string
  email?: string
  pwd?: string
  qq?: string
  url?: string
  settle_id?: number
  account?: string
  username?: string
  mode?: number
  pay?: number
  settle?: number
  status?: number
}): Promise<ApiResponse> {
  return request.post('/admin/user/add', data)
}

// 获取商户编辑信息
export function getUserEdit(uid: number): Promise<ApiResponse> {
  return request.get('/admin/user/edit', { params: { uid } })
}

// 更新商户
export function updateUser(data: {
  uid: number
  gid?: number
  phone?: string
  email?: string
  pwd?: string
  qq?: string
  url?: string
  settle_id?: number
  account?: string
  username?: string
  mode?: number
  pay?: number
  settle?: number
  status?: number
  money?: number
}): Promise<ApiResponse> {
  return request.post('/admin/user/update', data)
}

// 商户操作
export function userOp(data: { action: string; uid: number;[key: string]: any }): Promise<ApiResponse> {
  return request.post('/admin/user/op', data)
}

// 获取订单列表
export function getOrderList(params: { page?: number; limit?: number; status?: number }): Promise<ApiResponse> {
  return request.get('/admin/orders', { params })
}

// 订单操作
export function orderOp(data: { action: string; trade_no: string;[key: string]: any }): Promise<ApiResponse> {
  return request.post('/admin/order/op', data)
}

// 获取结算列表
export function getSettleList(params: { page?: number; limit?: number }): Promise<ApiResponse> {
  return request.get('/admin/settles', { params })
}

// 结算操作
export function settleOp(data: { action: string; id: number;[key: string]: any }): Promise<ApiResponse> {
  return request.post('/admin/settle/op', data)
}

// 获取转账列表
export function getTransferList(params: { page?: number; limit?: number; status?: string; search?: string }): Promise<ApiResponse> {
  return request.get('/admin/transfer', { params })
}

// 转账操作
export function transferOp(data: { action: string; biz_no: string;[key: string]: any }): Promise<ApiResponse> {
  return request.post('/admin/transfer/op', data)
}

// 获取通道列表
export function getChannelList(): Promise<ApiResponse> {
  return request.get('/admin/channel')
}

// 通道操作
export function channelOp(data: { action: string;[key: string]: any }): Promise<ApiResponse> {
  return request.post('/admin/channel/op', data)
}

// 获取插件列表
export function getPluginList(): Promise<ApiResponse> {
  return request.get('/admin/plugin')
}

// 插件操作
export function pluginOp(data: { action: string;[key: string]: any }): Promise<ApiResponse> {
  return request.post('/admin/plugin/op', data)
}

// 获取系统配置
export function getConfig(): Promise<ApiResponse> {
  return request.get('/admin/set/config')
}

// 保存系统配置
export function saveConfig(data: Record<string, string>): Promise<ApiResponse<{ token?: string }>> {
  return request.post('/admin/set/save', data)
}

// 获取邀请码列表
export function getInviteCodeList(params: { page?: number; limit?: number; search?: string }): Promise<ApiResponse> {
  return request.get('/admin/invitecode', { params })
}

// 生成邀请码
export function generateInviteCode(num: number): Promise<ApiResponse<{ codes: string[] }>> {
  return request.post('/admin/invitecode/generate', { num })
}

// 删除邀请码
export function deleteInviteCode(id: number): Promise<ApiResponse> {
  return request.post('/admin/invitecode/delete', { id })
}
