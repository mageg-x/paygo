<template>
  <div class="space-y-4">
    <div class="page-head">
      <div>
        <h1 class="page-title no-wrap">结算管理</h1>
        <p class="page-subtitle">管理商户结算申请，支持多退少补（补发差额 / 冲正扣回）</p>
      </div>
      <select v-model="filterStatus" @change="page = 1; fetchSettles()"
        class="form-input w-auto min-w-[132px] px-3">
        <option :value="-1">全部状态</option>
        <option :value="0">待处理</option>
        <option :value="1">已完成</option>
        <option :value="2">处理中</option>
        <option :value="3">已拒绝</option>
      </select>
    </div>

    <!-- 统计卡片 -->
    <div class="grid grid-cols-2 lg:grid-cols-4 gap-3 md:gap-4">
      <div class="card p-4">
        <div class="text-sm text-gray-500">全部申请</div>
        <div class="text-2xl font-bold text-slate-700 mt-1">{{ total }}</div>
      </div>
      <div class="card p-4">
        <div class="text-sm text-gray-500">待处理</div>
        <div class="text-2xl font-bold text-amber-600 mt-1">{{ statusCount(0) }}</div>
      </div>
      <div class="card p-4">
        <div class="text-sm text-gray-500">已完成</div>
        <div class="text-2xl font-bold text-emerald-600 mt-1">{{ statusCount(1) }}</div>
      </div>
      <div class="card p-4">
        <div class="text-sm text-gray-500">已拒绝</div>
        <div class="text-2xl font-bold text-rose-600 mt-1">{{ statusCount(3) }}</div>
      </div>
    </div>

    <div class="card p-3 md:p-4 text-sm text-slate-600">
      转账发起入口在本页面：待处理结算可“同意”执行打款；已处理结算可做“补发差额 / 冲正扣回”。
    </div>

    <div class="table-shell">
      <div class="overflow-x-auto">
        <table class="table min-w-[980px] whitespace-nowrap">
          <thead>
            <tr>
              <th class="text-left">ID</th>
              <th class="text-left">商户ID</th>
              <th class="text-left">结算方式</th>
              <th class="text-left">账号</th>
              <th class="text-left">姓名</th>
              <th class="text-right">申请金额</th>
              <th class="text-right">实际到账</th>
              <th>状态</th>
              <th class="text-left">申请时间</th>
              <th>操作</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="s in settles" :key="s.id">
              <td class="text-left text-gray-900 font-medium">{{ s.id }}</td>
              <td class="text-left text-gray-600">{{ s.uid }}</td>
              <td class="text-left">
                <div class="flex items-center gap-1.5">
                  <SvgIcon :name="s.type === 1 ? 'alipay' : 'wechatpay'" :size="16" />
                  <span class="text-sm font-medium" :class="s.type === 1 ? 'text-blue-600' : 'text-green-600'">{{
                    settleType(s.type) }}</span>
                </div>
              </td>
              <td class="text-left text-gray-600">{{ s.account }}</td>
              <td class="text-left text-gray-600">{{ s.username }}</td>
              <td class="text-right font-semibold text-gray-700">￥{{ s.money }}</td>
              <td class="text-right font-semibold text-emerald-600">￥{{ s.realmoney }}</td>
              <td>
                <span :class="['badge', statusClass(s.status)]">
                  {{ statusMap[s.status]?.text }}
                </span>
              </td>
              <td class="text-left text-gray-500 text-xs">{{ formatTime(s.addtime) }}</td>
              <td>
                <template v-if="s.status === 0">
                  <button @click="handleApprove(s.id)" class="action-link action-link-success">同意</button>
                  <button @click="handleReject(s.id)" class="action-link action-link-danger">拒绝</button>
                </template>
                <template v-else>
                  <div class="inline-flex items-center gap-1">
                    <button @click="openAdjustModal(s.id, 'compensate')" class="action-link action-link-success">补发差额</button>
                    <button @click="openAdjustModal(s.id, 'deduct')" class="action-link action-link-warning">冲正扣回</button>
                  </div>
                </template>
              </td>
            </tr>
            <tr v-if="settles.length === 0">
              <td colspan="10" class="py-12 text-center text-gray-400">
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

      <div class="px-4 py-3 border-t border-slate-200/70 flex flex-wrap items-center justify-between gap-2">
        <div class="text-sm text-gray-500">共 {{ total }} 条</div>
        <div class="flex items-center gap-2">
          <button class="pagination-item disabled:opacity-50 disabled:cursor-not-allowed"
            :disabled="page <= 1" @click="page--; fetchSettles()">上一页</button>
          <span class="px-3 py-1 text-sm">{{ page }} / {{ totalPages }}</span>
          <button class="pagination-item disabled:opacity-50 disabled:cursor-not-allowed"
            :disabled="page >= totalPages" @click="page++; fetchSettles()">下一页</button>
        </div>
      </div>
    </div>

    <div v-if="showRejectModal" class="dialog-backdrop">
      <div class="dialog-wrap">
        <div class="dialog-mask" @click="showRejectModal = false"></div>
        <div class="dialog-panel max-w-md">
          <div class="dialog-header">
            <div>
              <h3 class="dialog-title">拒绝结算申请</h3>
              <p class="dialog-subtitle">请填写拒绝原因</p>
            </div>
            <button class="dialog-close" @click="showRejectModal = false">✕</button>
          </div>
          <div class="dialog-body">
          <div>
            <label class="form-label">拒绝原因</label>
            <textarea v-model="rejectReason"
              class="form-input h-24 resize-none px-3"
              placeholder="请输入拒绝原因..."></textarea>
          </div>
          </div>
          <div class="dialog-footer">
            <button @click="showRejectModal = false" class="btn btn-outline">取消</button>
            <button @click="confirmReject" class="btn btn-danger">确认拒绝</button>
          </div>
        </div>
      </div>
    </div>

    <div v-if="showAdjustModal" class="dialog-backdrop">
      <div class="dialog-wrap">
        <div class="dialog-mask" @click="showAdjustModal = false"></div>
        <div class="dialog-panel max-w-md">
          <div class="dialog-header">
            <div>
              <h3 class="dialog-title">{{ adjustAction === 'compensate' ? '补发差额' : '冲正扣回' }}</h3>
              <p class="dialog-subtitle">请输入金额与原因</p>
            </div>
            <button class="dialog-close" @click="showAdjustModal = false">✕</button>
          </div>
          <div class="dialog-body space-y-4">
            <div>
              <label class="form-label">金额</label>
              <input v-model.number="adjustAmount" type="number" min="0.01" step="0.01" class="form-input px-3" />
            </div>
            <div>
              <label class="form-label">原因</label>
              <textarea v-model="adjustReason" class="form-input h-24 resize-none px-3" placeholder="请输入原因"></textarea>
            </div>
          </div>
          <div class="dialog-footer">
            <button @click="showAdjustModal = false" class="btn btn-outline">取消</button>
            <button @click="confirmAdjust" class="btn btn-primary">确认</button>
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
const showAdjustModal = ref(false)
const currentAdjustId = ref<number | null>(null)
const adjustAction = ref<'compensate' | 'deduct'>('compensate')
const adjustAmount = ref<number>(0.01)
const adjustReason = ref('')

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
    0: 'badge-warning',
    1: 'badge-success',
    2: 'badge-info',
    3: 'badge-danger'
  }
  return map[s] || 'badge'
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

function openAdjustModal(id: number, action: 'compensate' | 'deduct') {
  currentAdjustId.value = id
  adjustAction.value = action
  adjustAmount.value = 0.01
  adjustReason.value = ''
  showAdjustModal.value = true
}

async function confirmAdjust() {
  if (!currentAdjustId.value) return
  if (!adjustAmount.value || adjustAmount.value <= 0) {
    ElMessage.warning('请输入正确金额')
    return
  }
  if (!adjustReason.value.trim()) {
    ElMessage.warning('请输入原因')
    return
  }
  try {
    const res = await settleOp({
      action: adjustAction.value,
      id: currentAdjustId.value,
      amount: Number(adjustAmount.value),
      reason: adjustReason.value.trim()
    })
    ElMessage.success(res.msg || '操作成功')
    showAdjustModal.value = false
    fetchSettles()
  } catch (error) {
    console.error('调整失败:', error)
  }
}

onMounted(() => {
  fetchSettles()
})
</script>
