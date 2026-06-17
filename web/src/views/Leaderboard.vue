<template>
  <div class="leaderboard-page">
    <h2>排行榜</h2>
    <el-radio-group v-model="period" @change="fetch" style="margin-bottom:20px">
      <el-radio-button value="total">总榜</el-radio-button>
      <el-radio-button value="weekly">周榜</el-radio-button>
      <el-radio-button value="daily">日榜</el-radio-button>
    </el-radio-group>

    <div class="rank-list" v-loading="loading">
      <div v-for="entry in entries" :key="entry.user_id" class="rank-item">
        <span class="rank-num" :class="'rank-' + entry.rank">
          {{ entry.rank <= 3 ? ['🥇','🥈','🥉'][entry.rank-1] : entry.rank }}
        </span>
        <span class="name">{{ entry.nickname || entry.user_id?.slice(0,8) }}</span>
        <span class="score">{{ entry.score }} 分</span>
      </div>
      <el-empty v-if="!loading && entries.length === 0" description="暂无数据" />
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { quizApi } from '@/api/quiz'
import type { LeaderboardEntry } from '@/types'

const period = ref('total')
const entries = ref<LeaderboardEntry[]>([])
const loading = ref(false)

async function fetch() {
  loading.value = true
  try {
    const res = await quizApi.getLeaderboard(period.value, 50)
    entries.value = res.data.data!
  } catch { /* ignore */ }
  finally { loading.value = false }
}

onMounted(fetch)
</script>

<style scoped>
.leaderboard-page { max-width: 600px; margin: 0 auto; }
.rank-item { display: flex; align-items: center; padding: 12px 16px; background: #fff; border-radius: 8px; margin-bottom: 8px; }
.rank-num { width: 40px; font-size: 18px; }
.rank-1, .rank-2, .rank-3 { font-size: 24px; }
.name { flex: 1; font-size: 16px; }
.score { color: #409eff; font-weight: bold; }
</style>
