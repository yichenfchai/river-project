<script setup lang="ts">
const sampleStories = [
  { id: 1, emoji: '🏰', topic: '历史', title: '隋炀帝与通济渠', excerpt: '公元605年，隋炀帝下令开凿连通黄河与淮河的通济渠...', age: '青少年', likes: 128 },
  { id: 2, emoji: '🦅', topic: '生态', title: '运河边的白鹭家族', excerpt: '在扬州段的运河畔，生活着一个庞大的白鹭家族...', age: '儿童', likes: 95 },
  { id: 3, emoji: '🏗', topic: '工程', title: '古代船闸的秘密', excerpt: '京口闸、瓜洲闸——古人如何让船翻山越岭？', age: '成人', likes: 210 },
  { id: 4, emoji: '🎭', topic: '民俗', title: '船工号子的回响', excerpt: '运河两岸苍凉的船工号子，是非遗中的活态史诗', age: '成人', likes: 73 },
]</script>

<template>
  <div class="story-page">
    <h2>📖 科普故事会</h2>
    <p class="page-desc">AI 驱动的运河文化科普内容 —— 跨越 2500 年的运河故事</p>

    <div class="story-grid">
      <div v-for="story in sampleStories" :key="story.id" class="story-card">
        <div class="story-cover">{{ story.emoji }}</div>
        <div class="story-info">
          <span class="story-tag">{{ story.topic }}</span>
          <h3>{{ story.title }}</h3>
          <p>{{ story.excerpt }}</p>
          <div class="story-meta">
            <span>{{ story.age }}</span>
            <span>{{ story.likes }} 赞</span>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>


<style scoped>
.story-page {
  max-width: 1200px;
  margin: 0 auto;
}

.story-page h2 {
  font-size: 22px;
  color: #303133;
}

.page-desc {
  color: #909399;
  font-size: 14px;
  margin-bottom: 24px;
}

.story-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 20px;
}

.story-card {
  display: flex;
  background: #fff;
  border-radius: 12px;
  overflow: hidden;
  box-shadow: 0 1px 4px rgba(0, 0, 0, 0.06);
  cursor: pointer;
  transition: box-shadow 0.2s;
}

.story-card:hover {
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
}

.story-cover {
  width: 120px;
  min-height: 140px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 48px;
  background: #fff;
  flex-shrink: 0;
}

.story-info {
  padding: 16px 20px;
  flex: 1;
}

.story-tag {
  font-size: 11px;
  padding: 2px 8px;
  background: #eae7e0;
  color: #2c3e50;
  border-radius: 4px;
}

.story-info h3 {
  margin: 8px 0;
  font-size: 16px;
  color: #303133;
}

.story-info p {
  color: #909399;
  font-size: 13px;
  line-height: 1.6;
  margin: 0;
}

.story-meta {
  margin-top: 12px;
  display: flex;
  gap: 16px;
  font-size: 12px;
  color: #c0c4cc;
}
</style>
