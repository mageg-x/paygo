<template>
  <div class="space-y-4">
    <!-- 页面标题 -->
    <div class="flex items-center justify-between">
      <div>
        <h1 class="text-2xl font-bold text-gray-900">结算管理</h1>
        <p class="text-sm text-gray-500 mt-1">管理商户结算申请</p>
      </div>
      <select v-model="filterStatus" @change="page = 1; fetchSettles()"
        class="px-4 py-2 border border-gray-200 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-500">
        <option :value="-1">全部状态</option>
        <option :value="0">待处理</option>
        <option :value="1">已完成</option>
        <option :value="2">处理中</option>
        <option :value="3">已拒绝</option>
      </select>
    </div>

    <!-- 统计卡片 -->
    <div class="grid grid-cols-4 gap-4">
      <div class="bg-white rounded-xl p-4 border border-gray-100 shadow-sm border-l-4 border-l-slate-400">
        <div class="text-sm text-gray-500">全部申请</div>
        <div class="text-2xl font-bold text-slate-700 mt-1">{{ total }}</div>
      </div>
      <div class="bg-white rounded-xl p-4 border border-gray-100 shadow-sm border-l-4 border-l-amber-400">
        <div class="text-sm text-gray-500">待处理</div>
        <div class="text-2xl font-bold text-amber-600 mt-1">{{ statusCount(0) }}</div>
      </div>
      <div class="bg-white rounded-xl p-4 border border-gray-100 shadow-sm border-l-4 border-l-emerald-400">
        <div class="text-sm text-gray-500">已完成</div>
        <div class="text-2xl font-bold text-emerald-600 mt-1">{{ statusCount(1) }}</div>
      </div>
      <div class="bg-white rounded-xl p-4 border border-gray-100 shadow-sm border-l-4 border-l-rose-400">
        <div class="text-sm text-gray-500">已拒绝</div>
        <div class="text-2xl font-bold text-rose-600 mt-1">{{ statusCount(3) }}</div>
      </div>
    </div>

    <!-- 结算列表 -->
    <div class="bg-white rounded-xl border border-gray-100 shadow-sm overflow-hidden">
      <div class="overflow-x-auto">
        <table class="w-full text-sm whitespace-nowrap">
          <thead>
            <tr class="bg-gray-50 border-b border-gray-100">
              <th class="px-4 py-3 text-left font-semibold text-gray-600">ID</th>
              <th class="px-4 py-3 text-left font-semibold text-gray-600">商户ID</th>
              <th class="px-4 py-3 text-left font-semibold text-gray-600">结算方式</th>
              <th class="px-4 py-3 text-left font-semibold text-gray-600">账号</th>
              <th class="px-4 py-3 text-left font-semibold text-gray-600">姓名</th>
              <th class="px-4 py-3 text-right font-semibold text-gray-600">申请金额</th>
              <th class="px-4 py-3 text-right font-semibold text-gray-600">实际到账</th>
              <th class="px-4 py-3 text-center font-semibold text-gray-600">状态</th>
              <th class="px-4 py-3 text-left font-semibold text-gray-600">申请时间</th>
              <th class="px-4 py-3 text-center font-semibold text-gray-600">操作</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-gray-50">
            <tr v-for="s in settles" :key="s.id" class="hover:bg-gray-50/50 transition-colors">
              <td class="px-4 py-3 text-gray-900 font-medium">{{ s.id }}</td>
              <td class="px-4 py-3 text-gray-600">{{ s.uid }}</td>
              <td class="px-4 py-3">
                <div class="flex items-center gap-1.5">
                  <SvgIcon :name="s.type === 1 ? 'alipay' : 'wechatpay'" :size="16" />
                  <span class="text-sm font-medium" :class="s.type === 1 ? 'text-blue-600' : 'text-green-600'">{{
                    settleType(s.type) }}</span>
                </div>
              </td>
              <td class="px-4 py-3 text-gray-600">{{ s.account }}</td>
              <td class="px-4 py-3 text-gray-600">{{ s.username }}</td>
              <td class="px-4 py-3 text-right font-semibold text-gray-700">￥{{ s.money }}</td>
              <td class="px-4 py-3 text-right font-semibold text-emerald-600">￥{{ s.realmoney }}</td>
              <td class="px-4 py-3 text-center">
                <span
                  :class="['inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium', statusClass(s.status)]">
                  {{ statusMap[s.status]?.text }}
                </span>
              </td>
              <td class="px-4 py-3 text-gray-500 text-xs">{{ formatTime(s.addtime) }}</td>
              <td class="px-4 py-3 text-center">
                <template v-if="s.status === 0">
                  <button @click="handleApprove(s.id)"
                    class="px-3 py-1 text-xs text-green-600 hover:bg-green-50 rounded transition-colors">同意</button>
                  <button @click="handleReject(s.id)"
                    class="px-3 py-1 text-xs text-red-600 hover:bg-red-50 rounded transition-colors">拒绝</button>
                </template>
                <template v-else>
                  <span class="text-gray-400 text-xs">{{ statusMap[s.status]?.text }}</span>
                </template>
              </td>
            </tr>
            <tr v-if="settles.length === 0">
              <td colspan="10" class="px-4 py-12 text-center text-gray-400">
                <div class="flex flex-col items-center">
                  <svg class="w-12 h-12 text-gray-300 mb-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5"
                      d="M17 9V7a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2m2 4h10a2 2 0 002-2v-6a2 2 0 00-2-2H9a2 2 0 00-2 2v6a2 2 0 002 2zm7-5a2 2 0 11-4 0 2 2 0 014 0z" />
                  </svg>
                  <span>暂无结算申请</span>
                </div>
              </td>
            </tr>
          </tbody>
        </table>
      </div>

      <!-- 分页 -->
      <div class="px-4 py-3 border-t border-gray-100 flex items-center justify-between">
        <div class="text-sm text-gray-500">共 {{ total }} 条</div>
        <div class="flex items-center gap-2">
          <button
            class="px-3 py-1 text-sm border border-gray-200 rounded hover:bg-gray-50 disabled:opacity-50 disabled:cursor-not-allowed"
            :disabled="page <= 1" @click="page--; fetchSettles()">上一页</button>
          <span class="px-3 py-1 text-sm">{{ page }} / {{ totalPages }}</span>
          <button
            class="px-3 py-1 text-sm border border-gray-200 rounded hover:bg-gray-50 disabled:opacity-50 disabled:cursor-not-allowed"
            :disabled="page >= totalPages" @click="page++; fetchSettles()">下一页</button>
        </div>
      </div>
    </div>

    <!-- 拒绝原因弹窗 -->
    <div v-if="showRejectModal" class="fixed inset-0 z-50 overflow-y-auto">
      <div class="flex min-h-full items-center justify-center p-4">
        <div class="fixed inset-0 bg-black/50" @click="showRejectModal = false"></div>
        <div class="relative bg-white rounded-xl shadow-xl w-full max-w-md p-6">
          <h3 class="text-lg font-semibold text-gray-900 mb-4">拒绝结算申请</h3>
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-1">拒绝原因</label>
            <textarea v-model="rejectReason"
              class="w-full h-24 px-3 py-2 border border-gray-200 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 resize-none"
              placeholder="请输入拒绝原因..."></textarea>
          </div>
          <div class="flex justify-end gap-3 mt-6">
            <button @click="showRejectModal = false"
              class="px-4 py-2 text-sm text-gray-600 hover:bg-gray-100 rounded-lg transition-colors">取消</button>
            <button @click="confirmReject"
              class="px-4 py-2 text-sm bg-red-600 text-white rounded-lg hover:bg-red-700 transition-colors">确认拒绝</button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { getSettleList, settleOp } from '@/api/admin'
import { ElMessage, ElMessageBox } from 'element-plus'
import dayjs from 'dayjs'
import SvgIcon from '@/components/svgicon.vue'

const settles = ref<any[]>([])
const loading = ref(false)
const page = ref(1)
const pageSize = ref(20)
const total = ref(0)
const filterStatus = ref(-1)
const showRejectModal = ref(false)
const currentRejectId = ref<number | null>(null)
const rejectReason = ref('')

const totalPages = computed(() => Math.ceil(total.value / pageSize.value) || 1)

function statusCount(s: number) {
  return settles.value.filter(r => r.status === s).length
}

const statusMap: Record<number, { text: string }> = {
  0: { text: '待处理' },
  1: { text: '已完成' },
  2: { text: '处理中' },
  3: { text: '已拒绝' }
}

function statusClass(s: number) {
  const map: Record<number, string> = {
    0: 'bg-yellow-100 text-yellow-700',
    1: 'bg-green-100 text-green-700',
    2: 'bg-blue-100 text-blue-700',
    3: 'bg-red-100 text-red-700'
  }
  return map[s] || 'bg-gray-100 text-gray-700'
}

function settleType(type: number) {
  return ['', '支付宝', '微信'][type] || '未知'
}

function formatTime(time: string) {
  if (!time) return '-'
  return dayjs(time).format('YYYY-MM-DD HH:mm')
}

async function fetchSettles() {
  loading.value = true
  try {
    const params: any = { page: page.value, limit: pageSize.value }
    if (filterStatus.value !== -1) {
      params.status = filterStatus.value
    }
    const res = await getSettleList(params)
    if (res.code === 0) {
      settles.value = res.data || []
      total.value = res.count || 0
    }
  } catch (error) {
    console.error('获取结算列表失败:', error)
  } finally {
    loading.value = false
  }
}

async function handleApprove(id: number) {
  try {
    await ElMessageBox.confirm('确定同意该结算申请？', '提示', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    })
  } catch {
    return
  }
  try {
    const res = await settleOp({ action: 'approve', id })
    ElMessage.success(res.msg || '操作成功')
    fetchSettles()
  } catch (error) {
    console.error('操作失败:', error)
  }
}

function handleReject(id: number) {
  currentRejectId.value = id
  rejectReason.value = ''
  showRejectModal.value = true
}

async function confirmReject() {
  if (!currentRejectId.value) return
  if (!rejectReason.value.trim()) {
    ElMessage.warning('请输入拒绝原因')
    return
  }
  try {
    const res = await settleOp({ action: 'reject', id: currentRejectId.value, reason: rejectReason.value })
    ElMessage.success(res.msg || '已拒绝')
    showRejectModal.value = false
    fetchSettles()
  } catch (error) {
    console.error('操作失败:', error)
  }
}

onMounted(() => {
  fetchSettles()
})
</script>
