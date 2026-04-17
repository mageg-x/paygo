<template>
  <div class="space-y-4">
    <!-- 页面标题 -->
    <div class="flex items-center justify-between">
      <div>
        <h1 class="text-2xl font-bold text-gray-900">收款二维码</h1>
        <p class="text-sm text-gray-500 mt-1">生成聚合收款码，方便客户扫码支付</p>
      </div>
    </div>

    <!-- 收款码生成 -->
    <div class="bg-white rounded-xl border border-gray-100 shadow-sm p-6">
      <div class="grid grid-cols-1 md:grid-cols-2 gap-8">
        <!-- 配置区 -->
        <div class="space-y-4">
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-2">收款金额 (元)</label>
            <div class="relative">
              <span class="absolute left-4 top-1/2 -translate-y-1/2 text-gray-400">¥</span>
              <input v-model.number="amount" type="number" step="0.01" min="0.01"
                class="w-full pl-10 pr-4 py-3 border border-gray-200 rounded-lg text-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
                placeholder="请输入收款金额" />
            </div>
          </div>

          <div>
            <label class="block text-sm font-medium text-gray-700 mb-2">订单备注</label>
            <input v-model="remark" type="text"
              class="w-full px-4 py-2 border border-gray-200 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
              placeholder="可选填写" />
          </div>

          <div>
            <label class="block text-sm font-medium text-gray-700 mb-2">收款通道</label>
            <select v-model="selectedChannel"
              class="w-full px-4 py-2 border border-gray-200 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-500">
              <option value="">自动选择</option>
              <option v-for="ch in channels" :key="ch.id" :value="ch.id">{{ ch.name }}</option>
            </select>
          </div>

          <button @click="generateQRCode"
            class="w-full py-3 bg-blue-600 text-white rounded-lg font-medium hover:bg-blue-700 transition-colors flex items-center justify-center gap-2">
            <QRCode class="w-5 h-5" />
            生成收款码
          </button>
        </div>

        <!-- 预览区 -->
        <div class="flex flex-col items-center justify-center">
          <div class="bg-white border-2 border-dashed border-gray-200 rounded-2xl p-8 w-full max-w-xs">
            <div v-if="!qrCodeUrl" class="flex flex-col items-center text-gray-400">
              <QRCode class="w-32 h-32 mb-4 opacity-30" />
              <p class="text-sm">输入金额后点击生成</p>
            </div>
            <div v-else class="flex flex-col items-center">
              <img :src="qrCodeUrl" alt="收款二维码" class="w-48 h-48 mb-4" />
              <p class="text-2xl font-bold text-gray-800">¥{{ amount }}</p>
              <p v-if="remark" class="text-sm text-gray-500 mt-1">{{ remark }}</p>
              <div class="mt-4 flex gap-2">
                <button @click="downloadQRCode"
                  class="px-4 py-2 text-sm bg-green-600 text-white rounded-lg hover:bg-green-700 transition-colors">
                  下载二维码
                </button>
                <button @click="copyPayLink"
                  class="px-4 py-2 text-sm bg-gray-100 text-gray-700 rounded-lg hover:bg-gray-200 transition-colors">
                  复制链接
                </button>
              </div>
            </div>
          </div>
          <p class="text-xs text-gray-400 mt-4">扫码即可支付，资金直接进入您的账户</p>
        </div>
      </div>
    </div>

    <!-- 最近记录 -->
    <div class="bg-white rounded-xl border border-gray-100 shadow-sm overflow-hidden">
      <div class="px-4 py-3 border-b border-gray-100">
        <h3 class="font-semibold text-gray-700">最近收款记录</h3>
      </div>
      <div class="overflow-x-auto">
        <table class="w-full text-sm whitespace-nowrap">
          <thead>
            <tr class="bg-gray-50 border-b border-gray-100">
              <th class="px-4 py-3 text-left font-semibold text-gray-600">订单号</th>
              <th class="px-4 py-3 text-right font-semibold text-gray-600">金额</th>
              <th class="px-4 py-3 text-center font-semibold text-gray-600">通道</th>
              <th class="px-4 py-3 text-center font-semibold text-gray-600">状态</th>
              <th class="px-4 py-3 text-left font-semibold text-gray-600">时间</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-gray-50">
            <tr v-for="o in recentOrders" :key="o.id" class="hover:bg-gray-50/50 transition-colors">
              <td class="px-4 py-3 font-mono text-xs text-gray-600">{{ o.trade_no }}</td>
              <td class="px-4 py-3 text-right font-medium text-gray-900">¥{{ o.money }}</td>
              <td class="px-4 py-3 text-center text-gray-500">{{ o.type_name || '-' }}</td>
              <td class="px-4 py-3 text-center">
                <span :class="['inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium',
                  o.status === 1 ? 'bg-green-100 text-green-700' : o.status === 0 ? 'bg-yellow-100 text-yellow-700' : 'bg-red-100 text-red-700']">
                  {{ statusName(o.status) }}
                </span>
              </td>
              <td class="px-4 py-3 text-gray-500 text-xs">{{ formatTime(o.addtime) }}</td>
            </tr>
            <tr v-if="recentOrders.length === 0">
              <td colspan="5" class="px-4 py-8 text-center text-gray-400">
                <div class="flex flex-col items-center">
                  <Receipt class="w-10 h-10 text-gray-300 mb-2" />
                  <span>暂无收款记录</span>
                </div>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import QRCode from 'qrcode'
import { getUserOrders } from '@/api/user'
import { ElMessage } from 'element-plus'
import { QRCode as QRIcon, Receipt } from 'lucide-vue-next'

const amount = ref<number>()
const remark = ref('')
const selectedChannel = ref('')
const qrCodeUrl = ref('')
const payLink = ref('')
const recentOrders = ref<any[]>([])

const channels = [
  { id: 1, name: '微信支付' },
  { id: 2, name: '支付宝' },
  { id: 3, name: 'QQ钱包' },
  { id: 4, name: '银行卡' }
]

function statusName(status: number) {
  const map: Record<number, string> = {
    0: '待支付',
    1: '已支付',
    2: '已关闭'
  }
  return map[status] || '未知'
}

function formatTime(time: string) {
  if (!time) return '-'
  return new Date(time).toLocaleString('zh-CN')
}

async function generateQRCode() {
  if (!amount.value || amount.value <= 0) {
    ElMessage.warning('请输入有效的收款金额')
    return
  }

  // 生成订单号
  const tradeNo = 'QR' + Date.now() + Math.random().toString(36).substr(2, 6).toUpperCase()

  // 构建支付链接 - 跳转到收银台
  const baseUrl = window.location.origin
  payLink.value = `${baseUrl}/cashier/${tradeNo}?amount=${amount.value}&remark=${encodeURIComponent(remark.value)}`

  try {
    qrCodeUrl.value = await QRCode.toDataURL(payLink.value, {
      width: 200,
      margin: 2,
      color: {
        dark: '#000000',
        light: '#ffffff'
      }
    })
    ElMessage.success('收款码已生成')
  } catch (error) {
    console.error('生成二维码失败:', error)
    ElMessage.error('生成二维码失败')
  }
}

function downloadQRCode() {
  if (!qrCodeUrl.value) return

  const link = document.createElement('a')
  link.download = `收款码_${amount.value}元.png`
  link.href = qrCodeUrl.value
  link.click()
}

function copyPayLink() {
  if (!payLink.value) return
  navigator.clipboard.writeText(payLink.value).then(() => {
    ElMessage.success('支付链接已复制')
  }).catch(() => {
    ElMessage.error('复制失败')
  })
}

async function fetchRecentOrders() {
  try {
    const res = await getUserOrders({ page: 1, limit: 10 })
    if (res.code === 0) {
      // 过滤出二维码收款订单
      recentOrders.value = (res.data || []).filter((o: any) => o.trade_no.startsWith('QR')).slice(0, 5)
    }
  } catch (error) {
    console.error('获取订单失败:', error)
  }
}

onMounted(() => {
  fetchRecentOrders()
})
</script>
