<script setup lang="ts">
import { ref, onMounted, onUnmounted, nextTick } from 'vue'
import { ElCard, ElButton } from 'element-plus'
import { useRouter } from 'vue-router'

const router = useRouter()

/* ---- Banner Carousel ---- */
const slides = [
  {
    bg: 'linear-gradient(135deg, #0a1620 0%, #152d3a 40%, #1a5276 70%, #2c6e8e 100%)',
    bgImage: '/images/banner.jpg',
    title: '守护千年运河',
    subtitle: '传承生态文化',
    desc: '京杭大运河生态与文化保护公众参与平台',
  },
  {
    bg: 'linear-gradient(135deg, #1a3528 0%, #2d5040 40%, #4a6278 70%, #2c3e50 100%)',
    bgImage: '/images/banner.jpg',
    title: '探索时空地图',
    subtitle: '见证运河变迁',
    desc: '从隋唐至今，感受千年运河的历史脉络',
  },
  {
    bg: 'linear-gradient(135deg, #1a2532 0%, #2c3e50 40%, #4a6278 70%, #2c3e50 100%)',
    bgImage: '/images/banner.jpg',
    title: '趣味问答挑战',
    subtitle: '争当运河守护者',
    desc: '答题赢积分，解锁专属称号与成就',
  },
  {
    bg: 'linear-gradient(135deg, #1c2e28 0%, #2d4540 40%, #4a6278 70%, #2c3e50 100%)',
    bgImage: '/images/banner.jpg',
    title: '分享运河故事',
    subtitle: '记录美好时刻',
    desc: '发布图文，与数万用户共同守护运河文化',
  },
]

const currentSlide = ref(0)
const isTransitioning = ref(false)
let autoPlayTimer: ReturnType<typeof setInterval> | null = null

function nextSlide() {
  if (isTransitioning.value) return
  isTransitioning.value = true
  currentSlide.value = (currentSlide.value + 1) % slides.length
  setTimeout(() => { isTransitioning.value = false }, 800)
}

function goToSlide(index: number) {
  if (isTransitioning.value || index === currentSlide.value) return
  isTransitioning.value = true
  currentSlide.value = index
  setTimeout(() => { isTransitioning.value = false }, 800)
}

function startAutoPlay() {
  stopAutoPlay()
  autoPlayTimer = setInterval(nextSlide, 5000)
}

function stopAutoPlay() {
  if (autoPlayTimer) {
    clearInterval(autoPlayTimer)
    autoPlayTimer = null
  }
}

onMounted(() => {
  startAutoPlay()
})

onUnmounted(() => {
  stopAutoPlay()
})

/* ---- Scroll Reveal ---- */
const homeRef = ref<HTMLElement | null>(null)

function initScrollReveal() {
  if (!homeRef.value) return
  const observer = new IntersectionObserver(
    (entries) => {
      for (const entry of entries) {
        const el = entry.target as HTMLElement
        const animName = el.dataset.reveal || 'fade-up'
        if (entry.isIntersecting) {
          el.classList.add('reveal-active', `${animName}-active`)
          el.classList.remove(animName)
        } else {
          el.classList.remove('reveal-active', `${animName}-active`)
          el.classList.add(animName)
        }
      }
    },
    { threshold: 0.15, rootMargin: '0px 0px -50px 0px' },
  )
  const elements = homeRef.value.querySelectorAll('[data-reveal]')
  elements.forEach((el) => {
    el.classList.add('reveal')
    el.classList.add(el.getAttribute('data-reveal')!)
    observer.observe(el)
  })
}

/* ---- Animated Stats Counter ---- */
const stats = [
  { label: '注册用户', target: 12580, suffix: '' },
  { label: '运河 POI', target: 3200, suffix: '+' },
  { label: '分享帖子', target: 8900, suffix: '' },
  { label: '科普故事', target: 560, suffix: '+' },
]

const animatedStats = ref(stats.map(() => 0))

function animateCounters() {
  stats.forEach((stat, i) => {
    const duration = 2000
    const step = Math.ceil(stat.target / (duration / 16))
    const timer = setInterval(() => {
      animatedStats.value[i] = Math.min((animatedStats.value[i] ?? 0) + step, stat.target)
      if ((animatedStats.value[i] ?? 0) >= stat.target) {
        animatedStats.value[i] = stat.target
        clearInterval(timer)
      }
    }, 16)
  })
}

let statsObserver: IntersectionObserver | null = null

onMounted(() => {
  nextTick(() => {
    initScrollReveal()
    const statsEl = homeRef.value?.querySelector('.stats-section')
    if (statsEl) {
      statsObserver = new IntersectionObserver(
        (entries) => {
          if (entries[0]?.isIntersecting) {
            animateCounters()
            statsObserver?.disconnect()
          }
        },
        { threshold: 0.3 },
      )
      statsObserver.observe(statsEl)
    }
  })
})

onUnmounted(() => {
  statsObserver?.disconnect()
})

/* ---- Cards ---- */
const cards = [
  { emoji: '🗺', title: '时空地图', desc: '探索大运河从隋唐至今的历史变迁', path: '/map', icon: '📍' },
  { emoji: '📖', title: '科普故事', desc: 'AI 驱动的运河文化与生态科普', path: '/story', icon: '📚' },
  { emoji: '💬', title: '分享广场', desc: '发布图文记录运河美好时刻', path: '/plaza', icon: '💬' },
  { emoji: '❓', title: '趣味问答', desc: '答题赢积分，争夺运河守护者称号', path: '/quiz', icon: '🏆' },
]

/* ---- News mock ---- */
const news = [
  { date: '2026-06-12', title: '流动的书香：大运河上的文人朋友圈' },
  { date: '2026-06-12', title: '邀您共赴运河文化之约，同行恰五载' },
  { date: '2026-06-05', title: '盛辉与尘光：古希腊古罗马遗存特展预告' },
  { date: '2026-06-05', title: '《文物里的大运河》重磅上线' },
  { date: '2026-05-28', title: '千度淬炼，琉光永存：解密琉璃的匠造之美' },
  { date: '2026-05-22', title: '淄博琉璃非遗艺术展云导览上线' },
]
</script>

<template>
  <div ref="homeRef" class="home-page">
    <!-- ====== Banner Carousel ====== -->
    <section class="banner-section" @mouseenter="stopAutoPlay" @mouseleave="startAutoPlay">
      <div class="banner-slides">
        <div
          v-for="(slide, i) in slides"
          :key="i"
          class="banner-slide"
          :class="{ active: currentSlide === i }"
          :style="{ background: slide.bg }"
        >
          <div class="banner-bg-pattern"></div>
          <div
            v-if="slide.bgImage"
            class="banner-bg-image"
            :style="{ backgroundImage: `url(${slide.bgImage})` }"
          ></div>
          <div class="banner-overlay">
            <div class="banner-content" :class="{ 'slide-in': currentSlide === i }">
              <h1 class="banner-title">{{ slide.title }}</h1>
              <h2 class="banner-subtitle">{{ slide.subtitle }}</h2>
              <p class="banner-desc">{{ slide.desc }}</p>
              <div class="banner-actions">
                <ElButton type="primary" size="large" round @click="router.push('/map')">探索时空地图</ElButton>
                <ElButton size="large" round plain class="btn-outline" @click="router.push('/quiz')">挑战趣味问答</ElButton>
              </div>
            </div>
            <!-- 预约面板 -->
            <div class="banner-panel">
              <div class="panel-item" @click="router.push('/map')">
                <div class="panel-icon">📅</div>
                <div class="panel-text">
                  <strong>时空地图</strong>
                  <span>探索运河千年变迁</span>
                </div>
                <div class="panel-arrow">→</div>
              </div>
              <div class="panel-item" @click="router.push('/story')">
                <div class="panel-icon">📖</div>
                <div class="panel-text">
                  <strong>运河故事</strong>
                  <span>AI 科普即刻体验</span>
                </div>
                <div class="panel-arrow">→</div>
              </div>
            </div>
          </div>
        </div>
      </div>
      <!-- 指示器 -->
      <div class="banner-dots">
        <span
          v-for="(_, i) in slides"
          :key="i"
          class="dot"
          :class="{ active: currentSlide === i }"
          @click="goToSlide(i)"
        >
          <svg viewBox="0 0 16000 16000">
            <circle class="dot-circle" cx="8000" cy="8000" r="5800" />
          </svg>
        </span>
      </div>
      <!-- 箭头 -->
      <div class="banner-arrow left" @click="goToSlide((currentSlide - 1 + slides.length) % slides.length)">
        <svg viewBox="0 0 16000 16000">
          <polyline points="11040,1920 4960,8000 11040,14080" />
        </svg>
      </div>
      <div class="banner-arrow right" @click="nextSlide">
        <svg viewBox="0 0 16000 16000">
          <polyline points="4960,1920 11040,8000 4960,14080" />
        </svg>
      </div>
    </section>

    <!-- ====== 新闻 & 功能区 ====== -->
    <section class="content-row news-row">
      <div class="news-left" data-reveal="fade-up">
        <h3 class="section-title">运博动态</h3>
        <ul class="news-list">
          <li v-for="(item, i) in news" :key="i" :style="{ animationDelay: `${i * 0.1}s` }">
            <span class="news-date">{{ item.date }}</span>
            <span class="news-title">{{ item.title }}</span>
          </li>
        </ul>
        <div class="more-link">
          了解更多 <span class="arrow-icon">→</span>
        </div>
      </div>
      <div class="news-right" data-reveal="fade-up" :style="{ transitionDelay: '0.2s' }">
        <div class="news-card news-card-top" @click="router.push('/story')">
          <div class="card-img-placeholder img-story"></div>
          <div class="card-label">科普故事</div>
        </div>
        <div class="news-card" @click="router.push('/plaza')">
          <div class="card-img-placeholder img-plaza"></div>
          <div class="card-label">分享广场</div>
        </div>
      </div>
    </section>

    <!-- ====== 功能卡片 ====== -->
    <section class="content-row cards-section">
      <div class="section-header" data-reveal="fade-up">
        <h2>核心服务</h2>
        <p>四大功能模块，全方位守护运河文化</p>
      </div>
      <div class="cards-grid">
        <div
          v-for="(card, i) in cards"
          :key="card.title"
          class="feature-card-wrapper"
          :data-reveal="i % 2 === 0 ? 'fade-right' : 'fade-left'"
          :style="{ transitionDelay: `${i * 0.15}s` }"
        >
          <ElCard shadow="hover" class="feature-card" @click="router.push(card.path)">
            <div class="card-top-icon">{{ card.emoji }}</div>
            <h3>{{ card.title }}</h3>
            <p>{{ card.desc }}</p>
            <div class="card-bottom-icon">→</div>
          </ElCard>
        </div>
      </div>
    </section>

    <!-- ====== 展览区 ====== -->
    <section class="content-row exhibit-section">
      <div class="exhibit-col" data-reveal="fade-right" @click="router.push('/map')">
        <div class="exhibit-card">
          <div class="exhibit-bg exhibit-bg-1"></div>
          <div class="exhibit-info">
            <h3>时空地图</h3>
            <span>探索运河千年变迁</span>
          </div>
        </div>
      </div>
      <div class="exhibit-col" data-reveal="zoom-in" @click="router.push('/story')">
        <div class="exhibit-card">
          <div class="exhibit-bg exhibit-bg-2"></div>
          <div class="exhibit-info">
            <h3>科普故事</h3>
            <span>AI 运河文化科普</span>
          </div>
        </div>
      </div>
      <div class="exhibit-col" data-reveal="fade-left" @click="router.push('/quiz')">
        <div class="exhibit-card">
          <div class="exhibit-bg exhibit-bg-3"></div>
          <div class="exhibit-info">
            <h3>趣味问答</h3>
            <span>答题争当守护者</span>
          </div>
        </div>
      </div>
    </section>

    <!-- ====== 数据统计 ====== -->
    <section class="content-row stats-section">
      <div class="stats-inner">
        <div
          v-for="(stat, i) in stats"
          :key="stat.label"
          class="stat-item"
          data-reveal="fade-up"
          :style="{ transitionDelay: `${i * 0.1}s` }"
        >
          <span class="stat-number">{{ (animatedStats[i] ?? 0).toLocaleString() }}{{ stat.suffix }}</span>
          <span class="stat-label">{{ stat.label }}</span>
        </div>
      </div>
    </section>
  </div>
</template>

<style scoped>
.home-page {
  margin: -24px;
  overflow: hidden;
}

/* ====== Banner ====== */
.banner-section {
  position: relative;
  width: 100%;
  height: 700px;
  overflow: hidden;
  cursor: default;
}

.banner-slides {
  position: relative;
  width: 100%;
  height: 100%;
}

.banner-slide {
  position: absolute;
  inset: 0;
  opacity: 0;
  transition: opacity 0.8s ease;
}

.banner-slide.active {
  opacity: 1;
  z-index: 1;
}

.banner-bg-pattern {
  position: absolute;
  inset: 0;
  background:
    radial-gradient(ellipse at 20% 50%, rgba(255,255,255,0.06) 0%, transparent 50%),
    radial-gradient(ellipse at 80% 20%, rgba(255,255,255,0.04) 0%, transparent 50%);
}

.banner-bg-image {
  position: absolute;
  inset: 0;
  background-size: cover;
  background-position: center;
  background-repeat: no-repeat;
  opacity: 0.5;
  z-index: 0;
}

.banner-overlay {
  position: relative;
  z-index: 2;
  display: flex;
  align-items: center;
  justify-content: center;
  height: 100%;
  padding: 0 80px;
}

.banner-content {
  text-align: center;
  color: #fff;
  max-width: 700px;
}

.banner-content.slide-in {
  animation: slideInUp 0.8s ease forwards;
}

@keyframes slideInUp {
  from {
    opacity: 0;
    transform: translateY(40px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

.banner-content h1 {
  font-size: 104px;
  font-weight: 700;
  margin: 0 0 8px;
  letter-spacing: 4px;
  text-shadow: 0 2px 24px rgba(0,0,0,0.6);
}

.banner-content h2 {
  font-size: 28px;
  font-weight: 300;
  margin: 0 0 20px;
  opacity: 0.9;
  letter-spacing: 6px;
}

.banner-content p {
  font-size: 16px;
  opacity: 0.75;
  margin: 0 0 36px;
}

.banner-actions {
  display: flex;
  gap: 16px;
  justify-content: center;
}

.btn-outline {
  border-color: rgba(255,255,255,0.5) !important;
  color: #fff !important;
  background: rgba(255,255,255,0.1) !important;
}

.btn-outline:hover {
  border-color: #fff !important;
  background: rgba(255,255,255,0.2) !important;
}

/* 预约面板 */
.banner-panel {
  position: absolute;
  bottom: 60px;
  left: 80px;
  display: flex;
  gap: 16px;
}

.panel-item {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 16px 24px;
  background: rgba(255,255,255,0.12);
  backdrop-filter: blur(12px);
  border: 1px solid rgba(255,255,255,0.2);
  border-radius: 12px;
  color: #fff;
  cursor: pointer;
  transition: background 0.3s, transform 0.3s;
  min-width: 200px;
}

.panel-item:hover {
  background: rgba(255,255,255,0.22);
  transform: translateY(-2px);
}

.panel-icon {
  font-size: 28px;
}

.panel-text strong {
  display: block;
  font-size: 14px;
  margin-bottom: 2px;
}

.panel-text span {
  font-size: 12px;
  opacity: 0.7;
}

.panel-arrow {
  margin-left: auto;
  font-size: 18px;
  opacity: 0.6;
}

/* Dots */
.banner-dots {
  position: absolute;
  bottom: 24px;
  right: 24px;
  display: flex;
  gap: 10px;
  z-index: 5;
}

.dot {
  width: 16px;
  height: 16px;
  cursor: pointer;
  opacity: 0.5;
  transition: opacity 0.3s;
}

.dot.active {
  opacity: 1;
}

.dot svg {
  width: 100%;
  height: 100%;
}

.dot-circle {
  fill: none;
  stroke: #fff;
  stroke-width: 1600;
  stroke-dasharray: 0 10000;
  transition: stroke-dasharray 0.3s;
}

.dot.active .dot-circle {
  stroke-dasharray: 36000 10000;
}

/* Arrows */
.banner-arrow {
  position: absolute;
  top: 50%;
  width: 50px;
  height: 50px;
  transform: translateY(-50%);
  cursor: pointer;
  z-index: 5;
  opacity: 0;
  transition: opacity 0.3s;
}

.banner-section:hover .banner-arrow {
  opacity: 0.7;
}

.banner-arrow:hover {
  opacity: 1 !important;
}

.banner-arrow.left {
  left: 24px;
}

.banner-arrow.right {
  right: 24px;
}

.banner-arrow svg {
  width: 100%;
  height: 100%;
}

.banner-arrow polyline {
  fill: none;
  stroke: #fff;
  stroke-width: 1200;
  stroke-linecap: round;
  stroke-linejoin: round;
}

/* ====== Content Rows ====== */
.content-row {
  max-width: 1200px;
  margin: 0 auto;
  padding: 60px 24px;
}

/* ====== News Row ====== */
.news-row {
  display: flex;
  gap: 30px;
}

.news-left {
  flex: 1.2;
  background: #fff;
  border-radius: 12px;
  padding: 32px;
  box-shadow: 0 2px 12px rgba(0,0,0,0.04);
}

.section-title {
  font-size: 20px;
  color: #2c3e50;
  margin: 0 0 20px;
  padding-bottom: 12px;
  border-bottom: 2px solid #2c3e50;
  display: inline-block;
}

.news-list {
  list-style: none;
}

.news-list li {
  display: flex;
  align-items: baseline;
  gap: 12px;
  padding: 8px 0;
  border-bottom: 1px dashed #ebeef5;
  cursor: pointer;
  transition: color 0.2s;
}

.news-list li:hover {
  color: #2c3e50;
}

.news-date {
  font-size: 12px;
  color: #909399;
  white-space: nowrap;
  font-family: 'Courier New', monospace;
}

.news-title {
  font-size: 14px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.more-link {
  margin-top: 16px;
  font-size: 13px;
  color: #2c3e50;
  cursor: pointer;
  display: flex;
  align-items: center;
  gap: 4px;
  transition: gap 0.3s;
}

.more-link:hover {
  gap: 8px;
}

.arrow-icon {
  font-size: 12px;
}

/* 右侧卡片 */
.news-right {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.news-card {
  flex: 1;
  border-radius: 12px;
  overflow: hidden;
  cursor: pointer;
  position: relative;
  transition: transform 0.3s;
}

.news-card:hover {
  transform: translateY(-3px);
}

.card-img-placeholder {
  width: 100%;
  height: 100%;
  min-height: 130px;
  background-size: cover;
  background-position: center;
  transition: transform 0.5s ease;
}

.news-card:hover .card-img-placeholder {
  transform: scale(1.05);
}

.img-story {
  background: linear-gradient(135deg, #4a6278, #6b7b6e);
  background-image: repeating-linear-gradient(45deg, rgba(255,255,255,0.03) 0px, rgba(255,255,255,0.03) 2px, transparent 2px, transparent 8px);
}

.img-plaza {
  background: linear-gradient(135deg, #5a3a5f, #2d4a7a);
  background-image: repeating-linear-gradient(-30deg, rgba(255,255,255,0.03) 0px, rgba(255,255,255,0.03) 2px, transparent 2px, transparent 8px);
}

.card-label {
  position: absolute;
  bottom: 16px;
  left: 20px;
  color: #fff;
  font-size: 16px;
  font-weight: 600;
  text-shadow: 0 1px 6px rgba(0,0,0,0.4);
}

/* ====== Cards Section ====== */
.cards-section {
  background: #f5efe5;
  max-width: 100%;
  padding: 60px calc((100% - 1200px) / 2 + 24px);
}

.section-header {
  text-align: center;
  margin-bottom: 40px;
}

.section-header h2 {
  font-size: 28px;
  color: #303133;
  margin: 0 0 8px;
}

.section-header p {
  font-size: 14px;
  color: #909399;
  margin: 0;
}

.cards-grid {
  max-width: 1200px;
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 20px;
}

.feature-card-wrapper {
  cursor: pointer;
}

.feature-card {
  text-align: center;
  border-radius: 12px;
  overflow: hidden;
  transition: transform 0.3s ease, box-shadow 0.3s ease;
}

.feature-card:hover {
  transform: translateY(-6px);
  box-shadow: 0 8px 24px rgba(44, 62, 80, 0.12);
}

.card-top-icon {
  font-size: 44px;
  margin-bottom: 12px;
  animation: float 3s ease-in-out infinite;
}

.feature-card h3 {
  margin: 0 0 8px;
  color: #303133;
  font-size: 16px;
}

.feature-card p {
  margin: 0 0 12px;
  color: #909399;
  font-size: 13px;
  line-height: 1.6;
}

.card-bottom-icon {
  font-size: 18px;
  color: #2c3e50;
  opacity: 0;
  transform: translateX(-8px);
  transition: opacity 0.3s, transform 0.3s;
}

.feature-card:hover .card-bottom-icon {
  opacity: 1;
  transform: translateX(0);
}

/* ====== Exhibit Section ====== */
.exhibit-section {
  display: flex;
  gap: 20px;
}

.exhibit-col {
  flex: 1;
  cursor: pointer;
}

.exhibit-card {
  position: relative;
  height: 280px;
  border-radius: 12px;
  overflow: hidden;
  transition: transform 0.3s ease;
}

.exhibit-card:hover {
  transform: translateY(-4px);
}

.exhibit-bg {
  width: 100%;
  height: 100%;
  background-size: cover;
  background-position: center;
  transition: transform 0.6s ease;
}

.exhibit-card:hover .exhibit-bg {
  transform: scale(1.08);
}

.exhibit-bg-1 {
  background: linear-gradient(135deg, #2c3e50, #4a6278, #6b7b6e);
}

.exhibit-bg-2 {
  background: linear-gradient(135deg, #2d5a3f, #4a6278, #2c3e50);
}

.exhibit-bg-3 {
  background: linear-gradient(135deg, #5a2d5a, #4a2d6a, #2d1a5a);
}

.exhibit-info {
  position: absolute;
  bottom: 0;
  left: 0;
  right: 0;
  padding: 24px;
  background: linear-gradient(transparent, rgba(0,0,0,0.7));
  color: #fff;
}

.exhibit-info h3 {
  margin: 0 0 4px;
  font-size: 20px;
}

.exhibit-info span {
  font-size: 13px;
  opacity: 0.8;
}

/* ====== Stats Section ====== */
.stats-section {
  background: #fff;
  max-width: 100%;
  padding: 60px calc((100% - 1200px) / 2 + 24px);
  position: relative;
}

.stats-section::before {
  content: '';
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  height: 3px;
  background: linear-gradient(90deg, #8b7355, #2c3e50, #8b7355);
}

.stats-inner {
  max-width: 1200px;
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 24px;
}

.stat-item {
  text-align: center;
  padding: 32px 16px;
  border-radius: 8px;
  transition: transform 0.3s, box-shadow 0.3s;
}

.stat-item:hover {
  transform: translateY(-4px);
  box-shadow: 0 4px 16px rgba(0,0,0,0.06);
}

.stat-number {
  display: block;
  font-size: 36px;
  font-weight: 700;
  color: #2c3e50;
  font-family: 'Georgia', serif;
  margin-bottom: 8px;
}

.stat-label {
  font-size: 13px;
  color: #909399;
}

/* ====== Responsive ====== */
@media (max-width: 1024px) {
  .cards-grid {
    grid-template-columns: repeat(2, 1fr);
  }

  .news-row {
    flex-direction: column;
  }

  .banner-panel {
    left: 24px;
    bottom: 30px;
  }

  .banner-content h1 {
    font-size: 56px;
  }

  .banner-content h2 {
    font-size: 20px;
  }

  .exhibit-section {
    flex-direction: column;
  }
}

@media (max-width: 768px) {
  .banner-section {
    height: 500px;
  }

  .banner-panel {
    flex-direction: column;
    gap: 8px;
  }

  .panel-item {
    min-width: auto;
  }

  .cards-grid {
    grid-template-columns: 1fr;
  }

  .stats-inner {
    grid-template-columns: repeat(2, 1fr);
  }

  .exhibit-card {
    height: 200px;
  }
}
</style>
