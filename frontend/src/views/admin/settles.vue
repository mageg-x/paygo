<template>
  <div>
    <h2 class="text-2xl font-bold text-gray-800 mb-6">结算管理</h2>

    <div class="card">
      <div class="card-body">
        <table class="table">
          <thead>
            <tr>
              <th>ID</th>
              <th>商户ID</th>
              <th>结算方式</th>
              <th>账号</th>
              <th>姓名</th>
              <th>申请金额</th>
              <th>实际到账</th>
              <th>状态</th>
              <th>申请时间</th>
              <th>操作</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="s in settles" :key="s.id">
              <td>{{ s.id }}</td>
              <td>{{ s.uid }}</td>
              <td>{{ ['支付宝', '微信'][s.type - 1] || '未知' }}</td>
              <td>{{ s.account }}</td>
              <td>{{ s.username }}</td>
              <td class="text-warning font-medium">¥{{ s.money }}</td>
              <td class="text-success font-medium">¥{{ s.realmoney }}</td>
              <td>
                <span :class="['badge', statusMap[s.status]?.class]">
                  {{ statusMap[s.status]?.text }}
                </span>
              </td>
              <td>{{ dayjs(s.addtime).format('YYYY-MM-DD HH:mm') }}</td>
              <td>
                <template v-if="s.status === 0">
                  <button class="text-success hover:text-success mr-2" @click="handleApprove(s.id)">同意</button>
                  <button class="text-danger hover:text-danger" @click="handleReject(s.id)">拒绝</button>
                </template>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { getSettleList, settleOp } from '@/api/admin'
import dayjs from 'dayjs'

const settles = ref<any[]>([])
const loading = ref(false)
const page = ref(1)
const total = ref(0)

const statusMap: Record<number, { text: string; class: string }> = {
  0: { text: '待处理', class: 'badge-warning' },
  1: { text: '已完成', class: 'badge-success' },
  2: { text: '处理中', class: 'badge-info' },
  3: { text: '已拒绝', class: 'badge-danger' }
}

async function fetchSettles() {
  loading.value = true
  try {
    const res = await getSettleList({ page: page.value, limit: 20 })
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
  if (confirm('确定同意该结算申请？')) {
    await settleOp({ action: 'approve', id })
    fetchSettles()
  }
}

async function handleReject(id: number) {
  const reason = prompt('请输入拒绝原因：')
  if (reason) {
    await settleOp({ action: 'reject', id, reason })
    fetchSettles()
  }
}

onMounted(() => {
  fetchSettles()
})
</script>
