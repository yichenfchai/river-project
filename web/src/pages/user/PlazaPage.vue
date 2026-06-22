<script setup lang="ts">
const topics = ['全部', '生态', '文化', '历史', '民俗', '环保']
const samplePosts = [
  { id: 1, author: '运河小舟', title: '今日运河拍到了白鹭群飞', content: '傍晚在扬州段运河边散步，偶遇一大群白鹭从水面掠过...', time: '2 小时前', likes: 42, comments: 7, views: 230 },
  { id: 2, author: '遗产守望者', title: '拱宸桥的历史文化价值', content: '作为京杭大运河终点的标志性建筑，拱宸桥承载了太多记忆...', time: '5 小时前', likes: 28, comments: 12, views: 186 },
  { id: 3, author: '水质监测员老张', title: '运河水质月度报告', content: '本月对通州段进行了水质采样分析，各项指标总体向好...', time: '昨天', likes: 56, comments: 23, views: 412 },
]</script>

<template>
  <div class="plaza-page">
    <div class="plaza-header">
      <h2>💬 分享广场</h2>
      <el-button type="primary">发布帖子</el-button>
    </div>
    <p class="page-desc">记录运河之美，分享生态观察与文化遗产保护心得</p>

    <div class="topic-filter">
      <span v-for="t in topics" :key="t" class="topic-tag" :class="{ active: t === '全部' }">{{ t }}</span>
    </div>

    <div class="post-list">
      <div v-for="post in samplePosts" :key="post.id" class="post-card">
        <div class="post-author">
          <el-avatar :size="40">{{ post.author[0] }}</el-avatar>
          <div class="author-info">
            <span class="author-name">{{ post.author }}</span>
            <span class="post-time">{{ post.time }}</span>
          </div>
        </div>
        <h3 class="post-title">{{ post.title }}</h3>
        <p class="post-content">{{ post.content }}</p>
        <div class="post-footer">
          <span>❤ {{ post.likes }}</span>
          <span>💬 {{ post.comments }}</span>
          <span>👁 {{ post.views }}</span>
        </div>
      </div>
    </div>
  </div>
</template>


<style scoped>
.plaza-page {
  max-width: 800px;
  margin: 0 auto;
}

.plaza-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.plaza-header h2 {
  font-size: 22px;
  color: #303133;
  margin: 0;
}

.page-desc {
  color: #909399;
  font-size: 14px;
  margin: 8px 0 20px;
}

.topic-filter {
  display: flex;
  gap: 8px;
  margin-bottom: 20px;
  flex-wrap: wrap;
}

.topic-tag {
  padding: 4px 16px;
  border-radius: 16px;
  background: #f0f2f5;
  color: #606266;
  font-size: 13px;
  cursor: pointer;
  transition: all 0.2s;
}

.topic-tag.active,
.topic-tag:hover {
  background: #2c3e50;
  color: #fff;
}

.post-list {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.post-card {
  background: #fff;
  border-radius: 12px;
  padding: 20px;
  box-shadow: 0 1px 4px rgba(0, 0, 0, 0.04);
  cursor: pointer;
  transition: box-shadow 0.2s;
}

.post-card:hover {
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.08);
}

.post-author {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-bottom: 12px;
}

.author-info {
  display: flex;
  flex-direction: column;
}

.author-name {
  font-size: 14px;
  color: #303133;
  font-weight: 500;
}

.post-time {
  font-size: 12px;
  color: #c0c4cc;
}

.post-title {
  font-size: 18px;
  margin: 0 0 8px;
  color: #303133;
}

.post-content {
  color: #606266;
  font-size: 14px;
  line-height: 1.6;
  margin: 0 0 16px;
}

.post-footer {
  display: flex;
  gap: 20px;
  color: #c0c4cc;
  font-size: 13px;
}
</style>
