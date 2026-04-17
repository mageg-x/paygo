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

// 获取用户组列表
export function getGroupList(params?: { page?: number; limit?: number }): Promise<ApiResponse> {
  return request.get('/admin/group', { params })
}

// 用户组操作
export function groupOp(data: { action: string;[key: string]: any }): Promise<ApiResponse> {
  return request.post('/admin/group/op', data)
}

// 风控列表
export function riskList(params: { page?: number; limit?: number; uid?: string }): Promise<ApiResponse> {
  return request.get('/admin/risk', { params })
}

// 风控操作
export function riskOp(data: { action: string; id: number; status?: number }): Promise<ApiResponse> {
  return request.post('/admin/risk/op', data)
}

// 黑名单列表
export function blacklistList(params: { page?: number; limit?: number; type?: string }): Promise<ApiResponse> {
  return request.get('/admin/blacklist', { params })
}

// 黑名单操作
export function blacklistOp(data: { action: string; id?: number; type?: number; content?: string; remark?: string }): Promise<ApiResponse> {
  return request.post('/admin/blacklist/op', data)
}

// 域名授权列表
export function domainList(params: { page?: number; limit?: number; uid?: string }): Promise<ApiResponse> {
  return request.get('/admin/domain', { params })
}

// 域名授权操作
export function domainOp(data: { action: string; id?: number; uid?: number; domain?: string; status?: number }): Promise<ApiResponse> {
  return request.post('/admin/domain/op', data)
}

// 公告列表
export function anounceList(params: { page?: number; limit?: number }): Promise<ApiResponse> {
  return request.get('/admin/anounce', { params })
}

// 公告操作
export function anounceOp(data: { action: string; id?: number; content?: string; color?: string; sort?: number; status?: number }): Promise<ApiResponse> {
  return request.post('/admin/anounce/op', data)
}

// 操作日志列表
export function logList(params: { page?: number; limit?: number; uid?: string }): Promise<ApiResponse> {
  return request.get('/admin/log', { params })
}

// SSO单点登录
export function ssoLogin(data: { uid: number }): Promise<ApiResponse<{ token: string; uid: number }>> {
  return request.post('/admin/sso', data)
}

// 计划任务列表
export function cronList(): Promise<ApiResponse> {
  return request.get('/admin/cron')
}

// 计划任务操作
export function cronOp(data: { action: string; name?: string; enable?: boolean; spec?: string }): Promise<ApiResponse> {
  return request.post('/admin/cron/op', data)
}

// 通用获取配置(根据key数组)
export function getSettings(keys: string[]): Promise<ApiResponse<Record<string, string>>> {
  return request.get('/admin/set/get', { params: { keys: keys.join(',') } })
}

// 通用保存配置
export function saveSettings(data: Record<string, string>): Promise<ApiResponse> {
  return request.post('/admin/set/save', data)
}

// 支付类型列表
export function getPayTypeList(): Promise<ApiResponse> {
  return request.get('/admin/paytype')
}

// 支付类型操作
export function payTypeOp(data: { action: string; id?: number; name?: string; device?: number; showname?: string; status?: number }): Promise<ApiResponse> {
  return request.post('/admin/paytype/op', data)
}

// 轮询配置列表
export function getRollList(): Promise<ApiResponse> {
  return request.get('/admin/roll')
}

// 轮询配置操作
export function rollOp(data: { action: string; id?: number; type?: number; name?: string; kind?: number; info?: string; status?: number; index?: number }): Promise<ApiResponse> {
  return request.post('/admin/roll/op', data)
}

// 分账订单列表
export function profitOrderList(): Promise<ApiResponse> {
  return request.get('/admin/profit/order')
}

// 分账接收方列表
export function profitReceiverList(): Promise<ApiResponse> {
  return request.get('/admin/profit/receiver')
}

// 分账接收方操作
export function profitReceiverOp(data: { action: string; id?: number; uid?: number; name?: string; account?: string; rate?: string; status?: number }): Promise<ApiResponse> {
  return request.post('/admin/profit/receiver/op', data)
}

// 数据清理统计
export function getCleanStats(params: { order_timeout: string; max_retry: string; log_days: string }): Promise<ApiResponse> {
  return request.get('/admin/clean/stats', { params })
}

// 执行数据清理
export function runClean(data: { action: 'orders' | 'failed_notifies' | 'logs' | 'cache'; order_timeout?: number; max_retry?: number; log_days?: number }): Promise<ApiResponse<{ count: number }>> {
  return request.post('/admin/clean/run', data)
}

// 导出订单数据
export function exportOrders(params: {
  start_date: string
  end_date: string
  uid?: string
  status?: string
  type?: string
  limit?: number
}): Promise<ApiResponse> {
  return request.get('/admin/export/orders', { params })
}

// 执行分账
export function profitDo(data: { ps_no: string }): Promise<ApiResponse> {
  return request.post('/admin/profit/do', data)
}

// 批量转账记录列表
export function getTransferBatchList(): Promise<ApiResponse> {
  return request.get('/admin/transfer/batch')
}

// 创建批量转账
export function transferBatchCreate(data: { filename: string; data: string }): Promise<ApiResponse> {
  return request.post('/admin/transfer/batch/create', data)
}
