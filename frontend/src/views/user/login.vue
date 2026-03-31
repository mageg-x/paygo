<template>
  <div class="min-h-screen bg-gray-100 flex items-center justify-center">
    <div class="bg-white rounded-2xl shadow-xl p-8 w-full max-w-md">
      <div class="text-center mb-8">
        <h1 class="text-3xl font-bold text-gray-800">商户登录</h1>
        <p class="text-gray-500 mt-2">彩虹易支付</p>
      </div>

      <div class="mb-4">
        <div class="flex border rounded-lg overflow-hidden">
          <button
            :class="['flex-1 py-2 text-sm', loginType === 'password' ? 'bg-primary-50 text-primary-700' : 'text-gray-500']"
            @click="loginType = 'password'"
          >
            密码登录
          </button>
          <button
            :class="['flex-1 py-2 text-sm', loginType === 'key' ? 'bg-primary-50 text-primary-700' : 'text-gray-500']"
            @click="loginType = 'key'"
          >
            密钥登录
          </button>
        </div>
      </div>

      <form @submit.prevent="handleLogin" class="space-y-4">
        <div>
          <label class="form-label">商户ID</label>
          <input v-model="form.uid" type="number" class="form-input" placeholder="请输入商户ID" required />
        </div>

        <div v-if="loginType === 'password'">
          <label class="form-label">密码</label>
          <input v-model="form.password" type="password" class="form-input" placeholder="请输入密码" required />
        </div>

        <div v-else>
          <label class="form-label">API密钥</label>
          <input v-model="form.key" type="text" class="form-input" placeholder="请输入API密钥" required />
        </div>

        <button type="submit" class="w-full py-3 bg-primary-600 text-white rounded-lg font-medium hover:bg-primary-700">
          {{ loading ? '登录中...' : '登录' }}
        </button>
      </form>

      <div class="mt-4 text-center">
        <router-link to="/user/register" class="text-primary-600 hover:text-primary-700 text-sm">
          还没有账号？立即注册
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

const router = useRouter()
const appStore = useAppStore()

const form = ref({
  uid: '',
  password: '',
  key: ''
})
const loginType = ref('password')
const loading = ref(false)

async function handleLogin() {
  loading.value = true
  try {
    const data = loginType.value === 'password'
      ? { uid: parseInt(form.value.uid), password: form.value.password }
      : { uid: parseInt(form.value.uid), key: form.value.key }

    const res = await userLogin(data)
    if (res.code === 0 && res.token) {
      appStore.userLogin(res.token, res.data)
      router.push('/user/index')
    }
  } catch (error) {
    console.error('登录失败:', error)
  } finally {
    loading.value = false
  }
}
</script>
