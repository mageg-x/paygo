<template>
  <div>
    <h2 class="text-2xl font-bold text-gray-800 mb-6">后台首页</h2>

    <!-- 统计卡片 -->
    <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6 mb-8">
      <div class="card p-6">
        <div class="text-gray-500 text-sm mb-2">今日订单数</div>
        <div class="text-3xl font-bold text-gray-800">{{ stats.today_order_count }}</div>
      </div>

      <div class="card p-6">
        <div class="text-gray-500 text-sm mb-2">今日交易额</div>
        <div class="text-3xl font-bold text-primary-600">¥{{ stats.today_order_money.toFixed(2) }}</div>
      </div>

      <div class="card p-6">
        <div class="text-gray-500 text-sm mb-2">商户总数</div>
        <div class="text-3xl font-bold text-gray-800">{{ stats.user_count }}</div>
      </div>

      <div class="card p-6">
        <div class="text-gray-500 text-sm mb-2">昨日交易额</div>
        <div class="text-3xl font-bold text-gray-600">¥{{ stats.yesterday_order_money.toFixed(2) }}</div>
      </div>
    </div>

    <!-- 快捷操作 -->
    <div class="card">
      <div class="card-header">
        <h3 class="text-lg font-medium text-gray-800">快捷操作</h3>
      </div>
      <div class="card-body">
        <div class="grid grid-cols-2 md:grid-cols-4 gap-4">
          <router-link to="/admin/users" class="p-4 bg-gray-50 rounded-lg hover:bg-gray-100 transition-colors">
            <div class="text-primary-600 mb-2">
              <svg class="w-8 h-8" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                  d="M12 4.354a4 4 0 110 5.292M15 21H3v-1a6 6 0 0112 0v1zm0 0h6v-1a6 6 0 00-9-5.197m13.5-9a2.5 2.5 0 11-5 0 2.5 2.5 0 015 0z" />
              </svg>
            </div>
            <div class="font-medium text-gray-700">商户管理</div>
          </router-link>

          <router-link to="/admin/orders" class="p-4 bg-gray-50 rounded-lg hover:bg-gray-100 transition-colors">
            <div class="text-success mb-2">
              <svg class="w-8 h-8" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                  d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2" />
              </svg>
            </div>
            <div class="font-medium text-gray-700">订单管理</div>
          </router-link>

          <router-link to="/admin/settles" class="p-4 bg-gray-50 rounded-lg hover:bg-gray-100 transition-colors">
            <div class="text-warning mb-2">
              <svg class="w-8 h-8" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                  d="M17 9V7a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2m2 4h10a2 2 0 002-2v-6a2 2 0 00-2-2H9a2 2 0 00-2 2v6a2 2 0 002 2zm7-5a2 2 0 11-4 0 2 2 0 014 0z" />
              </svg>
            </div>
            <div class="font-medium text-gray-700">结算管理</div>
          </router-link>

          <router-link to="/admin/settings" class="p-4 bg-gray-50 rounded-lg hover:bg-gray-100 transition-colors">
            <div class="text-gray-600 mb-2">
              <svg class="w-8 h-8" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                  d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z" />
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                  d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
              </svg>
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
