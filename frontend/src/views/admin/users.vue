<template>
  <div>
    <div class="flex justify-between items-center mb-6">
      <div>
        <h2 class="text-2xl font-bold text-gray-800">商户管理</h2>
        <p class="text-sm text-gray-500 mt-1">共 {{ total }} 个商户</p>
      </div>
      <button class="btn btn-primary" @click="openAddDialog">+ 添加商户</button>
    </div>

    <div class="card mb-4">
      <div class="card-body py-3">
        <div class="flex items-center gap-4">
          <div class="relative flex-1 max-w-xs">
            <input v-model="searchKeyword" type="text" class="form-input pl-10" placeholder="搜索商户 ID、姓名、账号..." />
            <svg class="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-gray-400" fill="none"
              stroke="currentColor" viewBox="0 0 24 24">
              <circle cx="11" cy="11" r="8"></circle>
              <path d="m21 21-4.35-4.35"></path>
            </svg>
          </div>
          <span class="text-sm text-gray-500">筛选结果：{{ filteredUsers.length }} 条</span>
        </div>
      </div>
    </div>

    <div class="card overflow-hidden">
      <div class="card-body p-0 overflow-x-auto">
        <div v-if="loading" class="flex items-center justify-center py-12 text-gray-500">
          <svg class="animate-spin h-6 w-6 mr-2" fill="none" viewBox="0 0 24 24">
            <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
            <path class="opacity-75" fill="currentColor"
              d="m4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z">
            </path>
          </svg>
          加载中...
        </div>

        <table v-else class="table table-fixed">
          <thead>
            <tr>
              <th class="pl-6 w-8">ID</th>
              <th class="w-32">商户信息</th>
              <th class="w-32">结算账号</th>
              <th class="w-24">余额</th>
              <th class="w-20">模式</th>
              <th class="w-20">支付</th>
              <th class="w-20">结算</th>
              <th class="w-20">状态</th>
              <th class="w-36">注册时间</th>
              <th class="pr-6 w-56">操作</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="user in filteredUsers" :key="user.uid" class="align-middle">
              <td class="pl-6">
                <span
                  class="inline-flex items-center justify-center min-w-[32px] h-6 px-2 bg-blue-50 text-blue-600 text-xs font-semibold rounded">
                  {{ user.uid }}
                </span>
              </td>
              <td class="truncate">
                <div class="font-medium text-gray-900 truncate">{{ user.username || '-' }}</div>
                <div class="text-xs text-gray-500 truncate">
                  <span v-if="user.phone">📱{{ user.phone }}</span>
                  <span v-if="user.phone && user.email"> | </span>
                  <span v-if="user.email">✉️{{ user.email }}</span>
                </div>
              </td>
              <td class="truncate">
                <div class="text-gray-900 truncate">{{ user.account || '-' }}</div>
                <div class="text-xs text-gray-500 truncate" :title="user.url">{{ user.url || '-' }}</div>
              </td>
              <td class="whitespace-nowrap">
                <span
                  class="inline-flex items-center px-2 py-0.5 rounded-lg bg-emerald-50 text-emerald-700 font-semibold text-sm">
                  ¥{{ user.money }}
                </span>
              </td>
              <td>
                <span
                  :class="['inline-flex items-center px-2 py-0.5 rounded-full text-xs font-medium', user.mode === 1 ? 'bg-amber-100 text-amber-700' : 'bg-slate-100 text-slate-600']">
                  {{ user.mode === 1 ? '加费' : '减费' }}
                </span>
              </td>
              <td>
                <button @click="setStatus(user.uid, user.pay === 1 ? 0 : 1)" :class="[
                  'inline-flex items-center px-2 py-0.5 rounded-full text-xs font-medium cursor-pointer transition-colors',
                  user.pay === 1 ? 'bg-emerald-100 text-emerald-700 hover:bg-emerald-200' : 'bg-rose-100 text-rose-700 hover:bg-rose-200'
                ]">
                  {{ payMap[user.pay] }}
                </button>
              </td>
              <td>
                <button @click="setStatus(user.uid, user.settle === 1 ? 0 : 1)" :class="[
                  'inline-flex items-center px-2 py-0.5 rounded-full text-xs font-medium cursor-pointer transition-colors',
                  user.settle === 1 ? 'bg-emerald-100 text-emerald-700 hover:bg-emerald-200' : 'bg-rose-100 text-rose-700 hover:bg-rose-200'
                ]">
                  {{ settleMap[user.settle] }}
                </button>
              </td>
              <td>
                <span :class="['badge', statusMap[user.status]?.class]">
                  {{ statusMap[user.status]?.text || '未知' }}
                </span>
              </td>
              <td class="text-gray-500 text-sm whitespace-nowrap">{{ formatTime(user.addtime) }}</td>
              <td class="pr-6">
                <div class="flex items-center gap-1">
                  <button class="text-blue-600 hover:text-blue-800 text-xs font-medium px-1"
                    @click="openEditDialog(user)">编辑</button>
                  <button class="text-purple-600 hover:text-purple-800 text-xs font-medium px-1"
                    @click="resetKey(user.uid)">重置密钥</button>
                  <button v-if="user.status === 0" class="text-amber-600 hover:text-amber-800 text-xs font-medium px-1"
                    @click="setStatus(user.uid, 1)">
                    禁用
                  </button>
                  <button v-else-if="user.status === 1"
                    class="text-emerald-600 hover:text-emerald-800 text-xs font-medium px-1"
                    @click="setStatus(user.uid, 0)">
                    启用
                  </button>
                  <button class="text-red-600 hover:text-red-800 text-xs font-medium px-1"
                    @click="deleteUser(user.uid)">删除</button>
                </div>
              </td>
            </tr>
          </tbody>
        </table>

        <div v-if="!loading && filteredUsers.length === 0" class="text-center py-12 text-gray-500">
          <svg class="w-16 h-16 mx-auto mb-4 text-gray-300" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5"
              d="M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v-2c0-.656-.126-1.283-.356-1.857M7 20H2v-2a3 3 0 015.356-1.857M7 20v-2c0-.656.126-1.283.356-1.857m0 0a5.002 5.002 0 019.288 0M15 7a3 3 0 11-6 0 3 3 0 016 0zm6 3a2 2 0 11-4 0 2 2 0 014 0zM7 10a2 2 0 11-4 0 2 2 0 014 0z">
            </path>
          </svg>
          <p class="mb-3">暂无商户数据</p>
          <button class="btn btn-primary" @click="openAddDialog">添加第一个商户</button>
        </div>

        <div v-if="!loading && filteredUsers.length > 0" class="pagination border-t text-sm border-gray-100">
          <button class="pagination-item" :disabled="page === 1" @click="page--; fetchUsers()">
            上一页
          </button>
          <span class="px-4 py-1 text-gray-600">
            第 {{ page }} / {{ Math.ceil(total / 20) || 1 }} 页，共 {{ total }} 条
          </span>
          <button class="pagination-item" :disabled="page * 20 >= total" @click="page++; fetchUsers()">
            下一页
          </button>
        </div>
      </div>
    </div>

    <Teleport to="body">
      <div v-if="dialogVisible" class="modal-overlay" @click.self="dialogVisible = false">
        <div class="modal">
          <div class="modal-header">
            <div>
              <h3 class="text-lg font-semibold text-gray-900">{{ dialogTitle }}</h3>
              <p class="text-sm text-gray-500 mt-0.5">{{ isEdit ? '修改商户信息' : '创建新商户账户' }}</p>
            </div>
            <button class="modal-close-btn" @click="dialogVisible = false">&times;</button>
          </div>

          <div class="modal-body">
            <div class="grid grid-cols-3 gap-6">
              <div class="form-section">
                <h4 class="form-section-title">基本信息</h4>
                <div class="space-y-4">
                  <div>
                    <label class="form-label">用户组</label>
                    <select v-model="userForm.gid" class="form-input px-3">
                      <option v-for="g in groups" :key="g.gid" :value="g.gid">{{ g.name }}</option>
                    </select>
                  </div>
                  <div>
                    <label class="form-label">手机号</label>
                    <input v-model="userForm.phone" type="text" class="form-input px-3" placeholder="可留空">
                  </div>
                  <div>
                    <label class="form-label">邮箱</label>
                    <input v-model="userForm.email" type="email" class="form-input px-3" placeholder="可留空">
                  </div>
                  <div>
                    <label class="form-label">登录密码</label>
                    <input v-model="userForm.pwd" type="password" class="form-input px-3"
                      :placeholder="isEdit ? '留空则不修改' : '可留空'">
                  </div>
                  <div>
                    <label class="form-label">QQ</label>
                    <input v-model="userForm.qq" type="text" class="form-input px-3" placeholder="可留空">
                  </div>
                  <div>
                    <label class="form-label">网站域名</label>
                    <input v-model="userForm.url" type="text" class="form-input px-3" placeholder="可留空">
                  </div>
                </div>
              </div>

              <div class="form-section">
                <h4 class="form-section-title">结算信息</h4>
                <div class="space-y-4">
                  <div>
                    <label class="form-label">结算方式</label>
                    <select v-model="userForm.settle_id" class="form-input px-3">
                      <option :value="1">支付宝</option>
                      <option :value="2">微信</option>
                    </select>
                  </div>
                  <div>
                    <label class="form-label">
                      结算账号 <span class="text-red-500">*</span>
                    </label>
                    <input v-model="userForm.account" type="text" class="form-input px-3" placeholder="必填">
                  </div>
                  <div>
                    <label class="form-label">
                      结算姓名 <span class="text-red-500">*</span>
                    </label>
                    <input v-model="userForm.username" type="text" class="form-input px-3" placeholder="必填">
                  </div>
                </div>
              </div>

              <div class="form-section">
                <h4 class="form-section-title">功能开关</h4>
                <div class="space-y-4">
                  <div>
                    <label class="form-label">手续费模式</label>
                    <select v-model="userForm.mode" class="form-input px-3">
                      <option :value="0">余额扣费</option>
                      <option :value="1">订单加费</option>
                    </select>
                  </div>
                  <div>
                    <label class="form-label">商户状态</label>
                    <select v-model="userForm.status" class="form-input px-3">
                      <option :value="1">正常</option>
                      <option :value="0">禁用</option>
                      <option :value="2">待审核</option>
                    </select>
                  </div>
                  <div>
                    <label class="form-label">支付权限</label>
                    <select v-model="userForm.pay" class="form-input px-3">
                      <option :value="1">开启</option>
                      <option :value="0">关闭</option>
                    </select>
                  </div>
                  <div>
                    <label class="form-label">结算权限</label>
                    <select v-model="userForm.settle" class="form-input px-3">
                      <option :value="1">开启</option>
                      <option :value="0">关闭</option>
                    </select>
                  </div>
                </div>
              </div>
            </div>
          </div>

          <div class="modal-footer">
            <button class="btn btn-outline" @click="dialogVisible = false">取消</button>
            <button class="btn btn-primary" @click="submitForm">{{ isEdit ? '保存修改' : '创建商户' }}</button>
          </div>
        </div>
      </div>
    </Teleport>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, reactive, computed } from 'vue'
import { getUserList, addUser, updateUser, userOp, getUserEdit } from '@/api/admin'
import { ElMessage, ElMessageBox } from 'element-plus'
import dayjs from 'dayjs'

interface User {
  uid: number
  gid: number
  username: string
  email: string
  phone: string
  qq: string
  url: string
  account: string
  money: number
  pay: number
  settle: number
  status: number
  mode: number
  addtime: string
}

interface Group {
  gid: number
  name: string
}

const users = ref<User[]>([])
const loading = ref(false)
const page = ref(1)
const total = ref(0)
const searchKeyword = ref('')

const dialogVisible = ref(false)
const dialogTitle = ref('添加商户')
const isEdit = ref(false)
const editingUser = ref<User | null>(null)

const userForm = reactive({
  gid: 1,
  phone: '',
  email: '',
  pwd: '',
  qq: '',
  url: '',
  settle_id: 1,
  account: '',
  username: '',
  mode: 0,
  pay: 1,
  settle: 1,
  status: 1
})

const groups = ref<Group[]>([
  { gid: 1, name: '默认组' }
])

const statusMap: Record<number, { text: string; class: string }> = {
  0: { text: '禁用', class: 'badge-danger' },
  1: { text: '正常', class: 'badge-success' },
  2: { text: '待审核', class: 'badge-warning' }
}

const payMap: Record<number, string> = { 0: '关闭', 1: '开启' }
const settleMap: Record<number, string> = { 0: '关闭', 1: '开启' }

const filteredUsers = computed(() => {
  if (!searchKeyword.value) return users.value
  const kw = searchKeyword.value.toLowerCase()
  return users.value.filter(u =>
    u.username?.toLowerCase().includes(kw) ||
    u.account?.toLowerCase().includes(kw) ||
    u.uid.toString().includes(kw)
  )
})

async function fetchUsers() {
  loading.value = true
  try {
    const res = await getUserList({ page: page.value, limit: 20 })
    if (res.code === 0) {
      users.value = res.data || []
      total.value = res.count || 0
    }
  } catch (error) {
    console.error('获取商户列表失败:', error)
  } finally {
    loading.value = false
  }
}

function formatTime(time: string) {
  return dayjs(time).format('YYYY-MM-DD HH:mm')
}

function openAddDialog() {
  isEdit.value = false
  dialogTitle.value = '添加商户'
  editingUser.value = null
  resetForm()
  dialogVisible.value = true
}

async function openEditDialog(user: User) {
  isEdit.value = true
  dialogTitle.value = '编辑商户'
  editingUser.value = user

  try {
    const res = await getUserEdit(user.uid)
    if (res.code === 0) {
      const editUser = res.user
      userForm.gid = editUser.gid
      userForm.phone = editUser.phone || ''
      userForm.email = editUser.email || ''
      userForm.qq = editUser.qq || ''
      userForm.url = editUser.url || ''
      userForm.settle_id = editUser.settle_id || 1
      userForm.account = editUser.account || ''
      userForm.username = editUser.username || ''
      userForm.mode = editUser.mode || 0
      userForm.pay = editUser.pay || 1
      userForm.settle = editUser.settle || 1
      userForm.status = editUser.status || 1
      userForm.pwd = ''
    }
  } catch (error) {
    console.error('获取商户信息失败:', error)
    ElMessage.error('获取商户信息失败')
  }

  dialogVisible.value = true
}

function resetForm() {
  userForm.gid = 1
  userForm.phone = ''
  userForm.email = ''
  userForm.pwd = ''
  userForm.qq = ''
  userForm.url = ''
  userForm.settle_id = 1
  userForm.account = ''
  userForm.username = ''
  userForm.mode = 0
  userForm.pay = 1
  userForm.settle = 1
  userForm.status = 1
}

async function submitForm() {
  if (!userForm.account || !userForm.username) {
    ElMessage.warning('结算账号和姓名不能为空')
    return
  }

  try {
    if (isEdit.value && editingUser.value) {
      await updateUser({
        uid: editingUser.value.uid,
        gid: userForm.gid,
        phone: userForm.phone,
        email: userForm.email,
        pwd: userForm.pwd,
        qq: userForm.qq,
        url: userForm.url,
        settle_id: userForm.settle_id,
        account: userForm.account,
        username: userForm.username,
        mode: userForm.mode,
        pay: userForm.pay,
        settle: userForm.settle,
        status: userForm.status
      })
      ElMessage.success('更新成功')
    } else {
      await addUser({
        gid: userForm.gid,
        phone: userForm.phone,
        email: userForm.email,
        pwd: userForm.pwd,
        qq: userForm.qq,
        url: userForm.url,
        settle_id: userForm.settle_id,
        account: userForm.account,
        username: userForm.username,
        mode: userForm.mode,
        pay: userForm.pay,
        settle: userForm.settle,
        status: userForm.status
      })
      ElMessage.success('添加成功')
    }
    dialogVisible.value = false
    fetchUsers()
  } catch (error) {
    console.error('操作失败:', error)
  }
}

async function resetKey(uid: number) {
  try {
    await userOp({ action: 'reset_key', uid })
    ElMessage.success('密钥已重置')
  } catch (error) {
    console.error('重置密钥失败:', error)
  }
}

async function deleteUser(uid: number) {
  try {
    await ElMessageBox.confirm('确定要删除该商户吗？', '提示', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    })
  } catch {
    return
  }
  try {
    await userOp({ action: 'delete', uid })
    ElMessage.success('删除成功')
    fetchUsers()
  } catch (error) {
    console.error('删除失败:', error)
  }
}

async function setStatus(uid: number, status: number) {
  try {
    await userOp({ action: 'set_status', uid, status })
    ElMessage.success('状态已更新')
    fetchUsers()
  } catch (error) {
    console.error('更新状态失败:', error)
  }
}

onMounted(() => {
  fetchUsers()
})
</script>

<style scoped>
.modal-overlay {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.5);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
  padding: 1rem;
}

.modal {
  background: white;
  border-radius: 0.75rem;
  width: 100%;
  max-width: 900px;
  max-height: 90vh;
  overflow: hidden;
  display: flex;
  flex-direction: column;
  box-shadow: 0 25px 50px -12px rgba(0, 0, 0, 0.25);
}

.modal-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  padding: 1.25rem 1.5rem;
  border-bottom: 1px solid #e5e7eb;
}

.modal-close-btn {
  width: 2rem;
  height: 2rem;
  display: flex;
  align-items: center;
  justify-content: center;
  background: #f3f4f6;
  border: none;
  border-radius: 0.5rem;
  font-size: 1.25rem;
  color: #6b7280;
  cursor: pointer;
  transition: all 0.15s;
}

.modal-close-btn:hover {
  background: #fee2e2;
  color: #ef4444;
}

.modal-body {
  padding: 1.5rem;
  overflow-y: auto;
  flex: 1;
}

.modal-footer {
  display: flex;
  justify-content: flex-end;
  gap: 0.75rem;
  padding: 1rem 1.5rem;
  border-top: 1px solid #e5e7eb;
  background: #f9fafb;
}

.form-section {
  background: #f9fafb;
  border-radius: 0.5rem;
  padding: 1rem;
}

.form-section-title {
  font-size: 0.875rem;
  font-weight: 600;
  color: #3b82f6;
  margin-bottom: 1rem;
  padding-bottom: 0.5rem;
  border-bottom: 1px solid #e5e7eb;
}

.form-label {
  display: block;
  font-size: 0.875rem;
  font-weight: 500;
  color: #374151;
  margin-bottom: 0.375rem;
}

.form-input {
  width: 100%;
  padding: 0.5rem 0.75rem;
  border: 1px solid #d1d5db;
  border-radius: 0.375rem;
  font-size: 0.875rem;
  transition: border-color 0.15s, box-shadow 0.15s;
}

.form-input:focus {
  outline: none;
  border-color: #3b82f6;
  box-shadow: 0 0 0 2px rgba(59, 130, 246, 0.1);
}
</style>
