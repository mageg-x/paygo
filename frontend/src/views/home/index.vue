<template>
  <div class="min-h-screen bg-gradient-to-br from-blue-500 to-purple-600">
    <div class="container mx-auto px-4 py-16">
      <div class="text-center mb-12">
        <h1 class="text-4xl font-bold text-white mb-4">彩虹易支付</h1>
        <p class="text-blue-100">安全、快捷、稳定的聚合支付平台</p>
      </div>

      <div class="max-w-lg mx-auto bg-white rounded-2xl shadow-xl p-8">
        <h2 class="text-2xl font-bold text-gray-800 mb-6">发起支付</h2>

        <form @submit.prevent="handleSubmit" class="space-y-4">
          <div>
            <label class="form-label">商户ID</label>
            <input v-model="form.pid" type="text" class="form-input px-3" placeholder="请输入商户ID" required />
          </div>

          <div>
            <label class="form-label">支付方式</label>
            <div class="grid grid-cols-2 gap-2">
              <button v-for="pt in payTypes" :key="pt.id" type="button" :class="[
                'p-3 rounded-lg border-2 text-center transition-all',
                form.type === String(pt.id)
                  ? 'border-primary-500 bg-primary-50 text-primary-700'
                  : 'border-gray-200 hover:border-gray-300'
              ]" @click="form.type = String(pt.id)">
                <div class="text-lg">{{ pt.name }}</div>
              </button>
            </div>
          </div>

          <div>
            <label class="form-label">商户订单号</label>
            <input v-model="form.out_trade_no" type="text" class="form-input px-3" placeholder="唯一订单号" required />
          </div>

          <div>
            <label class="form-label">商品名称</label>
            <input v-model="form.name" type="text" class="form-input px-3" placeholder="商品名称" required />
          </div>

          <div>
            <label class="form-label">金额（元）</label>
            <input v-model="form.money" type="number" step="0.01" class="form-input px-3" placeholder="0.00" required />
          </div>

          <div>
            <label class="form-label">回调地址</label>
            <input v-model="form.notify_url" type="url" class="form-input px-3" placeholder="http://" />
          </div>

          <div>
            <label class="form-label">返回地址</label>
            <input v-model="form.return_url" type="url" class="form-input px-3" placeholder="http://" />
          </div>

          <button type="submit" :disabled="loading"
            class="w-full py-3 bg-primary-600 text-white rounded-lg font-medium hover:bg-primary-700 transition-colors disabled:opacity-50">
            {{ loading ? '处理中...' : '提交支付' }}
          </button>
        </form>
      </div>

      <div class="text-center mt-8 text-blue-100">
        <p>API文档 | 商户后台 | 联系我们</p>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'

const loading = ref(false)
const form = ref({
  pid: '',
  type: '1',
  out_trade_no: '',
  name: '',
  money: '',
  notify_url: '',
  return_url: ''
})

const payTypes = [
  { id: 1, name: '支付宝', icon: 'alipay' },
  { id: 2, name: '微信支付', icon: 'wechat' },
]

async function handleSubmit() {
  if (!form.value.pid || !form.value.out_trade_no || !form.value.money) {
    alert('请填写必填项')
    return
  }

  loading.value = true

  try {
    const res = await fetch('/api/pay/submit', {
      method: 'POST',
      headers: { 'Content-Type': 'application/x-www-form-urlencoded' },
      body: new URLSearchParams({
        ...form.value,
        money: form.value.money
      })
    })
    const data = await res.json()

    if (data.code === 0) {
      if (data.result.type === 'jump' && data.result.url) {
        window.location.href = data.result.url
      }
    } else {
      alert(data.msg)
    }
  } catch (error) {
    alert('请求失败')
  } finally {
    loading.value = false
  }
}
</script>
