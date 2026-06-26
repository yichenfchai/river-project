<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { ElButton, ElMessage } from 'element-plus'
import { getQuestions, submitBatch } from '@/api/modules/quiz'
import type { Question, SessionResponse, BatchSubmitResult } from '@/types/quiz'

const categories = [
  { label: '全部', value: '' },
  { label: '历史', value: 'history' },
  { label: '生态', value: 'ecology' },
  { label: '地理', value: 'geography' },
  { label: '水利', value: 'water_conservancy' },
]

const difficulties = [
  { label: '混合', value: 'mixed' },
  { label: '简单', value: 'easy' },
  { label: '中等', value: 'medium' },
  { label: '困难', value: 'hard' },
]

const loading = ref(false)
const answering = ref(false)
const sessionId = ref('')
const questions = ref<Question[]>([])
const selectedAnswers = ref<Record<string, string>>({})
const activeCategory = ref('')
const activeDifficulty = ref('mixed')
const result = ref<BatchSubmitResult | null>(null)

const ranks = [
  { icon: '🥉', name: '青铜守护者 0-500', color: '#cd7f32' },
  { icon: '🥈', name: '白银守护者 500-1500', color: '#c0c0c0' },
  { icon: '🥇', name: '黄金守护者 1500-5000', color: '#ffd700' },
  { icon: '💎', name: '钻石守护者 5000-15000', color: '#2c3e50' },
  { icon: '👑', name: '运河守护者 15000+', color: '#e6a23c' },
]

const currentQuestion = computed(() => questions.value[0])
const allAnswered = computed(() => questions.value.length > 0 && questions.value.every(q => selectedAnswers.value[q.id]))

async function fetchQuestions() {
  loading.value = true
  result.value = null
  selectedAnswers.value = {}
  try {
    const res = await getQuestions({
      count: 5,
      difficulty: activeDifficulty.value,
      category: activeCategory.value || undefined,
    })
    sessionId.value = res.data.session_id
    questions.value = res.data.questions || []
  } catch {
    ElMessage.error('获取题目失败')
  } finally {
    loading.value = false
  }
}

function selectAnswer(qid: string, letter: string) {
  selectedAnswers.value[qid] = letter
}

async function submit() {
  if (!allAnswered.value) {
    ElMessage.warning('请回答所有题目')
    return
  }
  answering.value = true
  try {
    const answers = Object.entries(selectedAnswers.value).map(([qid, ans]) => ({
      question_id: qid,
      answer: ans,
    }))
    const res = await submitBatch({ session_id: sessionId.value, answers })
    result.value = res.data
    ElMessage.success(`答题完成！正确 ${res.data.correct_count}/${res.data.total_count}`)
  } catch {
    ElMessage.error('提交失败')
  } finally {
    answering.value = false
  }
}

function restart() {
  result.value = null
  selectedAnswers.value = {}
  fetchQuestions()
}

onMounted(fetchQuestions)
</script>

<template>
  <div class="quiz-page">
    <h2>❓ 趣味问答</h2>
    <p class="page-desc">测试你对大运河的知识储备，赢取积分，争夺守护者称号</p>

    <div class="quiz-categories">
      <el-button v-for="c in categories" :key="c.value" :type="activeCategory === c.value ? 'primary' : 'default'" size="large" @click="activeCategory = c.value; fetchQuestions()">
        {{ c.label }}
      </el-button>
    </div>

    <div class="difficulty-bar">
      <span v-for="d in difficulties" :key="d.value" class="diff-tag" :class="{ active: activeDifficulty === d.value }" @click="activeDifficulty = d.value; fetchQuestions()">{{ d.label }}</span>
    </div>

    <!-- Result Screen -->
    <div v-if="result" class="quiz-card result-card">
      <h3>📊 答题结果</h3>
      <div class="result-summary">
        <div class="result-item"><span class="result-label">正确</span><span class="result-value">{{ result.correct_count }} / {{ result.total_count }}</span></div>
        <div class="result-item"><span class="result-label">获得积分</span><span class="result-value highlight">+{{ result.total_points_earned }}</span></div>
        <div class="result-item"><span class="result-label">总积分</span><span class="result-value">{{ result.new_total_points }}</span></div>
        <div class="result-item"><span class="result-label">段位</span><span class="result-value rank-title">{{ result.rank_title }}</span></div>
      </div>
      <el-button type="primary" size="large" @click="restart">再来一轮</el-button>
    </div>

    <!-- Question Screen -->
    <div v-else class="question-list" v-loading="loading">
      <div v-for="(q, qi) in questions" :key="q.id" class="quiz-card">
        <div class="quiz-question">
          <span class="q-num">第 {{ qi + 1 }} 题</span>
          <span class="difficulty" :class="q.difficulty">{{ { easy: '简单', medium: '中等', hard: '困难' }[q.difficulty] }}</span>
          <span class="category-tag">{{ q.category }}</span>
          <h3>{{ q.question }}</h3>
        </div>
        <div class="quiz-options">
          <div
            v-for="(opt, i) in q.options"
            :key="i"
            class="option-item"
            :class="{ selected: selectedAnswers[q.id] === String.fromCharCode(65 + i) }"
            @click="selectAnswer(q.id, String.fromCharCode(65 + i))"
          >
            {{ String.fromCharCode(65 + i) }}. {{ opt }}
          </div>
        </div>
      </div>
      <div class="quiz-action">
        <el-button type="primary" size="large" @click="submit" :loading="answering" :disabled="!allAnswered">
          {{ allAnswered ? '提交答案' : `请回答所有题目 (${Object.keys(selectedAnswers).length}/${questions.length})` }}
        </el-button>
        <span class="hint">连续答对有连击加成！</span>
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
.quiz-page { max-width:800px;margin:0 auto }
.quiz-page h2 { font-size:22px;color:#303133 }
.page-desc { color:#909399;font-size:14px;margin-bottom:12px }
.quiz-categories { display:flex;gap:10px;margin-bottom:12px;flex-wrap:wrap }
.difficulty-bar { display:flex;gap:8px;margin-bottom:20px }
.diff-tag { padding:4px 14px;border-radius:14px;background:#f0f2f5;color:#606266;font-size:12px;cursor:pointer;transition:.2s }
.diff-tag.active,.diff-tag:hover { background:#2c3e50;color:#fff }
.quiz-card { background:#fff;border-radius:12px;padding:32px;box-shadow:0 2px 8px rgba(0,0,0,.06);margin-bottom:16px }
.question-list { display:flex;flex-direction:column }
.q-num { font-size:13px;color:#909399;margin-right:8px;font-weight:600 }
.quiz-question { margin-bottom:24px }
.difficulty { display:inline-block;font-size:11px;padding:2px 8px;border-radius:4px;margin-right:8px }
.difficulty.easy { background:#f0f9eb;color:#67c23a }
.difficulty.medium { background:#fdf6ec;color:#e6a23c }
.difficulty.hard { background:#fef0f0;color:#f56c6c }
.category-tag { font-size:11px;padding:2px 8px;background:#eae7e0;color:#2c3e50;border-radius:4px }
.quiz-question h3 { margin-top:12px;font-size:20px;color:#303133 }
.quiz-options { display:flex;flex-direction:column;gap:12px }
.option-item { padding:14px 18px;border:1px solid #e4e7ed;border-radius:8px;cursor:pointer;transition:.2s;font-size:15px;user-select:none }
.option-item:hover { border-color:#2c3e50;background:#eae7e0 }
.option-item.selected { border-color:#2c3e50;background:#eae7e0;font-weight:600 }
.quiz-action { margin-top:24px;display:flex;align-items:center;gap:16px }
.hint { color:#909399;font-size:13px }
.result-card { text-align:center }
.result-card h3 { margin:0 0 20px }
.result-summary { display:flex;justify-content:center;gap:32px;margin-bottom:24px;flex-wrap:wrap }
.result-item { display:flex;flex-direction:column;align-items:center }
.result-label { font-size:13px;color:#909399 }
.result-value { font-size:24px;font-weight:700;color:#303133;margin-top:4px }
.result-value.highlight { color:#67c23a }
.result-value.rank-title { color:#e6a23c }
.rank-preview { margin-top:32px;text-align:center }
.rank-preview h3 { color:#303133;margin-bottom:12px }
.rank-list { display:flex;gap:8px;justify-content:center;flex-wrap:wrap }
.rank-badge { padding:4px 12px;border:1px solid;border-radius:16px;font-size:12px }
</style>
