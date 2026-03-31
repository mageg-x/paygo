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
              <th>支付方式</th>
              <th>金额</th>
              <th>状态</th>
              <th>时间</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="order in orders" :key="order.trade_no">
              <td class="text-xs">{{ order.trade_no }}</td>
              <td>{{ order.name }}</td>
              <td>
                <div class="flex items-center gap-1.5">
                  <SvgIcon :name="order.type === 1 ? 'alipay' : 'wechatpay'" :size="18" />
                  <span class="text-sm font-medium" :class="order.type === 1 ? 'text-blue-600' : 'text-green-600'">{{ order.type === 1 ? '支付宝' : '微信' }}</span>
                </div>
              </td>
              <td class="font-semibold text-emerald-600">¥{{ order.money }}</td>
              <td>
                <span v-if="order.status === 1" class="inline-flex items-center px-2.5 py-1 rounded-full text-xs font-medium bg-emerald-100 text-emerald-700 ring-1 ring-emerald-200">
                  <span class="w-1.5 h-1.5 rounded-full bg-emerald-500 mr-1.5"></span>已支付
                </span>
                <span v-else class="inline-flex items-center px-2.5 py-1 rounded-full text-xs font-medium bg-amber-100 text-amber-700 ring-1 ring-amber-200">
                  <span class="w-1.5 h-1.5 rounded-full bg-amber-500 mr-1.5"></span>待支付
                </span>
              </td>
              <td>{{ dayjs(order.addtime).format('YYYY-MM-DD HH:mm') }}</td>
            </tr>
            <tr v-if="orders.length === 0">
              <td colspan="6" class="text-center text-gray-500 py-8">暂无订单</td>
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
import SvgIcon from '@/components/svgicon.vue'

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
