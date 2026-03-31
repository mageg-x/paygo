<template>
  <div>
    <h2 class="text-2xl font-bold text-gray-800 mb-6">后台首页</h2>

    <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6 mb-8">
      <div class="card p-6 flex items-start gap-4">
        <div class="w-12 h-12 bg-blue-100 rounded-xl flex items-center justify-center flex-shrink-0">
          <ShoppingBag class="w-6 h-6 text-blue-600" />
        </div>
        <div>
          <div class="text-gray-500 text-sm mb-1">今日订单数</div>
          <div class="text-2xl font-bold text-gray-800">{{ stats.today_order_count }}</div>
        </div>
      </div>

      <div class="card p-6 flex items-start gap-4">
        <div class="w-12 h-12 bg-green-100 rounded-xl flex items-center justify-center flex-shrink-0">
          <TrendingUp class="w-6 h-6 text-green-600" />
        </div>
        <div>
          <div class="text-gray-500 text-sm mb-1">今日交易额</div>
          <div class="text-2xl font-bold text-green-600">¥{{ stats.today_order_money.toFixed(2) }}</div>
        </div>
      </div>

      <div class="card p-6 flex items-start gap-4">
        <div class="w-12 h-12 bg-purple-100 rounded-xl flex items-center justify-center flex-shrink-0">
          <Users class="w-6 h-6 text-purple-600" />
        </div>
        <div>
          <div class="text-gray-500 text-sm mb-1">商户总数</div>
          <div class="text-2xl font-bold text-gray-800">{{ stats.user_count }}</div>
        </div>
      </div>

      <div class="card p-6 flex items-start gap-4">
        <div class="w-12 h-12 bg-orange-100 rounded-xl flex items-center justify-center flex-shrink-0">
          <BarChart3 class="w-6 h-6 text-orange-600" />
        </div>
        <div>
          <div class="text-gray-500 text-sm mb-1">昨日交易额</div>
          <div class="text-2xl font-bold text-gray-600">¥{{ stats.yesterday_order_money.toFixed(2) }}</div>
        </div>
      </div>
    </div>

    <div class="card">
      <div class="card-header">
        <h3 class="text-lg font-medium text-gray-800">快捷操作</h3>
      </div>
      <div class="card-body">
        <div class="grid grid-cols-2 md:grid-cols-4 gap-4">
          <router-link to="/admin/users" class="p-4 bg-gray-50 rounded-xl hover:bg-gray-100 transition-colors group">
            <div class="w-10 h-10 bg-blue-100 rounded-lg flex items-center justify-center mb-3 group-hover:bg-blue-200 transition-colors">
              <Users class="w-5 h-5 text-blue-600" />
            </div>
            <div class="font-medium text-gray-700">商户管理</div>
          </router-link>

          <router-link to="/admin/orders" class="p-4 bg-gray-50 rounded-xl hover:bg-gray-100 transition-colors group">
            <div class="w-10 h-10 bg-green-100 rounded-lg flex items-center justify-center mb-3 group-hover:bg-green-200 transition-colors">
              <FileText class="w-5 h-5 text-green-600" />
            </div>
            <div class="font-medium text-gray-700">订单管理</div>
          </router-link>

          <router-link to="/admin/settles" class="p-4 bg-gray-50 rounded-xl hover:bg-gray-100 transition-colors group">
            <div class="w-10 h-10 bg-yellow-100 rounded-lg flex items-center justify-center mb-3 group-hover:bg-yellow-200 transition-colors">
              <Wallet class="w-5 h-5 text-yellow-600" />
            </div>
            <div class="font-medium text-gray-700">结算管理</div>
          </router-link>

          <router-link to="/admin/settings" class="p-4 bg-gray-50 rounded-xl hover:bg-gray-100 transition-colors group">
            <div class="w-10 h-10 bg-gray-200 rounded-lg flex items-center justify-center mb-3 group-hover:bg-gray-300 transition-colors">
              <Settings class="w-5 h-5 text-gray-600" />
            </div>
            <div class="font-medium text-gray-700">系统设置</div>
          </router-link>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { getAdminStats } from '@/api/admin'
import { ShoppingBag, TrendingUp, Users, BarChart3, FileText, Wallet, Settings } from 'lucide-vue-next'

const stats = ref({
  today_order_money: 0,
  today_order_count: 0,
  yesterday_order_money: 0,
  yesterday_order_count: 0,
  user_count: 0
})

const loading = ref(true)

onMounted(async () => {
  try {
    const res = await getAdminStats()
    if (res.code === 0) {
      stats.value = res.data
    }
  } catch (error) {
    console.error('获取统计失败:', error)
  } finally {
    loading.value = false
  }
})
</script>
