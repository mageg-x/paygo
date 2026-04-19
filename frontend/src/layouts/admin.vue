<template>
  <div class="h-[100dvh] flex flex-col bg-transparent overflow-hidden">
    <header class="shell-header flex-shrink-0 sticky top-0 z-40">
      <div class="flex items-center justify-between px-3 md:px-6 py-3">
        <div class="flex items-center gap-2 md:gap-3">
          <button
            class="md:hidden w-9 h-9 rounded-xl border border-slate-200/80 bg-white/80 text-slate-600 flex items-center justify-center"
            @click="drawerOpen = true"
          >
            <Menu class="w-5 h-5" />
          </button>
          <img src="@/assets/gopay.png" alt="Logo" class="w-7 h-7 md:w-8 md:h-8" />
          <h1 class="text-base md:text-xl font-bold text-gray-800 no-wrap">管理后台</h1>
        </div>
        <div class="flex items-center gap-2 md:gap-4">
          <span class="text-gray-600 hidden sm:flex items-center gap-1.5 no-wrap">
            <User class="w-4 h-4" />
            {{ appStore.adminUser || 'admin' }}
          </span>
          <button @click="handleLogout" class="btn btn-outline !px-3 !py-1.5 !min-h-0 text-sm">
            <LogOut class="w-4 h-4" />
            退出
          </button>
        </div>
      </div>
    </header>

    <div class="flex flex-1 min-h-0 overflow-hidden">
      <aside class="shell-sidebar hidden md:block w-56 flex-shrink-0 min-h-0 overflow-y-auto">
        <nav class="p-4 space-y-4">
          <section v-for="section in menuSections" :key="section.title" class="space-y-1">
            <h3 class="menu-section-title px-2 pb-1 text-[11px] font-semibold text-gray-400">{{ section.title }}</h3>
            <router-link v-for="menu in section.items" :key="menu.path" :to="menu.path" :class="[
              'menu-link',
              activeMenu === menu.path
                ? 'menu-link-active'
                : ''
            ]">
              <component :is="menu.icon" class="w-4 h-4 mr-2 flex-shrink-0" />
              <span>{{ menu.name }}</span>
            </router-link>
          </section>
        </nav>
      </aside>

      <main class="shell-main flex-1 min-h-0 mobile-content overflow-y-auto">
        <router-view />
      </main>
    </div>

    <div v-if="drawerOpen" class="fixed inset-0 z-50 md:hidden">
      <div class="absolute inset-0 bg-black/40" @click="drawerOpen = false"></div>
      <aside class="shell-sidebar absolute left-0 top-0 h-full w-[86%] max-w-[320px] shadow-xl overflow-y-auto">
        <div class="flex items-center justify-between px-4 py-3 border-b border-slate-100">
          <div class="flex items-center gap-2">
            <img src="@/assets/gopay.png" alt="Logo" class="w-7 h-7" />
            <span class="font-semibold text-slate-800 no-wrap">管理后台</span>
          </div>
          <button class="w-8 h-8 rounded-xl border border-slate-200/80 bg-white/80 text-slate-600" @click="drawerOpen = false">✕</button>
        </div>
        <nav class="p-4 space-y-4">
          <section v-for="section in menuSections" :key="'mobile-' + section.title" class="space-y-1">
            <h3 class="menu-section-title px-2 pb-1 text-[11px] font-semibold text-gray-400">{{ section.title }}</h3>
            <router-link
              v-for="menu in section.items"
              :key="'mobile-' + menu.path"
              :to="menu.path"
              :class="[
                'menu-link',
                activeMenu === menu.path
                  ? 'menu-link-active'
                  : ''
              ]"
              @click="drawerOpen = false"
            >
              <component :is="menu.icon" class="w-4 h-4 mr-2 flex-shrink-0" />
              <span class="no-wrap">{{ menu.name }}</span>
            </router-link>
          </section>
        </nav>
      </aside>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { adminLogout } from '@/api/admin'
import { useAppStore } from '@/stores/app'
import { Home, Users, FileText, Wallet, ArrowLeftRight, Strikethrough, Puzzle, Settings, User, LogOut, Ticket, Shield, Ban, Globe, Volume2, FileSearch, Key, Timer, MessageCircle, Trash2, Download, CreditCard, RefreshCw, Coins, QrCode, Menu } from 'lucide-vue-next'

const route = useRoute()
const router = useRouter()
const appStore = useAppStore()
const drawerOpen = ref(false)

const menuSections = [
  {
    title: '控制台',
    items: [{ path: '/admin/index', name: '首页', icon: Home }]
  },
  {
    title: '商户与账户',
    items: [
      { path: '/admin/users', name: '商户管理', icon: Users },
      { path: '/admin/groups', name: '用户组管理', icon: Users },
      { path: '/admin/invitecodes', name: '邀请码管理', icon: Ticket },
      { path: '/admin/sso', name: '商户代登录', icon: Key }
    ]
  },
  {
    title: '交易与资金',
    items: [
      { path: '/admin/orders', name: '订单管理', icon: FileText },
      { path: '/admin/settles', name: '结算管理', icon: Wallet },
      { path: '/admin/transfers', name: '转账记录', icon: ArrowLeftRight },
      { path: '/admin/profit', name: '分账管理', icon: Coins }
    ]
  },
  {
    title: '渠道与支付',
    items: [
      { path: '/admin/plugins', name: '插件管理', icon: Puzzle },
      { path: '/admin/channels', name: '通道管理', icon: Strikethrough },
      { path: '/admin/paytype', name: '支付方式管理', icon: CreditCard },
      { path: '/admin/paytest', name: '通道测试', icon: QrCode },
      { path: '/admin/payroll', name: '支付轮询规则', icon: RefreshCw }
    ]
  },
  {
    title: '风控与安全',
    items: [
      { path: '/admin/risk', name: '风控规则', icon: Shield },
      { path: '/admin/blacklist', name: '黑名单', icon: Ban },
      { path: '/admin/domains', name: '域名授权', icon: Globe },
      { path: '/admin/logs', name: '操作日志', icon: FileSearch }
    ]
  },
  {
    title: '系统运维',
    items: [
      { path: '/admin/crons', name: '计划任务', icon: Timer },
      { path: '/admin/settings', name: '系统设置', icon: Settings },
      { path: '/admin/export', name: '数据导出', icon: Download },
      { path: '/admin/clean', name: '数据清理', icon: Trash2 },
      { path: '/admin/announces', name: '公告管理', icon: Volume2 },
      { path: '/admin/wxkf', name: '微信客服', icon: MessageCircle }
    ]
  }
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
