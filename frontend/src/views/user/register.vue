<template>
  <div class="min-h-screen bg-gray-100 flex items-center justify-center">
    <div class="bg-white rounded-2xl shadow-xl p-8 w-full max-w-md">
      <div class="text-center mb-8">
        <img src="@/assets/paygo.png" alt="Logo" class="w-16 h-16 mx-auto mb-4" />
        <h1 class="text-3xl font-bold text-gray-800">商户注册</h1>
        <p class="text-gray-500 mt-2">彩虹易支付</p>
      </div>

      <form @submit.prevent="handleRegister" class="space-y-4">
        <div>
          <label class="form-label">邮箱</label>
          <div class="relative">
            <Mail class="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-gray-400 pointer-events-none z-10" />
            <input v-model="form.email" type="email" class="form-input pl-10 pr-3 relative" placeholder="请输入邮箱"
              required />
          </div>
        </div>

        <div>
          <label class="form-label">手机号（可选）</label>
          <div class="relative">
            <Phone class="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-gray-400 pointer-events-none z-10" />
            <input v-model="form.phone" type="tel" class="form-input pl-10 pr-3 relative" placeholder="请输入手机号" />
          </div>
        </div>

        <div>
          <label class="form-label">密码</label>
          <div class="relative">
            <Lock class="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-gray-400 pointer-events-none z-10" />
            <input v-model="form.password" type="password" class="form-input pl-10 pr-3 relative" placeholder="请输入密码"
              required />
          </div>
        </div>

        <div>
          <label class="form-label">邀请码（可选）</label>
          <div class="relative">
            <Ticket class="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-gray-400 pointer-events-none z-10" />
            <input v-model="form.invite_code" type="text" class="form-input pl-10 pr-3 relative" placeholder="请输入邀请码" />
          </div>
        </div>

        <button type="submit"
          class="w-full py-3 bg-primary-600 text-white rounded-lg font-medium hover:bg-primary-700 transition-colors flex items-center justify-center gap-2">
          <Loader2 v-if="loading" class="w-4 h-4 animate-spin" />
          <UserPlus v-else class="w-4 h-4" />
          {{ loading ? '注册中...' : '注册' }}
        </button>
      </form>

      <div class="mt-6 text-center">
        <router-link to="/user/login"
          class="text-primary-600 hover:text-primary-700 text-sm inline-flex items-center gap-1">
          <LogIn class="w-4 h-4" />
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
import { Mail, Phone, Lock, Ticket, Loader2, LogIn, UserPlus } from 'lucide-vue-next'

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
      ElMessage.success('注册成功，请登录')
      router.push('/user/login')
    }
  } catch (error: any) {
    ElMessage.error(error.message || '注册失败')
  } finally {
    loading.value = false
  }
}
</script>
