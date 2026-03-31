<template>
  <div class="min-h-screen bg-gray-100 flex items-center justify-center">
    <div class="bg-white rounded-2xl shadow-xl p-8 w-full max-w-md">
      <div class="text-center mb-8">
        <h1 class="text-3xl font-bold text-gray-800">商户注册</h1>
        <p class="text-gray-500 mt-2">彩虹易支付</p>
      </div>

      <form @submit.prevent="handleRegister" class="space-y-4">
        <div>
          <label class="form-label">邮箱</label>
          <input v-model="form.email" type="email" class="form-input" placeholder="请输入邮箱" required />
        </div>

        <div>
          <label class="form-label">手机号（可选）</label>
          <input v-model="form.phone" type="tel" class="form-input" placeholder="请输入手机号" />
        </div>

        <div>
          <label class="form-label">密码</label>
          <input v-model="form.password" type="password" class="form-input" placeholder="请输入密码" required />
        </div>

        <div>
          <label class="form-label">邀请码（可选）</label>
          <input v-model="form.invite_code" type="text" class="form-input" placeholder="请输入邀请码" />
        </div>

        <button type="submit" class="w-full py-3 bg-primary-600 text-white rounded-lg font-medium hover:bg-primary-700">
          {{ loading ? '注册中...' : '注册' }}
        </button>
      </form>

      <div class="mt-4 text-center">
        <router-link to="/user/login" class="text-primary-600 hover:text-primary-700 text-sm">
          已有账号？立即登录
        </router-link>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { userRegister } from '@/api/user'

const router = useRouter()

const form = ref({
  email: '',
  phone: '',
  password: '',
  invite_code: ''
})
const loading = ref(false)

async function handleRegister() {
  loading.value = true
  try {
    const res = await userRegister(form.value)
    if (res.code === 0) {
      alert('注册成功')
      router.push('/user/login')
    }
  } catch (error) {
    console.error('注册失败:', error)
  } finally {
    loading.value = false
  }
}
</script>
