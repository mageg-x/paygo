<template>
  <div>
    <h2 class="text-2xl font-bold text-gray-800 mb-6">我的订单</h2>

    <div class="card">
      <div class="card-body">
        <table class="table">
          <thead>
            <tr>
              <th>订单号</th>
              <th>商品名称</th>
              <th>金额</th>
              <th>状态</th>
              <th>时间</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="order in orders" :key="order.trade_no">
              <td class="text-xs">{{ order.trade_no }}</td>
              <td>{{ order.name }}</td>
              <td class="text-primary-600">¥{{ order.money }}</td>
              <td>
                <span :class="['badge', order.status === 1 ? 'badge-success' : 'badge-warning']">
                  {{ order.status === 1 ? '已支付' : '待支付' }}
                </span>
              </td>
              <td>{{ dayjs(order.addtime).format('YYYY-MM-DD HH:mm') }}</td>
            </tr>
            <tr v-if="orders.length === 0">
              <td colspan="5" class="text-center text-gray-500 py-8">暂无订单</td>
            </tr>
          </tbody>
        </table>

        <div class="pagination text-sm">
          <button class="pagination-item" :disabled="page === 1" @click="page--; fetchOrders()">上一页</button>
          <span class="px-4">第 {{ page }} 页</span>
          <button class="pagination-item" @click="page++; fetchOrders()">下一页</button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { getUserOrders } from '@/api/user'
import dayjs from 'dayjs'

const orders = ref<any[]>([])
const loading = ref(false)
const page = ref(1)
const total = ref(0)

async function fetchOrders() {
  loading.value = true
  try {
    const res = await getUserOrders({ page: page.value, limit: 20 })
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

onMounted(() => {
  fetchOrders()
})
</script>
