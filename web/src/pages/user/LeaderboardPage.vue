<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { ElAvatar } from 'element-plus'
import { getLeaderboard } from '@/api/modules/quiz'
import type { LeaderboardEntry, MyRank } from '@/types/quiz'

const periods = [
  { label: '日榜', value: 'daily' },
  { label: '周榜', value: 'weekly' },
  { label: '月榜', value: 'monthly' },
  { label: '总榜', value: 'total' },
]

const activePeriod = ref('total')
const leaderboard = ref<LeaderboardEntry[]>([])
const myRank = ref<MyRank | null>(null)
const loading = ref(false)

async function fetchLeaderboard() {
  loading.value = true
  try {
    const res = await getLeaderboard({ period: activePeriod.value, page: 1, page_size: 20 })
    leaderboard.value = res.data.leaderboard || []
    myRank.value = res.data.my_rank || null
  } catch {
    // handled
  } finally {
    loading.value = false
  }
}

function switchPeriod(period: string) {
  activePeriod.value = period
  fetchLeaderboard()
}

onMounted(fetchLeaderboard)
</script>

<template>
  <div class="leaderboard-page">
    <h2>🏆 积分排行榜</h2>
    <p class="page-desc">看看谁是最强运河守护者</p>

    <div class="period-tabs">
      <span v-for="p in periods" :key="p.value" class="period-tab" :class="{ active: activePeriod === p.value }" @click="switchPeriod(p.value)">{{ p.label }}</span>
    </div>

    <div class="leaderboard-list" v-loading="loading">
      <div v-if="leaderboard.length === 0 && !loading" class="empty-hint">暂无排行数据，快来答题吧</div>
      <div v-for="(entry, i) in leaderboard" :key="entry.user.id" class="leader-item" :class="{ top: i < 3 }">
        <span class="rank" :class="`rank-${i + 1}`">{{ i + 1 }}</span>
        <el-avatar :size="40" :src="entry.user.avatar_url">{{ entry.user.nickname[0] }}</el-avatar>
        <div class="entry-info">
          <span class="entry-name">{{ entry.user.nickname || entry.user.username }}</span>
          <span class="entry-title">{{ entry.rank_title }}</span>
        </div>
        <div class="entry-stats">
          <span class="points">{{ entry.total_points }} 分</span>
          <span class="accuracy">{{ entry.answer_count }} 题 · 正确率 {{ Math.round(entry.accuracy * 100) }}%</span>
        </div>
      </div>
    </div>

    <div v-if="myRank" class="my-rank-bar">
      我的排名：第 <strong>{{ myRank.rank }}</strong> 名 · <strong>{{ myRank.total_points }}</strong> 分
    </div>
  </div>
</template>

<style scoped>
.leaderboard-page { max-width:700px;margin:0 auto }
.leaderboard-page h2 { font-size:22px;color:#303133 }
.page-desc { color:#909399;font-size:14px;margin-bottom:20px }
.period-tabs { display:flex;gap:4px;background:#f0f2f5;border-radius:8px;padding:4px;margin-bottom:24px }
.period-tab { flex:1;text-align:center;padding:6px 0;border-radius:6px;font-size:13px;color:#606266;cursor:pointer;transition:.2s;user-select:none }
.period-tab.active { background:#fff;color:#2c3e50;font-weight:500;box-shadow:0 1px 2px rgba(0,0,0,.06) }
.leaderboard-list { display:flex;flex-direction:column;gap:8px }
.leader-item { display:flex;align-items:center;gap:14px;padding:16px;background:#fff;border-radius:10px;box-shadow:0 1px 4px rgba(0,0,0,.04) }
.leader-item.top { background:#fef9e7 }
.rank { width:32px;height:32px;display:flex;align-items:center;justify-content:center;border-radius:50%;font-weight:700;font-size:14px;color:#909399;background:#f0f2f5 }
.rank-1 { background:#ffd700;color:#fff }
.rank-2 { background:#c0c0c0;color:#fff }
.rank-3 { background:#cd7f32;color:#fff }
.entry-info { flex:1 }
.entry-name { font-size:15px;color:#303133;font-weight:500 }
.entry-title { display:block;font-size:12px;color:#c0c4cc;margin-top:2px }
.entry-stats { text-align:right }
.points { display:block;font-size:16px;font-weight:700;color:#e6a23c }
.accuracy { font-size:12px;color:#c0c4cc }
.empty-hint { text-align:center;color:#c0c4cc;padding:40px 0 }
.my-rank-bar { margin-top:20px;padding:12px 20px;background:#eae7e0;border-radius:10px;text-align:center;font-size:15px;color:#303133 }
</style>
