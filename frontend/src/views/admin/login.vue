<template>
  <div class="min-h-screen bg-gray-100 flex items-center justify-center">
    <div class="bg-white rounded-2xl shadow-xl p-8 w-full max-w-md">
      <div class="text-center mb-8">
        <h1 class="text-3xl font-bold text-gray-800">管理员登录</h1>
        <p class="text-gray-500 mt-2">支付系统管理后台</p>
      </div>

      <form @submit.prevent="handleLogin" class="space-y-4">
        <div>
          <label class="form-label">用户名</label>
          <input v-model="form.username" type="text" class="form-input" placeholder="请输入用户名" />
        </div>

        <div>
          <label class="form-label">密码</label>
          <input v-model="form.password" type="password" class="form-input" placeholder="请输入密码" />
        </div>

        <div v-if="errorMsg" class="p-3 bg-red-50 border border-red-200 rounded-lg">
          <p class="text-sm text-red-600">{{ errorMsg }}</p>
        </div>

        <button type="submit" :disabled="loading"
          class="w-full py-3 bg-primary-600 text-white rounded-lg font-medium hover:bg-primary-700 transition-colors disabled:opacity-50">
          {{ loading ? '登录中...' : '登录' }}
        </button>
      </form>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { adminLogin } from '@/api/admin'
import { useAppStore } from '@/stores/app'

const router = useRouter()
const appStore = useAppStore()

const form = ref({
  username: '',
  password: ''
})
const loading = ref(false)
const errorMsg = ref('')

async function handleLogin() {
  if (!form.value.username || !form.value.password) {
    errorMsg.value = '请填写用户名和密码'
    return
  }

  loading.value = true
  errorMsg.value = ''

  try {
    const res = await adminLogin(form.value)
    if (res.code === 0 && res.token) {
      appStore.adminLogin(res.token, form.value.username)
      router.push('/admin/index')
    } else {
      errorMsg.value = res.msg || '登录失败'
    }
  } catch (error: any) {
    console.error('登录失败:', error)
    errorMsg.value = error.message || '登录失败，请稍后重试'
  } finally {
    loading.value = false
  }
}
</script>
