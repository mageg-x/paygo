<template>
  <div class="min-h-screen bg-gray-100">
    <!-- 顶部导航 -->
    <header class="bg-white shadow-sm border-b border-gray-200">
      <div class="flex items-center justify-between px-6 py-3">
        <h1 class="text-xl font-bold text-gray-800">商户后台</h1>
        <div class="flex items-center gap-4">
          <span class="text-gray-600">UID: {{ appStore.userInfo?.uid || '-' }}</span>
          <button @click="handleLogout" class="text-gray-500 hover:text-gray-700">退出</button>
        </div>
      </div>
    </header>

    <div class="flex">
      <!-- 侧边栏 -->
      <aside class="w-56 bg-white border-r border-gray-200 min-h-screen">
        <nav class="p-4 space-y-1">
          <router-link
            v-for="menu in menus"
            :key="menu.path"
            :to="menu.path"
            :class="[
              'flex items-center px-4 py-2.5 rounded-lg text-sm font-medium transition-colors',
              activeMenu === menu.path
                ? 'bg-primary-50 text-primary-700'
                : 'text-gray-600 hover:bg-gray-50 hover:text-gray-900'
            ]"
          >
            <span class="mr-3">{{ menu.name }}</span>
          </router-link>
        </nav>
      </aside>

      <!-- 主内容 -->
      <main class="flex-1 p-6">
        <router-view />
      </main>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { userLogout } from '@/api/user'
import { useAppStore } from '@/stores/app'

const router = useRouter()
const route = useRoute()
const appStore = useAppStore()

const menus = [
  { path: '/user/index', name: '商户中心', icon: 'home' },
  { path: '/user/orders', name: '订单管理', icon: 'list' },
  { path: '/user/settles', name: '结算管理', icon: 'wallet' },
  { path: '/user/records', name: '资金记录', icon: 'credit-card' },
  { path: '/user/profile', name: '资料管理', icon: 'user' }
]

const activeMenu = computed(() => route.path)

async function handleLogout() {
  try {
    await userLogout()
    appStore.userLogout()
    router.push('/user/login')
  } catch (error) {
    console.error('logout failed:', error)
    appStore.userLogout()
    router.push('/user/login')
  }
}
</script>
