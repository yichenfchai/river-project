<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { ElCard, ElButton, ElTag, ElPagination, ElEmpty, ElMessage, ElMessageBox, ElIcon } from 'element-plus'
import { Present, Coin } from '@element-plus/icons-vue'
import { getShopItems, redeemItem, getHistory } from '@/api/modules/shop'
import { getMyProfile } from '@/api/modules/auth'
import { useAuthStore } from '@/stores/auth'
import type { ShopItem, Redemption } from '@/api/modules/shop'

const auth = useAuthStore()
const userPoints = ref(0)
const items = ref<ShopItem[]>([])
const loading = ref(false)
const redeeming = ref<string | null>(null)

const history = ref<Redemption[]>([])
const historyTotal = ref(0)
const historyPage = ref(1)
const showHistory = ref(false)
const historyLoading = ref(false)

async function fetchItems() {
  loading.value = true
  try {
    const res = await getShopItems()
    items.value = res.data
  } catch {
    items.value = []
  } finally {
    loading.value = false
  }
  if (auth.isLoggedIn && !auth.isGuest) {
    try {
      const profile = await getMyProfile()
      userPoints.value = profile.data.points
    } catch {
      // ignore
    }
  }
}

async function fetchHistory(page = 1) {
  historyLoading.value = true
  try {
    const res = await getHistory(page, 10)
    history.value = res.data.items
    historyTotal.value = res.data.pagination.total
    historyPage.value = page
  } catch {
    // handled by interceptor
  } finally {
    historyLoading.value = false
  }
}

async function handleRedeem(item: ShopItem) {
  if (auth.isGuest) {
    ElMessage.warning('请先登录后兑换')
    return
  }
  try {
    await ElMessageBox.confirm(
      `确认使用 ${item.points_cost} 积分兑换「${item.name}」？`,
      '确认兑换',
      { type: 'info', confirmButtonText: '确认兑换' },
    )
  } catch {
    return
  }

  redeeming.value = item.id
  try {
    const res =     await redeemItem(item.id)
    userPoints.value = res.data.user_points
    ElMessage.success(`兑换成功！「${item.name.slice(2)}」已获得，剩余积分 ${res.data.user_points}`)
  } catch {
    ElMessage.error('兑换失败，可能积分不足或商品已售罄')
  } finally {
    redeeming.value = null
  }
}

function toggleHistory() {
  showHistory.value = !showHistory.value
  if (showHistory.value) {
    fetchHistory(1)
  }
}

const stockLabel = computed(() => (stock: number) => {
  if (stock === -1) return '不限量'
  if (stock === 0) return '已售罄'
  return `剩余 ${stock}`
})

onMounted(fetchItems)
</script>

<template>
  <div class="shop-page">
    <div class="shop-header">
      <h2><el-icon :size="24"><Present /></el-icon> 积分兑换商店</h2>
      <div class="header-actions">
        <span v-if="auth.isLoggedIn && !auth.isGuest" class="my-points">
          <el-icon><Coin /></el-icon>
          我的积分：<strong>{{ userPoints }}</strong>
        </span>
        <el-button v-if="!auth.isGuest" text type="primary" @click="toggleHistory">
          {{ showHistory ? '返回商品' : '兑换记录' }}
        </el-button>
      </div>
    </div>

    <p class="shop-desc">用答题赢取的积分兑换专属称号、头像框和隐藏内容</p>

    <!-- Guest tip -->
    <div v-if="auth.isGuest" class="guest-tip">
      <span>游客模式下可以浏览商品，登录后即可兑换</span>
    </div>

    <!-- Items grid -->
    <div v-if="!showHistory" class="shop-grid" v-loading="loading">
      <el-empty v-if="!loading && items.length === 0" description="暂无商品" />
      <div v-for="item in items" :key="item.id" class="shop-item">
        <el-card shadow="hover" :body-style="{ padding: '0' }">
          <div class="item-body">
            <div class="item-icon">{{ item.name.slice(0, 2) }}</div>
            <h3 class="item-name">{{ item.name.slice(2) }}</h3>
            <p class="item-desc">{{ item.description }}</p>
            <div class="item-meta">
              <el-tag :type="item.stock === 0 ? 'danger' : 'success'" size="small">
                {{ stockLabel(item.stock) }}
              </el-tag>
              <span class="item-cost">
                <el-icon><Coin /></el-icon> {{ item.points_cost }}
              </span>
            </div>
          </div>
          <div class="item-footer">
            <el-button
              type="primary"
              :disabled="item.stock === 0 || auth.isGuest"
              :loading="redeeming === item.id"
              size="default"
              @click="handleRedeem(item)"
            >
              {{ auth.isGuest ? '登录后兑换' : item.stock === 0 ? '已售罄' : '立即兑换' }}
            </el-button>
          </div>
        </el-card>
      </div>
    </div>

    <!-- History -->
    <div v-else class="history-section" v-loading="historyLoading">
      <el-empty v-if="!historyLoading && history.length === 0" description="暂无兑换记录" />
      <div v-for="r in history" :key="r.id" class="history-item">
        <div class="history-left">
          <span class="history-name">{{ r.item_name.slice(2) }}</span>
          <span class="history-date">{{ new Date(r.created_at).toLocaleString('zh-CN') }}</span>
        </div>
        <div class="history-right">
          <el-tag size="small" type="warning">{{ r.points_spent }} 积分</el-tag>
        </div>
      </div>
      <el-pagination
        v-if="historyTotal > 10"
        v-model:current-page="historyPage"
        :page-size="10"
        :total="historyTotal"
        layout="prev, pager, next"
        small
        @current-change="fetchHistory"
        style="margin-top: 16px; justify-content: center"
      />
    </div>
  </div>
</template>

<style scoped>
.shop-page {
  max-width: 1100px;
  margin: 0 auto;
}

.shop-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 8px;
}

.shop-header h2 {
  margin: 0;
  font-size: 22px;
  display: flex;
  align-items: center;
  gap: 8px;
  color: #303133;
}

.header-actions {
  display: flex;
  align-items: center;
  gap: 16px;
}

.my-points {
  display: flex;
  align-items: center;
  gap: 4px;
  font-size: 14px;
  color: #606266;
}

.my-points strong {
  color: #e6a23c;
  font-size: 16px;
}

.shop-desc {
  color: #909399;
  font-size: 14px;
  margin: 0 0 24px;
}

.guest-tip {
  background: #fdf6ec;
  border: 1px solid #faecd8;
  border-radius: 8px;
  padding: 12px 20px;
  color: #e6a23c;
  font-size: 14px;
  margin-bottom: 20px;
}

.shop-grid {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 20px;
}

.shop-item :deep(.el-card) {
  border-radius: 12px;
  overflow: hidden;
  transition: transform 0.2s, box-shadow 0.2s;
}

.shop-item :deep(.el-card:hover) {
  transform: translateY(-4px);
}

.item-body {
  padding: 24px 20px 16px;
  text-align: center;
}

.item-icon {
  font-size: 40px;
  margin-bottom: 8px;
}

.item-name {
  margin: 0 0 8px;
  font-size: 15px;
  color: #303133;
}

.item-desc {
  font-size: 12px;
  color: #909399;
  margin: 0 0 16px;
  line-height: 1.5;
  min-height: 36px;
}

.item-meta {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 12px;
}

.item-cost {
  display: flex;
  align-items: center;
  gap: 2px;
  font-size: 18px;
  font-weight: 700;
  color: #e6a23c;
}

.item-footer {
  padding: 12px 20px;
  border-top: 1px solid #ebeef5;
  text-align: center;
}

.item-footer .el-button {
  width: 100%;
}

.history-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 14px 20px;
  background: #fff;
  border-radius: 8px;
  margin-bottom: 8px;
  border: 1px solid #ebeef5;
}

.history-left {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.history-name {
  font-size: 14px;
  color: #303133;
  font-weight: 500;
}

.history-date {
  font-size: 12px;
  color: #909399;
}

@media (max-width: 768px) {
  .shop-grid {
    grid-template-columns: repeat(2, 1fr);
  }
}

@media (max-width: 480px) {
  .shop-grid {
    grid-template-columns: 1fr;
  }
}
</style>
