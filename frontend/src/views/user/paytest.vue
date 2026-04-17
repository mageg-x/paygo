<template>
  <div class="space-y-4">
    <div>
      <h1 class="text-2xl font-bold text-gray-900">接口下单测试</h1>
      <p class="text-sm text-gray-500 mt-1">商户后台模拟调用平台支付接口，验证下单与支付链路</p>
    </div>

    <div class="bg-white rounded-xl border border-gray-100 shadow-sm p-6 space-y-4">
      <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
        <div>
          <label class="block text-sm font-medium text-gray-700 mb-1">当前商户ID</label>
          <input
            :value="pidText"
            type="text"
            readonly
            class="w-full px-3 py-2 bg-gray-50 border border-gray-200 rounded-lg text-sm text-gray-600"
          />
        </div>
        <div>
          <label class="block text-sm font-medium text-gray-700 mb-1">支付类型</label>
          <select
            v-model.number="form.type"
            class="w-full px-3 py-2 border border-gray-200 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
            @change="loadChannels"
          >
            <option :value="0">请选择支付类型</option>
            <option v-for="pt in payTypes" :key="pt.id" :value="Number(pt.id)">
              {{ pt.showname || pt.name || ('类型' + pt.id) }}
            </option>
          </select>
        </div>
        <div>
          <label class="block text-sm font-medium text-gray-700 mb-1">指定通道（可选）</label>
          <select
            v-model.number="form.channel"
            class="w-full px-3 py-2 border border-gray-200 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
          >
            <option :value="0">自动选择通道</option>
            <option v-for="ch in channels" :key="ch.id" :value="Number(ch.id)">
              {{ ch.name }}（ID: {{ ch.id }}）
            </option>
          </select>
        </div>
        <div>
          <label class="block text-sm font-medium text-gray-700 mb-1">金额 (元)</label>
          <input
            v-model.number="form.money"
            type="number"
            min="0.01"
            step="0.01"
            class="w-full px-3 py-2 border border-gray-200 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
          />
        </div>
      </div>

      <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
        <div>
          <label class="block text-sm font-medium text-gray-700 mb-1">商户订单号</label>
          <div class="flex gap-2">
            <input
              v-model.trim="form.out_trade_no"
              type="text"
              class="flex-1 px-3 py-2 border border-gray-200 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
            />
            <button
              class="px-3 py-2 text-xs text-blue-600 border border-blue-200 rounded-lg hover:bg-blue-50"
              @click="regenerateOutTradeNo"
            >
              重新生成
            </button>
          </div>
        </div>
        <div>
          <label class="block text-sm font-medium text-gray-700 mb-1">商品名称</label>
          <input
            v-model.trim="form.name"
            type="text"
            class="w-full px-3 py-2 border border-gray-200 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
          />
        </div>
        <div>
          <label class="block text-sm font-medium text-gray-700 mb-1">异步通知地址</label>
          <input
            v-model.trim="form.notify_url"
            type="text"
            class="w-full px-3 py-2 border border-gray-200 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
          />
        </div>
        <div>
          <label class="block text-sm font-medium text-gray-700 mb-1">同步返回地址</label>
          <input
            v-model.trim="form.return_url"
            type="text"
            class="w-full px-3 py-2 border border-gray-200 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
          />
        </div>
      </div>

      <div class="flex flex-wrap gap-2 pt-2">
        <button
          class="px-4 py-2 text-sm bg-gray-100 text-gray-700 rounded-lg hover:bg-gray-200"
          @click="loadPayTypes"
        >
          刷新类型/通道
        </button>
        <button
          class="px-4 py-2 text-sm bg-blue-600 text-white rounded-lg hover:bg-blue-700 disabled:opacity-60"
          :disabled="submitting || !pid"
          @click="submitTest"
        >
          {{ submitting ? '下单中...' : '提交接口测试' }}
        </button>
        <button
          class="px-4 py-2 text-sm bg-emerald-600 text-white rounded-lg hover:bg-emerald-700 disabled:opacity-60"
          :disabled="!hasPayAction"
          @click="openPayAction"
        >
          打开支付页
        </button>
        <button
          class="px-4 py-2 text-sm bg-amber-500 text-white rounded-lg hover:bg-amber-600 disabled:opacity-60"
          :disabled="!tradeNo && !form.out_trade_no"
          @click="queryOrder"
        >
          查询订单状态
        </button>
      </div>
    </div>

    <div v-if="tradeNo" class="bg-white rounded-xl border border-gray-100 shadow-sm p-6 space-y-4">
      <div class="text-sm text-gray-600">
        平台订单号：
        <span class="font-mono text-gray-900">{{ tradeNo }}</span>
      </div>

      <div v-if="qrCodeUrl" class="text-center">
        <img :src="qrCodeUrl" alt="支付二维码" class="mx-auto w-56 h-56 border rounded-lg p-2 bg-white" />
      </div>

      <div v-if="htmlPayload" class="text-xs text-orange-600 bg-orange-50 border border-orange-200 rounded-lg px-3 py-2">
        当前为 HTML 表单支付，点击“打开支付页”可在新窗口自动提交。
      </div>

      <div v-if="payUrl" class="text-xs text-blue-600 break-all bg-blue-50 border border-blue-200 rounded-lg px-3 py-2">
        {{ payUrl }}
      </div>

      <div>
        <h3 class="text-sm font-semibold text-gray-700 mb-2">提交返回</h3>
        <pre class="bg-gray-900 text-gray-100 rounded-lg p-3 text-xs overflow-auto">{{ prettyResult }}</pre>
      </div>
    </div>

    <div v-if="orderInfo" class="bg-white rounded-xl border border-gray-100 shadow-sm p-6 space-y-2">
      <h3 class="text-sm font-semibold text-gray-700">订单查询结果</h3>
      <div class="text-sm text-gray-700">商户订单号：{{ orderInfo.out_trade_no }}</div>
      <div class="text-sm text-gray-700">金额：{{ orderInfo.money }}</div>
      <div class="text-sm text-gray-700">
        状态：
        <span class="font-semibold">{{ statusText(orderInfo.status) }}</span>
      </div>
      <div class="text-sm text-gray-700">支付类型：{{ orderInfo.type }}</div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { ElMessage } from 'element-plus'
import QRCode from 'qrcode'
import { getPayChannels, getPayTypes, payQuery, paySubmit } from '@/api/pay'
import { getUserInfo } from '@/api/user'
import { useAppStore } from '@/stores/app'

const appStore = useAppStore()

const pid = ref<number>(0)
const payTypes = ref<any[]>([])
const channels = ref<any[]>([])
const submitting = ref(false)
const tradeNo = ref('')
const submitResult = ref<any>(null)
const orderInfo = ref<any>(null)
const qrCodeUrl = ref('')
const payUrl = ref('')
const htmlPayload = ref('')

const form = ref({
  type: 0,
  channel: 0,
  out_trade_no: '',
  name: '商户接口测试订单',
  money: 0.01,
  notify_url: '',
  return_url: '',
  param: ''
})

const pidText = computed(() => (pid.value ? String(pid.value) : '未获取到'))
const prettyResult = computed(() => JSON.stringify(submitResult.value || {}, null, 2))
const hasPayAction = computed(() => Boolean(payUrl.value || htmlPayload.value))

function regenerateOutTradeNo() {
  const rand = Math.random().toString(36).slice(2, 8).toUpperCase()
  form.value.out_trade_no = `USRTEST_${Date.now()}_${rand}`
}

function clearPayDisplay() {
  qrCodeUrl.value = ''
  payUrl.value = ''
  htmlPayload.value = ''
}

function normalizePayHtml(raw: string) {
  if (!raw) return raw
  let html = raw

  // 不修改任何签名参数，仅补充提交编码，避免中文在浏览器提交时被错误转码。
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

async function loadPayTypes() {
  if (!pid.value) return
  try {
    const res = await getPayTypes(pid.value)
    payTypes.value = res.data || []
    if (payTypes.value.length === 0) {
      channels.value = []
      form.value.type = 0
      form.value.channel = 0
      return
    }

    if (!form.value.type) {
      form.value.type = Number(payTypes.value[0].id)
    }
    await loadChannels()
  } catch (error: any) {
    ElMessage.error(error?.message || '加载支付类型失败')
  }
}

async function loadChannels() {
  if (!pid.value || !form.value.type) return
  try {
    const res = await getPayChannels(pid.value, Number(form.value.type))
    channels.value = res.data || []
    if (!channels.value.some((ch: any) => Number(ch.id) === Number(form.value.channel))) {
      form.value.channel = 0
    }
  } catch (error: any) {
    ElMessage.error(error?.message || '加载通道失败')
  }
}

async function submitTest() {
  if (!pid.value) {
    ElMessage.warning('未获取到商户信息，请重新登录')
    return
  }
  if (!form.value.type) {
    ElMessage.warning('请选择支付类型')
    return
  }
  if (!form.value.out_trade_no) {
    regenerateOutTradeNo()
  }
  if (!form.value.money || Number(form.value.money) <= 0) {
    ElMessage.warning('金额必须大于 0')
    return
  }

  submitting.value = true
  orderInfo.value = null
  clearPayDisplay()

  try {
    const res = await paySubmit({
      pid: pid.value,
      type: Number(form.value.type),
      channel: form.value.channel ? Number(form.value.channel) : undefined,
      out_trade_no: form.value.out_trade_no,
      name: form.value.name,
      money: Number(form.value.money),
      notify_url: form.value.notify_url,
      return_url: form.value.return_url,
      param: form.value.param
    })

    tradeNo.value = res.trade_no || ''
    submitResult.value = res.result || null
    await renderSubmitResult(res.result)
    ElMessage.success('接口测试下单成功')
  } catch (error: any) {
    ElMessage.error(error?.message || '下单失败')
  } finally {
    submitting.value = false
  }
}

function openPayAction() {
  if (payUrl.value) {
    window.open(payUrl.value, '_blank')
    return
  }

  if (htmlPayload.value) {
    const win = window.open('', '_blank')
    if (!win) {
      ElMessage.error('浏览器阻止了新窗口，请允许弹窗后重试')
      return
    }
    const html = normalizePayHtml(htmlPayload.value)
    win.document.open()
    win.document.write(html)
    win.document.close()
    return
  }

  ElMessage.warning('当前返回结果没有可直接打开的支付链接')
}

async function queryOrder() {
  if (!pid.value) {
    ElMessage.warning('未获取到商户信息')
    return
  }

  try {
    const res = await payQuery({
      pid: pid.value,
      trade_no: tradeNo.value || undefined,
      out_trade_no: tradeNo.value ? undefined : form.value.out_trade_no
    })
    orderInfo.value = res
    ElMessage.success('查询成功')
  } catch (error: any) {
    ElMessage.error(error?.message || '查询失败')
  }
}

function statusText(status: number) {
  const map: Record<number, string> = {
    0: '待支付',
    1: '已支付',
    2: '已关闭'
  }
  return map[status] || `未知(${status})`
}

async function initUser() {
  if (appStore.userInfo?.uid) {
    pid.value = appStore.userInfo.uid
  } else {
    const res = await getUserInfo()
    if (res.data?.uid) {
      pid.value = Number(res.data.uid)
    }
  }
}

onMounted(async () => {
  const origin = window.location.origin
  form.value.notify_url = `${origin}/api/pay/notify/test`
  form.value.return_url = `${origin}/`
  regenerateOutTradeNo()

  try {
    await initUser()
    await loadPayTypes()
  } catch (error: any) {
    ElMessage.error(error?.message || '初始化失败')
  }
})
</script>
