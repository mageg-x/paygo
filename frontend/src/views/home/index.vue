<template>
  <div class="min-h-screen bg-gradient-to-br from-blue-600 via-sky-500 to-indigo-600 text-white">
    <div class="mx-auto max-w-4xl px-6 py-14 md:py-20">
      <div class="text-center">
        <img :src="logo" alt="彩虹易支付" class="mx-auto h-20 w-20 drop-shadow-lg" />
        <h1 class="mt-5 text-4xl md:text-5xl font-black tracking-wide">彩虹易支付</h1>
        <p class="mt-3 text-blue-50 text-base md:text-lg">安全、快捷、稳定的聚合支付平台</p>
      </div>

      <div class="mt-10 grid gap-4 md:grid-cols-3">
        <div class="rounded-2xl bg-white/15 backdrop-blur p-5 border border-white/30 shadow-lg">
          <div class="text-sm text-blue-100">核心能力</div>
          <div class="mt-1 text-xl font-bold">统一收银</div>
          <p class="mt-2 text-sm text-blue-50/90">支付宝、微信等通道统一接入，快速完成支付闭环。</p>
        </div>
        <div class="rounded-2xl bg-white/15 backdrop-blur p-5 border border-white/30 shadow-lg">
          <div class="text-sm text-blue-100">核心能力</div>
          <div class="mt-1 text-xl font-bold">稳定可靠</div>
          <p class="mt-2 text-sm text-blue-50/90">支持状态轮询、回调重试与风控校验，保障订单可追踪。</p>
        </div>
        <div class="rounded-2xl bg-white/15 backdrop-blur p-5 border border-white/30 shadow-lg">
          <div class="text-sm text-blue-100">核心能力</div>
          <div class="mt-1 text-xl font-bold">开放接口</div>
          <p class="mt-2 text-sm text-blue-50/90">提供 OpenAPI 对接能力，兼容商户自有业务系统。</p>
        </div>
      </div>

      <div class="mt-10 rounded-2xl bg-white/95 text-gray-800 p-6 md:p-8 shadow-2xl">
        <h2 class="text-2xl font-bold">开始使用</h2>
        <p class="mt-2 text-sm text-gray-500">请选择入口进入系统控制台或收银体验页。</p>

        <div class="mt-6 grid gap-3 grid-cols-3">
          <button
            class="rounded-xl bg-blue-600 text-white px-4 py-3 font-semibold hover:bg-blue-700 transition-colors"
            @click="go('/user/login')"
          >
            商户后台
          </button>
          <button
            class="rounded-xl bg-slate-800 text-white px-4 py-3 font-semibold hover:bg-slate-900 transition-colors"
            @click="go('/admin/login')"
          >
            管理后台
          </button>
          <button
            class="rounded-xl bg-indigo-500 text-white px-4 py-3 font-semibold hover:bg-indigo-600 transition-colors"
            @click="go('/cashier/user/1')"
          >
            收银体验
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useRouter } from 'vue-router'
import logo from '@/assets/paygo.png'

const router = useRouter()

function go(path: string) {
  router.push(path)
}
</script>