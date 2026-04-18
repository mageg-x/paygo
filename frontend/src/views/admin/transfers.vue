<template>
  <div class="space-y-4">
    <div class="page-head">
      <div>
        <h1 class="page-title no-wrap">转账记录</h1>
        <p class="page-subtitle">查看转账执行结果；发起入口在结算管理（同意结算 / 补发差额）</p>
      </div>
    </div>

    <div class="grid grid-cols-2 lg:grid-cols-4 gap-3 md:gap-4">
      <div class="card p-4">
        <div class="text-sm text-gray-500">总转账笔数</div>
        <div class="text-2xl font-bold text-slate-700 mt-1">{{ total }}</div>
      </div>
      <div class="card p-4">
        <div class="text-sm text-gray-500">待处理</div>
        <div class="text-2xl font-bold text-amber-600 mt-1">{{ statusCount(0) }}</div>
      </div>
      <div class="card p-4">
        <div class="text-sm text-gray-500">成功</div>
        <div class="text-2xl font-bold text-emerald-600 mt-1">{{ statusCount(1) }}</div>
      </div>
      <div class="card p-4">
        <div class="text-sm text-gray-500">失败</div>
        <div class="text-2xl font-bold text-rose-600 mt-1">{{ statusCount(2) }}</div>
      </div>
    </div>

    <div class="panel-filter">
      <div class="card-body toolbar-wrap">
        <div class="w-full md:w-72">
          <input v-model="searchForm.search" type="text" placeholder="搜索交易号/账号/姓名..."
            class="form-input px-3"
            @keyup.enter="page = 1; fetchTransfers()" />
        </div>
        <select v-model="searchForm.status" class="form-input w-auto min-w-[132px] px-3">
          <option value="-1">全部状态</option>
          <option value="0">待处理</option>
          <option value="1">成功</option>
          <option value="2">失败</option>
        </select>
        <button class="btn btn-primary" @click="page = 1; fetchTransfers()">
          搜索
        </button>
      </div>
    </div>

    <div class="table-shell">
      <div class="overflow-x-auto">
        <table class="table min-w-[1100px] whitespace-nowrap">
          <thead>
            <tr>
              <th class="text-left">ID</th>
              <th class="text-left">商户</th>
              <th class="text-left">交易号</th>
              <th class="text-left">类型</th>
              <th class="text-left">账号</th>
              <th class="text-left">姓名</th>
              <th class="text-right">金额</th>
              <th>状态</th>
              <th class="text-left">时间</th>
              <th>操作</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="t in transfers" :key="t.biz_no">
              <td class="text-left text-gray-900">{{ t.id }}</td>
              <td class="text-left text-gray-600">{{ t.user_name || t.uid }}</td>
              <td class="text-left text-gray-500 font-mono text-xs">{{ t.biz_no }}</td>
              <td class="text-left">
                <div class="flex items-center gap-1.5">
                  <SvgIcon :name="t.type === 'alipay' ? 'alipay' : 'wechatpay'" :size="16" />
                  <span class="text-sm font-medium" :class="t.type === 'alipay' ? 'text-blue-600' : 'text-green-600'">{{
                    typeName(t.type) }}</span>
                </div>
              </td>
              <td class="text-left text-gray-600">{{ t.account }}</td>
              <td class="text-left text-gray-600">{{ t.username }}</td>
              <td class="text-right font-semibold text-emerald-600">￥{{ t.money }}</td>
              <td>
                <span :class="['badge', statusClass(t.status)]">
                  {{ statusName(t.status) }}
                </span>
              </td>
              <td class="text-left text-gray-500 text-xs">{{ formatTime(t.paytime) }}</td>
              <td>
                <div class="inline-flex items-center gap-1">
                  <button @click="queryStatus(t.biz_no)" class="action-link action-link-primary">查询</button>
                  <button @click="showSetStatus(t)" class="action-link action-link-primary">改状态</button>
                  <button @click="handleRefund(t.biz_no)" class="action-link action-link-warning">退回</button>
                  <button @click="handleDelete(t.biz_no)" class="action-link action-link-danger">删除</button>
                </div>
              </td>
            </tr>
            <tr v-if="transfers.length === 0">
              <td colspan="10" class="py-12 text-center text-gray-400">
                <div class="flex flex-col items-center">
                  <svg class="w-12 h-12 text-gray-300 mb-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5"
                      d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2" />
                  </svg>
                  <span>暂无转账记录</span>
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
            :disabled="page <= 1" @click="page--; fetchTransfers()">上一页</button>
          <span class="px-3 py-1 text-sm">{{ page }} / {{ totalPages }}</span>
          <button class="pagination-item disabled:opacity-50 disabled:cursor-not-allowed"
            :disabled="page >= totalPages" @click="page++; fetchTransfers()">下一页</button>
        </div>
      </div>
    </div>

    <div v-if="showStatusModal" class="dialog-backdrop">
      <div class="dialog-wrap">
        <div class="dialog-mask" @click="showStatusModal = false"></div>
        <div class="dialog-panel max-w-md">
          <div class="dialog-header">
            <div>
              <h3 class="dialog-title">修改转账状态</h3>
              <p class="dialog-subtitle">可手动修正转账处理结果</p>
            </div>
            <button class="dialog-close" @click="showStatusModal = false">✕</button>
          </div>
          <div class="dialog-body space-y-4">
            <div>
              <label class="form-label">状态</label>
              <select v-model="newStatus" class="form-input px-3">
                <option value="0">待处理</option>
                <option value="1">成功</option>
                <option value="2">失败</option>
              </select>
            </div>
            <div>
              <label class="form-label">失败原因（可选）</label>
              <input v-model="newResult" type="text" placeholder="请输入失败原因"
                class="form-input px-3" />
            </div>
          </div>
          <div class="dialog-footer">
            <button @click="showStatusModal = false" class="btn btn-outline">取消</button>
            <button @click="confirmSetStatus" class="btn btn-primary">确定</button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { getTransferList, transferOp } from '@/api/admin'
import { ElMessage, ElMessageBox } from 'element-plus'
import SvgIcon from '@/components/svgicon.vue'

const transfers = ref<any[]>([])
const page = ref(1)
const pageSize = ref(20)
const total = ref(0)
const searchForm = ref({
  search: '',
  status: '-1'
})
const showStatusModal = ref(false)
const currentBizNo = ref('')
const newStatus = ref('1')
const newResult = ref('')

const totalPages = computed(() => Math.ceil(total.value / pageSize.value) || 1)

function statusCount(status: number) {
  return transfers.value.filter(t => t.status === status).length
}

async function fetchTransfers() {
  try {
    const res = await getTransferList({
      page: page.value,
      limit: pageSize.value,
      status: searchForm.value.status,
      search: searchForm.value.search
    })
    if (res.code === 0) {
      transfers.value = res.data || []
      total.value = res.count || 0
    }
  } catch (error) {
    console.error('获取转账列表失败:', error)
  }
}

async function queryStatus(bizNo: string) {
  try {
    const res = await transferOp({ action: 'query', biz_no: bizNo })
    ElMessage.success(res.msg || '查询成功')
    fetchTransfers()
  } catch (error) {
    console.error('查询失败:', error)
  }
}

function showSetStatus(t: any) {
  currentBizNo.value = t.biz_no
  newStatus.value = String(t.status)
  newResult.value = ''
  showStatusModal.value = true
}

async function confirmSetStatus() {
  try {
    const res = await transferOp({
      action: 'set_status',
      biz_no: currentBizNo.value,
      status: parseInt(newStatus.value),
      result: newResult.value
    })
    ElMessage.success(res.msg || '状态已更新')
    showStatusModal.value = false
    fetchTransfers()
  } catch (error) {
    console.error('更新状态失败:', error)
  }
}

async function handleRefund(bizNo: string) {
  try {
    await ElMessageBox.confirm('确定要退回这笔转账吗？', '提示', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    })
  } catch {
    return
  }
  try {
    const res = await transferOp({ action: 'refund', biz_no: bizNo })
    ElMessage.success(res.msg || '退回成功')
    fetchTransfers()
  } catch (error) {
    console.error('退回失败:', error)
  }
}

async function handleDelete(bizNo: string) {
  try {
    await ElMessageBox.confirm('确定要删除这条记录吗？', '提示', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    })
  } catch {
    return
  }
  try {
    const res = await transferOp({ action: 'delete', biz_no: bizNo })
    ElMessage.success(res.msg || '删除成功')
    fetchTransfers()
  } catch (error) {
    console.error('删除失败:', error)
  }
}

function typeName(type: string) {
  const map: Record<string, string> = {
    alipay: '支付宝',
    wxpay: '微信',
  }
  return map[type] || type
}

function statusName(status: number) {
  const map: Record<number, string> = {
    0: '待处理',
    1: '成功',
    2: '失败'
  }
  return map[status] || '未知'
}

function statusClass(status: number) {
  const map: Record<number, string> = {
    0: 'badge-warning',
    1: 'badge-success',
    2: 'badge-danger'
  }
  return map[status] || 'badge'
}

function formatTime(time: string) {
  if (!time) return '-'
  return new Date(time).toLocaleString('zh-CN')
}

onMounted(() => {
  fetchTransfers()
})
</script>
