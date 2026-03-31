<template>
  <div>
    <h2 class="text-2xl font-bold text-gray-800 mb-6">通道管理</h2>

    <div class="card">
      <div class="card-body">
        <table class="table">
          <thead>
            <tr>
              <th>ID</th>
              <th>通道名称</th>
              <th>插件</th>
              <th>费率</th>
              <th>状态</th>
              <th>操作</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="ch in channels" :key="ch.id">
              <td>{{ ch.id }}</td>
              <td>{{ ch.name }}</td>
              <td>{{ ch.plugin }}</td>
              <td>{{ ch.rate }}%</td>
              <td>
                <span :class="['badge', ch.status ? 'badge-success' : 'badge-danger']">
                  {{ ch.status ? '开启' : '关闭' }}
                </span>
              </td>
              <td>
                <button class="text-primary-600 hover:text-primary-800 mr-2">编辑</button>
                <button class="text-danger hover:text-danger">删除</button>
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
import { getChannelList } from '@/api/admin'

const channels = ref<any[]>([])
const loading = ref(false)

async function fetchChannels() {
  loading.value = true
  try {
    const res = await getChannelList()
    if (res.code === 0) {
      channels.value = res.data || []
    }
  } catch (error) {
    console.error('获取通道列表失败:', error)
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  fetchChannels()
})
</script>
