<template>
  <div class="space-y-4">
    <!-- 页面标题 -->
    <div class="flex items-center justify-between">
      <div>
        <h1 class="text-2xl font-bold text-gray-900">余额充值</h1>
        <p class="text-sm text-gray-500 mt-1">查看账户余额和充值记录</p>
      </div>
    </div>

    <!-- 余额卡片 -->
    <div class="grid grid-cols-1 md:grid-cols-3 gap-4">
      <div class="bg-gradient-to-br from-blue-500 to-blue-600 rounded-xl p-6 text-white">
        <div class="flex items-center gap-3 mb-4">
          <div class="w-10 h-10 bg-white/20 rounded-lg flex items-center justify-center">
            <Wallet class="w-5 h-5" />
          </div>
          <div>
            <p class="text-blue-100 text-sm">当前余额</p>
            <p class="text-2xl font-bold">¥{{ userInfo?.money?.toFixed(2) || '0.00' }}</p>
          </div>
        </div>
      </div>

      <div class="bg-white rounded-xl p-6 border border-gray-100 shadow-sm">
        <div class="flex items-center gap-3 mb-4">
          <div class="w-10 h-10 bg-green-100 rounded-lg flex items-center justify-center">
            <TrendingUp class="w-5 h-5 text-green-600" />
          </div>
          <div>
            <p class="text-gray-500 text-sm">累计充值</p>
            <p class="text-xl font-bold text-gray-800">¥{{ stats.totalRecharge?.toFixed(2) || '0.00' }}</p>
          </div>
        </div>
      </div>

      <div class="bg-white rounded-xl p-6 border border-gray-100 shadow-sm">
        <div class="flex items-center gap-3 mb-4">
          <div class="w-10 h-10 bg-orange-100 rounded-lg flex items-center justify-center">
            <ArrowDownCircle class="w-5 h-5 text-orange-600" />
          </div>
          <div>
            <p class="text-gray-500 text-sm">累计支出</p>
            <p class="text-xl font-bold text-gray-800">¥{{ stats.totalExpense?.toFixed(2) || '0.00' }}</p>
          </div>
        </div>
      </div>
    </div>

    <!-- 充值说明 -->
    <div class="bg-blue-50 rounded-xl p-4 border border-blue-100">
      <div class="flex gap-3">
        <Info class="w-5 h-5 text-blue-600 flex-shrink-0 mt-0.5" />
        <div class="text-sm text-blue-800">
          <p class="font-medium mb-1">充值说明</p>
          <ul class="list-disc list-inside space-y-1 text-blue-700">
            <li>余额用于支付交易手续费和开通增值服务</li>
            <li>充值后余额可在结算时抵扣手续费</li>
            <li>如有疑问请联系客服</li>
          </ul>
        </div>
      </div>
    </div>

    <!-- 充值记录 -->
    <div class="bg-white rounded-xl border border-gray-100 shadow-sm overflow-hidden">
      <div class="px-4 py-3 border-b border-gray-100 flex items-center justify-between">
        <h3 class="font-semibold text-gray-700">充值记录</h3>
        <div class="flex items-center gap-2">
          <select v-model="filterType"
            class="px-3 py-1.5 border border-gray-200 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-500">
            <option value="">全部类型</option>
            <option value="1">充值</option>
            <option value="2">退款</option>
            <option value="3">提现</option>
            <option value="4">消费</option>
          </select>
          <button @click="page = 1; fetchRecords()"
            class="px-3 py-1.5 bg-blue-600 text-white rounded-lg text-sm hover:bg-blue-700">
            筛选
          </button>
        </div>
      </div>
      <div class="overflow-x-auto">
        <table class="w-full text-sm whitespace-nowrap">
          <thead>
            <tr class="bg-gray-50 border-b border-gray-100">
              <th class="px-4 py-3 text-left font-semibold text-gray-600">时间</th>
              <th class="px-4 py-3 text-center font-semibold text-gray-600">类型</th>
              <th class="px-4 py-3 text-right font-semibold text-gray-600">金额</th>
              <th class="px-4 py-3 text-right font-semibold text-gray-600">余额</th>
              <th class="px-4 py-3 text-left font-semibold text-gray-600">备注</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-gray-50">
            <tr v-for="r in records" :key="r.id" class="hover:bg-gray-50/50 transition-colors">
              <td class="px-4 py-3 text-gray-500 text-xs">{{ formatTime(r.date) }}</td>
              <td class="px-4 py-3 text-center">
                <span :class="['inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium',
                  typeClass(r.action)]">
                  {{ typeName(r.action) }}
                </span>
              </td>
              <td class="px-4 py-3 text-right">
                <span :class="['font-medium', r.money >= 0 ? 'text-green-600' : 'text-red-600']">
                  {{ r.money >= 0 ? '+' : '' }}{{ r.money.toFixed(2) }}
                </span>
              </td>
              <td class="px-4 py-3 text-right text-gray-600">¥{{ r.newmoney.toFixed(2) }}</td>
              <td class="px-4 py-3 text-gray-500 text-xs">{{ r.trade_no || '-' }}</td>
            </tr>
            <tr v-if="records.length === 0">
              <td colspan="5" class="px-4 py-12 text-center text-gray-400">
                <div class="flex flex-col items-center">
                  <Receipt class="w-10 h-10 text-gray-300 mb-2" />
                  <span>暂无充值记录</span>
                </div>
              </td>
            </tr>
          </tbody>
        </table>
      </div>

      <!-- 分页 -->
      <div class="px-4 py-3 border-t border-gray-100 flex items-center justify-between">
        <div class="text-sm text-gray-500">共 {{ total }} 条</div>
        <div class="flex items-center gap-2">
          <button @click="page--; fetchRecords()" :disabled="page <= 1"
            class="px-3 py-1 text-sm border border-gray-200 rounded hover:bg-gray-50 disabled:opacity-50">
            上一页
          </button>
          <span class="px-3 py-1 text-sm">{{ page }} / {{ totalPages }}</span>
          <button @click="page++; fetchRecords()" :disabled="page >= totalPages"
            class="px-3 py-1 text-sm border border-gray-200 rounded hover:bg-gray-50 disabled:opacity-50">
            下一页
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { getUserRecords } from '@/api/user'
import { useAppStore } from '@/stores/app'
import { Wallet, TrendingUp, ArrowDownCircle, Info, Receipt } from 'lucide-vue-next'

const appStore = useAppStore()
const userInfo = computed(() => appStore.userInfo)

const records = ref<any[]>([])
const page = ref(1)
const pageSize = ref(20)
const total = ref(0)
const filterType = ref('')

const stats = ref({
  totalRecharge: 0,
  totalExpense: 0
})

const totalPages = computed(() => Math.ceil(total.value / pageSize.value) || 1)

function typeName(action: number) {
  const map: Record<number, string> = {
    1: '充值',
    2: '退款',
    3: '提现',
    4: '消费',
    5: '结算',
    6: '返现'
  }
  return map[action] || '其他'
}

function typeClass(action: number) {
  const map: Record<number, string> = {
    1: 'bg-green-100 text-green-700',
    2: 'bg-blue-100 text-blue-700',
    3: 'bg-orange-100 text-orange-700',
    4: 'bg-red-100 text-red-700',
    5: 'bg-purple-100 text-purple-700',
    6: 'bg-yellow-100 text-yellow-700'
  }
  return map[action] || 'bg-gray-100 text-gray-700'
}

function formatTime(time: string) {
  if (!time) return '-'
  return new Date(time).toLocaleString('zh-CN')
}

async function fetchRecords() {
  try {
    const params: any = { page: page.value, limit: pageSize.value }
    if (filterType.value) {
      params.action = filterType.value
    }
    const res = await getUserRecords(params)
    if (res.code === 0) {
      records.value = res.data || []
      total.value = res.count || 0

      // 计算统计
      let recharge = 0
      let expense = 0
      records.value.forEach(r => {
        if (r.action === 1 && r.money > 0) recharge += r.money
        if (r.action === 4 && r.money < 0) expense += Math.abs(r.money)
      })
      stats.value.totalRecharge = recharge
      stats.value.totalExpense = expense
    }
  } catch (error) {
    console.error('获取记录失败:', error)
  }
}

onMounted(() => {
  fetchRecords()
})
</script>
