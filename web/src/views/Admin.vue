<template>
  <div class="admin-page">
    <h2>管理后台</h2>
    <el-table :data="users" v-loading="loading" stripe>
      <el-table-column prop="username" label="用户名" />
      <el-table-column prop="nickname" label="昵称" />
      <el-table-column prop="email" label="邮箱" />
      <el-table-column prop="role" label="角色" width="100">
        <template #default="{ row }">
          <el-tag :type="row.role === 'admin' ? 'danger' : row.role === 'monitor' ? 'warning' : ''" size="small">
            {{ row.role }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="status" label="状态" width="80">
        <template #default="{ row }">
          <el-tag :type="row.status === 'banned' ? 'danger' : 'success'" size="small">
            {{ row.status === 'banned' ? '封禁' : '正常' }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column label="操作" width="200">
        <template #default="{ row }">
          <el-select v-model="row._newRole" placeholder="改角色" size="small" @change="(v: string) => changeRole(row, v)" style="width:90px">
            <el-option label="user" value="user" />
            <el-option label="monitor" value="monitor" />
            <el-option label="admin" value="admin" />
          </el-select>
          <el-button
            :type="row.status === 'banned' ? 'success' : 'danger'"
            size="small"
            @click="toggleBan(row)"
          >
            {{ row.status === 'banned' ? '解封' : '封禁' }}
          </el-button>
        </template>
      </el-table-column>
    </el-table>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { adminApi } from '@/api/admin'
import { ElMessage } from 'element-plus'
import type { User } from '@/types'

interface UserEx extends User { _newRole?: string }
const users = ref<UserEx[]>([])
const loading = ref(false)

async function fetchUsers() {
  loading.value = true
  try {
    const res = await adminApi.listUsers(1, 100)
    users.value = (res.data.data!.items as UserEx[]).map(u => ({ ...u, _newRole: u.role }))
  } catch { /* ignore */ }
  finally { loading.value = false }
}

async function changeRole(user: UserEx, role: string) {
  try {
    await adminApi.changeRole(user.id, role)
    user.role = role as User['role']
    ElMessage.success('角色已更新')
  } catch { ElMessage.error('操作失败') }
}

async function toggleBan(user: UserEx) {
  const ban = user.status !== 'banned'
  try {
    await adminApi.banUser(user.id, ban)
    user.status = ban ? 'banned' : 'active'
    ElMessage.success(ban ? '已封禁' : '已解封')
  } catch { ElMessage.error('操作失败') }
}

onMounted(fetchUsers)
</script>

<style scoped>
.admin-page { max-width: 1000px; margin: 0 auto; }
</style>
