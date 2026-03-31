<template>
  <div>
    <div class="flex justify-between items-center mb-6">
      <h2 class="text-2xl font-bold text-gray-800 ">订单管理</h2>
      <select v-model="status" @change="fetchOrders" class="form-input w-32 max-w-28">
        <option :value="-1">全部</option>
        <option :value="0">待支付</option>
        <option :value="1">已支付</option>
        <option :value="2">已退款</option>
        <option :value="3">已冻结</option>
      </select>
    </div>

    <div class="card">
      <div class="card-body">
        <table class="table">
          <thead>
            <tr>
              <th>订单号</th>
              <th>商户订单号</th>
              <th>商户ID</th>
              <th>商品名称</th>
              <th>金额</th>
              <th>支付方式</th>
              <th>状态</th>
              <th>下单时间</th>
              <th>操作</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="order in orders" :key="order.trade_no">
              <td class="text-xs">{{ order.trade_no }}</td>
              <td class="text-xs">{{ order.out_trade_no }}</td>
              <td>{{ order.uid }}</td>
              <td>{{ order.name }}</td>
              <td class="text-primary-600 font-medium">¥{{ order.money }}</td>
              <td>{{ order.typename || '未知' }}</td>
              <td>
                <span :class="['badge', statusMap[order.status]?.class]">
                  {{ statusMap[order.status]?.text || '未知' }}
                </span>
              </td>
              <td class="text-xs">{{ formatTime(order.addtime) }}</td>
              <td>
                <button v-if="order.status === 1" class="text-danger hover:text-danger mr-2"
                  @click="handleOp('refund', order.trade_no)">
                  退款
                </button>
                <button v-if="order.status === 0" class="text-warning hover:text-warning mr-2"
                  @click="handleOp('freeze', order.trade_no)">
                  冻结
                </button>
                <button v-if="order.status === 3" class="text-success hover:text-success"
                  @click="handleOp('unfreeze', order.trade_no)">
                  解冻
                </button>
              </td>
            </tr>
          </tbody>
        </table>

        <!-- 分页 -->
        <div class="pagination text-sm">
          <button class="pagination-item" :disabled="page === 1" @click="page--; fetchOrders()">
            上一页
          </button>
          <span class="px-4 py-1">第 {{ page }} / {{ Math.ceil(total / 20) || 1 }} 页</span>
          <button class="pagination-item" :disabled="page * 20 >= total" @click="page++; fetchOrders()">
            下一页
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { getOrderList, orderOp } from '@/api/admin'
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
const total = ref(0)
const status = ref(-1)

const statusMap: Record<number, { text: string; class: string }> = {
  0: { text: '待支付', class: 'badge-warning' },
  1: { text: '已支付', class: 'badge-success' },
  2: { text: '已退款', class: 'badge-info' },
  3: { text: '已冻结', class: 'badge-danger' }
}

async function fetchOrders() {
  loading.value = true
  try {
    const res = await getOrderList({ page: page.value, limit: 20, status: status.value })
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
  if (confirm(`确定要执行该操作吗？`)) {
    try {
      const res = await orderOp({ action, trade_no: tradeNo })
      if (res.code === 0) {
        alert('操作成功')
        fetchOrders()
      }
    } catch (error) {
      console.error('操作失败:', error)
    }
  }
}

function formatTime(time: string) {
  return dayjs(time).format('YYYY-MM-DD HH:mm')
}

onMounted(() => {
  fetchOrders()
})
</script>
