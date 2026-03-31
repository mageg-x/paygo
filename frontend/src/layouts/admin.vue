<template>
  <div class="min-h-screen bg-gray-100">
    <!-- 顶部导航 -->
    <header class="bg-white shadow-sm border-b border-gray-200">
      <div class="flex items-center justify-between px-6 py-3">
        <h1 class="text-xl font-bold text-gray-800">支付系统 - 管理后台</h1>
        <div class="flex items-center gap-4">
          <span class="text-gray-600">{{ appStore.adminUser || 'admin' }}</span>
          <button @click="handleLogout" class="text-gray-500 hover:text-gray-700">退出</button>
        </div>
      </div>
    </header>

    <div class="flex">
      <!-- 侧边栏 -->
      <aside class="w-56 flex-shrink-0 bg-white border-r border-gray-200 min-h-screen">
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
      <main class="flex-1 p-6 overflow-x-auto">
        <router-view />
      </main>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { adminLogout } from '@/api/admin'
import { useAppStore } from '@/stores/app'

const route = useRoute()
const router = useRouter()
const appStore = useAppStore()

const menus = [
  { path: '/admin/index', name: '首页', icon: 'home' },
  { path: '/admin/users', name: '商户管理', icon: 'users' },
  { path: '/admin/orders', name: '订单管理', icon: 'list' },
  { path: '/admin/settles', name: '结算管理', icon: 'wallet' },
  { path: '/admin/transfers', name: '转账管理', icon: 'exchange' },
  { path: '/admin/channels', name: '通道管理', icon: 'link' },
  { path: '/admin/plugins', name: '插件管理', icon: 'puzzle' },
  { path: '/admin/settings', name: '系统设置', icon: 'cog' }
]

const activeMenu = computed(() => route.path)

async function handleLogout() {
  try {
    await adminLogout()
    appStore.adminLogout()
    router.push('/admin/login')
  } catch (error) {
    console.error('logout failed:', error)
    appStore.adminLogout()
    router.push('/admin/login')
  }
}
</script>
