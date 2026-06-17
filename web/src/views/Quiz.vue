<template>
  <div class="quiz-page">
    <template v-if="!session">
      <div class="quiz-start">
        <h2>趣味问答</h2>
        <p>测试你对大运河的了解！</p>
        <div class="start-options">
          <el-select v-model="difficulty" placeholder="难度">
            <el-option label="全部" value="all" />
            <el-option label="简单" value="easy" />
            <el-option label="中等" value="medium" />
            <el-option label="困难" value="hard" />
          </el-select>
          <el-button type="primary" size="large" @click="start" :loading="starting">
            开始答题
          </el-button>
        </div>
      </div>
    </template>

    <template v-else>
      <div class="quiz-progress">
        <el-progress :percentage="progress" :stroke-width="8" />
        <span>{{ session.total_index }} / {{ session.total_questions }}</span>
        <el-tag>得分: {{ session.total_score }}</el-tag>
      </div>

      <div class="question-card" v-if="currentQuestion">
        <h3>{{ currentQuestion.content }}</h3>
        <div class="options">
          <el-button
            v-for="(opt, idx) in currentQuestion.options"
            :key="idx"
            :type="optionType(idx)"
            :disabled="answered"
            @click="select(idx)"
            class="option-btn"
          >
            {{ ['A', 'B', 'C', 'D'][idx] }}. {{ opt }}
          </el-button>
        </div>
        <div v-if="answered" class="result">
          <el-tag :type="lastResult?.correct ? 'success' : 'danger'" size="large">
            {{ lastResult?.correct ? '✓ 正确' : '✗ 错误' }}
            <template v-if="lastResult?.correct"> +{{ lastResult?.score }} 分</template>
          </el-tag>
          <el-button type="primary" @click="nextQuestion" style="margin-left:12px">
            {{ session.has_more ? '下一题' : '查看结果' }}
          </el-button>
        </div>
      </div>
    </template>

    <div v-if="finished" class="quiz-result">
      <h2>🎉 答题完成</h2>
      <p>总分: {{ session.total_score }}</p>
      <el-button type="primary" @click="reset">再来一轮</el-button>
      <el-button @click="$router.push('/leaderboard')">查看排行榜</el-button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, reactive } from 'vue'
import { quizApi } from '@/api/quiz'
import { ElMessage } from 'element-plus'
import type { Question, SubmitResponse } from '@/types'

const difficulty = ref('all')
const starting = ref(false)
const sessionId = ref('')
const questions = ref<Question[]>([])
const currentIdx = ref(0)
const answered = ref(false)
const finished = ref(false)
const lastResult = ref<SubmitResponse | null>(null)
const selectedOption = ref(-1)
const isCorrect = ref(false)

const session = reactive({
  total_score: 0, streak: 0, total_index: 0, total_questions: 0, has_more: true,
})

const currentQuestion = computed(() => questions.value[currentIdx.value] || null)
const progress = computed(() => {
  if (session.total_questions === 0) return 0
  return Math.round((currentIdx.value / session.total_questions) * 100)
})

function optionType(idx: number) {
  if (!answered.value) return ''
  if (idx === selectedOption.value) return isCorrect.value ? 'success' : 'danger'
  return ''
}

async function start() {
  starting.value = true
  try {
    const res = await quizApi.startSession(difficulty.value, 10)
    const data = res.data.data!
    sessionId.value = data.session_id
    questions.value = data.questions
    session.total_questions = data.questions.length
    currentIdx.value = 0
    session.total_score = 0
    session.total_index = 0
    session.has_more = true
    finished.value = false
  } catch { ElMessage.error('加载题目失败') }
  finally { starting.value = false }
}

async function select(idx: number) {
  if (answered.value) return
  selectedOption.value = idx
  answered.value = true

  try {
    const res = await quizApi.submit(sessionId.value, currentQuestion.value!.question_id, idx)
    const data = res.data.data!
    lastResult.value = data
    isCorrect.value = data.correct
    session.total_score = data.total_score
    session.streak = data.streak
    session.total_index = data.total_index
    session.has_more = data.has_more
  } catch { ElMessage.error('提交失败') }
}

function nextQuestion() {
  if (!session.has_more) {
    finished.value = true
    return
  }
  answered.value = false
  selectedOption.value = -1
  lastResult.value = null
  currentIdx.value++
}

function reset() {
  sessionId.value = ''
  questions.value = []
  currentIdx.value = 0
  answered.value = false
  finished.value = false
  lastResult.value = null
  session.total_score = 0
}

// 未登录时隐藏 start options 中的登录要求 — 已由后端中间件处理
</script>

<style scoped>
.quiz-page { max-width: 700px; margin: 0 auto; }
.quiz-start { text-align: center; padding: 60px 0; }
.quiz-start h2 { font-size: 28px; margin-bottom: 8px; }
.quiz-start p { color: #909399; margin-bottom: 24px; }
.start-options { display: flex; gap: 12px; justify-content: center; }
.quiz-progress { display: flex; align-items: center; gap: 16px; margin-bottom: 24px; }
.quiz-progress .el-progress { flex: 1; }
.question-card { background: #fff; border-radius: 8px; padding: 32px; }
.question-card h3 { font-size: 18px; margin-bottom: 24px; line-height: 1.6; }
.options { display: flex; flex-direction: column; gap: 12px; }
.option-btn { justify-content: flex-start; height: auto; padding: 12px 16px; white-space: normal; text-align: left; }
.result { display: flex; align-items: center; margin-top: 20px; }
.quiz-result { text-align: center; padding: 60px 0; }
.quiz-result h2 { margin-bottom: 12px; }
.quiz-result p { font-size: 24px; color: #409eff; margin-bottom: 24px; }
</style>
