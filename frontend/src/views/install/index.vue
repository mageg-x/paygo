<template>
  <div class="min-h-screen bg-gradient-to-br from-blue-50 to-indigo-100 flex items-center justify-center p-4">
    <div class="bg-white rounded-2xl shadow-xl w-full max-w-lg overflow-hidden">
      <!-- 头部 -->
      <div class="bg-gradient-to-r from-blue-600 to-indigo-600 px-8 py-6">
        <div class="flex items-center gap-3">
          <img src="@/assets/paygo.png" alt="Logo" class="w-10 h-10" />
          <div>
            <h1 class="text-xl font-bold text-white">PayGo支付系统</h1>
            <p class="text-blue-100 text-sm">安装向导</p>
          </div>
        </div>
      </div>

      <!-- 步骤指示器 -->
      <div class="px-8 py-4 border-b border-gray-100">
        <div class="flex items-center justify-between">
          <div v-for="(s, idx) in steps" :key="idx" class="flex items-center">
            <div :class="['w-8 h-8 rounded-full flex items-center justify-center text-sm font-medium',
              currentStep > idx ? 'bg-green-500 text-white' :
              currentStep === idx ? 'bg-blue-600 text-white' : 'bg-gray-200 text-gray-500']">
              <svg v-if="currentStep > idx" class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7" />
              </svg>
              <span v-else>{{ idx + 1 }}</span>
            </div>
            <span v-if="idx < steps.length - 1" :class="['mx-2 text-sm', currentStep > idx ? 'text-green-600' : 'text-gray-400']">
              <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7" />
              </svg>
            </span>
          </div>
        </div>
        <div class="flex justify-between mt-2 text-xs text-gray-500">
          <span v-for="(s, idx) in steps" :key="idx">{{ s }}</span>
        </div>
      </div>

      <!-- 内容区 -->
      <div class="px-8 py-6">
        <!-- 步骤1: 环境检查 -->
        <div v-if="currentStep === 0">
          <h2 class="text-lg font-semibold text-gray-900 mb-4">环境检查</h2>
          <div class="space-y-3">
            <div v-for="check in envChecks" :key="check.name" class="flex items-center justify-between p-3 bg-gray-50 rounded-lg">
              <span class="text-sm text-gray-700">{{ check.name }}</span>
              <span v-if="check.ok" class="text-green-600 flex items-center gap-1">
                <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7" />
                </svg>
                {{ check.value }}
              </span>
              <span v-else class="text-red-600 flex items-center gap-1">
                <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
                </svg>
                {{ check.value }}
              </span>
            </div>
          </div>
        </div>

        <!-- 步骤2: 数据库配置 -->
        <div v-if="currentStep === 1">
          <h2 class="text-lg font-semibold text-gray-900 mb-4">数据库配置</h2>
          <div class="space-y-4">
            <div>
              <label class="block text-sm font-medium text-gray-700 mb-1">数据库路径</label>
              <input v-model="form.dbPath" type="text"
                placeholder="/path/to/pay.db"
                class="w-full px-4 py-2.5 border border-gray-200 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-500" />
              <p class="text-xs text-gray-500 mt-1">默认为 ./data/pay.db</p>
            </div>
          </div>
        </div>

        <!-- 步骤3: 管理员配置 -->
        <div v-if="currentStep === 2">
          <h2 class="text-lg font-semibold text-gray-900 mb-4">管理员配置</h2>
          <div class="space-y-4">
            <div>
              <label class="block text-sm font-medium text-gray-700 mb-1">管理员用户名</label>
              <input v-model="form.adminUser" type="text"
                placeholder="admin"
                class="w-full px-4 py-2.5 border border-gray-200 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-500" />
            </div>
            <div>
              <label class="block text-sm font-medium text-gray-700 mb-1">管理员密码</label>
              <input v-model="form.adminPwd" type="password"
                placeholder="输入密码"
                class="w-full px-4 py-2.5 border border-gray-200 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-500" />
            </div>
            <div>
              <label class="block text-sm font-medium text-gray-700 mb-1">系统密钥</label>
              <input v-model="form.sysKey" type="text"
                placeholder="系统密钥"
                class="w-full px-4 py-2.5 border border-gray-200 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-500" />
              <p class="text-xs text-gray-500 mt-1">用于API签名验证，建议使用随机字符串</p>
            </div>
          </div>
        </div>

        <!-- 步骤4: 完成 -->
        <div v-if="currentStep === 3">
          <div class="text-center py-8">
            <div class="w-16 h-16 bg-green-100 rounded-full flex items-center justify-center mx-auto mb-4">
              <svg class="w-8 h-8 text-green-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7" />
              </svg>
            </div>
            <h2 class="text-xl font-semibold text-gray-900 mb-2">安装成功！</h2>
            <p class="text-gray-500 mb-6">PayGo支付系统已安装完成</p>
            <button @click="goToAdmin"
              class="px-6 py-2.5 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors text-sm font-medium">
              前往管理后台
            </button>
          </div>
        </div>
      </div>

      <!-- 底部按钮 -->
      <div v-if="currentStep < 3" class="px-8 py-4 bg-gray-50 border-t border-gray-100 flex justify-between">
        <button v-if="currentStep > 0" @click="currentStep--"
          class="px-6 py-2 text-gray-600 hover:bg-gray-100 rounded-lg transition-colors text-sm">
          上一步
        </button>
        <div v-else></div>
        <button v-if="currentStep < 2" @click="nextStep"
          class="px-6 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors text-sm">
          下一步
        </button>
        <button v-else @click="doInstall" :disabled="installing"
          class="px-6 py-2 bg-green-600 text-white rounded-lg hover:bg-green-700 transition-colors text-sm disabled:opacity-50">
          {{ installing ? '安装中...' : '完成安装' }}
        </button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import request from '@/api/request'

const router = useRouter()
const currentStep = ref(0)
const installing = ref(false)
const steps = ['环境检查', '数据库配置', '管理员配置', '安装完成']
const envChecks = ref([
  { name: 'Go 环境', ok: true, value: '正常' },
  { name: 'SQLite 支持', ok: true, value: '正常' },
  { name: '静态文件', ok: true, value: '正常' }
])
const form = ref({
  dbPath: './data/pay.db',
  adminUser: 'admin',
  adminPwd: '12345678',
  sysKey: 'paygosyskey2024'
})

async function checkInstallStatus() {
  try {
    const res = await request.get('/install/status')
    if (res.code === 0 && res.status === 1) {
      // 已安装，跳过安装向导
      router.replace('/admin/login')
    }
  } catch (error) {
    console.error('检查安装状态失败:', error)
  }
}

function nextStep() {
  if (currentStep.value < steps.length - 1) {
    currentStep.value++
  }
}

async function doInstall() {
  installing.value = true
  try {
    const res = await request.post('/install/do', {
      db_path: form.value.dbPath,
      admin_user: form.value.adminUser,
      admin_pwd: form.value.adminPwd,
      sys_key: form.value.sysKey
    })
    if (res.code === 0) {
      currentStep.value = 3
    } else {
      alert(res.msg || '安装失败')
    }
  } catch (error) {
    console.error('安装失败:', error)
    alert('安装失败')
  } finally {
    installing.value = false
  }
}

function goToAdmin() {
  router.push('/admin/login')
}

onMounted(() => {
  checkInstallStatus()
})
</script>
