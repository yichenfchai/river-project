<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { ElTable, ElTableColumn, ElTag, ElButton, ElPagination, ElDialog, ElForm, ElFormItem, ElInput, ElInputNumber, ElSwitch, ElMessage, ElMessageBox, ElIcon } from 'element-plus'
import { Plus, Edit, Delete } from '@element-plus/icons-vue'
import { getShopItems, createShopItem, updateShopItem, deleteShopItem } from '@/api/modules/shop'
import type { ShopItem } from '@/api/modules/shop'

const items = ref<ShopItem[]>([])
const loading = ref(false)
const dialogVisible = ref(false)
const dialogTitle = ref('新增商品')
const formLoading = ref(false)
const editingId = ref<string | null>(null)
const form = ref({
  name: '',
  description: '',
  image_url: '',
  points_cost: 50,
  stock: -1,
  is_active: true,
})

async function fetchItems() {
  loading.value = true
  try {
    const res = await getShopItems()
    items.value = res.data
  } catch {
    // handled by interceptor
  } finally {
    loading.value = false
  }
}

function openCreate() {
  editingId.value = null
  dialogTitle.value = '新增商品'
  form.value = { name: '', description: '', image_url: '', points_cost: 50, stock: -1, is_active: true }
  dialogVisible.value = true
}

function openEdit(item: ShopItem) {
  editingId.value = item.id
  dialogTitle.value = '编辑商品'
  form.value = {
    name: item.name,
    description: item.description,
    image_url: item.image_url,
    points_cost: item.points_cost,
    stock: item.stock,
    is_active: item.is_active,
  }
  dialogVisible.value = true
}

async function handleSubmit() {
  formLoading.value = true
  try {
    if (editingId.value) {
      await updateShopItem(editingId.value, {
        name: form.value.name,
        description: form.value.description,
        image_url: form.value.image_url,
        points_cost: form.value.points_cost,
        stock: form.value.stock,
        is_active: form.value.is_active,
      })
      ElMessage.success('商品已更新')
    } else {
      await createShopItem({
        name: form.value.name,
        description: form.value.description,
        image_url: form.value.image_url,
        points_cost: form.value.points_cost,
        stock: form.value.stock,
        is_active: true,
      })
      ElMessage.success('商品已创建')
    }
    dialogVisible.value = false
    fetchItems()
  } catch {
    // handled by interceptor
  } finally {
    formLoading.value = false
  }
}

async function handleDelete(item: ShopItem) {
  try {
    await ElMessageBox.confirm(`确认删除「${item.name}」？此操作不可恢复。`, '删除商品', { type: 'warning' })
    await deleteShopItem(item.id)
    ElMessage.success('已删除')
    fetchItems()
  } catch {
    // cancelled
  }
}

function stockText(stock: number) {
  if (stock === -1) return '不限量'
  return `剩余 ${stock}`
}

onMounted(fetchItems)
</script>

<template>
  <div class="shop-management">
    <div class="page-header">
      <h2>🏪 兑换商店管理</h2>
      <el-button type="primary" :icon="Plus" @click="openCreate">新增商品</el-button>
    </div>

    <el-table :data="items" v-loading="loading" stripe row-key="id" style="margin-top: 16px">
      <el-table-column prop="name" label="商品名称" min-width="180" />
      <el-table-column prop="description" label="描述" min-width="200" show-overflow-tooltip />
      <el-table-column prop="points_cost" label="所需积分" width="100" align="center">
        <template #default="{ row }">🪙 {{ row.points_cost }}</template>
      </el-table-column>
      <el-table-column label="库存" width="100" align="center">
        <template #default="{ row }">
          <el-tag :type="row.stock === 0 ? 'danger' : row.stock === -1 ? 'info' : 'warning'" size="small">
            {{ stockText(row.stock) }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column label="状态" width="90" align="center">
        <template #default="{ row }">
          <el-tag :type="row.is_active ? 'success' : 'info'" size="small">
            {{ row.is_active ? '上架' : '下架' }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column label="操作" width="160" align="center" fixed="right">
        <template #default="{ row }">
          <el-button type="primary" link :icon="Edit" size="small" @click="openEdit(row as ShopItem)">编辑</el-button>
          <el-button type="danger" link :icon="Delete" size="small" @click="handleDelete(row as ShopItem)">删除</el-button>
        </template>
      </el-table-column>
    </el-table>

    <!-- Dialog -->
    <el-dialog v-model="dialogVisible" :title="dialogTitle" width="500px" destroy-on-close>
      <el-form :model="form" label-width="80px">
        <el-form-item label="名称" required>
          <el-input v-model="form.name" placeholder="含 emoji，如 🏅 运河守护者称号" />
        </el-form-item>
        <el-form-item label="描述">
          <el-input v-model="form.description" type="textarea" :rows="2" placeholder="商品描述" />
        </el-form-item>
        <el-form-item label="所需积分" required>
          <el-input-number v-model="form.points_cost" :min="1" :max="99999" />
        </el-form-item>
        <el-form-item label="库存">
          <el-input-number v-model="form.stock" :min="-1" :max="9999" />
          <span style="margin-left: 8px; color: #909399; font-size: 12px">-1 表示不限量</span>
        </el-form-item>
        <el-form-item v-if="editingId" label="上架">
          <el-switch v-model="form.is_active" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="formLoading" @click="handleSubmit">
          {{ editingId ? '保存' : '创建' }}
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<style scoped>
.shop-management {
  max-width: 1200px;
}

.page-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.page-header h2 {
  margin: 0;
  font-size: 22px;
}
</style>
