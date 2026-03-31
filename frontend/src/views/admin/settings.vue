<template>
  <div>
    <h2 class="text-2xl font-bold text-gray-800 mb-6">系统设置</h2>

    <div class="flex gap-6">
      <!-- 标签导航 -->
      <div class="w-48">
        <div class="bg-white rounded-lg border p-2">
          <button v-for="tab in tabs" :key="tab.id" :class="[
            'w-full text-left px-4 py-2 rounded-lg text-sm transition-colors',
            activeTab === tab.id
              ? 'bg-primary-50 text-primary-700'
              : 'text-gray-600 hover:bg-gray-50'
          ]" @click="activeTab = tab.id">
            {{ tab.name }}
          </button>
        </div>
      </div>

      <!-- 设置表单 -->
      <div class="flex-1">
        <div class="card">
          <div class="card-body">
            <div v-if="successMsg" class="mb-4 p-3 bg-green-50 border border-green-200 rounded-lg">
              <p class="text-sm text-green-600">{{ successMsg }}</p>
            </div>

            <div v-if="errorMsg" class="mb-4 p-3 bg-red-50 border border-red-200 rounded-lg">
              <p class="text-sm text-red-600">{{ errorMsg }}</p>
            </div>

            <form v-if="activeTab !== 'account'" @submit.prevent="handleSave" class="space-y-4">
              <div v-if="activeTab === 'site'">
                <div class="mb-4">
                  <label class="form-label">网站名称</label>
                  <input v-model="form.sitename" type="text" class="form-input" />
                </div>
                <div class="mb-4">
                  <label class="form-label">本地地址</label>
                  <input v-model="form.localurl" type="text" class="form-input" />
                </div>
                <div class="mb-4">
                  <label class="form-label">API地址</label>
                  <input v-model="form.apiurl" type="text" class="form-input" />
                </div>
                <div class="mb-4">
                  <label class="form-label">客服QQ</label>
                  <input v-model="form.kfqq" type="text" class="form-input" />
                </div>
                <div class="mb-4">
                  <label class="form-label">开放注册</label>
                  <select v-model="form.reg_open" class="form-input">
                    <option value="0">关闭</option>
                    <option value="1">开放</option>
                    <option value="2">需要邀请码</option>
                  </select>
                </div>
              </div>

              <div v-if="activeTab === 'settle'">
                <div class="mb-4">
                  <label class="form-label">最低结算金额</label>
                  <input v-model="form.settle_money" type="text" class="form-input" />
                </div>
                <div class="mb-4">
                  <label class="form-label">支付宝结算</label>
                  <select v-model="form.settle_alipay" class="form-input">
                    <option value="1">开启</option>
                    <option value="0">关闭</option>
                  </select>
                </div>
                <div class="mb-4">
                  <label class="form-label">微信结算</label>
                  <select v-model="form.settle_wxpay" class="form-input">
                    <option value="1">开启</option>
                    <option value="0">关闭</option>
                  </select>
                </div>
              </div>

              <button type="submit" :disabled="saving" class="btn btn-primary">
                {{ saving ? '保存中...' : '保存设置' }}
              </button>
            </form>

            <form v-else @submit.prevent="handlePasswordChange" class="space-y-4">
              <div class="mb-4">
                <label class="form-label">原密码</label>
                <input v-model="passwordForm.old_pwd" type="password" class="form-input" placeholder="请输入原密码" />
              </div>
              <div class="mb-4">
                <label class="form-label">新密码</label>
                <input v-model="passwordForm.new_pwd" type="password" class="form-input" placeholder="请输入新密码（至少8位）" />
              </div>
              <div class="mb-4">
                <label class="form-label">确认新密码</label>
                <input v-model="passwordForm.confirm_pwd" type="password" class="form-input" placeholder="请再次输入新密码" />
              </div>

              <button type="submit" :disabled="passwordSaving" class="btn btn-primary">
                {{ passwordSaving ? '修改中...' : '修改密码' }}
              </button>
            </form>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive } from 'vue'
import { saveConfig } from '@/api/admin'

const activeTab = ref('site')

const tabs = [
  { id: 'site', name: '网站设置' },
  { id: 'pay', name: '支付设置' },
  { id: 'settle', name: '结算设置' },
  { id: 'transfer', name: '转账设置' },
  { id: 'oauth', name: '快捷登录' },
  { id: 'notice', name: '通知设置' },
  { id: 'account', name: '账户设置' }
]

const form = reactive({
  sitename: '',
  localurl: '',
  apiurl: '',
  kfqq: '',
  reg_open: '1',
  settle_money: '30',
  settle_alipay: '1',
  settle_wxpay: '1'
})

const passwordForm = reactive({
  old_pwd: '',
  new_pwd: '',
  confirm_pwd: ''
})

const saving = ref(false)
const passwordSaving = ref(false)
const successMsg = ref('')
const errorMsg = ref('')

async function handleSave() {
  saving.value = true
  successMsg.value = ''
  errorMsg.value = ''

  try {
    const res = await saveConfig(form)
    if (res.code === 0) {
      successMsg.value = '保存成功'
    } else {
      errorMsg.value = res.msg || '保存失败'
    }
  } catch (error: any) {
    console.error('保存失败:', error)
    errorMsg.value = error.message || '保存失败'
  } finally {
    saving.value = false
  }
}

async function handlePasswordChange() {
  if (!passwordForm.old_pwd || !passwordForm.new_pwd || !passwordForm.confirm_pwd) {
    errorMsg.value = '请填写所有密码字段'
    return
  }

  if (passwordForm.new_pwd !== passwordForm.confirm_pwd) {
    errorMsg.value = '两次输入的新密码不一致'
    return
  }

  if (passwordForm.new_pwd.length < 8) {
    errorMsg.value = '新密码长度至少8位'
    return
  }

  passwordSaving.value = true
  successMsg.value = ''
  errorMsg.value = ''

  try {
    const res = await saveConfig({ mod: 'account', ...passwordForm })
    if (res.code === 0) {
      successMsg.value = '密码修改成功'
      passwordForm.old_pwd = ''
      passwordForm.new_pwd = ''
      passwordForm.confirm_pwd = ''
      if (res.token) {
        localStorage.setItem('admin_token', res.token)
      }
    } else {
      errorMsg.value = res.msg || '密码修改失败'
    }
  } catch (error: any) {
    console.error('密码修改失败:', error)
    errorMsg.value = error.message || '密码修改失败'
  } finally {
    passwordSaving.value = false
  }
}
</script>
