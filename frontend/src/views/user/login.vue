<template>
  <div class="min-h-screen bg-gray-100 flex items-center justify-center">
    <div class="bg-white rounded-2xl shadow-xl p-8 w-full max-w-md">
      <div class="text-center mb-8">
        <img src="@/assets/paygo.png" alt="Logo" class="w-16 h-16 mx-auto mb-4" />
        <h1 class="text-3xl font-bold text-gray-800">商户登录</h1>
        <p class="text-gray-500 mt-2">GoPay支付</p>
      </div>

      <div class="mb-4">
        <div class="flex border rounded-lg overflow-hidden">
          <button
            :class="['flex-1 py-2.5 text-sm font-medium flex items-center justify-center gap-2', loginType === 'password' ? 'bg-primary-50 text-primary-700' : 'text-gray-500 hover:bg-gray-50']"
            @click="loginType = 'password'">
            <KeyRound class="w-4 h-4" />
            密码登录
          </button>
          <button
            :class="['flex-1 py-2.5 text-sm font-medium flex items-center justify-center gap-2', loginType === 'key' ? 'bg-primary-50 text-primary-700' : 'text-gray-500 hover:bg-gray-50']"
            @click="loginType = 'key'">
            <Key class="w-4 h-4" />
            密钥登录
          </button>
        </div>
      </div>

      <form @submit.prevent="handleLogin" class="space-y-4">
        <div v-if="errorMsg" class="p-3 bg-red-50 border border-red-200 rounded-lg flex items-center gap-2">
          <AlertCircle class="w-4 h-4 text-red-500 flex-shrink-0" />
          <p class="text-sm text-red-600">{{ errorMsg }}</p>
        </div>

        <div>
          <label class="form-label">商户ID</label>
          <div class="relative">
            <Hash class="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-gray-400 pointer-events-none z-10" />
            <input v-model="form.uid" type="number" class="form-input pl-10 pr-3 relative" placeholder="请输入商户ID"
              required />
          </div>
        </div>

        <div v-if="loginType === 'password'">
          <label class="form-label">密码</label>
          <div class="relative">
            <Lock class="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-gray-400 pointer-events-none z-10" />
            <input v-model="form.password" type="password" class="form-input pl-10 pr-3 relative" placeholder="请输入密码"
              required />
          </div>
        </div>

        <div v-else>
          <label class="form-label">API密钥</label>
          <div class="relative">
            <Key class="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-gray-400 pointer-events-none z-10" />
            <input v-model="form.key" type="text" class="form-input pl-10 pr-3 relative" placeholder="请输入API密钥"
              required />
          </div>
        </div>

        <button type="submit"
          class="w-full py-3 bg-primary-600 text-white rounded-lg font-medium hover:bg-primary-700 transition-colors flex items-center justify-center gap-2">
          <Loader2 v-if="loading" class="w-4 h-4 animate-spin" />
          <LogIn v-else class="w-4 h-4" />
          {{ loading ? '登录中...' : '登录' }}
        </button>
      </form>

      <div class="mt-6 text-center space-y-2">
        <router-link to="/user/register"
          class="text-primary-600 hover:text-primary-700 text-sm inline-flex items-center gap-1">
          <UserPlus class="w-4 h-4" />
          还没有账号？立即注册
        </router-link>
        <br />
        <router-link to="/user/findpwd"
          class="text-gray-500 hover:text-gray-700 text-sm inline-flex items-center gap-1">
          忘记密码？
        </router-link>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { userLogin } from '@/api/user'
import { useAppStore } from '@/stores/app'
import { KeyRound, Key, Hash, Lock, LogIn, Loader2, UserPlus, AlertCircle } from 'lucide-vue-next'

const router = useRouter()
const appStore = useAppStore()

const form = ref({
  uid: '',
  password: '',
  key: ''
})
const loginType = ref('password')
const loading = ref(false)
const errorMsg = ref('')

async function handleLogin() {
  loading.value = true
  errorMsg.value = ''
  try {
    const data = loginType.value === 'password'
      ? { uid: parseInt(form.value.uid), password: form.value.password }
      : { uid: parseInt(form.value.uid), key: form.value.key }

    const res = await userLogin(data)
    if (res.code === 0 && res.token) {
      const userInfo = {
        uid: res.uid || parseInt(form.value.uid),
        username: '',
        email: '',
        phone: '',
        money: 0,
        status: 1
      }
      appStore.userLogin(res.token, userInfo)
      router.push('/user/index')
    }
  } catch (error: any) {
    console.error('登录失败:', error)
    errorMsg.value = error.message || '登录失败'
  } finally {
    loading.value = false
  }
}
</script>
