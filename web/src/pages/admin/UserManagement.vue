<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { ElTable, ElTableColumn, ElTag, ElButton, ElPagination, ElInput, ElSelect, ElOption, ElMessageBox, ElMessage } from 'element-plus'
import { Search } from '@element-plus/icons-vue'
import { getUsers, updateUserRole, banUser } from '@/api/modules/admin'
import type { User, UserRole, Pagination } from '@/types'

const users = ref<User[]>([])
const pagination = ref<Pagination>({ page: 1, page_size: 10, total: 0, total_pages: 0 })
const loading = ref(false)
const filters = ref({ keyword: '', role: '' as UserRole | '' })

async function fetchUsers(page = 1) {
  loading.value = true
  try {
    const res = await getUsers({ page, page_size: 10, ...(filters.value.role ? { role: filters.value.role } : {}), keyword: filters.value.keyword || undefined })
    users.value = res.data.users
    pagination.value = res.data.pagination
  } catch {
    // handled by interceptor
  } finally {
    loading.value = false
  }
}

async function handleRoleChange(row: User, newRole: UserRole) {
  try {
    await ElMessageBox.confirm(`确认将 ${row.nickname || row.username} 的角色改为 ${newRole}？`, '修改角色', { type: 'warning' })
    await updateUserRole(row.id, newRole)
    row.role = newRole
    ElMessage.success('角色修改成功')
  } catch {
    // cancelled
  }
}

async function handleBan(row: User) {
  try {
    await ElMessageBox.prompt('请输入封禁原因', '封禁用户', { type: 'warning', inputPlaceholder: '违反社区规则等' })
    await banUser(row.id, { banned: true, reason: '管理员操作' })
    ElMessage.success('用户已封禁')
  } catch {
    // cancelled
  }
}

function onSearch() {
  fetchUsers(1)
}

onMounted(() => fetchUsers())
</script>

<template>
  <div class="user-management">
    <h2>👥 用户管理</h2>

    <div class="search-bar">
      <el-input v-model="filters.keyword" placeholder="搜索用户名/邮箱" clearable class="search-input" :prefix-icon="Search" @keyup.enter="onSearch" />
      <el-select v-model="filters.role" placeholder="角色筛选" clearable @change="onSearch" style="width: 140px">
        <el-option label="普通用户" value="user" />
        <el-option label="监测人员" value="monitor" />
        <el-option label="管理员" value="admin" />
      </el-select>
      <el-button type="primary" @click="onSearch">搜索</el-button>
    </div>

    <div class="table-wrap">
      <el-table :data="users" v-loading="loading" stripe>
        <el-table-column prop="username" label="用户名" width="120" />
        <el-table-column prop="nickname" label="昵称" width="120" />
        <el-table-column prop="email" label="邮箱" min-width="180" />
        <el-table-column prop="role" label="角色" width="100">
          <template #default="{ row }">
            <el-tag size="small" :type="row.role === 'admin' ? 'danger' : row.role === 'monitor' ? 'warning' : 'info'">
              {{ row.role === 'admin' ? '管理员' : row.role === 'monitor' ? '监测人员' : '普通用户' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="created_at" label="注册时间" width="180" />
        <el-table-column label="操作" width="220" fixed="right">
          <template #default="{ row }">
            <el-button size="small" @click="handleRoleChange(row as User, 'user')" :disabled="(row as User).role === 'user'">设为普通用户</el-button>
            <el-button size="small" type="warning" @click="handleRoleChange(row as User, 'monitor')" :disabled="(row as User).role === 'monitor'">设为监测员</el-button>
            <el-button size="small" type="danger" @click="handleBan(row as User)">封禁</el-button>
          </template>
        </el-table-column>
      </el-table>
    </div>

    <div class="pagination-wrap" v-if="pagination.total_pages > 1">
      <el-pagination
        background
        layout="prev, pager, next"
        :total="pagination.total"
        :page-size="pagination.page_size"
        :current-page="pagination.page"
        @current-change="fetchUsers"
      />
    </div>
  </div>
</template>

<style scoped>
.user-management {
  max-width: 1200px;
}

.user-management h2 {
  font-size: 22px;
  color: #303133;
  margin: 0 0 20px;
}

.search-bar {
  display: flex;
  gap: 10px;
  margin-bottom: 16px;
}

.search-input {
  width: 280px;
}

.table-wrap {
  background: #fff;
  border-radius: 10px;
  overflow: hidden;
  box-shadow: 0 1px 4px rgba(0, 0, 0, 0.04);
}

.pagination-wrap {
  margin-top: 20px;
  display: flex;
  justify-content: center;
}
</style>
