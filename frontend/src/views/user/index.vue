<template>
  <div>
    <h2 class="text-2xl font-bold text-gray-800 mb-6">商户中心</h2>

    <div class="grid grid-cols-1 md:grid-cols-4 gap-6 mb-6">
      <div class="card p-6 flex items-start gap-4 border-l-4 border-l-emerald-400">
        <div class="w-12 h-12 bg-gradient-to-br from-emerald-400 to-emerald-500 rounded-xl flex items-center justify-center flex-shrink-0 shadow-md shadow-emerald-200">
          <TrendingUp class="w-6 h-6 text-white" />
        </div>
        <div>
          <div class="text-gray-500 text-sm mb-1">今日收入</div>
          <div class="text-2xl font-bold text-emerald-600">¥{{ stats.today_money.toFixed(2) }}</div>
        </div>
      </div>
      <div class="card p-6 flex items-start gap-4 border-l-4 border-l-blue-400">
        <div class="w-12 h-12 bg-gradient-to-br from-blue-400 to-blue-500 rounded-xl flex items-center justify-center flex-shrink-0 shadow-md shadow-blue-200">
          <ShoppingBag class="w-6 h-6 text-white" />
        </div>
        <div>
          <div class="text-gray-500 text-sm mb-1">今日订单</div>
          <div class="text-2xl font-bold text-blue-600">{{ stats.today_count }}</div>
        </div>
      </div>
      <div class="card p-6 flex items-start gap-4 border-l-4 border-l-violet-400">
        <div class="w-12 h-12 bg-gradient-to-br from-violet-400 to-violet-500 rounded-xl flex items-center justify-center flex-shrink-0 shadow-md shadow-violet-200">
          <Wallet class="w-6 h-6 text-white" />
        </div>
        <div>
          <div class="text-gray-500 text-sm mb-1">累计收入</div>
          <div class="text-2xl font-bold text-violet-600">¥{{ stats.total_money.toFixed(2) }}</div>
        </div>
      </div>
      <div class="card p-6 flex items-start gap-4 border-l-4 border-l-amber-400">
        <div class="w-12 h-12 bg-gradient-to-br from-amber-400 to-amber-500 rounded-xl flex items-center justify-center flex-shrink-0 shadow-md shadow-amber-200">
          <CheckCircle class="w-6 h-6 text-white" />
        </div>
        <div>
          <div class="text-gray-500 text-sm mb-1">已结算</div>
          <div class="text-2xl font-bold text-amber-600">¥{{ stats.settle_money.toFixed(2) }}</div>
        </div>
      </div>
    </div>

    <div class="card mb-6">
      <div class="card-header">
        <h3 class="font-medium text-gray-800">快捷操作</h3>
      </div>
      <div class="card-body">
        <div class="grid grid-cols-4 gap-4">
          <router-link to="/user/orders" class="p-4 bg-gradient-to-br from-blue-50 to-blue-100 rounded-xl hover:from-blue-100 hover:to-blue-150 transition-all group text-center border border-blue-200/50">
            <div class="w-10 h-10 bg-gradient-to-br from-blue-400 to-blue-500 rounded-lg flex items-center justify-center mb-3 mx-auto group-hover:scale-110 transition-transform shadow-sm shadow-blue-200">
              <FileText class="w-5 h-5 text-white" />
            </div>
            <div class="font-medium text-blue-700">我的订单</div>
          </router-link>
          <router-link to="/user/settles" class="p-4 bg-gradient-to-br from-amber-50 to-amber-100 rounded-xl hover:from-amber-100 hover:to-amber-150 transition-all group text-center border border-amber-200/50">
            <div class="w-10 h-10 bg-gradient-to-br from-amber-400 to-amber-500 rounded-lg flex items-center justify-center mb-3 mx-auto group-hover:scale-110 transition-transform shadow-sm shadow-amber-200">
              <Wallet class="w-5 h-5 text-white" />
            </div>
            <div class="font-medium text-amber-700">结算记录</div>
          </router-link>
          <router-link to="/user/records" class="p-4 bg-gradient-to-br from-emerald-50 to-emerald-100 rounded-xl hover:from-emerald-100 hover:to-emerald-150 transition-all group text-center border border-emerald-200/50">
            <div class="w-10 h-10 bg-gradient-to-br from-emerald-400 to-emerald-500 rounded-lg flex items-center justify-center mb-3 mx-auto group-hover:scale-110 transition-transform shadow-sm shadow-emerald-200">
              <Receipt class="w-5 h-5 text-white" />
            </div>
            <div class="font-medium text-emerald-700">资金记录</div>
          </router-link>
          <router-link to="/user/profile" class="p-4 bg-gradient-to-br from-violet-50 to-violet-100 rounded-xl hover:from-violet-100 hover:to-violet-150 transition-all group text-center border border-violet-200/50">
            <div class="w-10 h-10 bg-gradient-to-br from-violet-400 to-violet-500 rounded-lg flex items-center justify-center mb-3 mx-auto group-hover:scale-110 transition-transform shadow-sm shadow-violet-200">
              <User class="w-5 h-5 text-white" />
            </div>
            <div class="font-medium text-violet-700">资料管理</div>
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
