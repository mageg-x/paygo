<template>
  <div>
    <h2 class="text-2xl font-bold text-gray-800 mb-6">结算记录</h2>

    <div class="card">
      <div class="card-body">
        <table class="table whitespace-nowrap">
          <thead>
            <tr>
              <th>结算方式</th>
              <th>账号</th>
              <th>申请金额</th>
              <th>实际到账</th>
              <th>状态</th>
              <th>时间</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="s in settles" :key="s.id">
              <td>
                <div class="flex items-center gap-1.5">
                  <SvgIcon :name="s.type === 1 ? 'alipay' : 'wechatpay'" :size="16" />
                  <span>{{ s.type === 1 ? '支付宝' : '微信' }}</span>
                </div>
              </td>
              <td>{{ s.account }}</td>
              <td class="text-warning">¥{{ s.money }}</td>
              <td class="text-success">¥{{ s.realmoney }}</td>
              <td>
                <span :class="['badge', s.status === 1 ? 'badge-success' : 'badge-warning']">
                  {{ s.status === 1 ? '已完成' : '待处理' }}
                </span>
              </td>
              <td>{{ dayjs(s.addtime).format('YYYY-MM-DD HH:mm') }}</td>
            </tr>
            <tr v-if="settles.length === 0">
              <td colspan="6" class="text-center text-gray-500 py-8">暂无结算记录</td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { getUserSettles } from '@/api/user'
import dayjs from 'dayjs'
import SvgIcon from '@/components/svgicon.vue'

const settles = ref<any[]>([])
const loading = ref(false)

async function fetchSettles() {
  loading.value = true
  try {
    const res = await getUserSettles({ page: 1, limit: 20 })
    if (res.code === 0) {
      settles.value = res.data || []
    }
  } catch (error) {
    console.error('获取结算记录失败:', error)
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  fetchSettles()
})
</script>
