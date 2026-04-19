import request from './request'

export interface PaySubmitResult {
  trade_no?: string
  result?: {
    pay_url?: string
    qr_url?: string
    html?: string
    app_payload?: string
    method?: string
  }
}

export interface PaySubmitResponse {
  code: number
  msg: string
  data: any
  trade_no?: string
  result?: PaySubmitResult['result']
}

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
  sign?: string
  sign_type?: string
}): Promise<PaySubmitResponse> {
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

export function payCashierSubmit(data: {
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
}): Promise<PaySubmitResponse> {
  const form = new URLSearchParams()
  Object.entries(data).forEach(([key, value]) => {
    if (value !== undefined && value !== null && value !== '') {
      form.append(key, String(value))
    }
  })
  return request.post('/pay/cashier_submit', form, {
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
export function payQuery(params: { pid: number; trade_no?: string; out_trade_no?: string; sign?: string; sign_type?: string }) {
  return request.get('/pay/query', { params })
}

// 退款
export function payRefund(data: { pid: number; trade_no: string; money: number; sign?: string; sign_type?: string }) {
  return request.post('/pay/refund', data)
}

// 获取可用支付方式
export function getPayTypes(pid: number): Promise<{
  code: number
  msg?: string
  data: any[]
  pid?: number
  requested_pid?: number
  fallback?: boolean
}> {
  return request.get('/pay/types', { params: { pid } })
}

// 获取可用通道
export function getPayChannels(pid: number, type: number) {
  return request.get('/pay/channels', { params: { pid, type } })
}
