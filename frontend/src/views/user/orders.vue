<template>
  <div>
    <h2 class="text-2xl font-bold text-gray-800 mb-6">我的订单</h2>

    <!-- 筛选 -->
    <div class="card mb-4">
      <div class="card-body flex items-center gap-4">
        <select v-model="filterStatus"
          class="px-4 py-2 border border-gray-200 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-500">
          <option value="">全部状态</option>
          <option value="0">待支付</option>
          <option value="1">已支付</option>
          <option value="2">已退款</option>
          <option value="3">已冻结</option>
        </select>
        <input v-model="searchTradeNo" type="text" placeholder="订单号"
          class="px-4 py-2 border border-gray-200 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-500" />
        <button @click="page = 1; fetchOrders()"
          class="px-4 py-2 bg-blue-600 text-white rounded-lg text-sm hover:bg-blue-700">
          搜索
        </button>
      </div>
    </div>

    <div class="card">
      <div class="card-body">
        <table class="table whitespace-nowrap">
          <thead>
            <tr>
              <th>订单号</th>
              <th>商品名称</th>
              <th>支付方式</th>
              <th>金额</th>
              <th>状态</th>
              <th>时间</th>
              <th>操作</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="order in orders" :key="order.trade_no">
              <td class="text-xs font-mono">{{ order.trade_no }}</td>
              <td>{{ order.name }}</td>
              <td>
                <div class="flex items-center gap-1.5">
                  <span class="text-lg">{{ order.type === 1 ? '💙' : order.type === 2 ? '🟢' : '💜' }}</span>
                  <span class="text-sm font-medium">{{ typeName(order.type) }}</span>
                </div>
              </td>
              <td class="font-semibold text-emerald-600">¥{{ order.money }}</td>
              <td>
                <span v-if="order.status === 1"
                  class="inline-flex items-center px-2.5 py-1 rounded-full text-xs font-medium bg-emerald-100 text-emerald-700">
                  已支付
                </span>
                <span v-else-if="order.status === 0"
                  class="inline-flex items-center px-2.5 py-1 rounded-full text-xs font-medium bg-amber-100 text-amber-700">
                  待支付
                </span>
                <span v-else-if="order.status === 2"
                  class="inline-flex items-center px-2.5 py-1 rounded-full text-xs font-medium bg-blue-100 text-blue-700">
                  已退款
                </span>
                <span v-else-if="order.status === 3"
                  class="inline-flex items-center px-2.5 py-1 rounded-full text-xs font-medium bg-red-100 text-red-700">
                  已冻结
                </span>
                <span v-else
                  class="inline-flex items-center px-2.5 py-1 rounded-full text-xs font-medium bg-gray-100 text-gray-700">
                  未知
                </span>
              </td>
              <td>{{ dayjs(order.addtime).format('YYYY-MM-DD HH:mm') }}</td>
              <td>
                <button @click="showDetail(order)"
                  class="px-3 py-1 text-xs text-blue-600 hover:bg-blue-50 rounded transition-colors">
                  详情
                </button>
              </td>
            </tr>
            <tr v-if="orders.length === 0">
              <td colspan="7" class="text-center text-gray-500 py-8">暂无订单</td>
            </tr>
          </tbody>
        </table>

        <!-- 分页 -->
        <div class="flex items-center justify-between mt-4">
          <div class="text-sm text-gray-500">共 {{ total }} 条</div>
          <div class="flex items-center gap-2">
            <button class="pagination-item" :disabled="page === 1" @click="page--; fetchOrders()">上一页</button>
            <span class="px-4">{{ page }} / {{ totalPages }}</span>
            <button class="pagination-item" :disabled="page >= totalPages" @click="page++; fetchOrders()">下一页</button>
          </div>
        </div>
      </div>
    </div>

    <!-- 订单详情弹窗 -->
    <div v-if="detailVisible" class="fixed inset-0 z-50 overflow-y-auto">
      <div class="flex min-h-full items-center justify-center p-4">
        <div class="fixed inset-0 bg-black/50" @click="detailVisible = false"></div>
        <div class="relative bg-white rounded-xl shadow-xl w-full max-w-lg p-6">
          <h3 class="text-lg font-semibold text-gray-900 mb-4">订单详情</h3>

          <div v-if="currentOrder" class="space-y-3 text-sm">
            <div class="grid grid-cols-2 gap-2">
              <div class="text-gray-500">订单号:</div>
              <div class="font-mono text-gray-900">{{ currentOrder.trade_no }}</div>
              <div class="text-gray-500">商户订单号:</div>
              <div class="font-mono text-gray-900">{{ currentOrder.out_trade_no || '-' }}</div>
              <div class="text-gray-500">商品名称:</div>
              <div class="text-gray-900">{{ currentOrder.name }}</div>
              <div class="text-gray-500">支付方式:</div>
              <div class="text-gray-900">{{ typeName(currentOrder.type) }}</div>
              <div class="text-gray-500">订单金额:</div>
              <div class="font-bold text-emerald-600">¥{{ currentOrder.money }}</div>
              <div class="text-gray-500">实付金额:</div>
              <div class="font-bold text-emerald-600">¥{{ currentOrder.realmoney || currentOrder.money }}</div>
              <div class="text-gray-500">商户所得:</div>
              <div class="text-blue-600">¥{{ currentOrder.getmoney || '-' }}</div>
              <div class="text-gray-500">状态:</div>
              <div>
                <span v-if="currentOrder.status === 1" class="text-emerald-600">已支付</span>
                <span v-else-if="currentOrder.status === 0" class="text-amber-600">待支付</span>
                <span v-else-if="currentOrder.status === 2" class="text-blue-600">已退款</span>
                <span v-else-if="currentOrder.status === 3" class="text-red-600">已冻结</span>
                <span v-else class="text-gray-500">未知</span>
              </div>
              <div class="text-gray-500">创建时间:</div>
              <div>{{ dayjs(currentOrder.addtime).format('YYYY-MM-DD HH:mm:ss') }}</div>
              <div class="text-gray-500">支付时间:</div>
              <div>{{ currentOrder.endtime ? dayjs(currentOrder.endtime).format('YYYY-MM-DD HH:mm:ss') : '-' }}</div>
              <div class="text-gray-500">回调状态:</div>
              <div>
                <span v-if="currentOrder.notify === 1" class="text-emerald-600">已回调</span>
                <span v-else class="text-amber-600">未回调</span>
              </div>
              <div class="text-gray-500">订单类型:</div>
              <div>{{ currentOrder.isrecharge ? '余额充值' : '普通订单' }}</div>
            </div>

            <div v-if="currentOrder.param" class="border-t pt-3 mt-3">
              <div class="text-gray-500 mb-1">订单备注:</div>
              <div class="text-gray-900">{{ currentOrder.param }}</div>
            </div>
          </div>

          <div class="flex justify-end gap-3 mt-6">
            <button @click="handleNotify(currentOrder)" v-if="currentOrder?.status === 1"
              class="px-4 py-2 text-sm bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors">
              重新通知
            </button>
            <button @click="handleRefund(currentOrder)" v-if="currentOrder?.status === 1"
              class="px-4 py-2 text-sm bg-red-600 text-white rounded-lg hover:bg-red-700 transition-colors">
              退款
            </button>
            <button @click="detailVisible = false"
              class="px-4 py-2 text-sm text-gray-600 hover:bg-gray-100 rounded-lg transition-colors">
              关闭
            </button>
          </div>
        </div>
      </div>
    </div>

    <!-- 退款弹窗 -->
    <div v-if="refundVisible" class="fixed inset-0 z-50 overflow-y-auto">
      <div class="flex min-h-full items-center justify-center p-4">
        <div class="fixed inset-0 bg-black/50" @click="refundVisible = false"></div>
        <div class="relative bg-white rounded-xl shadow-xl w-full max-w-md p-6">
          <h3 class="text-lg font-semibold text-gray-900 mb-4">订单退款</h3>
          <div class="space-y-4">
            <div class="bg-amber-50 rounded-lg p-4">
              <p class="text-amber-800 text-sm">订单号: {{ refundForm.trade_no }}</p>
              <p class="text-amber-800 text-sm mt-1">订单金额: ¥{{ refundForm.money }}</p>
            </div>
            <div>
              <label class="block text-sm font-medium text-gray-700 mb-1">退款金额</label>
              <input v-model="refundForm.amount" type="number" step="0.01"
                class="w-full px-3 py-2 border border-gray-200 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
                placeholder="请输入退款金额" />
            </div>
          </div>
          <div class="flex justify-end gap-3 mt-6">
            <button @click="refundVisible = false"
              class="px-4 py-2 text-sm text-gray-600 hover:bg-gray-100 rounded-lg transition-colors">取消</button>
            <button @click="submitRefund"
              class="px-4 py-2 text-sm bg-red-600 text-white rounded-lg hover:bg-red-700 transition-colors">确认退款</button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { getUserInfo, getUserOrders, userOrderOp } from '@/api/user'
import { ElMessage, ElMessageBox } from 'element-plus'
import dayjs from 'dayjs'
import { useAppStore } from '@/stores/app'

const appStore = useAppStore()

const orders = ref<any[]>([])
const page = ref(1)
const pageSize = ref(20)
const total = ref(0)
const filterStatus = ref('')
const searchTradeNo = ref('')
const detailVisible = ref(false)
const refundVisible = ref(false)
const currentOrder = ref<any>(null)

const refundForm = ref({
  trade_no: '',
  money: 0,
  amount: ''
})

const totalPages = computed(() => Math.ceil(total.value / pageSize.value) || 1)

function typeName(type: number) {
  const map: Record<number, string> = {
    1: '支付宝',
    2: '微信支付',
    3: 'QQ钱包',
    4: '银行卡'
  }
  return map[type] || '其他'
}

async function fetchOrders() {
  try {
    const params: any = { page: page.value, limit: pageSize.value }
    if (filterStatus.value !== '') {
      params.status = filterStatus.value
    }
    if (searchTradeNo.value) {
      params.trade_no = searchTradeNo.value.trim()
    }
    const res = await getUserOrders(params)
    orders.value = Array.isArray(res.data) ? res.data : []
    total.value = res.count || 0
  } catch (error) {
    console.error('获取订单列表失败:', error)
    ElMessage.error((error as any)?.message || '获取订单列表失败')
  }
}

async function initUser() {
  const res = await getUserInfo()
  if (!res.data?.uid) {
    throw new Error('未获取到当前商户信息')
  }
  appStore.userInfo = {
    ...(appStore.userInfo || {
      uid: Number(res.data.uid),
      username: '',
      email: '',
      phone: '',
      money: 0,
      status: 1
    }),
    uid: Number(res.data.uid)
  }
}

function showDetail(order: any) {
  currentOrder.value = order
  detailVisible.value = true
}

async function handleNotify(order: any) {
  try {
    await ElMessageBox.confirm('确定要重新通知该订单吗？', '提示', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    })
  } catch {
    return
  }

  try {
    const res = await userOrderOp({ action: 'notify', trade_no: order.trade_no })
    ElMessage.success(res.msg || '已触发重新通知')
    fetchOrders()
  } catch (error) {
    console.error('重新通知失败:', error)
  }
}

function handleRefund(order: any) {
  const remain = Math.max(0, Number(order.money || 0) - Number(order.refundmoney || 0))
  if (remain <= 0) {
    ElMessage.warning('可退款金额为0')
    return
  }
  refundForm.value = {
    trade_no: order.trade_no,
    money: order.money,
    amount: remain.toFixed(2)
  }
  detailVisible.value = false
  refundVisible.value = true
}

async function submitRefund() {
  const amount = parseFloat(refundForm.value.amount)
  const remain = Math.max(0, Number(refundForm.value.money || 0) - Number(currentOrder.value?.refundmoney || 0))
  if (isNaN(amount) || amount <= 0) {
    ElMessage.warning('请输入有效的退款金额')
    return
  }
  if (amount > remain) {
    ElMessage.warning(`退款金额不能超过可退金额 ${remain.toFixed(2)} 元`)
    return
  }
  try {
    await ElMessageBox.confirm(`确定要退款 ¥${amount} 吗？`, '退款确认', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    })
    const res = await userOrderOp({
      action: 'refund',
      trade_no: refundForm.value.trade_no,
      money: amount
    })
    ElMessage.success(res.msg || '退款成功')
    refundVisible.value = false
    fetchOrders()
  } catch {
    return
  }
}

onMounted(async () => {
  try {
    await initUser()
    await fetchOrders()
  } catch (error: any) {
    ElMessage.error(error?.message || '初始化失败')
  }
})
</script>
