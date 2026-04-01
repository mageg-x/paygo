<template>
  <div class="h-screen flex flex-col bg-gray-100">
    <header class="bg-white shadow-sm border-b border-gray-200 flex-shrink-0">
      <div class="flex items-center justify-between px-6 py-3">
        <div class="flex items-center gap-3">
          <img src="@/assets/paygo.png" alt="Logo" class="w-8 h-8" />
          <h1 class="text-xl font-bold text-gray-800">商户后台</h1>
        </div>
        <div class="flex items-center gap-4">
          <span class="text-gray-600 flex items-center gap-1.5">
            <User class="w-4 h-4" />
            UID: {{ appStore.userInfo?.uid || '-' }}
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
      <aside class="w-48 bg-white border-r border-gray-200 overflow-y-auto">
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
import { computed, onMounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { userLogout, getUserInfo } from '@/api/user'
import { useAppStore } from '@/stores/app'
import { Home, FileText, Wallet, Receipt, User, LogOut, Gift, QrCode, CreditCard, HelpCircle, CreditCardIcon, Link, ArrowLeftRight } from 'lucide-vue-next'

const router = useRouter()
const route = useRoute()
const appStore = useAppStore()

const menus = [
  { path: '/user/index', name: '商户中心', icon: Home },
  { path: '/user/orders', name: '订单管理', icon: FileText },
  { path: '/user/settles', name: '结算管理', icon: Wallet },
  { path: '/user/records', name: '资金记录', icon: Receipt },
  { path: '/user/qrcode', name: '收款二维码', icon: QrCode },
  { path: '/user/recharge', name: '余额充值', icon: CreditCard },
  { path: '/user/invite', name: '邀请推广', icon: Gift },
  { path: '/user/groupbuy', name: '会员购买', icon: CreditCardIcon },
  { path: '/user/openid', name: '账号绑定', icon: Link },
  { path: '/user/transfer-add', name: '转让用户组', icon: ArrowLeftRight },
  { path: '/user/profile', name: '资料管理', icon: User },
  { path: '/user/help', name: '帮助中心', icon: HelpCircle }
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
        })
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
