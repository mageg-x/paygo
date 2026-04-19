<template>
  <div class="min-h-screen bg-gradient-to-br from-blue-500 via-sky-500 to-indigo-600 px-4 py-8">
    <div class="mx-auto w-full max-w-md space-y-4">
      <div class="text-center text-white">
        <img :src="logo" alt="GoPay支付" class="w-16 h-16 mx-auto mb-3 drop-shadow-md" />
        <!-- <h1 class="text-3xl font-bold tracking-wide">GoPay支付</h1> -->
        <p class="text-blue-50 text-sm mt-1">安全、快捷、稳定的聚合支付平台</p>
      </div>

      <div class="bg-white/95 backdrop-blur rounded-2xl shadow-xl p-6 md:p-8 w-full space-y-5">
      <template v-if="isFixedMode">
        <div>
          <h2 class="text-2xl font-bold text-gray-800">收银台</h2>
          <p class="text-sm text-gray-500 mt-1">商户 ID：{{ effectivePid || pid }}</p>
          <p v-if="cashierTip" class="text-xs text-amber-700 bg-amber-50 border border-amber-200 rounded-lg px-2 py-1 mt-2">
            {{ cashierTip }}
          </p>
        </div>

        <div>
          <label class="block text-sm font-medium text-gray-700 mb-2">收款金额 (元)</label>
          <input
            v-model.number="form.money"
            type="number"
            min="0.01"
            step="0.01"
            class="w-full px-4 py-2 border border-gray-200 rounded-lg text-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
            placeholder="请输入金额"
          />
        </div>

        <div>
          <label class="block text-sm font-medium text-gray-700 mb-2">支付方式</label>
          <div class="grid grid-cols-2 gap-3">
            <button
              v-for="pt in payTypes"
              :key="pt.id"
              type="button"
              :class="[
                'rounded-xl border-2 p-3 text-left transition-all',
                Number(form.type) === Number(pt.id)
                  ? 'border-blue-500 bg-blue-50 shadow-sm'
                  : 'border-gray-200 bg-white hover:border-blue-300 hover:bg-blue-50/40'
              ]"
              @click="form.type = Number(pt.id)"
            >
              <div class="flex items-center gap-3">
                <div
                  :class="[
                    'w-10 h-10 rounded-lg flex items-center justify-center',
                    Number(form.type) === Number(pt.id) ? 'bg-blue-100 text-blue-600' : 'bg-gray-100 text-gray-500'
                  ]"
                >
                  <SvgIcon :name="payTypeIcon(pt)" :size="24" />
                </div>
                <div class="min-w-0">
                  <div class="text-sm font-semibold text-gray-800 truncate">
                    {{ pt.showname || pt.name || ('类型' + pt.id) }}
                  </div>
                  <div class="text-xs text-gray-500 truncate">
                    {{ Number(form.type) === Number(pt.id) ? '已选择' : '点击选择' }}
                  </div>
                </div>
              </div>
            </button>
          </div>
          <div v-if="payTypes.length === 0" class="text-xs text-amber-600 bg-amber-50 border border-amber-200 rounded-lg px-3 py-2 mt-2">
            当前商户暂无可用支付方式
          </div>
        </div>

        <div>
          <label class="block text-sm font-medium text-gray-700 mb-2">订单备注（可选）</label>
          <input
            v-model.trim="form.remark"
            type="text"
            class="w-full px-4 py-2 border border-gray-200 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
            placeholder="请输入备注"
          />
        </div>

        <div class="flex gap-2">
          <button
            class="flex-1 py-2.5 bg-blue-600 text-white rounded-lg font-medium hover:bg-blue-700 disabled:opacity-60"
            :disabled="submitting"
            @click="submitOrder"
          >
            {{ submitting ? '创建中...' : '确认并支付' }}
          </button>
        </div>

        <div v-if="tradeNo" class="bg-gray-50 border border-gray-200 rounded-lg p-3 space-y-2">
          <div class="text-xs text-gray-500">订单号</div>
          <div class="font-mono text-sm text-gray-800 break-all">{{ tradeNo }}</div>

          <div v-if="qrCodeUrl" class="text-center pt-2">
            <img :src="qrCodeUrl" alt="支付二维码" class="mx-auto w-56 h-56 border rounded-lg p-2 bg-white" />
          </div>

          <div v-if="payUrl" class="text-xs text-blue-600 break-all bg-blue-50 border border-blue-200 rounded-lg px-3 py-2">
            {{ payUrl }}
          </div>

          <div v-if="htmlPayload" class="text-xs text-orange-600 bg-orange-50 border border-orange-200 rounded-lg px-3 py-2">
            正在拉起支付页，如未自动跳转请稍后重试。
          </div>
        </div>

        <div v-if="orderInfo" class="text-sm rounded-lg border border-gray-200 p-3 bg-white">
          <div>订单状态：<span class="font-semibold">{{ statusText(orderInfo.status) }}</span></div>
          <div v-if="orderInfo.trade_no" class="mt-1 text-gray-500 font-mono text-xs break-all">{{ orderInfo.trade_no }}</div>
        </div>
      </template>

      <template v-else>
        <div>
          <h2 class="text-2xl font-bold text-gray-800">订单查询</h2>
          <p class="text-sm text-gray-500 mt-1">订单号：{{ tradeNoParam }}</p>
        </div>

        <div v-if="loading" class="text-center text-gray-500 py-8">加载中...</div>

        <template v-else-if="orderInfo">
          <div class="space-y-2 text-sm">
            <div>商品名称：{{ orderInfo.name || '-' }}</div>
            <div>金额：¥{{ orderInfo.money || '-' }}</div>
            <div>状态：<span class="font-semibold">{{ statusText(orderInfo.status) }}</span></div>
          </div>
          <button
            class="w-full py-2.5 bg-blue-600 text-white rounded-lg font-medium hover:bg-blue-700"
            @click="queryOrderByTradeNo"
          >
            刷新状态
          </button>
        </template>

        <div v-else class="text-center py-8 text-gray-500">订单不存在或已过期</div>
      </template>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref } from 'vue'
import { useRoute } from 'vue-router'
import QRCode from 'qrcode'
import { ElMessage } from 'element-plus'
import SvgIcon from '@/components/svgicon.vue'
import logo from '@/assets/gopay.png'
import { getPayTypes, payCashierSubmit, payQuery, paySubmit } from '@/api/pay'
import { makeOpenAPISign } from '@/utils/sign'

const route = useRoute()
const loading = ref(true)
const submitting = ref(false)
const payTypes = ref<any[]>([])
const tradeNo = ref('')
const payUrl = ref('')
const htmlPayload = ref('')
const qrCodeUrl = ref('')
const orderInfo = ref<any>(null)
const apiKey = ref('')
const effectivePid = ref(0)
const cashierTip = ref('')

const form = ref({
  type: 0,
  money: 1.00,
  remark: ''
})
const autoPayTriggered = ref(false)

let pollTimer: number | null = null
let pollStartAt = 0

const uidParam = computed(() => String(route.params.uid || '').trim())
const tradeNoParam = computed(() => String(route.params.trade_no || '').trim())
const pid = computed(() => Number(uidParam.value || 0))
const isFixedMode = computed(() => !!uidParam.value)

function payTypeIcon(pt: any) {
  const text = String(pt?.showname || pt?.name || '').toLowerCase()
  if (text.includes('微信')) return 'wechatpay'
  if (text.includes('支付宝') || text.includes('alipay')) return 'alipay'
  return 'bankcard'
}

function statusText(status: number) {
  const map: Record<number, string> = {
    0: '待支付',
    1: '已支付',
    2: '已退款',
    3: '已冻结'
  }
  return map[status] || '未知'
}

function clearPayDisplay() {
  payUrl.value = ''
  htmlPayload.value = ''
  qrCodeUrl.value = ''
}

function normalizePayHtml(raw: string) {
  if (!raw) return raw
  let html = raw
  html = html.replace(/<form\b([^>]*)>/i, (all, attrs) => {
    if (/accept-charset\s*=/i.test(attrs)) return all
    return `<form${attrs} accept-charset="UTF-8">`
  })
  html = html.replace(
    /<script>\s*document\.getElementById\('payform'\)\.submit\(\);\s*<\/script>/i,
    "<script>(function(){var f=document.getElementById('payform');if(f){f.acceptCharset='UTF-8';f.submit();}})();<\\/script>"
  )
  return html
}

async function renderSubmitResult(result: any) {
  clearPayDisplay()
  if (!result) return

  const resultType = String(result.type || result.Type || '').toLowerCase()
  const resultURL = result.url || result.URL || ''
  const resultData = result.data ?? result.Data

  if (resultType === 'qrcode' && resultURL) {
    qrCodeUrl.value = await QRCode.toDataURL(resultURL, { width: 220, margin: 1 })
    return
  }

  if (resultType === 'jump' && resultURL) {
    payUrl.value = resultURL
    return
  }

  if (resultType === 'html' && typeof resultData === 'string') {
    htmlPayload.value = resultData
    return
  }

  if (resultURL) {
    payUrl.value = resultURL
  }
}

function goToPayPage() {
  if (payUrl.value) {
    window.location.href = payUrl.value
    return true
  }
  if (htmlPayload.value) {
    const html = normalizePayHtml(htmlPayload.value)
    document.body.innerHTML = html
    return true
  }
  return false
}

function genOutTradeNo() {
  const rand = Math.random().toString(36).slice(2, 8).toUpperCase()
  return `CASHIER_${Date.now()}_${rand}`
}

async function queryOrderByTradeNo() {
  if (!tradeNo.value && !tradeNoParam.value) return
  const targetTradeNo = tradeNo.value || tradeNoParam.value

  try {
    const queryPid = pid.value || Number(route.query.pid || 0)
    if (!queryPid) {
      if (!isFixedMode.value) {
        orderInfo.value = null
      }
      return
    }
    const queryParams = {
      pid: queryPid,
      trade_no: targetTradeNo
    }
    const sign = apiKey.value ? await makeOpenAPISign(queryParams, apiKey.value) : ''
    const res = await payQuery({
      ...queryParams,
      sign: sign || undefined,
      sign_type: sign ? 'HMAC-SHA256' : undefined
    })
    orderInfo.value = res
  } catch (error: any) {
    if (!isFixedMode.value) {
      orderInfo.value = null
    }
    console.error('查单失败:', error)
  }
}

function startPollingOrder() {
  stopPollingOrder()
  pollStartAt = Date.now()
  pollTimer = window.setInterval(async () => {
    if (!tradeNo.value) {
      stopPollingOrder()
      return
    }
    await queryOrderByTradeNo()
    if (orderInfo.value && Number(orderInfo.value.status) === 1) {
      ElMessage.success('支付成功')
      stopPollingOrder()
      return
    }
    if (Date.now() - pollStartAt > 2 * 60 * 1000) {
      stopPollingOrder()
    }
  }, 3000)
}

function stopPollingOrder() {
  if (pollTimer !== null) {
    window.clearInterval(pollTimer)
    pollTimer = null
  }
}

async function submitOrder() {
  const targetPid = effectivePid.value || pid.value
  if (!targetPid) {
    ElMessage.error('无效的商户ID')
    return
  }
  if (!form.value.type) {
    ElMessage.warning('请选择支付方式')
    return
  }
  if (!form.value.money || Number(form.value.money) <= 0) {
    ElMessage.warning('请输入有效金额')
    return
  }

  submitting.value = true
  try {
    const submitParams = {
      pid: targetPid,
      type: form.value.type,
      out_trade_no: genOutTradeNo(),
      name: form.value.remark || '收银台订单',
      money: Number(form.value.money),
      notify_url: '',
      return_url: '',
      param: `cashier_user_${targetPid}`
    }
    const sign = apiKey.value ? await makeOpenAPISign(submitParams, apiKey.value) : ''
    const res = sign
      ? await paySubmit({
        ...submitParams,
        sign: sign || undefined,
        sign_type: sign ? 'HMAC-SHA256' : undefined
      })
      : await payCashierSubmit(submitParams)

    tradeNo.value = res.trade_no || ''
    await renderSubmitResult(res.result)
    await queryOrderByTradeNo()
    startPollingOrder()
    const redirected = goToPayPage()
    if (!redirected) {
      ElMessage.success('订单已创建，请扫码完成支付')
    }
  } catch (error: any) {
    ElMessage.error(error?.message || '下单失败')
  } finally {
    submitting.value = false
  }
}

async function initFixedCashier() {
  if (!pid.value) {
    ElMessage.error('无效收款码')
    return
  }

  const res = await getPayTypes(pid.value)
  effectivePid.value = Number((res as any)?.pid || pid.value)
  cashierTip.value = String((res as any)?.msg || '').trim()
  payTypes.value = Array.isArray(res.data) ? res.data : []

  const queryKey = String(route.query.key || '').trim()
  if (queryKey) {
    apiKey.value = queryKey
  }

  const queryType = Number(route.query.type || 0)
  if (queryType > 0 && payTypes.value.some((pt: any) => Number(pt.id) === queryType)) {
    form.value.type = queryType
  } else if (payTypes.value.length > 0) {
    form.value.type = Number(payTypes.value[0].id)
  }

  const queryRemark = String(route.query.remark || '').trim()
  if (queryRemark) {
    form.value.remark = queryRemark
  }

  const queryAmount = Number(route.query.amount || 0)
  if (queryAmount > 0) {
    form.value.money = Number(queryAmount.toFixed(2))
  }

  const autoPay = String(route.query.autopay || '').trim() === '1'
  if (autoPay && queryAmount > 0 && form.value.type > 0 && !autoPayTriggered.value) {
    autoPayTriggered.value = true
    await submitOrder()
  }
}

onMounted(async () => {
  try {
    if (isFixedMode.value) {
      await initFixedCashier()
      loading.value = false
      return
    }

    if (!tradeNoParam.value) {
      loading.value = false
      return
    }
    await queryOrderByTradeNo()
  } catch (error: any) {
    console.error('初始化失败:', error)
  } finally {
    loading.value = false
  }
})

onBeforeUnmount(() => {
  stopPollingOrder()
})
</script>
