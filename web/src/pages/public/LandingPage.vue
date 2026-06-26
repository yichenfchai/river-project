<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElButton } from 'element-plus'
import { getCanalIntro, getCanalIntroByCoords } from '@/api/modules/geo'
import type { CanalIntro } from '@/api/modules/geo'

const defaultContent = `京杭大运河始建于公元前486年，是世界上开凿最早、里程最长的人工运河。它南起余杭（今杭州），北至涿郡（今北京），全长约1797公里，贯穿海河、黄河、淮河、长江、钱塘江五大水系。2500年来，大运河见证了中华民族的兴衰荣辱，承载了南北经济文化交流的辉煌篇章。2014年，中国大运河成功入选世界遗产名录。无论你身在何处，运河的精神与文脉始终奔流不息。`

const router = useRouter()
const intro = ref<CanalIntro>({
  is_canal_city: false,
  city: '', province: '', ip: '',
  intro: { title: '千里运河，流动的史诗', content: defaultContent, era: '公元前486年至今', poi: '京杭大运河' },
})

function tryApplyIntro(res: { data: CanalIntro } | null) {
  if (res && res.data && res.data.is_canal_city) {
    intro.value = res.data
  }
}

async function tryGeolocation() {
  if (!navigator.geolocation) return false

  try {
    const pos = await new Promise<GeolocationPosition>((resolve, reject) => {
      navigator.geolocation.getCurrentPosition(resolve, reject, {
        timeout: 3000,
        enableHighAccuracy: false,
      })
    })
    const res = await getCanalIntroByCoords(pos.coords.latitude, pos.coords.longitude)
    if (res.data?.is_canal_city) {
      tryApplyIntro(res)
      return true
    }
  } catch {
    // geolocation failed or denied, fall through to IP
  }
  return false
}

onMounted(async () => {
  const gpsOk = await tryGeolocation()
  if (!gpsOk) {
    try {
      const res = await getCanalIntro()
      tryApplyIntro(res)
    } catch {}
  }
})
</script>

<template>
  <div class="landing">
    <!-- 水墨背景层 -->
    <div class="bg-layer">
      <div class="mountain mountain-1" />
      <div class="mountain mountain-2" />
      <div class="mountain mountain-3" />
      <div class="river" />
      <div class="fog fog-1" />
      <div class="fog fog-2" />
    </div>

    <!-- 动态元素 -->
    <div class="birds">
      <svg viewBox="0 0 120 40" class="bird-group">
        <path d="M0 20 Q10 10 20 20 Q30 10 40 20" fill="none" stroke="rgba(255,255,255,0.5)" stroke-width="1.5" />
        <path d="M15 15 Q22 8 30 15" fill="none" stroke="rgba(255,255,255,0.4)" stroke-width="1" />
        <path d="M10 22 Q18 14 26 22" fill="none" stroke="rgba(255,255,255,0.35)" stroke-width="1" />
      </svg>
    </div>

    <!-- 飘浮灯笼 -->
    <div class="lanterns">
      <div class="lantern l1">&#x1F3EE;</div>
      <div class="lantern l2">&#x1F3EE;</div>
      <div class="lantern l3">&#x1F3EE;</div>
    </div>

    <!-- 主卡片 -->
    <div class="main-card loaded">
      <div class="card-header">
        <h1 class="title">
          <span class="char" v-for="(c, i) in '大运河生态与文化保护平台'" :key="i" :style="{ animationDelay: `${0.8 + i * 0.06}s` }">{{ c }}</span>
        </h1>
        <p class="subtitle">Grand Canal Guardian</p>
      </div>

      <!-- 城市介绍卡片 -->
      <div v-if="intro && intro.is_canal_city" class="city-card">
        <div class="city-name">
          <span class="city-icon">&#x1F30A;</span>
          <span>{{ intro.city }}</span>
        </div>
        <div class="city-province">{{ intro.province }}</div>
        <h3 class="city-title">{{ intro.intro.title }}</h3>
        <p class="city-content">{{ intro.intro.content }}</p>
        <div class="city-meta">
          <span class="meta-tag">&#x1F3DB; {{ intro.intro.poi }}</span>
          <span class="meta-tag">&#x23F3; {{ intro.intro.era }}</span>
        </div>
      </div>

      <div v-else class="city-card default-intro">
        <h3 class="city-title">{{ intro.intro.title }}</h3>
        <p class="city-content">{{ intro.intro.content }}</p>
        <div class="city-meta">
          <span class="meta-tag">&#x1F3DB; {{ intro.intro.poi }}</span>
          <span class="meta-tag">&#x23F3; {{ intro.intro.era }}</span>
        </div>
      </div>

      <div class="enter-wrap">
        <el-button class="enter-btn" size="large" round @click="router.push('/login')">
          进入运河世界
        </el-button>
      </div>
    </div>

    <!-- 底部 -->
    <p class="bottom-hint">京杭大运河 · 世界文化遗产 · 流动的中华史诗</p>
  </div>
</template>

<style scoped>
.landing {
  min-height: 100vh;
  min-height: 100dvh;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  position: relative;
  overflow: hidden;
  background: linear-gradient(180deg, #1a1a2e 0%, #16213e 40%, #0f3460 100%);
  padding: 20px;
  box-sizing: border-box;
}

/* ─── 水墨背景 ─── */
.bg-layer { position:absolute;inset:0;pointer-events:none;z-index:0 }
.mountain { position:absolute;bottom:0;width:100%;height:45% }
.mountain-1 {
  background: radial-gradient(ellipse 80% 100% at 20% bottom, rgba(44,62,80,0.5) 0%, transparent 70%),
              radial-gradient(ellipse 60% 90% at 80% bottom, rgba(44,62,80,0.4) 0%, transparent 70%);
  animation: mountainFade 2s ease-out;
}
.mountain-2 {
  bottom:5%;
  background: radial-gradient(ellipse 50% 70% at 40% bottom, rgba(30,40,55,0.5) 0%, transparent 70%);
  animation: mountainFade 2.5s ease-out;
}
.mountain-3 {
  bottom:8%;
  background: radial-gradient(ellipse 40% 50% at 60% bottom, rgba(20,30,45,0.4) 0%, transparent 70%);
  animation: mountainFade 3s ease-out;
}
.river {
  position:absolute;bottom:0;width:120%;height:15%;left:-10%;
  background: linear-gradient(180deg, transparent 0%, rgba(100,140,180,0.08) 40%, rgba(80,120,170,0.12) 100%);
  animation: riverFlow 8s ease-in-out infinite;
}
.fog { position:absolute;width:200%;height:30%;opacity:0.06 }
.fog-1 { top:20%;left:-50%;background: radial-gradient(ellipse, rgba(200,200,200,1) 0%, transparent 70%);animation: fogDrift 20s linear infinite }
.fog-2 { top:35%;left:-80%;background: radial-gradient(ellipse, rgba(180,190,200,1) 0%, transparent 70%);animation: fogDrift 25s linear infinite reverse }

/* ─── 飞鸟 ─── */
.birds { position:absolute;top:22%;left:-10%;z-index:1;pointer-events:none }
.bird-group { width:80px;height:30px;animation: birdFly 8s linear infinite }
.lanterns { position:absolute;inset:0;pointer-events:none;z-index:1 }
.lantern {
  position:absolute;font-size:20px;opacity:0.4;
  animation: lanternFloat 12s ease-in-out infinite;
}
.l1 { left:15%;animation-delay:0s }
.l2 { left:50%;animation-delay:4s }
.l3 { left:78%;animation-delay:8s }

/* ─── 主卡片 ─── */
.main-card {
  position:relative;z-index:10;max-width:520px;width:100%;
  opacity:0;transform:translateY(30px);
  transition: opacity 1s ease-out, transform 1s ease-out;
}
.main-card.loaded { opacity:1;transform:translateY(0) }
.card-header { text-align:center;margin-bottom:28px }
.title { font-size:clamp(20px,5vw,32px);color:#e0d6c0;margin:0;letter-spacing:4px;font-weight:700 }
.char { display:inline-block;animation: charIn 0.6s ease-out both }
@keyframes charIn { from{opacity:0;transform:translateY(10px)} to{opacity:1;transform:translateY(0)} }
.subtitle { color:rgba(255,255,255,0.35);font-size:13px;letter-spacing:6px;margin:8px 0 0;text-transform:uppercase }

.city-card {
  background: rgba(255,255,255,0.04);border:1px solid rgba(201,184,150,0.15);
  border-radius:16px;padding:clamp(20px,4vw,32px);margin-bottom:24px;
  backdrop-filter: blur(16px);-webkit-backdrop-filter: blur(16px);
  animation: cardReveal 1s ease-out .5s both;max-height:45vh;overflow-y:auto
}
.city-card.default-intro { border-color:rgba(255,255,255,0.06) }
.city-name { display:flex;align-items:center;gap:8px;font-size:24px;color:#c9b896;font-weight:700;margin-bottom:4px }
.city-icon { font-size:28px }
.city-province { font-size:13px;color:rgba(255,255,255,0.4);margin-bottom:16px }
.city-title { font-size:17px;color:rgba(255,255,255,0.85);margin:0 0 12px;font-weight:600 }
.city-content { font-size:14px;line-height:1.9;color:rgba(255,255,255,0.6);margin:0 0 16px;white-space:pre-line }
.city-meta { display:flex;gap:12px;flex-wrap:wrap }
.meta-tag { font-size:12px;color:rgba(201,184,150,0.7);padding:3px 10px;border:1px solid rgba(201,184,150,0.2);border-radius:12px }

.enter-wrap { text-align:center }
.enter-btn {
  background: linear-gradient(135deg, #c9b896 0%, #a8926a 100%) !important;
  border: none !important;color: #1a1a2e !important;font-weight:700;
  font-size:16px !important;padding:14px 48px !important;height:auto !important;
  letter-spacing:2px;transition: all .3s !important;
  box-shadow: 0 4px 20px rgba(201,184,150,0.3);
}
.enter-btn:hover { transform:translateY(-2px);box-shadow:0 8px 30px rgba(201,184,150,0.45) !important }

.bottom-hint { position:absolute;bottom:16px;font-size:11px;color:rgba(255,255,255,0.2);letter-spacing:2px;z-index:10 }

/* ─── Animations ─── */
@keyframes mountainFade { from{opacity:0;transform:translateY(40px)} to{opacity:1;transform:translateY(0)} }
@keyframes riverFlow { 0%,100%{transform:translateX(0)} 50%{transform:translateX(3%)} }
@keyframes fogDrift { from{transform:translateX(0)} to{transform:translateX(50%)} }
@keyframes birdFly { 0%{transform:translateX(0) translateY(0)} 50%{transform:translateX(40vw) translateY(-8px)} 100%{transform:translateX(100vw) translateY(0)} }
@keyframes lanternFloat { 0%{transform:translateY(100vh) translateX(0);opacity:0} 20%{opacity:0.4} 80%{opacity:0.3} 100%{transform:translateY(-10vh) translateX(30px);opacity:0} }
@keyframes cardReveal { from{opacity:0;transform:translateY(20px)} to{opacity:1;transform:translateY(0)} }

/* ─── Mobile ─── */
@media (max-width: 480px) {
  .landing { padding:12px;justify-content:flex-start;padding-top:10vh }
  .city-card { padding:16px;max-height:42vh }
  .title { font-size:20px;letter-spacing:1px }
  .subtitle { font-size:10px;letter-spacing:3px }
  .city-content { font-size:13px;line-height:1.7 }
  .enter-btn { font-size:14px !important;padding:10px 28px !important }
  .bottom-hint { font-size:10px }
  .lanterns { display:none }
  .birds { display:none }
  .fog { display:none }
}

@media (prefers-reduced-motion: reduce) {
  .mountain, .river, .fog, .birds, .bird-group, .lantern, .lanterns {
    animation: none !important
  }
  .mountain { opacity:1;transform:none !important }
}
</style>
