<template>
  <div class="space-y-4">
    <div class="flex items-center justify-between">
      <div>
        <h1 class="text-2xl font-bold text-gray-900">微信客服设置</h1>
        <p class="text-sm text-gray-500 mt-1">配置商户端微信客服显示</p>
      </div>
    </div>

    <!-- 客服配置 -->
    <div class="bg-white rounded-xl border border-gray-100 shadow-sm p-6">
      <h3 class="text-lg font-semibold text-gray-800 mb-4">客服配置</h3>
      <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
        <div>
          <label class="block text-sm font-medium text-gray-700 mb-2">客服二维码</label>
          <div class="flex items-start gap-4">
            <div class="w-32 h-32 border border-gray-200 rounded-lg overflow-hidden bg-gray-50 flex items-center justify-center">
              <img v-if="form.qrcode" :src="form.qrcode" alt="客服二维码" class="w-full h-full object-contain" />
              <span v-else class="text-gray-400 text-sm">未上传</span>
            </div>
            <div class="flex-1">
              <button @click="triggerUpload"
                class="px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors text-sm font-medium">
                上传二维码
              </button>
              <input ref="fileInput" type="file" accept="image/*" class="hidden" @change="handleFileChange" />
              <p class="text-xs text-gray-500 mt-2">支持 JPG、PNG 格式</p>
            </div>
          </div>
        </div>

        <div class="space-y-4">
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-2">客服链接</label>
            <input v-model="form.link" type="text" placeholder="输入客服微信或链接"
              class="w-full px-3 py-2 border border-gray-200 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-500" />
          </div>

          <div>
            <label class="block text-sm font-medium text-gray-700 mb-2">客服名称</label>
            <input v-model="form.name" type="text" placeholder="显示的客服名称"
              class="w-full px-3 py-2 border border-gray-200 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-500" />
          </div>
        </div>
      </div>

      <div class="mt-6 pt-4 border-t">
        <div class="flex items-center justify-between">
          <div class="flex items-center gap-3">
            <span class="text-sm font-medium text-gray-700">启用客服</span>
            <button @click="form.enabled = !form.enabled"
              :class="['relative inline-flex h-6 w-11 items-center rounded-full transition-colors', form.enabled ? 'bg-blue-600' : 'bg-gray-200']">
              <span :class="['inline-block h-4 w-4 transform rounded-full bg-white transition-transform', form.enabled ? 'translate-x-6' : 'translate-x-1']" />
            </button>
          </div>
          <button @click="handleSave"
            class="px-6 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors text-sm font-medium">
            保存配置
          </button>
        </div>
      </div>
    </div>

    <!-- 客服显示预览 -->
    <div class="bg-white rounded-xl border border-gray-100 shadow-sm p-6">
      <h3 class="text-lg font-semibold text-gray-800 mb-4">收银台预览</h3>
      <div class="border border-gray-200 rounded-lg p-4 bg-gray-50">
        <div class="flex items-center gap-4">
          <div class="w-16 h-16 bg-gray-200 rounded flex items-center justify-center">
            <span class="text-gray-400 text-xs">商品</span>
          </div>
          <div class="flex-1">
            <p class="text-sm font-medium text-gray-800">测试商品</p>
            <p class="text-lg font-bold text-red-600">¥100.00</p>
          </div>
        </div>
        <div class="mt-4 pt-4 border-t border-gray-200 flex items-center justify-between">
          <div class="flex items-center gap-2">
            <img v-if="form.enabled && form.qrcode" :src="form.qrcode" alt="客服" class="w-10 h-10 object-contain" />
            <div v-if="form.enabled">
              <p class="text-sm text-gray-700">{{ form.name || '客服' }}</p>
              <p class="text-xs text-gray-500">{{ form.link || '未设置链接' }}</p>
            </div>
            <span v-else class="text-sm text-gray-400">客服已关闭</span>
          </div>
        </div>
      </div>
    </div>

    <!-- 使用说明 -->
    <div class="bg-white rounded-xl border border-gray-100 shadow-sm p-6">
      <h3 class="text-lg font-semibold text-gray-800 mb-4">使用说明</h3>
      <div class="space-y-2 text-sm text-gray-600">
        <p>1. 上传客服二维码图片，建议尺寸 200x200 像素</p>
        <p>2. 填写客服链接，可以是微信群链接或个人微信</p>
        <p>3. 启用客服后，商户端收银台将显示客服入口</p>
        <p>4. 用户点击客服可直接复制微信号或打开链接</p>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { getSettings, saveSettings } from '@/api/admin'

const form = ref({
  qrcode: '',
  link: '',
  name: '',
  enabled: false
})

const fileInput = ref<HTMLInputElement>()

async function fetchConfig() {
  try {
    const res = await getSettings(['wxkf_qrcode', 'wxkf_link', 'wxkf_name', 'wxkf_enabled'])
    if (res.code === 0) {
      form.value.qrcode = res.data.wxkf_qrcode || ''
      form.value.link = res.data.wxkf_link || ''
      form.value.name = res.data.wxkf_name || ''
      form.value.enabled = res.data.wxkf_enabled === '1'
    }
  } catch (error) {
    console.error('获取配置失败:', error)
  }
}

function triggerUpload() {
  fileInput.value?.click()
}

function handleFileChange(e: Event) {
  const file = (e.target as HTMLInputElement).files?.[0]
  if (!file) return

  if (!file.type.startsWith('image/')) {
    ElMessage.error('请上传图片文件')
    return
  }

  const reader = new FileReader()
  reader.onload = (e) => {
    form.value.qrcode = e.target?.result as string
  }
  reader.readAsDataURL(file)
}

async function handleSave() {
  try {
    await saveSettings({
      wxkf_qrcode: form.value.qrcode,
      wxkf_link: form.value.link,
      wxkf_name: form.value.name,
      wxkf_enabled: form.value.enabled ? '1' : '0'
    })
    ElMessage.success('保存成功')
  } catch (error) {
    console.error('保存配置失败:', error)
    ElMessage.error('保存失败')
  }
}

onMounted(() => {
  fetchConfig()
})
</script>
