<template>
  <div class="home">
    <div class="hero">
      <h1>🏯 大运河生态与文化保护平台</h1>
      <p>守护千年运河，传承中华文明</p>
      <div class="hero-actions">
        <el-button type="primary" size="large" @click="$router.push('/quiz')">开始答题</el-button>
        <el-button size="large" @click="$router.push('/posts')">逛分享广场</el-button>
      </div>
    </div>

    <el-row :gutter="20" class="cards">
      <el-col :span="8">
        <el-card shadow="hover" @click="$router.push('/posts')">
          <el-icon :size="40" color="#409eff"><ChatDotRound /></el-icon>
          <h3>分享广场</h3>
          <p>分享大运河的美景、故事和生态观察</p>
        </el-card>
      </el-col>
      <el-col :span="8">
        <el-card shadow="hover" @click="$router.push('/quiz')">
          <el-icon :size="40" color="#67c23a"><Trophy /></el-icon>
          <h3>趣味问答</h3>
          <p>挑战运河知识，赢取积分，冲击排行榜</p>
        </el-card>
      </el-col>
      <el-col :span="8">
        <el-card shadow="hover" @click="$router.push('/leaderboard')">
          <el-icon :size="40" color="#e6a23c">< Medal /></el-icon>
          <h3>排行榜</h3>
          <p>看看谁是运河守护者中的佼佼者</p>
        </el-card>
      </el-col>
    </el-row>
  </div>
</template>

<script setup lang="ts">
</script>

<style scoped>
.hero { text-align: center; padding: 60px 20px; }
.hero h1 { font-size: 32px; margin-bottom: 12px; color: #303133; }
.hero p { font-size: 16px; color: #909399; margin-bottom: 24px; }
.hero-actions { display: flex; gap: 12px; justify-content: center; }
.cards { margin-top: 40px; }
.cards .el-card { text-align: center; cursor: pointer; padding: 24px; transition: transform .2s; }
.cards .el-card:hover { transform: translateY(-4px); }
.cards h3 { margin: 12px 0 8px; }
.cards p { color: #909399; font-size: 14px; }
</style>
