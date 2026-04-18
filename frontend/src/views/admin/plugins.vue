<template>
  <div class="space-y-4">
    <div class="page-head">
      <div>
        <h1 class="page-title no-wrap">插件管理</h1>
        <p class="page-subtitle">管理支付插件和配置</p>
      </div>
      <button @click="handleRefresh" class="btn btn-outline" :disabled="refreshing">
        <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
            d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
        </svg>
        {{ refreshing ? '刷新中...' : '刷新插件' }}
      </button>
    </div>

    <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-5">
      <div v-for="p in plugins" :key="p.name"
        class="card overflow-hidden hover:-translate-y-0.5">
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
            <span :class="['badge', p.status ? 'badge-success' : 'badge-info']">
              <span class="w-1.5 h-1.5 rounded-full mr-1.5" :class="p.status ? 'bg-green-500' : 'bg-gray-400'"></span>
              {{ p.status ? '已启用' : '已禁用' }}
            </span>
          </div>
        </div>

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

        <div class="px-5 py-3 bg-slate-50/75 border-t border-slate-200/70 flex gap-2">
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

      <div v-if="plugins.length === 0" class="col-span-full">
        <div class="card p-12 text-center">
          <svg class="w-16 h-16 text-gray-300 mx-auto mb-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5"
              d="M19.428 15.428a2 2 0 00-1.022-.547l-2.387-.477a6 6 0 00-3.86.517l-.318.158a6 6 0 01-3.86.517L6.05 15.21a2 2 0 00-1.806.547M8 4h8l-1 1v5.172a2 2 0 00.586 1.414l5 5c1.26 1.26.367 3.414-1.415 3.414H4.828c-1.782 0-2.674-2.154-1.414-3.414l5-5A2 2 0 009 10.172V5L8 4z" />
          </svg>
          <h3 class="text-lg font-medium text-gray-900 mb-1">暂无插件</h3>
          <p class="text-gray-500">未检测到插件，请先点击“刷新插件”扫描内置插件</p>
        </div>
      </div>
    </div>

    <div v-if="showConfigModal" class="dialog-backdrop">
      <div class="dialog-wrap">
        <div class="dialog-mask" @click="showConfigModal = false"></div>
        <div class="dialog-panel max-w-lg overflow-hidden">
          <div class="dialog-header">
            <div>
              <h3 class="dialog-title">配置插件</h3>
              <p class="dialog-subtitle">{{ currentPlugin?.showname || currentPlugin?.name }}</p>
            </div>
            <button class="dialog-close" @click="showConfigModal = false">✕</button>
          </div>

          <div class="dialog-body" v-if="!currentPlugin">
            <div class="text-center py-8 text-gray-500">
            加载中...
            </div>
          </div>

          <div class="dialog-body space-y-4" v-else>
            <div class="section-card">
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

            <div class="space-y-3">
              <label class="block text-sm font-medium text-gray-700">插件配置</label>
              <template v-if="currentPluginInputs.length > 0">
                <div
                  v-for="field in currentPluginInputs"
                  :key="field.key"
                  class="space-y-1"
                >
                  <label class="block text-sm text-gray-700">
                    {{ field.input?.name || field.key }}
                    <span class="text-xs text-gray-400 font-mono ml-1">({{ field.key }})</span>
                  </label>

                  <select
                    v-if="field.input?.type === 'select'"
                    v-model="pluginForm[field.key]"
                    class="w-full rounded-lg border border-gray-300 px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-400"
                  >
                    <option value="">请选择</option>
                    <option
                      v-for="(label, val) in field.input.options || {}"
                      :key="val"
                      :value="val"
                    >
                      {{ label }}
                    </option>
                  </select>

                  <textarea
                    v-else-if="field.input?.type === 'textarea'"
                    v-model="pluginForm[field.key]"
                    rows="6"
                    class="w-full rounded-lg border border-gray-300 px-3 py-2 text-sm font-mono focus:outline-none focus:ring-2 focus:ring-blue-400"
                  />

                  <input
                    v-else
                    v-model="pluginForm[field.key]"
                    :type="isSensitiveField(field.key) ? 'password' : 'text'"
                    class="w-full rounded-lg border border-gray-300 px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-400"
                    autocomplete="off"
                  />

                  <div v-if="field.input?.note" class="text-xs text-gray-500">
                    {{ field.input.note }}
                  </div>
                </div>
              </template>
              <div v-else class="text-sm text-gray-500 bg-gray-50 border border-gray-200 rounded-lg px-3 py-2">
                当前插件未声明可视化配置字段
              </div>
            </div>

            <div v-if="currentPlugin?.note" class="section-card text-sm">
              <div class="font-medium text-blue-700 mb-2">配置说明</div>
              <div class="text-blue-600 text-sm whitespace-pre-wrap break-words">{{ currentPlugin.note }}</div>
            </div>

            <div v-if="currentPlugin?.inputs && Object.keys(currentPlugin.inputs).length > 0" class="section-card">
              <div class="font-medium text-gray-700 mb-3">参数说明</div>
              <table class="w-full text-sm">
                <thead>
                  <tr class="border-b border-gray-200">
                    <th class="text-left py-1.5 pr-4 font-semibold text-gray-600 w-[140px]">参数名</th>
                    <th class="text-left py-1.5 pr-4 font-semibold text-gray-600">说明</th>
                    <th class="text-left py-1.5 font-semibold text-gray-600">备注</th>
                  </tr>
                </thead>
                <tbody>
                  <tr v-for="(input, key) in currentPlugin.inputs" :key="key" class="border-b border-gray-100 last:border-b-0">
                    <td class="py-1.5 pr-4 font-mono text-gray-700 bg-gray-50 rounded px-1.5 align-top">{{ key }}</td>
                    <td class="py-1.5 pr-4 text-gray-800 align-top whitespace-nowrap">{{ input.name }}</td>
                    <td class="py-1.5 text-gray-400 text-xs align-top break-all">{{ input.note || '-' }}</td>
                  </tr>
                </tbody>
              </table>
            </div>
          </div>

          <div class="dialog-footer justify-between">
            <button @click="showConfigModal = false" class="btn !h-9 btn-outline">关闭</button>
            <button v-if="currentPlugin?.testable"
              @click="testPluginConfig" class="btn !h-9 btn-success flex items-center gap-2">
              <TestIcon class="w-4 h-4" />
              测试配置
            </button>
            <button @click="savePluginConfig" :disabled="!currentPlugin"
              class="btn !h-9 btn-primary disabled:opacity-50">保存配置</button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, ref, onMounted, watch } from 'vue'
import { getPluginList, pluginOp } from '@/api/admin'
import { ElMessage, ElMessageBox } from 'element-plus'
import SvgIcon from '@/components/svgicon.vue'
import { Beaker } from 'lucide-vue-next'

const plugins = ref<any[]>([])
const showConfigModal = ref(false)
const currentPlugin = ref<any>(null)
const pluginForm = ref<Record<string, string>>({})
const extraConfig = ref<Record<string, any>>({})
const refreshing = ref(false)
const TestIcon = Beaker

const currentPluginInputs = computed(() => {
  const inputs = currentPlugin.value?.inputs || {}
  return Object.keys(inputs).map((key) => ({ key, input: inputs[key] }))
})

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
  if (refreshing.value) return
  refreshing.value = true
  try {
    const res = await pluginOp({ action: 'refresh' })
    ElMessage.success(res.msg || '刷新成功')
    await fetchPlugins()
  } catch (error) {
    console.error('刷新失败:', error)
  } finally {
    refreshing.value = false
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
  const parsed = parseConfigJSON(p.config)
  const inputs = p?.inputs || {}
  const form: Record<string, string> = {}
  for (const key of Object.keys(inputs)) {
    const val = parsed[key]
    form[key] = val === undefined || val === null ? '' : String(val)
  }
  pluginForm.value = form

  const extras: Record<string, any> = {}
  for (const [k, v] of Object.entries(parsed)) {
    if (!inputs[k]) {
      extras[k] = v
    }
  }
  extraConfig.value = extras
  showConfigModal.value = true
}

function parseConfigJSON(raw: string): Record<string, any> {
  const text = (raw || '').trim()
  if (!text) return {}
  try {
    const data = JSON.parse(text)
    if (data && typeof data === 'object' && !Array.isArray(data)) {
      return data
    }
    return {}
  } catch {
    return {}
  }
}

function buildPluginConfigString(): string {
  const nextConfig: Record<string, any> = { ...extraConfig.value }
  for (const [k, v] of Object.entries(pluginForm.value)) {
    if (v === undefined || v === null) continue
    const text = String(v)
    if (text.trim() === '') continue
    nextConfig[k] = text
  }
  return JSON.stringify(nextConfig)
}

function isSensitiveField(key: string): boolean {
  const k = String(key || '').toLowerCase()
  if (!k) return false
  return (
    k.includes('key') ||
    k.includes('secret') ||
    k.includes('pwd') ||
    k.includes('password') ||
    k.includes('token')
  )
}

function decodePemToDer(pem: string): Uint8Array | null {
  let text = String(pem || '').trim()
  if (!text) return null
  if ((text.startsWith('"') && text.endsWith('"')) || (text.startsWith("'") && text.endsWith("'"))) {
    text = text.slice(1, -1)
  }
  text = text.replace(/\\r\\n/g, '\n').replace(/\\n/g, '\n').trim()

  const match = text.match(/-----BEGIN CERTIFICATE-----([\s\S]*?)-----END CERTIFICATE-----/)
  const base64Body = (match ? match[1] : text).replace(/[^A-Za-z0-9+/=]/g, '')
  if (!base64Body) return null

  try {
    const binary = atob(base64Body)
    const bytes = new Uint8Array(binary.length)
    for (let i = 0; i < binary.length; i++) {
      bytes[i] = binary.charCodeAt(i)
    }
    return bytes
  } catch {
    return null
  }
}

function readAsn1Length(bytes: Uint8Array, offset: number): { length: number; next: number } | null {
  if (offset >= bytes.length) return null
  const first = bytes[offset]
  if ((first & 0x80) === 0) {
    return { length: first, next: offset + 1 }
  }
  const n = first & 0x7f
  if (n <= 0 || n > 4 || offset+1+n > bytes.length) return null
  let len = 0
  for (let i = 0; i < n; i++) {
    len = (len << 8) | bytes[offset + 1 + i]
  }
  return { length: len, next: offset + 1 + n }
}

function readAsn1Tlv(bytes: Uint8Array, offset: number): { tag: number; valueStart: number; valueEnd: number; next: number } | null {
  if (offset + 2 > bytes.length) return null
  const tag = bytes[offset]
  const lenInfo = readAsn1Length(bytes, offset + 1)
  if (!lenInfo) return null
  const valueStart = lenInfo.next
  const valueEnd = valueStart + lenInfo.length
  if (valueEnd > bytes.length) return null
  return { tag, valueStart, valueEnd, next: valueEnd }
}

function bytesToHexUpper(bytes: Uint8Array): string {
  let out = ''
  for (let i = 0; i < bytes.length; i++) {
    out += bytes[i].toString(16).padStart(2, '0')
  }
  return out.toUpperCase()
}

function extractSerialFromMerchantCert(pem: string): string | null {
  const der = decodePemToDer(pem)
  if (!der) return null

  // Certificate ::= SEQUENCE { tbsCertificate, signatureAlgorithm, signatureValue }
  const certSeq = readAsn1Tlv(der, 0)
  if (!certSeq || certSeq.tag !== 0x30) return null

  // tbsCertificate ::= SEQUENCE { [0] version OPTIONAL, serialNumber INTEGER, ... }
  const tbs = readAsn1Tlv(der, certSeq.valueStart)
  if (!tbs || tbs.tag !== 0x30) return null

  let cursor = tbs.valueStart
  const first = readAsn1Tlv(der, cursor)
  if (!first) return null

  // 可选 version [0] EXPLICIT
  if (first.tag === 0xa0) {
    cursor = first.next
  }

  const serialTlv = readAsn1Tlv(der, cursor)
  if (!serialTlv || serialTlv.tag !== 0x02) return null

  let serialBytes = der.slice(serialTlv.valueStart, serialTlv.valueEnd)
  // ASN.1 INTEGER 可能有符号扩展前导 00
  while (serialBytes.length > 1 && serialBytes[0] === 0x00) {
    serialBytes = serialBytes.slice(1)
  }
  return bytesToHexUpper(serialBytes)
}

function getCertInputText(): string {
  const merchantCert = String(pluginForm.value.merchant_cert || '').trim()
  if (merchantCert) return merchantCert
  const platformCert = String(pluginForm.value.platform_cert || '').trim()
  if (platformCert) return platformCert
  const certPath = String(pluginForm.value.cert_path || '').trim()
  if (certPath.includes('BEGIN CERTIFICATE')) return certPath
  return ''
}

async function testPluginConfig() {
  if (!currentPlugin.value) return
  try {
    const config = buildPluginConfigString()
    const res = await pluginOp({
      action: 'test_config',
      name: currentPlugin.value.name,
      config
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
    await ElMessageBox.confirm(
      `即将保存插件【${currentPlugin.value.showname || currentPlugin.value.name}】配置，确认继续？`,
      '确认保存',
      {
        type: 'warning',
        confirmButtonText: '确认保存',
        cancelButtonText: '取消'
      }
    )
    const config = buildPluginConfigString()
    const res = await pluginOp({
      action: 'save_config',
      name: currentPlugin.value.name,
      config
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

watch(
  () => [currentPlugin.value?.name, pluginForm.value.merchant_cert, pluginForm.value.platform_cert, pluginForm.value.cert_path],
  ([pluginName]) => {
    if (pluginName !== 'wxpay') return
    const certText = getCertInputText()
    const serial = extractSerialFromMerchantCert(certText)
    if (serial) {
      pluginForm.value.serial_no = serial
      return
    }
    if (!certText) {
      pluginForm.value.serial_no = ''
    }
  }
  , { immediate: true }
)
</script>
