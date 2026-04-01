import { createRouter, createWebHistory } from 'vue-router'
import type { RouteRecordRaw } from 'vue-router'
import { useAppStore } from '@/stores/app'

const routes: RouteRecordRaw[] = [
  // 首页
  {
    path: '/',
    name: 'Home',
    component: () => import('@/views/home/index.vue')
  },
  // 收银台
  {
    path: '/cashier/:trade_no',
    name: 'Cashier',
    component: () => import('@/views/cashier/index.vue')
  },
  // 管理员登录
  {
    path: '/admin/login',
    name: 'AdminLogin',
    component: () => import('@/views/admin/login.vue')
  },
  // 管理后台
  {
    path: '/admin',
    component: () => import('@/layouts/admin.vue'),
    children: [
      {
        path: '',
        redirect: '/admin/index'
      },
      {
        path: 'index',
        name: 'AdminIndex',
        component: () => import('@/views/admin/index.vue')
      },
      {
        path: 'users',
        name: 'AdminUsers',
        component: () => import('@/views/admin/users.vue')
      },
      {
        path: 'orders',
        name: 'AdminOrders',
        component: () => import('@/views/admin/orders.vue')
      },
      {
        path: 'settles',
        name: 'AdminSettles',
        component: () => import('@/views/admin/settles.vue')
      },
      {
        path: 'transfers',
        name: 'AdminTransfers',
        component: () => import('@/views/admin/transfers.vue')
      },
      {
        path: 'channels',
        name: 'AdminChannels',
        component: () => import('@/views/admin/channels.vue')
      },
      {
        path: 'invitecodes',
        name: 'AdminInviteCodes',
        component: () => import('@/views/admin/invitecodes.vue')
      },
      {
        path: 'plugins',
        name: 'AdminPlugins',
        component: () => import('@/views/admin/plugins.vue')
      },
      {
        path: 'settings',
        name: 'AdminSettings',
        component: () => import('@/views/admin/settings.vue')
      }
    ]
  },
  // 商户登录
  {
    path: '/user/login',
    name: 'UserLogin',
    component: () => import('@/views/user/login.vue')
  },
  // 商户注册
  {
    path: '/user/register',
    name: 'UserRegister',
    component: () => import('@/views/user/register.vue')
  },
  // 商户后台
  {
    path: '/user',
    component: () => import('@/layouts/user.vue'),
    children: [
      {
        path: '',
        redirect: '/user/index'
      },
      {
        path: 'index',
        name: 'UserIndex',
        component: () => import('@/views/user/index.vue')
      },
      {
        path: 'orders',
        name: 'UserOrders',
        component: () => import('@/views/user/orders.vue')
      },
      {
        path: 'settles',
        name: 'UserSettles',
        component: () => import('@/views/user/settles.vue')
      },
      {
        path: 'records',
        name: 'UserRecords',
        component: () => import('@/views/user/records.vue')
      },
      {
        path: 'profile',
        name: 'UserProfile',
        component: () => import('@/views/user/profile.vue')
      },
      {
        path: 'invite',
        name: 'UserInvite',
        component: () => import('@/views/user/invite.vue')
      }
    ]
  }
]

const router = createRouter({
  history: createWebHistory(),
  routes
})

// 导航守卫
router.beforeEach((to, _from, next) => {
  const appStore = useAppStore()

  // 如果访问 /admin 开头但不是登录页，需要检查登录状态
  if (to.path.startsWith('/admin') && !to.path.startsWith('/admin/login')) {
    if (!appStore.adminToken) {
      next('/admin/login')
      return
    }
  }

  // 如果访问 /user 开头但不是登录/注册页，需要检查登录状态
  if (to.path.startsWith('/user') && !to.path.startsWith('/user/login') && !to.path.startsWith('/user/register')) {
    if (!appStore.userToken) {
      next('/user/login')
      return
    }
  }

  next()
})

export default router
