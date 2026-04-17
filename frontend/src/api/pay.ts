import request from './request'

// 支付提交
export function paySubmit(data: {
  pid: number
  type: number
  channel?: number
  out_trade_no: string
  name: string
  money: number
  notify_url: string
  return_url?: string
  param?: string
  method?: string
  device?: string
}) {
  const form = new URLSearchParams()
  Object.entries(data).forEach(([key, value]) => {
    if (value !== undefined && value !== null && value !== '') {
      form.append(key, String(value))
    }
  })
  return request.post('/pay/submit', form, {
    headers: {
      'Content-Type': 'application/x-www-form-urlencoded'
    }
  })
}

// JSON API创建订单
export function payCreate(data: {
  pid: number
  type: number
  out_trade_no: string
  name: string
  money: number
  notify_url: string
  return_url?: string
  clientip?: string
  device?: string
  param?: string
  sign?: string
  sign_type?: string
}) {
  return request.post('/pay/create', data)
}

// 订单查询
export function payQuery(params: { pid: number; trade_no?: string; out_trade_no?: string; sign?: string }) {
  return request.get('/pay/query', { params })
}

// 退款
export function payRefund(data: { pid: number; trade_no: string; money: number; sign?: string }) {
  return request.post('/pay/refund', data)
}

// 获取可用支付方式
export function getPayTypes(pid: number) {
  return request.get('/pay/types', { params: { pid } })
}

// 获取可用通道
export function getPayChannels(pid: number, type: number) {
  return request.get('/pay/channels', { params: { pid, type } })
}
