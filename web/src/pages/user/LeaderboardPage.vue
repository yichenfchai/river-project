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
.leaderboard-page h2 { font-family: 'Noto Serif SC', 'STSong', serif; font-size:22px;color:var(--text-primary) }
.page-desc { color:var(--text-secondary);font-size:14px;margin-bottom:20px }
.period-tabs { display:flex;gap:4px;background:var(--paper-light);border-radius:10px;padding:4px;margin-bottom:24px }
.period-tab { flex:1;text-align:center;padding:10px 0;border-radius:8px;font-size:13px;color:var(--text-secondary);cursor:pointer;transition:.2s;user-select:none;min-height:44px;display:flex;align-items:center;justify-content:center }
.period-tab.active { background:var(--surface-card);color:var(--ink-primary);font-weight:600;box-shadow:var(--shadow-sm) }
.leaderboard-list { display:flex;flex-direction:column;gap:8px }
.leader-item { display:flex;align-items:center;gap:14px;padding:16px;background:var(--surface-card);border-radius:var(--radius-md);box-shadow:var(--shadow-xs);border:1px solid var(--border-hairline) }
.leader-item.top { background:linear-gradient(135deg,#fef9e7,#fdf4d6);border-color:rgba(201,184,150,0.3) }
.rank { width:36px;height:36px;display:flex;align-items:center;justify-content:center;border-radius:50%;font-weight:700;font-size:14px;color:var(--text-muted);background:var(--paper-light) }
.rank-1 { background:linear-gradient(135deg,#ffd700,#f0c000);color:#fff;box-shadow:0 2px 8px rgba(255,215,0,0.3);font-size:16px }
.rank-2 { background:linear-gradient(135deg,#c0c0c0,#a8a8a8);color:#fff;box-shadow:0 2px 6px rgba(192,192,192,0.3) }
.rank-3 { background:linear-gradient(135deg,#cd7f32,#b87028);color:#fff;box-shadow:0 2px 6px rgba(205,127,50,0.3) }
.entry-info { flex:1 }
.entry-name { font-size:15px;color:var(--text-primary);font-weight:500 }
.entry-title { display:block;font-size:12px;color:var(--text-muted);margin-top:2px }
.entry-stats { text-align:right }
.points { display:block;font-size:16px;font-weight:700;color:var(--gold-dark) }
.accuracy { font-size:12px;color:var(--text-muted) }
.empty-hint { text-align:center;color:var(--text-muted);padding:40px 0 }
.my-rank-bar { margin-top:20px;padding:14px 20px;background:var(--paper-light);border-radius:var(--radius-md);text-align:center;font-size:15px;color:var(--text-primary);border:1px solid var(--gold);border-left:4px solid var(--gold) }

@media (max-width: 768px) {
  .leaderboard-page { padding: 0 8px }
  .leader-item { padding: 12px; gap: 10px }
  .entry-stats { text-align: center }
  .points { font-size: 14px }
}

@media (max-width: 480px) {
  .leader-item { flex-wrap: wrap }
  .entry-stats { width: 100%; display: flex; justify-content: space-between; padding-top: 8px }
}
</style>
