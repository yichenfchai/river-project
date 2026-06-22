<script setup lang="ts">
const categories = [
  { label: '全部', value: 'all' },
  { label: '历史', value: 'history' },
  { label: '生态', value: 'ecology' },
  { label: '地理', value: 'geography' },
  { label: '水利', value: 'water' },
]
const ranks = [
  { icon: '🥉', name: '青铜守护者 0-500', color: '#cd7f32' },
  { icon: '🥈', name: '白银守护者 500-1500', color: '#c0c0c0' },
  { icon: '🥇', name: '黄金守护者 1500-5000', color: '#ffd700' },
  { icon: '💎', name: '钻石守护者 5000-15000', color: '#2c3e50' },
  { icon: '👑', name: '运河守护者 15000+', color: '#e6a23c' },
]</script>

<template>
  <div class="quiz-page">
    <h2>❓ 趣味问答</h2>
    <p class="page-desc">测试你对大运河的知识储备，赢取积分，争夺守护者称号</p>

    <div class="quiz-categories">
      <el-button v-for="c in categories" :key="c.value" :type="c.value === 'all' ? 'primary' : 'default'" size="large">
        {{ c.label }}
      </el-button>
    </div>

    <div class="quiz-card">
      <div class="quiz-question">
        <span class="difficulty easy">简单</span>
        <span class="category-tag">历史</span>
        <h3>京杭大运河全长约多少公里？</h3>
      </div>
      <div class="quiz-options">
        <div v-for="(opt, i) in ['A. 1200 公里', 'B. 1794 公里', 'C. 2500 公里', 'D. 3200 公里']" :key="i" class="option-item">
          {{ opt }}
        </div>
      </div>
      <div class="quiz-action">
        <el-button type="primary" size="large">提交答案</el-button>
        <span class="hint">答题获得积分，连续答对有加成！</span>
      </div>
    </div>

    <div class="rank-preview">
      <h3>段位体系</h3>
      <div class="rank-list">
        <span v-for="r in ranks" :key="r.name" class="rank-badge" :style="{ color: r.color, borderColor: r.color }">
          {{ r.icon }} {{ r.name }}
        </span>
      </div>
    </div>
  </div>
</template>


<style scoped>
.quiz-page {
  max-width: 800px;
  margin: 0 auto;
}

.quiz-page h2 {
  font-size: 22px;
  color: #303133;
}

.page-desc {
  color: #909399;
  font-size: 14px;
  margin-bottom: 20px;
}

.quiz-categories {
  display: flex;
  gap: 10px;
  margin-bottom: 24px;
  flex-wrap: wrap;
}

.quiz-card {
  background: #fff;
  border-radius: 12px;
  padding: 32px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.06);
}

.quiz-question {
  margin-bottom: 24px;
}

.difficulty {
  display: inline-block;
  font-size: 11px;
  padding: 2px 8px;
  border-radius: 4px;
  margin-right: 8px;
}

.difficulty.easy { background: #f0f9eb; color: #67c23a; }
.difficulty.medium { background: #fdf6ec; color: #e6a23c; }
.difficulty.hard { background: #fef0f0; color: #f56c6c; }

.category-tag {
  font-size: 11px;
  padding: 2px 8px;
  background: #eae7e0;
  color: #2c3e50;
  border-radius: 4px;
}

.quiz-question h3 {
  margin-top: 12px;
  font-size: 20px;
  color: #303133;
}

.quiz-options {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.option-item {
  padding: 14px 18px;
  border: 1px solid #e4e7ed;
  border-radius: 8px;
  cursor: pointer;
  transition: all 0.2s;
  font-size: 15px;
}

.option-item:hover {
  border-color: #2c3e50;
  background: #eae7e0;
}

.quiz-action {
  margin-top: 24px;
  display: flex;
  align-items: center;
  gap: 16px;
}

.hint {
  color: #909399;
  font-size: 13px;
}

.rank-preview {
  margin-top: 32px;
  text-align: center;
}

.rank-preview h3 {
  color: #303133;
  margin-bottom: 12px;
}

.rank-list {
  display: flex;
  gap: 8px;
  justify-content: center;
  flex-wrap: wrap;
}

.rank-badge {
  padding: 4px 12px;
  border: 1px solid;
  border-radius: 16px;
  font-size: 12px;
}
</style>
