<template>
  <div>
    <h2 class="text-2xl font-bold text-gray-800 mb-6">商户中心</h2>

    <div class="grid grid-cols-1 md:grid-cols-4 gap-6 mb-6">
      <div class="card p-6">
        <div class="text-gray-500 text-sm mb-2">今日收入</div>
        <div class="text-2xl font-bold text-success">¥{{ stats.today_money.toFixed(2) }}</div>
      </div>
      <div class="card p-6">
        <div class="text-gray-500 text-sm mb-2">今日订单</div>
        <div class="text-2xl font-bold text-gray-800">{{ stats.today_count }}</div>
      </div>
      <div class="card p-6">
        <div class="text-gray-500 text-sm mb-2">累计收入</div>
        <div class="text-2xl font-bold text-primary-600">¥{{ stats.total_money.toFixed(2) }}</div>
      </div>
      <div class="card p-6">
        <div class="text-gray-500 text-sm mb-2">已结算</div>
        <div class="text-2xl font-bold text-gray-600">¥{{ stats.settle_money.toFixed(2) }}</div>
      </div>
    </div>

    <div class="card mb-6">
      <div class="card-header">
        <h3 class="font-medium text-gray-800">快捷操作</h3>
      </div>
      <div class="card-body">
        <div class="grid grid-cols-4 gap-4">
          <router-link to="/user/orders" class="p-4 bg-gray-50 rounded-lg hover:bg-gray-100 text-center">
            <div class="text-gray-600">我的订单</div>
          </router-link>
          <router-link to="/user/settles" class="p-4 bg-gray-50 rounded-lg hover:bg-gray-100 text-center">
            <div class="text-gray-600">结算记录</div>
          </router-link>
          <router-link to="/user/records" class="p-4 bg-gray-50 rounded-lg hover:bg-gray-100 text-center">
            <div class="text-gray-600">资金记录</div>
          </router-link>
          <router-link to="/user/profile" class="p-4 bg-gray-50 rounded-lg hover:bg-gray-100 text-center">
            <div class="text-gray-600">资料管理</div>
          </router-link>
        </div>
      </div>
    </div>

    <div class="card">
      <div class="card-header">
        <h3 class="font-medium text-gray-800">最新订单</h3>
      </div>
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
            <tr v-for="order in recentOrders" :key="order.trade_no">
              <td>{{ order.trade_no }}</td>
              <td>{{ order.name }}</td>
              <td>¥{{ order.money }}</td>
              <td>
                <span :class="['badge', order.status === 1 ? 'badge-success' : 'badge-warning']">
                  {{ order.status === 1 ? '已支付' : '待支付' }}
                </span>
              </td>
              <td>{{ order.addtime }}</td>
            </tr>
            <tr v-if="recentOrders.length === 0">
              <td colspan="5" class="text-center text-gray-500 py-8">暂无订单</td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'

const stats = ref({
  today_money: 0,
  today_count: 0,
  total_money: 0,
  total_count: 0,
  settle_money: 0
})

const recentOrders = ref<any[]>([])

onMounted(async () => {
})
</script>
