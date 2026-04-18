<template>
  <div class="space-y-5">
    <div class="page-head">
      <div>
        <h2 class="page-title no-wrap">后台首页</h2>
        <p class="page-subtitle">平台运营总览与关键入口</p>
      </div>
    </div>

    <div class="card overflow-hidden">
      <div class="relative px-4 py-4 md:px-6 md:py-6">
        <div class="absolute inset-0 bg-gradient-to-r from-sky-500/15 via-blue-500/10 to-teal-400/15 pointer-events-none"></div>
        <div class="relative grid grid-cols-2 gap-2 md:gap-4">
          <div class="rounded-xl md:rounded-2xl border border-slate-200/70  bg-slate-50 px-3 py-2.5 md:px-6 md:py-5 shadow-sm min-h-[78px] md:min-h-[116px] flex flex-col justify-between">
            <div class="text-xs md:text-sm text-slate-500 no-wrap">昨日订单</div>
            <div class="text-lg md:text-3xl font-semibold text-slate-800 leading-none">{{ stats.yesterday_order_count }}</div>
          </div>
          <div class="rounded-xl md:rounded-2xl border border-slate-200/70 bg-slate-50 px-3 py-2.5 md:px-6 md:py-5 shadow-sm min-h-[78px] md:min-h-[116px] flex flex-col justify-between">
            <div class="text-xs md:text-sm text-slate-500 no-wrap">昨日交易额</div>
            <div class="text-lg md:text-3xl font-semibold text-amber-600 leading-none">¥{{ stats.yesterday_order_money.toFixed(2) }}</div>
          </div>
        </div>
      </div>
    </div>

    <div class="grid grid-cols-1 sm:grid-cols-2 xl:grid-cols-4 gap-3 md:gap-6 mb-6 md:mb-8">
      <div class="card p-4 md:p-6 flex items-start gap-3 md:gap-4">
        <div
          class="w-12 h-12 bg-gradient-to-br from-blue-400 to-blue-500 rounded-xl flex items-center justify-center flex-shrink-0 shadow-md shadow-blue-200">
          <ShoppingBag class="w-6 h-6 text-white" />
        </div>
        <div>
          <div class="text-gray-500 text-sm mb-1 no-wrap">今日订单数</div>
          <div class="text-2xl font-bold text-blue-600">{{ stats.today_order_count }}</div>
        </div>
      </div>

      <div class="card p-4 md:p-6 flex items-start gap-3 md:gap-4">
        <div
          class="w-12 h-12 bg-gradient-to-br from-emerald-400 to-emerald-500 rounded-xl flex items-center justify-center flex-shrink-0 shadow-md shadow-emerald-200">
          <TrendingUp class="w-6 h-6 text-white" />
        </div>
        <div>
          <div class="text-gray-500 text-sm mb-1 no-wrap">今日交易额</div>
          <div class="text-2xl font-bold text-emerald-600">¥{{ stats.today_order_money.toFixed(2) }}</div>
        </div>
      </div>

      <div class="card p-4 md:p-6 flex items-start gap-3 md:gap-4">
        <div
          class="w-12 h-12 bg-gradient-to-br from-cyan-400 to-cyan-500 rounded-xl flex items-center justify-center flex-shrink-0 shadow-md shadow-cyan-200">
          <Users class="w-6 h-6 text-white" />
        </div>
        <div>
          <div class="text-gray-500 text-sm mb-1 no-wrap">商户总数</div>
          <div class="text-2xl font-bold text-cyan-600">{{ stats.user_count }}</div>
        </div>
      </div>

      <div class="card p-4 md:p-6 flex items-start gap-3 md:gap-4">
        <div
          class="w-12 h-12 bg-gradient-to-br from-amber-400 to-amber-500 rounded-xl flex items-center justify-center flex-shrink-0 shadow-md shadow-amber-200">
          <BarChart3 class="w-6 h-6 text-white" />
        </div>
        <div>
          <div class="text-gray-500 text-sm mb-1 no-wrap">昨日交易额</div>
          <div class="text-2xl font-bold text-amber-600">¥{{ stats.yesterday_order_money.toFixed(2) }}</div>
        </div>
      </div>
    </div>

    <div class="card">
      <div class="card-header">
        <h3 class="text-lg font-medium text-gray-800 no-wrap">快捷操作</h3>
      </div>
      <div class="card-body">
        <div class="grid grid-cols-2 md:grid-cols-4 gap-3 md:gap-4">
          <router-link to="/admin/users"
            class="p-4 bg-gradient-to-br from-blue-50 to-blue-100 rounded-xl hover:from-blue-100 hover:to-blue-150 transition-all group border border-blue-200/50">
            <div
              class="w-10 h-10 bg-gradient-to-br from-blue-400 to-blue-500 rounded-lg flex items-center justify-center mb-3 group-hover:scale-110 transition-transform shadow-sm shadow-blue-200">
              <Users class="w-5 h-5 text-white" />
            </div>
            <div class="font-medium text-blue-700 no-wrap">商户管理</div>
          </router-link>

          <router-link to="/admin/orders"
            class="p-4 bg-gradient-to-br from-emerald-50 to-emerald-100 rounded-xl hover:from-emerald-100 hover:to-emerald-150 transition-all group border border-emerald-200/50">
            <div
              class="w-10 h-10 bg-gradient-to-br from-emerald-400 to-emerald-500 rounded-lg flex items-center justify-center mb-3 group-hover:scale-110 transition-transform shadow-sm shadow-emerald-200">
              <FileText class="w-5 h-5 text-white" />
            </div>
            <div class="font-medium text-emerald-700 no-wrap">订单管理</div>
          </router-link>

          <router-link to="/admin/settles"
            class="p-4 bg-gradient-to-br from-amber-50 to-amber-100 rounded-xl hover:from-amber-100 hover:to-amber-150 transition-all group border border-amber-200/50">
            <div
              class="w-10 h-10 bg-gradient-to-br from-amber-400 to-amber-500 rounded-lg flex items-center justify-center mb-3 group-hover:scale-110 transition-transform shadow-sm shadow-amber-200">
              <Wallet class="w-5 h-5 text-white" />
            </div>
            <div class="font-medium text-amber-700 no-wrap">结算管理</div>
          </router-link>

          <router-link to="/admin/settings"
            class="p-4 bg-gradient-to-br from-slate-50 to-slate-100 rounded-xl hover:from-slate-100 hover:to-slate-150 transition-all group border border-slate-200/50">
            <div
              class="w-10 h-10 bg-gradient-to-br from-slate-400 to-slate-500 rounded-lg flex items-center justify-center mb-3 group-hover:scale-110 transition-transform shadow-sm shadow-slate-200">
              <Settings class="w-5 h-5 text-white" />
            </div>
            <div class="font-medium text-slate-700 no-wrap">系统设置</div>
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
