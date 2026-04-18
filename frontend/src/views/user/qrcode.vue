<template>
  <div class="space-y-4">
    <div class="flex items-center justify-between">
      <div>
        <h1 class="text-2xl font-bold text-gray-900">固定收款码</h1>
        <p class="text-sm text-gray-500 mt-1">生成一个可反复扫码的收款码，支持自定义金额或固定额度直付</p>
      </div>
    </div>

    <div class="bg-white rounded-xl border border-gray-100 shadow-sm p-6">
      <div class="grid grid-cols-1 md:grid-cols-2 gap-8">
        <div class="space-y-4">
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-2">收款商户ID</label>
            <input
              :value="uidText"
              type="text"
              readonly
              class="w-full px-4 py-2 bg-gray-50 border border-gray-200 rounded-lg text-sm text-gray-600"
            />
          </div>

          <div>
            <label class="block text-sm font-medium text-gray-700 mb-2">默认支付方式（可选）</label>
            <select
              v-model.number="defaultType"
              class="w-full px-4 py-2 border border-gray-200 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
            >
              <option :value="0">不指定（扫码页自行选择）</option>
              <option v-for="pt in payTypes" :key="pt.id" :value="Number(pt.id)">
                {{ pt.showname || pt.name || ('类型' + pt.id) }}
              </option>
            </select>
          </div>

          <div>
            <label class="block text-sm font-medium text-gray-700 mb-2">默认备注（可选）</label>
            <input
              v-model.trim="defaultRemark"
              type="text"
              placeholder="例如：门店收款"
              class="w-full px-4 py-2 border border-gray-200 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
            />
          </div>

          <div class="space-y-2">
            <label class="block text-sm font-medium text-gray-700">金额模式</label>
            <div class="grid grid-cols-2 gap-2">
              <button
                type="button"
                :class="[
                  'px-3 py-2 text-sm rounded-lg border transition-colors',
                  fixedAmountMode ? 'border-gray-200 text-gray-600 bg-gray-50 hover:bg-gray-100' : 'border-blue-500 text-blue-700 bg-blue-50'
                ]"
                @click="fixedAmountMode = false"
              >
                扫码输入金额
              </button>
              <button
                type="button"
                :class="[
                  'px-3 py-2 text-sm rounded-lg border transition-colors',
                  fixedAmountMode ? 'border-blue-500 text-blue-700 bg-blue-50' : 'border-gray-200 text-gray-600 bg-gray-50 hover:bg-gray-100'
                ]"
                @click="fixedAmountMode = true"
              >
                固定额度直付
              </button>
            </div>
          </div>

          <div v-if="fixedAmountMode">
            <label class="block text-sm font-medium text-gray-700 mb-2">固定收款金额（元）</label>
            <input
              v-model.number="fixedAmount"
              type="number"
              min="0.01"
              step="0.01"
              placeholder="例如：10.00"
              class="w-full px-4 py-2 border border-gray-200 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
            />
          </div>

          <button
            @click="generateQRCode"
            class="w-full py-3 bg-blue-600 text-white rounded-lg font-medium hover:bg-blue-700 transition-colors"
          >
            生成/刷新固定收款码
          </button>

          <div class="text-xs text-gray-500 leading-5 bg-blue-50 border border-blue-100 rounded-lg p-3">
            该二维码是固定入口，可长期使用。固定额度模式下，付款人扫码后将自动发起支付。
          </div>
        </div>

        <div class="flex flex-col items-center justify-center">
          <div class="bg-white border-2 border-dashed border-gray-200 rounded-2xl p-8 w-full max-w-xs">
            <div v-if="!qrCodeUrl" class="flex flex-col items-center text-gray-400">
              <QrCode class="w-32 h-32 mb-4 opacity-30" />
              <p class="text-sm">点击左侧按钮生成二维码</p>
            </div>
            <div v-else class="flex flex-col items-center">
              <img :src="qrCodeUrl" alt="固定收款二维码" class="w-48 h-48 mb-4" />
              <p class="text-sm text-gray-600 text-center break-all">{{ payLink }}</p>
              <div class="mt-4 flex gap-2">
                <button
                  @click="downloadQRCode"
                  class="px-4 py-2 text-sm bg-green-600 text-white rounded-lg hover:bg-green-700 transition-colors"
                >
                  下载二维码
                </button>
                <button
                  @click="copyPayLink"
                  class="px-4 py-2 text-sm bg-gray-100 text-gray-700 rounded-lg hover:bg-gray-200 transition-colors"
                >
                  复制链接
                </button>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>

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
              <th class="px-4 py-3 text-center font-semibold text-gray-600">支付方式</th>
              <th class="px-4 py-3 text-center font-semibold text-gray-600">状态</th>
              <th class="px-4 py-3 text-left font-semibold text-gray-600">时间</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-gray-50">
            <tr v-for="o in recentOrders" :key="o.trade_no" class="hover:bg-gray-50/50 transition-colors">
              <td class="px-4 py-3 font-mono text-xs text-gray-600">{{ o.trade_no }}</td>
              <td class="px-4 py-3 text-right font-medium text-gray-900">¥{{ o.money }}</td>
              <td class="px-4 py-3 text-center text-gray-500">{{ o.typename || '-' }}</td>
              <td class="px-4 py-3 text-center">
                <span
                  :class="[
                    'inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium',
                    o.status === 1
                      ? 'bg-green-100 text-green-700'
                      : o.status === 0
                        ? 'bg-yellow-100 text-yellow-700'
                        : o.status === 2
                          ? 'bg-blue-100 text-blue-700'
                          : 'bg-red-100 text-red-700'
                  ]"
                >
                  {{ statusName(o.status) }}
                </span>
              </td>
              <td class="px-4 py-3 text-gray-500 text-xs">{{ formatTime(o.addtime) }}</td>
            </tr>
            <tr v-if="recentOrders.length === 0">
              <td colspan="5" class="px-4 py-8 text-center text-gray-400">暂无收款记录</td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import QRCode from 'qrcode'
import { getUserInfo, getUserOrders, getUserProfileAPI } from '@/api/user'
import { getPayTypes } from '@/api/pay'
import { ElMessage } from 'element-plus'
import { QrCode } from 'lucide-vue-next'

const uid = ref(0)
const payTypes = ref<any[]>([])
const defaultType = ref(0)
const defaultRemark = ref('')
const fixedAmountMode = ref(false)
const fixedAmount = ref<number | null>(null)
const qrCodeUrl = ref('')
const payLink = ref('')
const recentOrders = ref<any[]>([])
const apiKey = ref('')

const uidText = computed(() => (uid.value ? String(uid.value) : '未获取到'))

function statusName(status: number) {
  const map: Record<number, string> = {
    0: '待支付',
    1: '已支付',
    2: '已退款',
    3: '已冻结'
  }
  return map[status] || '未知'
}

function formatTime(time: string) {
  if (!time) return '-'
  return new Date(time).toLocaleString('zh-CN')
}

function buildPayLink() {
  if (!uid.value) {
    payLink.value = ''
    return
  }
  const baseUrl = window.location.origin
  const url = new URL(`${baseUrl}/cashier/user/${uid.value}`)
  if (defaultType.value > 0) {
    url.searchParams.set('type', String(defaultType.value))
  }
  if (defaultRemark.value) {
    url.searchParams.set('remark', defaultRemark.value)
  }
  if (fixedAmountMode.value) {
    const amount = Number(fixedAmount.value || 0)
    if (amount <= 0) {
      throw new Error('固定金额必须大于0')
    }
    url.searchParams.set('amount', amount.toFixed(2))
    url.searchParams.set('autopay', '1')
  }
  payLink.value = url.toString()
}

async function generateQRCode() {
  if (!uid.value) {
    ElMessage.warning('未获取到商户ID')
    return
  }

  try {
    buildPayLink()
  } catch (e: any) {
    ElMessage.warning(e?.message || '请检查二维码参数')
    return
  }
  if (!payLink.value) {
    ElMessage.error('生成链接失败')
    return
  }

  try {
    qrCodeUrl.value = await QRCode.toDataURL(payLink.value, {
      width: 220,
      margin: 2,
      color: {
        dark: '#000000',
        light: '#ffffff'
      }
    })
    ElMessage.success('固定收款码已生成')
  } catch (error) {
    console.error('生成二维码失败:', error)
    ElMessage.error('生成二维码失败')
  }
}

function downloadQRCode() {
  if (!qrCodeUrl.value) return

  const link = document.createElement('a')
  link.download = `固定收款码_UID${uid.value}.png`
  link.href = qrCodeUrl.value
  link.click()
}

function copyPayLink() {
  if (!payLink.value) return
  navigator.clipboard
    .writeText(payLink.value)
    .then(() => ElMessage.success('收款链接已复制'))
    .catch(() => ElMessage.error('复制失败'))
}

async function fetchRecentOrders() {
  try {
    const res = await getUserOrders({ page: 1, limit: 20 })
    if (res.code === 0) {
      const list = Array.isArray(res.data) ? res.data : []
      recentOrders.value = list
        .filter((o: any) => String(o?.param || '').startsWith('cashier_user_'))
        .slice(0, 10)
    }
  } catch (error) {
    console.error('获取订单失败:', error)
  }
}

async function initData() {
  try {
    const infoRes = await getUserInfo()
    if (infoRes.code === 0 && infoRes.data?.uid) {
      uid.value = Number(infoRes.data.uid)
      const profileRes = await getUserProfileAPI()
      if (profileRes.code === 0 && profileRes.data?.key) {
        apiKey.value = String(profileRes.data.key)
      }
      try {
        const typeRes = await getPayTypes(uid.value)
        if (typeRes.code === 0) {
          payTypes.value = Array.isArray(typeRes.data) ? typeRes.data : []
        }
      } catch (error) {
        console.error('获取支付方式失败:', error)
      }
      buildPayLink()
      await generateQRCode()
    }
    await fetchRecentOrders()
  } catch (error) {
    console.error('初始化失败:', error)
  }
}

onMounted(() => {
  initData()
})
</script>
