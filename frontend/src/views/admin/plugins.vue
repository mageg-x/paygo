<template>
  <div class="space-y-4">
    <!-- 页面标题 -->
    <div class="flex items-center justify-between">
      <div>
        <h1 class="text-2xl font-bold text-gray-900">插件管理</h1>
        <p class="text-sm text-gray-500 mt-1">管理支付插件和配置</p>
      </div>
      <button @click="handleRefresh"
        class="px-5 py-2.5 bg-white border border-gray-200 text-gray-700 rounded-lg hover:bg-gray-50 transition-colors text-sm font-medium flex items-center gap-2">
        <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
            d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
        </svg>
        刷新插件
      </button>
    </div>

    <!-- 插件列表 -->
    <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
      <div v-for="p in plugins" :key="p.name"
        class="bg-white rounded-xl border border-gray-100 shadow-sm p-5 hover:shadow-md transition-shadow">
        <!-- 插件头部 -->
        <div class="flex items-start justify-between mb-4">
          <div class="flex-1">
            <h3 class="font-semibold text-gray-900">{{ p.showname || p.name }}</h3>
            <span class="inline-block mt-1 px-2 py-0.5 bg-gray-100 text-gray-500 rounded text-xs font-mono">{{ p.name
              }}</span>
          </div>
          <span
            :class="['inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium', p.status ? 'bg-green-100 text-green-700' : 'bg-gray-100 text-gray-600']">
            {{ p.status ? '启用' : '禁用' }}
          </span>
        </div>

        <!-- 插件信息 -->
        <div class="space-y-2 mb-4">
          <div class="flex items-center text-sm text-gray-600">
            <svg class="w-5 h-5 mr-2 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                d="M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z" />
            </svg>
            {{ p.author || '未知作者' }}
          </div>
          <div class="flex items-center text-sm text-gray-600">
            <svg class="w-5 h-5 mr-2 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                d="M3 8l7.89 5.26a2 2 0 002.22 0L21 8M5 19h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z" />
            </svg>
            <span class="truncate">{{ p.types || '未知' }}</span>
          </div>
          <div v-if="p.transtypes" class="flex items-center text-sm text-gray-500">
            <svg class="w-5 h-5 mr-2 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                d="M12 8c-1.657 0-3 .895-3 2s1.343 2 3 2 3 .895 3 2-1.343 2-3 2m0-8c1.11 0 2.08.402 2.599 1M12 8V7m0 1v8m0 0v1m0-1c-1.11 0-2.08-.402-2.599-1M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
            </svg>
            <span class="truncate">转账: {{ p.transtypes }}</span>
          </div>
        </div>

        <!-- 操作按钮 -->
        <div class="flex gap-2 pt-3 border-t border-gray-100">
          <button @click="toggleStatus(p)"
            :class="['flex-1 py-2 text-sm font-medium rounded-lg transition-colors', p.status ? 'text-yellow-600 bg-yellow-50 hover:bg-yellow-100' : 'text-green-600 bg-green-50 hover:bg-green-100']">
            {{ p.status ? '禁用' : '启用' }}
          </button>
          <button @click="showConfig(p)"
            class="flex-1 py-2 text-sm font-medium text-blue-600 bg-blue-50 hover:bg-blue-100 rounded-lg transition-colors">
            配置
          </button>
        </div>
      </div>

      <!-- 空状态 -->
      <div v-if="plugins.length === 0" class="col-span-full">
        <div class="bg-white rounded-xl border border-gray-100 shadow-sm p-12 text-center">
          <svg class="w-16 h-16 text-gray-300 mx-auto mb-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5"
              d="M19.428 15.428a2 2 0 00-1.022-.547l-2.387-.477a6 6 0 00-3.86.517l-.318.158a6 6 0 01-3.86.517L6.05 15.21a2 2 0 00-1.806.547M8 4h8l-1 1v5.172a2 2 0 00.586 1.414l5 5c1.26 1.26.367 3.414-1.415 3.414H4.828c-1.782 0-2.674-2.154-1.414-3.414l5-5A2 2 0 009 10.172V5L8 4z" />
          </svg>
          <h3 class="text-lg font-medium text-gray-900 mb-1">暂无插件</h3>
          <p class="text-gray-500">请刷新或联系管理员安装插件</p>
        </div>
      </div>
    </div>

    <!-- 插件配置弹窗 -->
    <div v-if="showConfigModal" class="fixed inset-0 z-50 overflow-y-auto">
      <div class="flex min-h-full items-center justify-center p-4">
        <div class="fixed inset-0 bg-black/50" @click="showConfigModal = false"></div>
        <div class="relative bg-white rounded-xl shadow-xl w-full max-w-lg p-6">
          <h3 class="text-lg font-semibold text-gray-900 mb-1">配置插件</h3>
          <p class="text-sm text-gray-500 mb-4">{{ currentPlugin?.showname || currentPlugin?.name }}</p>

          <div v-if="!currentPlugin" class="text-center py-8 text-gray-500">
            加载中...
          </div>
          <div v-else class="space-y-4">
            <!-- 基本信息 -->
            <div class="bg-gray-50 rounded-lg p-4">
              <h4 class="text-sm font-medium text-gray-700 mb-3">基本信息</h4>
              <div class="grid grid-cols-2 gap-2 text-sm">
                <div class="text-gray-500">插件名称：<span class="text-gray-900">{{ currentPlugin.showname ||
                    currentPlugin.name }}</span></div>
                <div class="text-gray-500">作者：<span class="text-gray-900">{{ currentPlugin.author || '未知' }}</span>
                </div>
                <div class="text-gray-500">支付方式：<span class="text-gray-900">{{ currentPlugin.types || '未知' }}</span>
                </div>
                <div class="text-gray-500">转账方式：<span class="text-gray-900">{{ currentPlugin.transtypes || '无' }}</span>
                </div>
              </div>
            </div>

            <!-- 配置输入 -->
            <div>
              <label class="block text-sm font-medium text-gray-700 mb-1">插件配置 (JSON)</label>
              <textarea v-model="pluginConfig"
                class="w-full h-32 px-3 py-2 border border-gray-200 rounded-lg text-sm font-mono focus:outline-none focus:ring-2 focus:ring-blue-500 resize-none"
                placeholder='{"key": "value"}'></textarea>
              <p class="text-xs text-gray-400 mt-1">请输入有效的 JSON 格式配置</p>
            </div>
          </div>

          <div class="flex justify-end gap-3 mt-6">
            <button @click="showConfigModal = false"
              class="px-4 py-2 text-sm text-gray-600 hover:bg-gray-100 rounded-lg transition-colors">关闭</button>
            <button @click="savePluginConfig" :disabled="!currentPlugin"
              class="px-4 py-2 text-sm bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors disabled:opacity-50">保存配置</button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { getPluginList, pluginOp } from '@/api/admin'
import { ElMessage } from 'element-plus'

const plugins = ref<any[]>([])
const showConfigModal = ref(false)
const currentPlugin = ref<any>(null)
const pluginConfig = ref('')

async function fetchPlugins() {
  try {
    const res = await getPluginList()
    if (res.code === 0) {
      plugins.value = res.data || []
    }
  } catch (error) {
    console.error('获取插件列表失败:', error)
  }
}

async function handleRefresh() {
  try {
    const res = await pluginOp({ action: 'refresh' })
    ElMessage.success(res.msg || '刷新成功')
    fetchPlugins()
  } catch (error) {
    console.error('刷新失败:', error)
  }
}

async function toggleStatus(p: any) {
  try {
    const res = await pluginOp({
      action: 'set_status',
      name: p.name,
      status: p.status ? 0 : 1
    })
    ElMessage.success(res.msg || '状态已更新')
    fetchPlugins()
  } catch (error) {
    console.error('更新状态失败:', error)
  }
}

function showConfig(p: any) {
  currentPlugin.value = p
  pluginConfig.value = p.config || '{}'
  showConfigModal.value = true
}

async function savePluginConfig() {
  if (!currentPlugin.value) return
  try {
    try {
      JSON.parse(pluginConfig.value)
    } catch {
      ElMessage.warning('JSON 格式不正确')
      return
    }
    const res = await pluginOp({
      action: 'save_config',
      name: currentPlugin.value.name,
      config: pluginConfig.value
    })
    ElMessage.success(res.msg || '保存成功')
    showConfigModal.value = false
    fetchPlugins()
  } catch (error) {
    console.error('保存失败:', error)
  }
}

onMounted(() => {
  fetchPlugins()
})
</script>
