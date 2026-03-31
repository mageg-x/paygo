<template>
  <div class="space-y-4">
    <!-- 页面标题 -->
    <div class="flex items-center justify-between">
      <div>
        <h1 class="text-2xl font-bold text-gray-900">通道管理</h1>
        <p class="text-sm text-gray-500 mt-1">配置支付通道和费率</p>
      </div>
      <button @click="showAddModal"
        class="px-5 py-2.5 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors text-sm font-medium flex items-center gap-2">
        <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
        </svg>
        添加通道
      </button>
    </div>

    <!-- 通道列表 -->
    <div class="bg-white rounded-xl border border-gray-100 shadow-sm overflow-hidden">
      <div class="overflow-x-auto">
        <table class="w-full text-sm">
          <thead>
            <tr class="bg-gray-50 border-b border-gray-100">
              <th class="px-4 py-3 text-left font-semibold text-gray-600">ID</th>
              <th class="px-4 py-3 text-left font-semibold text-gray-600">通道名称</th>
              <th class="px-4 py-3 text-left font-semibold text-gray-600">插件</th>
              <th class="px-4 py-3 text-left font-semibold text-gray-600">支付类型</th>
              <th class="px-4 py-3 text-right font-semibold text-gray-600">费率</th>
              <th class="px-4 py-3 text-right font-semibold text-gray-600">成本</th>
              <th class="px-4 py-3 text-center font-semibold text-gray-600">限额</th>
              <th class="px-4 py-3 text-center font-semibold text-gray-600">状态</th>
              <th class="px-4 py-3 text-center font-semibold text-gray-600">操作</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-gray-50">
            <tr v-for="ch in channels" :key="ch.id" class="hover:bg-gray-50/50 transition-colors">
              <td class="px-4 py-3 text-gray-900 font-medium">{{ ch.id }}</td>
              <td class="px-4 py-3 text-gray-900">{{ ch.name }}</td>
              <td class="px-4 py-3 text-gray-600">{{ ch.plugin_showname || ch.plugin }}</td>
              <td class="px-4 py-3">
                <span class="inline-flex items-center px-2 py-0.5 rounded text-xs font-medium bg-blue-50 text-blue-700">
                  {{ typeName(ch.type) }}
                </span>
              </td>
              <td class="px-4 py-3 text-right">
                <span class="font-semibold text-green-600">{{ ch.rate }}%</span>
              </td>
              <td class="px-4 py-3 text-right text-gray-500">{{ ch.costrate }}%</td>
              <td class="px-4 py-3 text-center text-gray-500 text-xs">
                <div>￥{{ ch.paymin }} - ￥{{ ch.paymax }}</div>
                <div class="text-gray-400">日限 ￥{{ ch.daytop }}</div>
              </td>
              <td class="px-4 py-3 text-center">
                <span
                  :class="['inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium', ch.status ? 'bg-green-100 text-green-700' : 'bg-gray-100 text-gray-600']">
                  {{ ch.status ? '开启' : '关闭' }}
                </span>
              </td>
              <td class="px-4 py-3 text-center">
                <div class="inline-flex items-center gap-1">
                  <button @click="showEditModal(ch)"
                    class="px-3 py-1 text-xs text-blue-600 hover:bg-blue-50 rounded transition-colors">编辑</button>
                  <button @click="handleDelete(ch.id)"
                    class="px-3 py-1 text-xs text-red-600 hover:bg-red-50 rounded transition-colors">删除</button>
                  <button @click="toggleStatus(ch)"
                    :class="['px-3 py-1 text-xs rounded transition-colors', ch.status ? 'text-yellow-600 hover:bg-yellow-50' : 'text-green-600 hover:bg-green-50']">
                    {{ ch.status ? '关闭' : '开启' }}
                  </button>
                </div>
              </td>
            </tr>
            <tr v-if="channels.length === 0">
              <td colspan="9" class="px-4 py-12 text-center text-gray-400">
                <div class="flex flex-col items-center">
                  <svg class="w-12 h-12 text-gray-300 mb-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5"
                      d="M13.828 10.172a4 4 0 00-5.656 0l-4 4a4 4 0 105.656 5.656l1.102-1.101m-.758-4.899a4 4 0 005.656 0l4-4a4 4 0 00-5.656-5.656l-1.1 1.1" />
                  </svg>
                  <span>暂无通道配置</span>
                </div>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

    <!-- 添加/编辑通道弹窗 -->
    <div v-if="showModal" class="fixed inset-0 z-50 overflow-y-auto">
      <div class="flex min-h-full items-center justify-center p-4">
        <div class="fixed inset-0 bg-black/50" @click="showModal = false"></div>
        <div class="relative bg-white rounded-xl shadow-xl w-full max-w-2xl p-6">
          <h3 class="text-lg font-semibold text-gray-900 mb-6">{{ isEdit ? '编辑通道' : '添加通道' }}</h3>
          <div class="grid grid-cols-2 gap-x-6 gap-y-4">
            <div class="col-span-2">
              <label class="block text-sm font-medium text-gray-700 mb-1">通道名称</label>
              <input v-model="form.name" type="text" placeholder="例如：支付宝通道A"
                class="w-full px-3 py-2 border border-gray-200 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-500" />
            </div>
            <div>
              <label class="block text-sm font-medium text-gray-700 mb-1">支付类型</label>
              <select v-model="form.type"
                class="w-full px-3 py-2 border border-gray-200 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-500">
                <option :value="1">支付宝</option>
                <option :value="2">微信支付</option>
              </select>
            </div>
            <div>
              <label class="block text-sm font-medium text-gray-700 mb-1">插件</label>
              <select v-model="form.plugin"
                class="w-full px-3 py-2 border border-gray-200 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-500">
                <option value="">请选择插件</option>
                <option v-for="p in plugins" :key="p.name" :value="p.name">{{ p.showname || p.name }}</option>
              </select>
            </div>
            <div>
              <label class="block text-sm font-medium text-gray-700 mb-1">通道模式</label>
              <select v-model="form.mode"
                class="w-full px-3 py-2 border border-gray-200 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-500">
                <option :value="0">平台代收</option>
                <option :value="1">商户直清</option>
              </select>
            </div>
            <div>
              <label class="block text-sm font-medium text-gray-700 mb-1">状态</label>
              <select v-model="form.status"
                class="w-full px-3 py-2 border border-gray-200 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-500">
                <option :value="1">开启</option>
                <option :value="0">关闭</option>
              </select>
            </div>
            <div>
              <label class="block text-sm font-medium text-gray-700 mb-1">分成比例 (%)</label>
              <input v-model.number="form.rate" type="number" step="0.01"
                class="w-full px-3 py-2 border border-gray-200 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-500" />
            </div>
            <div>
              <label class="block text-sm font-medium text-gray-700 mb-1">成本费率 (%)</label>
              <input v-model.number="form.costrate" type="number" step="0.01"
                class="w-full px-3 py-2 border border-gray-200 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-500" />
            </div>
            <div>
              <label class="block text-sm font-medium text-gray-700 mb-1">单笔最小 (元)</label>
              <input v-model="form.paymin" type="number"
                class="w-full px-3 py-2 border border-gray-200 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-500" />
            </div>
            <div>
              <label class="block text-sm font-medium text-gray-700 mb-1">单笔最大 (元)</label>
              <input v-model="form.paymax" type="number"
                class="w-full px-3 py-2 border border-gray-200 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-500" />
            </div>
            <div>
              <label class="block text-sm font-medium text-gray-700 mb-1">单日限额 (元)</label>
              <input v-model.number="form.daytop" type="number"
                class="w-full px-3 py-2 border border-gray-200 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-500" />
            </div>
          </div>
          <div class="flex justify-end gap-3 mt-8">
            <button @click="showModal = false"
              class="px-4 py-2 text-sm text-gray-600 hover:bg-gray-100 rounded-lg transition-colors">取消</button>
            <button @click="handleSave"
              class="px-4 py-2 text-sm bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors">保存</button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { getChannelList, channelOp, getPluginList } from '@/api/admin'
import { ElMessage } from 'element-plus'

const channels = ref<any[]>([])
const plugins = ref<any[]>([])
const showModal = ref(false)
const isEdit = ref(false)
const form = ref({
  id: 0,
  name: '',
  plugin: '',
  type: 1,
  mode: 0,
  rate: 0.5,
  costrate: 0.3,
  daytop: 100000,
  paymin: 10,
  paymax: 5000,
  apptype: '',
  status: 1
})

async function fetchChannels() {
  try {
    const res = await getChannelList()
    if (res.code === 0) {
      channels.value = res.data || []
    }
  } catch (error) {
    console.error('获取通道列表失败:', error)
  }
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

function showAddModal() {
  isEdit.value = false
  form.value = {
    id: 0,
    name: '',
    plugin: '',
    type: 1,
    mode: 0,
    rate: 0.5,
    costrate: 0.3,
    daytop: 100000,
    paymin: 10,
    paymax: 5000,
    apptype: '',
    status: 1
  }
  showModal.value = true
}

function showEditModal(ch: any) {
  isEdit.value = true
  form.value = {
    id: ch.id,
    name: ch.name,
    plugin: ch.plugin,
    type: ch.type,
    mode: ch.mode,
    rate: ch.rate,
    costrate: ch.costrate,
    daytop: ch.daytop,
    paymin: ch.paymin,
    paymax: ch.paymax,
    apptype: ch.apptype || '',
    status: ch.status
  }
  showModal.value = true
}

async function handleSave() {
  if (!form.value.name) {
    ElMessage.warning('请输入通道名称')
    return
  }
  if (!form.value.plugin) {
    ElMessage.warning('请选择插件')
    return
  }
  try {
    const res = await channelOp({
      action: isEdit.value ? 'edit' : 'add',
      ...form.value
    })
    ElMessage.success(res.msg || '保存成功')
    showModal.value = false
    fetchChannels()
  } catch (error) {
    console.error('保存失败:', error)
  }
}

async function handleDelete(id: number) {
  if (!confirm('确定要删除这个通道吗？')) return
  try {
    const res = await channelOp({ action: 'delete', id })
    ElMessage.success(res.msg || '删除成功')
    fetchChannels()
  } catch (error) {
    console.error('删除失败:', error)
  }
}

async function toggleStatus(ch: any) {
  try {
    const res = await channelOp({
      action: 'set_status',
      id: ch.id,
      status: ch.status ? 0 : 1
    })
    ElMessage.success(res.msg || '状态已更新')
    fetchChannels()
  } catch (error) {
    console.error('更新状态失败:', error)
  }
}

function typeName(type: number) {
  const map: Record<number, string> = {
    1: '支付宝',
    2: '微信支付',
  }
  return map[type] || '未知'
}

onMounted(() => {
  fetchChannels()
  fetchPlugins()
})
</script>
