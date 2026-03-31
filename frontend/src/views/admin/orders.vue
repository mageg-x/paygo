<template>
  <div class="space-y-4">
    <!-- 页面标题 -->
    <div class="flex items-center justify-between">
      <div>
        <h1 class="text-2xl font-bold text-gray-900">订单管理</h1>
        <p class="text-sm text-gray-500 mt-1">查看和处理所有支付订单</p>
      </div>
      <select v-model="status" @change="page = 1; fetchOrders()"
        class="px-4 py-2 border border-gray-200 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-500">
        <option :value="-1">全部状态</option>
        <option :value="0">待支付</option>
        <option :value="1">已支付</option>
        <option :value="2">已退款</option>
        <option :value="3">已冻结</option>
      </select>
    </div>

    <!-- 统计卡片 -->
    <div class="grid grid-cols-4 gap-4">
      <div class="bg-white rounded-xl p-4 border border-gray-100 shadow-sm">
        <div class="text-sm text-gray-500">全部订单</div>
        <div class="text-2xl font-bold text-gray-900 mt-1">{{ total }}</div>
      </div>
      <div class="bg-white rounded-xl p-4 border border-gray-100 shadow-sm">
        <div class="text-sm text-gray-500">待支付</div>
        <div class="text-2xl font-bold text-yellow-600 mt-1">{{ statusCount(0) }}</div>
      </div>
      <div class="bg-white rounded-xl p-4 border border-gray-100 shadow-sm">
        <div class="text-sm text-gray-500">已支付</div>
        <div class="text-2xl font-bold text-green-600 mt-1">{{ statusCount(1) }}</div>
      </div>
      <div class="bg-white rounded-xl p-4 border border-gray-100 shadow-sm">
        <div class="text-sm text-gray-500">已退款/冻结</div>
        <div class="text-2xl font-bold text-red-600 mt-1">{{ statusCount(2) + statusCount(3) }}</div>
      </div>
    </div>

    <!-- 订单列表 -->
    <div class="bg-white rounded-xl border border-gray-100 shadow-sm overflow-hidden">
      <div class="overflow-x-auto">
        <table class="w-full text-sm">
          <thead>
            <tr class="bg-gray-50 border-b border-gray-100">
              <th class="px-4 py-3 text-left font-semibold text-gray-600">订单号</th>
              <th class="px-4 py-3 text-left font-semibold text-gray-600">商户订单号</th>
              <th class="px-4 py-3 text-left font-semibold text-gray-600">商户ID</th>
              <th class="px-4 py-3 text-left font-semibold text-gray-600">商品名称</th>
              <th class="px-4 py-3 text-right font-semibold text-gray-600">金额</th>
              <th class="px-4 py-3 text-center font-semibold text-gray-600">支付方式</th>
              <th class="px-4 py-3 text-center font-semibold text-gray-600">状态</th>
              <th class="px-4 py-3 text-left font-semibold text-gray-600">下单时间</th>
              <th class="px-4 py-3 text-center font-semibold text-gray-600">操作</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-gray-50">
            <tr v-for="order in orders" :key="order.trade_no" class="hover:bg-gray-50/50 transition-colors">
              <td class="px-4 py-3 text-gray-500 font-mono text-xs">{{ order.trade_no }}</td>
              <td class="px-4 py-3 text-gray-500 font-mono text-xs">{{ order.out_trade_no || '-' }}</td>
              <td class="px-4 py-3 text-gray-900">{{ order.uid }}</td>
              <td class="px-4 py-3 text-gray-900">{{ order.name || '-' }}</td>
              <td class="px-4 py-3 text-right font-semibold text-gray-900">￥{{ order.money }}</td>
              <td class="px-4 py-3 text-center">
                <span class="inline-flex items-center px-2 py-0.5 rounded text-xs font-medium bg-blue-50 text-blue-700">
                  {{ order.typename || '未知' }}
                </span>
              </td>
              <td class="px-4 py-3 text-center">
                <span
                  :class="['inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium', statusClass(order.status)]">
                  {{ statusMap[order.status]?.text || '未知' }}
                </span>
              </td>
              <td class="px-4 py-3 text-gray-500 text-xs">{{ formatTime(order.addtime) }}</td>
              <td class="px-4 py-3 text-center">
                <template v-if="order.status === 1">
                  <button @click="handleOp('refund', order.trade_no)"
                    class="px-3 py-1 text-xs text-red-600 hover:bg-red-50 rounded transition-colors">退款</button>
                </template>
                <template v-if="order.status === 0">
                  <button @click="handleOp('freeze', order.trade_no)"
                    class="px-3 py-1 text-xs text-yellow-600 hover:bg-yellow-50 rounded transition-colors">冻结</button>
                </template>
                <template v-if="order.status === 3">
                  <button @click="handleOp('unfreeze', order.trade_no)"
                    class="px-3 py-1 text-xs text-green-600 hover:bg-green-50 rounded transition-colors">解冻</button>
                </template>
              </td>
            </tr>
            <tr v-if="orders.length === 0">
              <td colspan="9" class="px-4 py-12 text-center text-gray-400">
                <div class="flex flex-col items-center">
                  <svg class="w-12 h-12 text-gray-300 mb-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5"
                      d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2" />
                  </svg>
                  <span>暂无订单数据</span>
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
          <button
            class="px-3 py-1 text-sm border border-gray-200 rounded hover:bg-gray-50 disabled:opacity-50 disabled:cursor-not-allowed"
            :disabled="page <= 1" @click="page--; fetchOrders()">上一页</button>
          <span class="px-3 py-1 text-sm">{{ page }} / {{ totalPages }}</span>
          <button
            class="px-3 py-1 text-sm border border-gray-200 rounded hover:bg-gray-50 disabled:opacity-50 disabled:cursor-not-allowed"
            :disabled="page >= totalPages" @click="page++; fetchOrders()">下一页</button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { getOrderList, orderOp } from '@/api/admin'
import { ElMessage } from 'element-plus'
import dayjs from 'dayjs'

interface Order {
  trade_no: string
  out_trade_no: string
  uid: number
  name: string
  money: number
  status: number
  type: number
  typename: string
  addtime: string
  endtime: string
}

const orders = ref<Order[]>([])
const loading = ref(false)
const page = ref(1)
const pageSize = ref(20)
const total = ref(0)
const status = ref(-1)

const totalPages = computed(() => Math.ceil(total.value / pageSize.value) || 1)

function statusCount(s: number) {
  return orders.value.filter(o => o.status === s).length
}

const statusMap: Record<number, { text: string }> = {
  0: { text: '待支付' },
  1: { text: '已支付' },
  2: { text: '已退款' },
  3: { text: '已冻结' }
}

function statusClass(s: number) {
  const map: Record<number, string> = {
    0: 'bg-yellow-100 text-yellow-700',
    1: 'bg-green-100 text-green-700',
    2: 'bg-blue-100 text-blue-700',
    3: 'bg-red-100 text-red-700'
  }
  return map[s] || 'bg-gray-100 text-gray-700'
}

async function fetchOrders() {
  loading.value = true
  try {
    const res = await getOrderList({ page: page.value, limit: pageSize.value, status: status.value })
    if (res.code === 0) {
      orders.value = res.data || []
      total.value = res.count || 0
    }
  } catch (error) {
    console.error('获取订单列表失败:', error)
  } finally {
    loading.value = false
  }
}

async function handleOp(action: string, tradeNo: string) {
  const actionText = { refund: '退款', freeze: '冻结', unfreeze: '解冻' }[action] || '操作'
  if (confirm(`确定要${actionText}该订单吗？`)) {
    try {
      const res = await orderOp({ action, trade_no: tradeNo })
      ElMessage.success(res.msg || `${actionText}成功`)
      fetchOrders()
    } catch (error) {
      console.error('操作失败:', error)
    }
  }
}

function formatTime(time: string) {
  if (!time) return '-'
  return dayjs(time).format('YYYY-MM-DD HH:mm')
}

onMounted(() => {
  fetchOrders()
})
</script>
