<template>
  <div>
    <h2 class="text-2xl font-bold text-gray-800 mb-6">插件管理</h2>

    <div class="card">
      <div class="card-body">
        <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
          <div v-for="p in plugins" :key="p.name" class="border rounded-lg p-4">
            <div class="flex justify-between items-start mb-2">
              <h3 class="font-medium text-gray-800">{{ p.showname }}</h3>
              <span class="badge badge-info">{{ p.name }}</span>
            </div>
            <p class="text-sm text-gray-500 mb-2">{{ p.author }}</p>
            <p class="text-xs text-gray-400">{{ p.types }}</p>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { getPluginList } from '@/api/admin'

const plugins = ref<any[]>([])
const loading = ref(false)

async function fetchPlugins() {
  loading.value = true
  try {
    const res = await getPluginList()
    if (res.code === 0) {
      plugins.value = res.data || []
    }
  } catch (error) {
    console.error('获取插件列表失败:', error)
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  fetchPlugins()
})
</script>
