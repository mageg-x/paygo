<template>
  <div class="space-y-4">
    <!-- 页面标题 -->
    <div class="flex items-center justify-between">
      <div>
        <h1 class="text-2xl font-bold text-gray-900">支付类型管理</h1>
        <p class="text-sm text-gray-500 mt-1">配置支付方式名称、图标和状态</p>
      </div>
      <button @click="showAddModal"
        class="px-5 py-2.5 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors text-sm font-medium flex items-center gap-2">
        <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
        </svg>
        添加类型
      </button>
    </div>

    <!-- 类型列表 -->
    <div class="bg-white rounded-xl border border-gray-100 shadow-sm overflow-hidden">
      <div class="overflow-x-auto">
        <table class="w-full text-sm">
          <thead>
            <tr class="bg-gray-50 border-b border-gray-100">
              <th class="px-4 py-3 text-left font-semibold text-gray-600">ID</th>
              <th class="px-4 py-3 text-left font-semibold text-gray-600">标识</th>
              <th class="px-4 py-3 text-left font-semibold text-gray-600">显示名称</th>
              <th class="px-4 py-3 text-left font-semibold text-gray-600">设备类型</th>
              <th class="px-4 py-3 text-center font-semibold text-gray-600">状态</th>
              <th class="px-4 py-3 text-center font-semibold text-gray-600">操作</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-gray-50">
            <tr v-for="pt in payTypes" :key="pt.id" class="hover:bg-gray-50/50 transition-colors">
              <td class="px-4 py-3 text-gray-900 font-medium">{{ pt.id }}</td>
              <td class="px-4 py-3 text-gray-900">
                <div class="flex items-center gap-1.5">
                  <SvgIcon :name="getIconName(pt.name)" :size="16" />
                  <span class="font-medium">{{ pt.name }}</span>
                </div>
              </td>
              <td class="px-4 py-3 text-gray-600">{{ pt.showname || '-' }}</td>
              <td class="px-4 py-3 text-gray-500">{{ deviceName(pt.device) }}</td>
              <td class="px-4 py-3 text-center">
                <span
                  :class="['inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium', pt.status ? 'bg-green-100 text-green-700' : 'bg-gray-100 text-gray-600']">
                  {{ pt.status ? '开启' : '关闭' }}
                </span>
              </td>
              <td class="px-4 py-3 text-center">
                <div class="inline-flex items-center gap-1">
                  <button @click="showEditModal(pt)"
                    class="px-3 py-1 text-xs text-blue-600 hover:bg-blue-50 rounded transition-colors">编辑</button>
                  <button @click="toggleStatus(pt)"
                    :class="['px-3 py-1 text-xs rounded transition-colors', pt.status ? 'text-yellow-600 hover:bg-yellow-50' : 'text-green-600 hover:bg-green-50']">
                    {{ pt.status ? '关闭' : '开启' }}
                  </button>
                  <button @click="handleDelete(pt.id)"
                    class="px-3 py-1 text-xs text-red-600 hover:bg-red-50 rounded transition-colors">删除</button>
                </div>
              </td>
            </tr>
            <tr v-if="payTypes.length === 0">
              <td colspan="6" class="px-4 py-12 text-center text-gray-400">
                <div class="flex flex-col items-center">
                  <svg class="w-12 h-12 text-gray-300 mb-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5"
                      d="M3 10h18M7 15h1m4 0h1m-7 4h12a3 3 0 003-3V8a3 3 0 00-3-3H6a3 3 0 00-3 3v8a3 3 0 003 3z" />
                  </svg>
                  <span>暂无支付类型配置</span>
                </div>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

    <!-- 添加/编辑弹窗 -->
    <div v-if="showModal" class="fixed inset-0 z-50 overflow-y-auto">
      <div class="flex min-h-full items-center justify-center p-4">
        <div class="fixed inset-0 bg-black/50" @click="showModal = false"></div>
        <div class="relative bg-white rounded-xl shadow-xl w-full max-w-md p-6">
          <h3 class="text-lg font-semibold text-gray-900 mb-6">{{ isEdit ? '编辑支付类型' : '添加支付类型' }}</h3>
          <div class="space-y-4">
            <div>
              <label class="block text-sm font-medium text-gray-700 mb-1">标识名称</label>
              <input v-model="form.name" type="text" placeholder="如：alipay, wechatpay"
                class="w-full px-3 py-2 border border-gray-200 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-500" />
            </div>
            <div>
              <label class="block text-sm font-medium text-gray-700 mb-1">显示名称</label>
              <input v-model="form.showname" type="text" placeholder="如：支付宝、微信支付"
                class="w-full px-3 py-2 border border-gray-200 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-500" />
            </div>
            <div>
              <label class="block text-sm font-medium text-gray-700 mb-1">设备类型</label>
              <select v-model="form.device"
                class="w-full px-3 py-2 border border-gray-200 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-500">
                <option :value="0">PC</option>
                <option :value="1">手机H5</option>
                <option :value="2">APP</option>
                <option :value="3">全端</option>
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
import { getPayTypeList, payTypeOp } from '@/api/admin'
import { ElMessage, ElMessageBox } from 'element-plus'
import SvgIcon from '@/components/svgicon.vue'

const payTypes = ref<any[]>([])
const showModal = ref(false)
const isEdit = ref(false)
const form = ref({
  id: 0,
  name: '',
  device: 0,
  showname: '',
  status: 1
})

async function fetchPayTypes() {
  try {
    const res = await getPayTypeList()
    if (res.code === 0) {
      payTypes.value = res.data || []
    }
  } catch (error) {
    console.error('获取支付类型失败:', error)
  }
}

function showAddModal() {
  isEdit.value = false
  form.value = {
    id: 0,
    name: '',
    device: 0,
    showname: '',
    status: 1
  }
  showModal.value = true
}

function showEditModal(pt: any) {
  isEdit.value = true
  form.value = {
    id: pt.id,
    name: pt.name,
    device: pt.device,
    showname: pt.showname || '',
    status: pt.status
  }
  showModal.value = true
}

async function handleSave() {
  if (!form.value.name) {
    ElMessage.warning('请输入标识名称')
    return
  }
  try {
    const res = await payTypeOp({
      action: isEdit.value ? 'edit' : 'add',
      ...form.value
    })
    ElMessage.success(res.msg || '保存成功')
    showModal.value = false
    fetchPayTypes()
  } catch (error) {
    console.error('保存失败:', error)
  }
}

async function handleDelete(id: number) {
  try {
    await ElMessageBox.confirm('确定要删除这个支付类型吗？', '提示', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    })
  } catch {
    return
  }
  try {
    const res = await payTypeOp({ action: 'delete', id })
    ElMessage.success(res.msg || '删除成功')
    fetchPayTypes()
  } catch (error) {
    console.error('删除失败:', error)
  }
}

async function toggleStatus(pt: any) {
  try {
    const res = await payTypeOp({
      action: 'set_status',
      id: pt.id,
      status: pt.status ? 0 : 1
    })
    ElMessage.success(res.msg || '状态已更新')
    fetchPayTypes()
  } catch (error) {
    console.error('更新状态失败:', error)
  }
}

function getIconName(name: string): string {
  const map: Record<string, string> = {
    alipay: 'alipay',
    wechatpay: 'wechatpay',
    qqpay: 'qqpay',
    unionpay: 'unionpay',
    jdpay: 'jdpay'
  }
  return map[name] || 'creditcard'
}

function deviceName(device: number): string {
  const map: Record<number, string> = {
    0: 'PC',
    1: '手机H5',
    2: 'APP',
    3: '全端'
  }
  return map[device] || '未知'
}

onMounted(() => {
  fetchPayTypes()
})
</script>
