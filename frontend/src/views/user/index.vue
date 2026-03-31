<template>
  <div>
    <h2 class="text-2xl font-bold text-gray-800 mb-6">商户中心</h2>

    <div class="grid grid-cols-1 md:grid-cols-4 gap-6 mb-6">
      <div class="card p-6 flex items-start gap-4">
        <div class="w-12 h-12 bg-green-100 rounded-xl flex items-center justify-center flex-shrink-0">
          <TrendingUp class="w-6 h-6 text-green-600" />
        </div>
        <div>
          <div class="text-gray-500 text-sm mb-1">今日收入</div>
          <div class="text-2xl font-bold text-green-600">¥{{ stats.today_money.toFixed(2) }}</div>
        </div>
      </div>
      <div class="card p-6 flex items-start gap-4">
        <div class="w-12 h-12 bg-blue-100 rounded-xl flex items-center justify-center flex-shrink-0">
          <ShoppingBag class="w-6 h-6 text-blue-600" />
        </div>
        <div>
          <div class="text-gray-500 text-sm mb-1">今日订单</div>
          <div class="text-2xl font-bold text-gray-800">{{ stats.today_count }}</div>
        </div>
      </div>
      <div class="card p-6 flex items-start gap-4">
        <div class="w-12 h-12 bg-purple-100 rounded-xl flex items-center justify-center flex-shrink-0">
          <Wallet class="w-6 h-6 text-purple-600" />
        </div>
        <div>
          <div class="text-gray-500 text-sm mb-1">累计收入</div>
          <div class="text-2xl font-bold text-primary-600">¥{{ stats.total_money.toFixed(2) }}</div>
        </div>
      </div>
      <div class="card p-6 flex items-start gap-4">
        <div class="w-12 h-12 bg-orange-100 rounded-xl flex items-center justify-center flex-shrink-0">
          <CheckCircle class="w-6 h-6 text-orange-600" />
        </div>
        <div>
          <div class="text-gray-500 text-sm mb-1">已结算</div>
          <div class="text-2xl font-bold text-gray-600">¥{{ stats.settle_money.toFixed(2) }}</div>
        </div>
      </div>
    </div>

    <div class="card mb-6">
      <div class="card-header">
        <h3 class="font-medium text-gray-800">快捷操作</h3>
      </div>
      <div class="card-body">
        <div class="grid grid-cols-4 gap-4">
          <router-link to="/user/orders" class="p-4 bg-gray-50 rounded-xl hover:bg-gray-100 transition-colors group text-center">
            <div class="w-10 h-10 bg-blue-100 rounded-lg flex items-center justify-center mb-3 mx-auto group-hover:bg-blue-200 transition-colors">
              <FileText class="w-5 h-5 text-blue-600" />
            </div>
            <div class="font-medium text-gray-700">我的订单</div>
          </router-link>
          <router-link to="/user/settles" class="p-4 bg-gray-50 rounded-xl hover:bg-gray-100 transition-colors group text-center">
            <div class="w-10 h-10 bg-yellow-100 rounded-lg flex items-center justify-center mb-3 mx-auto group-hover:bg-yellow-200 transition-colors">
              <Wallet class="w-5 h-5 text-yellow-600" />
            </div>
            <div class="font-medium text-gray-700">结算记录</div>
          </router-link>
          <router-link to="/user/records" class="p-4 bg-gray-50 rounded-xl hover:bg-gray-100 transition-colors group text-center">
            <div class="w-10 h-10 bg-green-100 rounded-lg flex items-center justify-center mb-3 mx-auto group-hover:bg-green-200 transition-colors">
              <Receipt class="w-5 h-5 text-green-600" />
            </div>
            <div class="font-medium text-gray-700">资金记录</div>
          </router-link>
          <router-link to="/user/profile" class="p-4 bg-gray-50 rounded-xl hover:bg-gray-100 transition-colors group text-center">
            <div class="w-10 h-10 bg-purple-100 rounded-lg flex items-center justify-center mb-3 mx-auto group-hover:bg-purple-200 transition-colors">
              <User class="w-5 h-5 text-purple-600" />
            </div>
            <div class="font-medium text-gray-700">资料管理</div>
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
              <th>支付方式</th>
              <th>金额</th>
              <th>状态</th>
              <th>时间</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="order in recentOrders" :key="order.trade_no">
              <td class="text-xs">{{ order.trade_no }}</td>
              <td>{{ order.name }}</td>
              <td>
                <div class="flex items-center gap-1.5">
                  <SvgIcon :name="order.type === 1 ? 'alipay' : 'wechatpay'" :size="16" />
                  <span class="text-sm">{{ order.type === 1 ? '支付宝' : '微信' }}</span>
                </div>
              </td>
              <td class="text-primary-600 font-medium">¥{{ order.money }}</td>
              <td>
                <span :class="['badge', order.status === 1 ? 'badge-success' : 'badge-warning']">
                  {{ order.status === 1 ? '已支付' : '待支付' }}
                </span>
              </td>
              <td>{{ order.addtime }}</td>
            </tr>
            <tr v-if="recentOrders.length === 0">
              <td colspan="6" class="text-center text-gray-500 py-8">暂无订单</td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { TrendingUp, ShoppingBag, Wallet, CheckCircle, FileText, Receipt, User } from 'lucide-vue-next'
import SvgIcon from '@/components/svgicon.vue'

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
