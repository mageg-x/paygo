<template>
  <div class="space-y-4">
    <!-- 页面标题 -->
    <div class="flex items-center justify-between">
      <div>
        <h1 class="text-2xl font-bold text-gray-900">分账管理</h1>
        <p class="text-sm text-gray-500 mt-1">分账订单管理和接收方配置</p>
      </div>
      <div class="flex gap-2">
        <button @click="showReceiverModal = true"
          class="px-5 py-2.5 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors text-sm font-medium flex items-center gap-2">
          <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
          </svg>
          添加接收方
        </button>
      </div>
    </div>

    <!-- 标签页 -->
    <div class="bg-white rounded-xl border border-gray-100 shadow-sm">
      <div class="border-b border-gray-100">
        <nav class="flex -mb-px">
          <button v-for="tab in tabs" :key="tab.key" @click="activeTab = tab.key"
            :class="['px-6 py-3 text-sm font-medium border-b-2 transition-colors',
              activeTab === tab.key
                ? 'border-blue-500 text-blue-600'
                : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300']">
            {{ tab.name }}
          </button>
        </nav>
      </div>

      <!-- 分账订单 -->
      <div v-if="activeTab === 'orders'" class="p-6">
        <div class="overflow-x-auto">
          <table class="w-full text-sm whitespace-nowrap">
            <thead>
              <tr class="bg-gray-50 border-b border-gray-100">
                <th class="px-4 py-3 text-left font-semibold text-gray-600">订单号</th>
                <th class="px-4 py-3 text-left font-semibold text-gray-600">原订单</th>
                <th class="px-4 py-3 text-right font-semibold text-gray-600">订单金额</th>
                <th class="px-4 py-3 text-right font-semibold text-gray-600">分账金额</th>
                <th class="px-4 py-3 text-center font-semibold text-gray-600">状态</th>
                <th class="px-4 py-3 text-left font-semibold text-gray-600">创建时间</th>
                <th class="px-4 py-3 text-center font-semibold text-gray-600">操作</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-gray-50">
              <tr v-for="order in psOrders" :key="order.id" class="hover:bg-gray-50/50 transition-colors">
                <td class="px-4 py-3 text-gray-900 font-mono text-xs">{{ order.ps_no }}</td>
                <td class="px-4 py-3 text-gray-500 font-mono text-xs">{{ order.trade_no }}</td>
                <td class="px-4 py-3 text-right font-semibold text-gray-900">￥{{ order.money }}</td>
                <td class="px-4 py-3 text-right font-semibold text-green-600">￥{{ order.ps_money }}</td>
                <td class="px-4 py-3 text-center">
                  <span :class="['inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium',
                    order.status === 1 ? 'bg-green-100 text-green-700' :
                    order.status === 2 ? 'bg-yellow-100 text-yellow-700' :
                    order.status === 3 ? 'bg-blue-100 text-blue-700' :
                    'bg-gray-100 text-gray-600']">
                    {{ statusName(order.status) }}
                  </span>
                </td>
                <td class="px-4 py-3 text-gray-500 text-xs">{{ order.created_at }}</td>
                <td class="px-4 py-3 text-center">
                  <button v-if="order.status === 1" @click="handleProfit(order)"
                    class="px-3 py-1 text-xs text-blue-600 hover:bg-blue-50 rounded transition-colors">
                    发起分账
                  </button>
                  <button v-else-if="order.status === 3" @click="viewProfitDetail(order)"
                    class="px-3 py-1 text-xs text-gray-600 hover:bg-gray-50 rounded transition-colors">
                    查看详情
                  </button>
                </td>
              </tr>
              <tr v-if="psOrders.length === 0">
                <td colspan="7" class="px-4 py-12 text-center text-gray-400">暂无分账订单</td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>

      <!-- 接收方管理 -->
      <div v-if="activeTab === 'receivers'" class="p-6">
        <div class="overflow-x-auto">
          <table class="w-full text-sm whitespace-nowrap">
            <thead>
              <tr class="bg-gray-50 border-b border-gray-100">
                <th class="px-4 py-3 text-left font-semibold text-gray-600">ID</th>
                <th class="px-4 py-3 text-left font-semibold text-gray-600">接收方名称</th>
                <th class="px-4 py-3 text-left font-semibold text-gray-600">账号</th>
                <th class="px-4 py-3 text-left font-semibold text-gray-600">类型</th>
                <th class="px-4 py-3 text-right font-semibold text-gray-600">分账比例</th>
                <th class="px-4 py-3 text-center font-semibold text-gray-600">状态</th>
                <th class="px-4 py-3 text-center font-semibold text-gray-600">操作</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-gray-50">
              <tr v-for="r in receivers" :key="r.id" class="hover:bg-gray-50/50 transition-colors">
                <td class="px-4 py-3 text-gray-900 font-medium">{{ r.id }}</td>
                <td class="px-4 py-3 text-gray-900">{{ r.name }}</td>
                <td class="px-4 py-3 text-gray-500 font-mono text-xs">{{ r.account }}</td>
                <td class="px-4 py-3 text-gray-500">{{ r.type === 1 ? '商户' : '个人' }}</td>
                <td class="px-4 py-3 text-right font-semibold text-green-600">{{ r.rate }}%</td>
                <td class="px-4 py-3 text-center">
                  <span
                    :class="['inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium', r.status ? 'bg-green-100 text-green-700' : 'bg-gray-100 text-gray-600']">
                    {{ r.status ? '开启' : '关闭' }}
                  </span>
                </td>
                <td class="px-4 py-3 text-center">
                  <button @click="editReceiver(r)"
                    class="px-3 py-1 text-xs text-blue-600 hover:bg-blue-50 rounded transition-colors">编辑</button>
                  <button @click="deleteReceiver(r.id)"
                    class="px-3 py-1 text-xs text-red-600 hover:bg-red-50 rounded transition-colors">删除</button>
                </td>
              </tr>
              <tr v-if="receivers.length === 0">
                <td colspan="7" class="px-4 py-12 text-center text-gray-400">暂无接收方配置</td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>
    </div>

    <!-- 添加/编辑接收方弹窗 -->
    <div v-if="showReceiverModal" class="fixed inset-0 z-50 overflow-y-auto">
      <div class="flex min-h-full items-center justify-center p-4">
        <div class="fixed inset-0 bg-black/50" @click="showReceiverModal = false"></div>
        <div class="relative bg-white rounded-xl shadow-xl w-full max-w-md p-6">
          <h3 class="text-lg font-semibold text-gray-900 mb-6">{{ isEditReceiver ? '编辑接收方' : '添加接收方' }}</h3>
          <div class="space-y-4">
            <div>
              <label class="block text-sm font-medium text-gray-700 mb-1">接收方名称</label>
              <input v-model="receiverForm.name" type="text" placeholder="如：商户A"
                class="w-full px-3 py-2 border border-gray-200 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-500" />
            </div>
            <div>
              <label class="block text-sm font-medium text-gray-700 mb-1">账号</label>
              <input v-model="receiverForm.account" type="text" placeholder="支付宝账号或商户ID"
                class="w-full px-3 py-2 border border-gray-200 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-500" />
            </div>
            <div>
              <label class="block text-sm font-medium text-gray-700 mb-1">类型</label>
              <select v-model="receiverForm.type"
                class="w-full px-3 py-2 border border-gray-200 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-500">
                <option :value="1">商户</option>
                <option :value="2">个人</option>
              </select>
            </div>
            <div>
              <label class="block text-sm font-medium text-gray-700 mb-1">分账比例 (%)</label>
              <input v-model.number="receiverForm.rate" type="number" step="0.01" max="100"
                class="w-full px-3 py-2 border border-gray-200 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-500" />
            </div>
            <div>
              <label class="block text-sm font-medium text-gray-700 mb-1">状态</label>
              <select v-model="receiverForm.status"
                class="w-full px-3 py-2 border border-gray-200 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-500">
                <option :value="1">开启</option>
                <option :value="0">关闭</option>
              </select>
            </div>
          </div>
          <div class="flex justify-end gap-3 mt-8">
            <button @click="showReceiverModal = false"
              class="px-4 py-2 text-sm text-gray-600 hover:bg-gray-100 rounded-lg transition-colors">取消</button>
            <button @click="saveReceiver"
              class="px-4 py-2 text-sm bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors">保存</button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { profitOrderList, profitReceiverList, profitReceiverOp, profitDo } from '@/api/admin'
import { ElMessage, ElMessageBox } from 'element-plus'

const tabs = [
  { key: 'orders', name: '分账订单' },
  { key: 'receivers', name: '接收方管理' }
]

const activeTab = ref('orders')
const psOrders = ref<any[]>([])
const receivers = ref<any[]>([])
const showReceiverModal = ref(false)
const isEditReceiver = ref(false)
const receiverForm = ref({
  id: 0,
  name: '',
  account: '',
  type: 1,
  rate: 0,
  status: 1
})

async function fetchPsOrders() {
  try {
    const res = await profitOrderList()
    if (res.code === 0) {
      psOrders.value = res.data || []
    }
  } catch (error) {
    console.error('获取分账订单失败:', error)
  }
}

async function fetchReceivers() {
  try {
    const res = await profitReceiverList()
    if (res.code === 0) {
      receivers.value = res.data || []
    }
  } catch (error) {
    console.error('获取接收方失败:', error)
  }
}

function editReceiver(r: any) {
  isEditReceiver.value = true
  receiverForm.value = {
    id: r.id,
    name: r.name,
    account: r.account,
    type: r.type,
    rate: r.rate,
    status: r.status
  }
  showReceiverModal.value = true
}

async function saveReceiver() {
  if (!receiverForm.value.name) {
    ElMessage.warning('请输入接收方名称')
    return
  }
  try {
    const res = await profitReceiverOp({
      action: isEditReceiver.value ? 'edit' : 'add',
      ...receiverForm.value
    })
    ElMessage.success(res.msg || '保存成功')
    showReceiverModal.value = false
    fetchReceivers()
  } catch (error) {
    console.error('保存失败:', error)
  }
}

async function deleteReceiver(id: number) {
  try {
    await ElMessageBox.confirm('确定要删除这个接收方吗？', '提示', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    })
  } catch {
    return
  }
  try {
    const res = await profitReceiverOp({ action: 'delete', id })
    ElMessage.success(res.msg || '删除成功')
    fetchReceivers()
  } catch (error) {
    console.error('删除失败:', error)
  }
}

async function handleProfit(order: any) {
  try {
    await ElMessageBox.confirm('确定要发起分账吗？', '提示', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    })
  } catch {
    return
  }
  try {
    const res = await profitDo({ ps_no: order.ps_no })
    if (res.code === 0) {
      ElMessage.success('分账请求已提交')
      fetchPsOrders()
    } else {
      ElMessage.error(res.msg || '分账失败')
    }
  } catch (error) {
    console.error('分账失败:', error)
  }
}

function viewProfitDetail(order: any) {
  ElMessage.info('分账详情: ' + order.ps_no)
}

function statusName(status: number): string {
  const map: Record<number, string> = {
    0: '待处理',
    1: '已提交',
    2: '分账中',
    3: '已完成'
  }
  return map[status] || '未知'
}

onMounted(() => {
  fetchPsOrders()
  fetchReceivers()
})
</script>
