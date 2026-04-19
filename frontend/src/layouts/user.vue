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
          <h1 class="text-base md:text-xl font-bold text-gray-800 no-wrap">商户后台</h1>
        </div>
        <div class="flex items-center gap-2 md:gap-4">
          <span class="text-gray-600 hidden sm:flex items-center gap-1.5 no-wrap">
            <User class="w-4 h-4" />
            UID: {{ appStore.userInfo?.uid || '-' }}
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
            <span class="font-semibold text-slate-800 no-wrap">商户后台</span>
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
import { computed, onMounted, ref } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { userLogout, getUserInfo } from '@/api/user'
import { useAppStore } from '@/stores/app'
import { Home, FileText, Wallet, Receipt, User, LogOut, Gift, QrCode, CreditCard, HelpCircle, CreditCardIcon, ArrowLeftRight, Menu } from 'lucide-vue-next'

const router = useRouter()
const route = useRoute()
const appStore = useAppStore()
const drawerOpen = ref(false)

const menuSections = [
  {
    title: '概览',
    items: [{ path: '/user/index', name: '商户中心', icon: Home }]
  },
  {
    title: '收款与交易',
    items: [
      { path: '/user/orders', name: '订单管理', icon: FileText },
      { path: '/user/qrcode', name: '收款二维码', icon: QrCode },
      { path: '/user/settles', name: '结算管理', icon: Wallet },
      { path: '/user/records', name: '资金记录', icon: Receipt }
    ]
  },
  {
    title: '开发接入',
    items: [
      { path: '/user/paytest', name: 'API调试', icon: CreditCard },
      { path: '/user/profile', name: '资料管理', icon: User }
    ]
  },
  {
    title: '增值服务',
    items: [
      { path: '/user/recharge', name: '余额充值', icon: CreditCard },
      { path: '/user/invite', name: '邀请推广', icon: Gift },
      { path: '/user/groupbuy', name: '会员购买', icon: CreditCardIcon },
      { path: '/user/transfer-add', name: '转让用户组', icon: ArrowLeftRight },
      { path: '/user/help', name: '帮助中心', icon: HelpCircle }
    ]
  }
]

const activeMenu = computed(() => route.path)

// 页面加载时检查用户信息
onMounted(async () => {
  // 如果有 token 但没有 userInfo，获取用户信息
  if (appStore.userToken && !appStore.userInfo) {
    try {
      const res = await getUserInfo()
      if (res.code === 0 && res.data) {
        const u = res.data
        appStore.userLogin(appStore.userToken, {
          uid: u.uid,
          username: u.username || '',
          email: u.email || '',
          phone: u.phone || '',
          money: u.money || 0,
          status: u.status || 1
        }, sessionStorage.getItem('user_csrf_token') || '')
      }
    } catch (error) {
      console.error('获取用户信息失败:', error)
    }
  }
})

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
