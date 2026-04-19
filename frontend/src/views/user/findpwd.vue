<template>
  <div class="min-h-screen bg-gradient-to-br from-blue-50 to-indigo-100 flex items-center justify-center p-4">
    <div class="bg-white rounded-2xl shadow-xl w-full max-w-md overflow-hidden">
      <!-- 头部 -->
      <div class="bg-gradient-to-r from-blue-600 to-indigo-600 px-8 py-6 text-center">
        <img src="@/assets/gopay.png" alt="Logo" class="w-16 h-16 mx-auto mb-3" />
        <h1 class="text-xl font-bold text-white">找回密码</h1>
        <p class="text-blue-100 text-sm mt-1">通过注册邮箱重置密码</p>
      </div>

      <!-- 步骤指示器 -->
      <div class="px-8 py-4 border-b border-gray-100">
        <div class="flex items-center justify-between">
          <div class="flex items-center">
            <div :class="['w-8 h-8 rounded-full flex items-center justify-center text-sm font-medium',
              step >= 1 ? 'bg-blue-600 text-white' : 'bg-gray-200 text-gray-500']">1</div>
            <span class="ml-2 text-sm" :class="step >= 1 ? 'text-blue-600' : 'text-gray-500'">输入邮箱</span>
          </div>
          <div class="flex-1 h-px bg-gray-200 mx-4"></div>
          <div class="flex items-center">
            <div :class="['w-8 h-8 rounded-full flex items-center justify-center text-sm font-medium',
              step >= 2 ? 'bg-blue-600 text-white' : 'bg-gray-200 text-gray-500']">2</div>
            <span class="ml-2 text-sm" :class="step >= 2 ? 'text-blue-600' : 'text-gray-500'">验证验证码</span>
          </div>
          <div class="flex-1 h-px bg-gray-200 mx-4"></div>
          <div class="flex items-center">
            <div :class="['w-8 h-8 rounded-full flex items-center justify-center text-sm font-medium',
              step >= 3 ? 'bg-blue-600 text-white' : 'bg-gray-200 text-gray-500']">3</div>
            <span class="ml-2 text-sm" :class="step >= 3 ? 'text-blue-600' : 'text-gray-500'">重置密码</span>
          </div>
        </div>
      </div>

      <!-- 表单内容 -->
      <div class="px-8 py-6">
        <!-- 步骤1: 输入邮箱 -->
        <div v-if="step === 1">
          <div class="space-y-4">
            <div>
              <label class="block text-sm font-medium text-gray-700 mb-1">注册邮箱</label>
              <input v-model="form.email" type="email"
                class="w-full px-4 py-3 border border-gray-200 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                placeholder="请输入注册时的邮箱" />
            </div>
          </div>
          <button @click="handleSendCode" :disabled="sending"
            class="w-full mt-6 bg-blue-600 text-white py-3 rounded-lg font-medium hover:bg-blue-700 transition-colors disabled:opacity-50">
            {{ sending ? '发送中...' : '发送验证码' }}
          </button>
          <p class="text-center text-sm text-gray-500 mt-4">
            记起密码了？
            <router-link to="/user/login" class="text-blue-600 hover:underline">返回登录</router-link>
          </p>
        </div>

        <!-- 步骤2: 输入验证码 -->
        <div v-if="step === 2">
          <div class="space-y-4">
            <div>
              <label class="block text-sm font-medium text-gray-700 mb-1">邮箱验证码</label>
              <div class="flex gap-2">
                <input v-model="form.code" type="text"
                  class="flex-1 px-4 py-3 border border-gray-200 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                  placeholder="请输入6位验证码" maxlength="6" />
                <button @click="handleSendCode" :disabled="countdown > 0"
                  class="px-4 py-3 bg-gray-100 text-gray-600 rounded-lg hover:bg-gray-200 transition-colors disabled:opacity-50 text-sm">
                  {{ countdown > 0 ? `${countdown}s` : '重新发送' }}
                </button>
              </div>
              <p class="text-xs text-gray-400 mt-2">验证码已发送至：{{ form.email }}</p>
            </div>
          </div>
          <button @click="step = 3" :disabled="form.code.length < 6"
            class="w-full mt-6 bg-blue-600 text-white py-3 rounded-lg font-medium hover:bg-blue-700 transition-colors disabled:opacity-50">
            下一步
          </button>
          <p class="text-center text-sm text-gray-500 mt-4">
            <button @click="step = 1" class="text-blue-600 hover:underline">重新输入邮箱</button>
          </p>
        </div>

        <!-- 步骤3: 设置新密码 -->
        <div v-if="step === 3">
          <div class="space-y-4">
            <div>
              <label class="block text-sm font-medium text-gray-700 mb-1">新密码</label>
              <input v-model="form.password" type="password"
                class="w-full px-4 py-3 border border-gray-200 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                placeholder="请输入新密码（至少6位）" />
            </div>
            <div>
              <label class="block text-sm font-medium text-gray-700 mb-1">确认密码</label>
              <input v-model="form.confirmPassword" type="password"
                class="w-full px-4 py-3 border border-gray-200 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                placeholder="请再次输入新密码" />
            </div>
          </div>
          <button @click="handleReset" :disabled="resetting"
            class="w-full mt-6 bg-blue-600 text-white py-3 rounded-lg font-medium hover:bg-blue-700 transition-colors disabled:opacity-50">
            {{ resetting ? '重置中...' : '确认重置' }}
          </button>
          <p class="text-center text-sm text-gray-500 mt-4">
            <button @click="step = 2" class="text-blue-600 hover:underline">上一步</button>
          </p>
        </div>

        <!-- 成功提示 -->
        <div v-if="success" class="text-center py-4">
          <div class="w-16 h-16 bg-green-100 rounded-full flex items-center justify-center mx-auto mb-4">
            <svg class="w-8 h-8 text-green-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7" />
            </svg>
          </div>
          <h3 class="text-lg font-medium text-gray-900">密码重置成功</h3>
          <p class="text-gray-500 text-sm mt-2">请使用新密码登录</p>
          <router-link to="/user/login"
            class="inline-block mt-4 px-6 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors">
            前往登录
          </router-link>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive } from 'vue'
import { findPwdSendCode, findPwdReset } from '@/api/user'
import { ElMessage } from 'element-plus'

const step = ref(1)
const sending = ref(false)
const resetting = ref(false)
const success = ref(false)
const countdown = ref(0)

const form = reactive({
  email: '',
  code: '',
  password: '',
  confirmPassword: ''
})

async function handleSendCode() {
  if (!form.email) {
    ElMessage.warning('请输入邮箱')
    return
  }
  if (!/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(form.email)) {
    ElMessage.warning('请输入有效的邮箱地址')
    return
  }

  sending.value = true
  try {
    const res = await findPwdSendCode(form.email)
    if (res.code === 0) {
      ElMessage.success('验证码已发送')
      step.value = 2
      startCountdown()
    } else {
      ElMessage.error(res.msg || '发送失败')
    }
  } catch (error) {
    console.error('发送验证码失败:', error)
    ElMessage.error('发送失败，请稍后重试')
  } finally {
    sending.value = false
  }
}

function startCountdown() {
  countdown.value = 60
  const timer = setInterval(() => {
    countdown.value--
    if (countdown.value <= 0) {
      clearInterval(timer)
    }
  }, 1000)
}

async function handleReset() {
  if (form.password.length < 6) {
    ElMessage.warning('密码长度至少6位')
    return
  }
  if (form.password !== form.confirmPassword) {
    ElMessage.warning('两次输入的密码不一致')
    return
  }

  resetting.value = true
  try {
    const res = await findPwdReset({
      email: form.email,
      code: form.code,
      password: form.password
    })
    if (res.code === 0) {
      success.value = true
    } else {
      ElMessage.error(res.msg || '重置失败')
    }
  } catch (error) {
    console.error('重置密码失败:', error)
    ElMessage.error('重置失败，请稍后重试')
  } finally {
    resetting.value = false
  }
}
</script>
