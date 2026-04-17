<template>
  <div>
    <h2 class="text-2xl font-bold text-gray-800 mb-6">资金记录</h2>

    <div class="card">
      <div class="card-body">
        <table class="table whitespace-nowrap">
          <thead>
            <tr>
              <th>类型</th>
              <th>金额</th>
              <th>余额</th>
              <th>关联订单</th>
              <th>时间</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="r in records" :key="r.id">
              <td>{{ actionMap[r.action] || '未知' }}</td>
              <td :class="r.money >= 0 ? 'text-success' : 'text-danger'">
                {{ r.money >= 0 ? '+' : '' }}{{ r.money }}
              </td>
              <td>¥{{ r.newmoney }}</td>
              <td>{{ r.trade_no || '-' }}</td>
              <td>{{ dayjs(r.date).format('YYYY-MM-DD HH:mm') }}</td>
            </tr>
            <tr v-if="records.length === 0">
              <td colspan="5" class="text-center text-gray-500 py-8">暂无资金记录</td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { getUserRecords } from '@/api/user'
import dayjs from 'dayjs'

const records = ref<any[]>([])
const loading = ref(false)

const actionMap: Record<number, string> = {
  1: '订单收入',
  2: '结算扣款',
  3: '转账',
  4: '退款',
  5: '后台加款',
  6: '后台扣款',
  7: '邀请返现',
  8: '结算返还',
  9: '转账退款'
}

async function fetchRecords() {
  loading.value = true
  try {
    const res = await getUserRecords({ page: 1, limit: 50 })
    if (res.code === 0) {
      records.value = res.data || []
    }
  } catch (error) {
    console.error('获取资金记录失败:', error)
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  fetchRecords()
})
</script>
