<template>
  <div class="h-screen flex flex-col bg-gray-100">
    <header class="bg-white shadow-sm border-b border-gray-200 flex-shrink-0">
      <div class="flex items-center justify-between px-6 py-3">
        <div class="flex items-center gap-3">
          <img src="@/assets/paygo.png" alt="Logo" class="w-8 h-8" />
          <h1 class="text-xl font-bold text-gray-800">支付系统 - 管理后台</h1>
        </div>
        <div class="flex items-center gap-4">
          <span class="text-gray-600 flex items-center gap-1.5">
            <User class="w-4 h-4" />
            {{ appStore.adminUser || 'admin' }}
          </span>
          <button @click="handleLogout"
            class="text-gray-500 hover:text-red-600 flex items-center gap-1 transition-colors">
            <LogOut class="w-4 h-4" />
            退出
          </button>
        </div>
      </div>
    </header>

    <div class="flex flex-1 overflow-hidden">
      <aside class="w-48 flex-shrink-0 bg-white border-r border-gray-200 overflow-y-auto">
        <nav class="p-4 space-y-1">
          <router-link v-for="menu in menus" :key="menu.path" :to="menu.path" :class="[
            'flex items-center gap-1 px-4 py-2.5 rounded-lg text-sm font-medium transition-colors',
            activeMenu === menu.path
              ? 'bg-primary-50 text-primary-700'
              : 'text-gray-600 hover:bg-gray-200 hover:text-gray-900'
          ]">
            <component :is="menu.icon" class="w-4 h-4 mr-2 flex-shrink-0" />
            <span>{{ menu.name }}</span>
          </router-link>
        </nav>
      </aside>

      <main class="flex-1 p-6 overflow-y-auto">
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
import { Home, Users, FileText, Wallet, ArrowLeftRight, Strikethrough, Puzzle, Settings, User, LogOut, Ticket, Shield, Ban, Globe, Volume2, FileSearch, Key, Timer, MessageCircle, Trash2, Download, CreditCard, RefreshCw, Coins, QrCode } from 'lucide-vue-next'

const route = useRoute()
const router = useRouter()
const appStore = useAppStore()

const menus = [
  { path: '/admin/index', name: '首页', icon: Home },
  { path: '/admin/users', name: '商户管理', icon: Users },
  { path: '/admin/groups', name: '用户组', icon: Users },
  { path: '/admin/orders', name: '订单管理', icon: FileText },
  { path: '/admin/settles', name: '结算管理', icon: Wallet },
  { path: '/admin/transfers', name: '转账管理', icon: ArrowLeftRight },
  { path: '/admin/channels', name: '通道管理', icon: Strikethrough },
  { path: '/admin/paytest', name: '通道测试', icon: QrCode },
  { path: '/admin/invitecodes', name: '邀请码', icon: Ticket },
  { path: '/admin/plugins', name: '插件管理', icon: Puzzle },
  { path: '/admin/risk', name: '风控管理', icon: Shield },
  { path: '/admin/blacklist', name: '黑名单', icon: Ban },
  { path: '/admin/domains', name: '域名授权', icon: Globe },
  { path: '/admin/announces', name: '公告管理', icon: Volume2 },
  { path: '/admin/logs', name: '操作日志', icon: FileSearch },
  { path: '/admin/sso', name: 'SSO登录', icon: Key },
  { path: '/admin/crons', name: '计划任务', icon: Timer },
  { path: '/admin/export', name: '数据导出', icon: Download },
  { path: '/admin/clean', name: '数据清理', icon: Trash2 },
  { path: '/admin/wxkf', name: '微信客服', icon: MessageCircle },
  { path: '/admin/paytype', name: '支付类型', icon: CreditCard },
  { path: '/admin/payroll', name: '轮询配置', icon: RefreshCw },
  { path: '/admin/profit', name: '分账管理', icon: Coins },
  { path: '/admin/settings', name: '系统设置', icon: Settings }
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
