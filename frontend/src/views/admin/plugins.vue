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
    <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-5">
      <div v-for="p in plugins" :key="p.name"
        class="bg-white rounded-2xl border border-gray-100 shadow-sm overflow-hidden hover:shadow-lg transition-all hover:-translate-y-0.5">
        <!-- 插件头部 - 带图标背景 -->
        <div class="relative px-5 pt-5 pb-4">
          <div class="flex items-start justify-between">
            <div class="flex items-center gap-3">
              <div class="w-12 h-12 rounded-xl flex items-center justify-center" :class="getPluginBgClass(p.types)">
                <SvgIcon v-if="p.types?.includes('支付宝')" name="alipay" :size="28" />
                <SvgIcon v-else-if="p.types?.includes('微信')" name="wechatpay" :size="28" />
                <svg v-else class="w-6 h-6 text-gray-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                    d="M19.428 15.428a2 2 0 00-1.022-.547l-2.387-.477a6 6 0 00-3.86.517l-.318.158a6 6 0 01-3.86.517L6.05 15.21a2 2 0 00-1.806.547M8 4h8l-1 1v5.172a2 2 0 00.586 1.414l5 5c1.26 1.26.367 3.414-1.415 3.414H4.828c-1.782 0-2.674-2.154-1.414-3.414l5-5A2 2 0 009 10.172V5L8 4z" />
                </svg>
              </div>
              <div>
                <h3 class="font-semibold text-gray-900 text-lg">{{ p.showname || p.name }}</h3>
                <span class="text-xs text-gray-400 font-mono">{{ p.name }}</span>
              </div>
            </div>
            <span
              :class="['inline-flex items-center px-2.5 py-1 rounded-full text-xs font-medium',
                p.status ? 'bg-green-100 text-green-700 ring-1 ring-green-200' : 'bg-gray-100 text-gray-500 ring-1 ring-gray-200']">
              <span class="w-1.5 h-1.5 rounded-full mr-1.5" :class="p.status ? 'bg-green-500' : 'bg-gray-400'"></span>
              {{ p.status ? '已启用' : '已禁用' }}
            </span>
          </div>
        </div>

        <!-- 插件信息 -->
        <div class="px-5 pb-4 space-y-2">
          <div class="flex items-center text-sm text-gray-500">
            <svg class="w-4 h-4 mr-2 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                d="M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z" />
            </svg>
            <span>{{ p.author || '未知作者' }}</span>
          </div>
          <div class="flex items-center text-sm text-gray-500">
            <svg class="w-4 h-4 mr-2 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
            </svg>
            <span>{{ p.types || '未知类型' }}</span>
          </div>
          <div v-if="p.transtypes" class="flex items-center text-sm text-gray-500">
            <svg class="w-4 h-4 mr-2 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                d="M8 7h12m0 0l-4-4m4 4l-4 4m0 6H4m0 0l4 4m-4-4l4-4" />
            </svg>
            <span>转账: {{ p.transtypes }}</span>
          </div>
        </div>

        <!-- 操作按钮 -->
        <div class="px-5 py-3 bg-gray-50 border-t border-gray-100 flex gap-2">
          <button @click="toggleStatus(p)" :class="['flex-1 py-2 text-sm font-medium rounded-lg transition-all',
            p.status
              ? 'text-amber-600 bg-amber-50 hover:bg-amber-100 ring-1 ring-amber-200'
              : 'text-emerald-600 bg-emerald-50 hover:bg-emerald-100 ring-1 ring-emerald-200']">
            {{ p.status ? '禁用' : '启用' }}
          </button>
          <button @click="showConfig(p)"
            class="flex-1 py-2 text-sm font-medium text-blue-600 bg-blue-50 hover:bg-blue-100 rounded-lg transition-all ring-1 ring-blue-200">
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
                <div class="text-gray-500">支付方式：
                  <span class="text-gray-900 inline-flex items-center gap-1">
                    <SvgIcon v-if="currentPlugin.types?.includes('支付宝')" name="alipay" :size="14" />
                    <SvgIcon v-if="currentPlugin.types?.includes('微信')" name="wechatpay" :size="14" />
                    {{ currentPlugin.types || '未知' }}
                  </span>
                </div>
                <div class="text-gray-500">转账方式：<span class="text-gray-900">{{ currentPlugin.transtypes || '无' }}</span>
                </div>
              </div>
            </div>

            <!-- 配置输入 -->
            <div>
              <label class="block text-sm font-medium text-gray-700 mb-1">插件配置 (JSON)</label>
              <JsonEditor v-model="pluginConfig" placeholder='{"appid": "xxx", "appkey": "xxx"}' />
            </div>

            <!-- 证书上传（仅微信支付需要） -->
            <div v-if="currentPlugin?.name === 'wxpay'" class="bg-amber-50 rounded-lg p-4">
              <div class="flex items-center justify-between mb-3">
                <div>
                  <div class="font-medium text-amber-800">微信支付证书</div>
                  <div class="text-xs text-amber-600 mt-1">微信退款、转账等功能需要上传证书</div>
                </div>
              </div>
              <div class="flex items-center gap-3">
                <input type="file" ref="certInput" accept=".pem,.cert,.key,.p12,.pfx" @change="handleCertSelect"
                  class="hidden" />
                <button @click="$refs.certInput.click()"
                  class="px-4 py-2 text-sm bg-white border border-amber-300 text-amber-700 rounded-lg hover:bg-amber-100 transition-colors">
                  选择证书文件
                </button>
                <span class="text-sm text-amber-700">{{ certFileName || '未选择文件' }}</span>
              </div>
              <div v-if="certFileName" class="mt-3 p-2 bg-white rounded border border-amber-200">
                <div class="text-xs text-gray-500">证书路径（自动填入配置）:</div>
                <div class="font-mono text-sm text-gray-700 mt-1">{{ certPath }}</div>
              </div>
              <div v-if="uploading" class="mt-3 text-sm text-amber-600">
                上传中...
              </div>
            </div>

            <!-- 配置帮助 -->
            <div v-if="currentPlugin?.note" class="bg-blue-50 rounded-lg p-4 text-sm">
              <div class="font-medium text-blue-700 mb-2">配置说明</div>
              <div class="text-blue-600 prose prose-sm max-w-none" v-html="currentPlugin.note"></div>
            </div>

            <!-- 配置参数说明 -->
            <div v-if="currentPlugin?.inputs && Object.keys(currentPlugin.inputs).length > 0" class="bg-gray-50 rounded-lg p-4">
              <div class="font-medium text-gray-700 mb-2">参数说明</div>
              <div class="space-y-2">
                <div v-for="(input, key) in currentPlugin.inputs" :key="key" class="flex items-start gap-2 text-sm">
                  <span class="font-mono text-gray-600 bg-gray-100 px-1.5 py-0.5 rounded min-w-[80px]">{{ key }}</span>
                  <span class="text-gray-500">{{ input.name }}</span>
                  <span v-if="input.note" class="text-gray-400 text-xs">({{ input.note }})</span>
                </div>
              </div>
            </div>
          </div>

          <div class="flex justify-between gap-3 mt-6">
            <button @click="showConfigModal = false"
              class="px-4 py-2 text-sm text-gray-600 hover:bg-gray-100 rounded-lg transition-colors">关闭</button>
            <button v-if="currentPlugin?.name === 'alipay' || currentPlugin?.name === 'wxpay'"
              @click="testPluginConfig"
              class="px-4 py-2 text-sm bg-emerald-600 text-white rounded-lg hover:bg-emerald-700 transition-colors flex items-center gap-2">
              <TestIcon class="w-4 h-4" />
              测试配置
            </button>
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
import request from '@/api/request'
import { getPluginList, pluginOp } from '@/api/admin'
import { ElMessage } from 'element-plus'
import SvgIcon from '@/components/svgicon.vue'
import JsonEditor from '@/components/json-editor.vue'
import { Beaker } from 'lucide-vue-next'

const plugins = ref<any[]>([])
const showConfigModal = ref(false)
const currentPlugin = ref<any>(null)
const pluginConfig = ref('')
const certInput = ref<HTMLInputElement | null>(null)
const certFileName = ref('')
const certPath = ref('')
const uploading = ref(false)
const TestIcon = Beaker

function getPluginBgClass(types: string) {
  if (types?.includes('支付宝')) return 'bg-blue-50'
  if (types?.includes('微信')) return 'bg-green-50'
  return 'bg-gray-100'
}

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
  certFileName.value = ''
  certPath.value = ''
  showConfigModal.value = true
}

async function testPluginConfig() {
  if (!currentPlugin.value) return
  try {
    const res = await pluginOp({
      action: 'test_config',
      name: currentPlugin.value.name,
      config: pluginConfig.value
    })
    if (res.code === 0) {
      ElMessage.success(res.msg || '测试成功')
    } else {
      ElMessage.error(res.msg || '测试失败')
    }
  } catch (error) {
    console.error('测试失败:', error)
    const msg = (error as any)?.message || '测试失败，请检查网络连接'
    ElMessage.error(`测试失败: ${msg}`)
  }
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

async function handleCertSelect(e: Event) {
  const target = e.target as HTMLInputElement
  const file = target.files?.[0]
  if (!file) return

  certFileName.value = file.name
  uploading.value = true

  const formData = new FormData()
  formData.append('cert', file)

  try {
    const res = await request.post('/admin/upload/cert', formData, {
      headers: { 'Content-Type': 'multipart/form-data' }
    })
    if (res.code === 0) {
      certPath.value = res.data.path
      ElMessage.success('证书上传成功')

      // 自动填入配置
      try {
        const config = JSON.parse(pluginConfig.value || '{}')
        if (file.name.endsWith('.p12') || file.name.endsWith('.pfx')) {
          config.cert_path = res.data.path
        } else if (file.name.endsWith('.pem') || file.name.endsWith('.cert')) {
          config.cert_path = res.data.path
        } else if (file.name.endsWith('.key')) {
          config.key_path = res.data.path
        }
        pluginConfig.value = JSON.stringify(config, null, 2)
      } catch {
        pluginConfig.value = JSON.stringify({ cert_path: res.data.path }, null, 2)
      }
    } else {
      ElMessage.error(res.msg || '上传失败')
    }
  } catch (error) {
    console.error('证书上传失败:', error)
    ElMessage.error('证书上传失败')
  } finally {
    uploading.value = false
  }
}

onMounted(() => {
  fetchPlugins()
})
</script>
