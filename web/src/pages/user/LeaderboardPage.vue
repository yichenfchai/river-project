<script setup lang="ts">
const periods = [
  { label: '日榜', value: 'daily' },
  { label: '周榜', value: 'weekly' },
  { label: '月榜', value: 'monthly' },
  { label: '总榜', value: 'total' },
]
const leaderboard = [
  { name: '运河守望者老李', title: '钻石守护者', points: 8750, accuracy: 92 },
  { name: '生态小卫士', title: '黄金守护者', points: 6210, accuracy: 88 },
  { name: '历史迷小王', title: '黄金守护者', points: 5940, accuracy: 85 },
  { name: '水系专家', title: '白银守护者', points: 4120, accuracy: 79 },
  { name: '运河小舟', title: '白银守护者', points: 3880, accuracy: 76 },
]</script>

<template>
  <div class="leaderboard-page">
    <h2>🏆 积分排行榜</h2>
    <p class="page-desc">看看谁是最强运河守护者</p>

    <div class="period-tabs">
      <span v-for="p in periods" :key="p.value" class="period-tab" :class="{ active: p.value === 'total' }">{{ p.label }}</span>
    </div>

    <div class="leaderboard-list">
      <div v-for="(entry, i) in leaderboard" :key="i" class="leader-item" :class="{ top: i < 3 }">
        <span class="rank" :class="`rank-${i + 1}`">{{ i + 1 }}</span>
        <el-avatar :size="40">{{ entry.name[0] }}</el-avatar>
        <div class="entry-info">
          <span class="entry-name">{{ entry.name }}</span>
          <span class="entry-title">{{ entry.title }}</span>
        </div>
        <div class="entry-stats">
          <span class="points">{{ entry.points }} 分</span>
          <span class="accuracy">正确率 {{ entry.accuracy }}%</span>
        </div>
      </div>
    </div>
  </div>
</template>


<style scoped>
.leaderboard-page {
  max-width: 700px;
  margin: 0 auto;
}

.leaderboard-page h2 {
  font-size: 22px;
  color: #303133;
}

.page-desc {
  color: #909399;
  font-size: 14px;
  margin-bottom: 20px;
}

.period-tabs {
  display: flex;
  gap: 4px;
  background: #f0f2f5;
  border-radius: 8px;
  padding: 4px;
  margin-bottom: 24px;
}

.period-tab {
  flex: 1;
  text-align: center;
  padding: 6px 0;
  border-radius: 6px;
  font-size: 13px;
  color: #606266;
  cursor: pointer;
  transition: all 0.2s;
}

.period-tab.active {
  background: #fff;
  color: #2c3e50;
  font-weight: 500;
  box-shadow: 0 1px 2px rgba(0, 0, 0, 0.06);
}

.leaderboard-list {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.leader-item {
  display: flex;
  align-items: center;
  gap: 14px;
  padding: 16px;
  background: #fff;
  border-radius: 10px;
  box-shadow: 0 1px 4px rgba(0, 0, 0, 0.04);
}

.leader-item.top {
  background: #fef9e7;
}

.rank {
  width: 32px;
  height: 32px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 50%;
  font-weight: 700;
  font-size: 14px;
  color: #909399;
  background: #f0f2f5;
}

.rank-1 { background: #ffd700; color: #fff; }
.rank-2 { background: #c0c0c0; color: #fff; }
.rank-3 { background: #cd7f32; color: #fff; }

.entry-info {
  flex: 1;
}

.entry-name {
  font-size: 15px;
  color: #303133;
  font-weight: 500;
}

.entry-title {
  display: block;
  font-size: 12px;
  color: #c0c4cc;
  margin-top: 2px;
}

.entry-stats {
  text-align: right;
}

.points {
  display: block;
  font-size: 16px;
  font-weight: 700;
  color: #e6a23c;
}

.accuracy {
  font-size: 12px;
  color: #c0c4cc;
}
</style>
