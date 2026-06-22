<script setup lang="ts">
import { ref } from 'vue'
import { ElForm, ElFormItem, ElInput, ElSelect, ElOption, ElButton, ElMessage } from 'element-plus'
import { createQuestion } from '@/api/modules/admin'
import type { CreateQuestionRequest } from '@/types/admin'

const form = ref<CreateQuestionRequest>({
  question: '',
  options: ['', '', '', ''],
  answer: '',
  difficulty: 'easy',
  category: 'history',
})

const submitting = ref(false)

async function handleSubmit() {
  const validOptions = form.value.options.filter((o) => o.trim())
  if (validOptions.length < 2) {
    ElMessage.warning('至少填写 2 个选项')
    return
  }
  submitting.value = true
  try {
    await createQuestion({ ...form.value, options: validOptions })
    ElMessage.success('题目添加成功')
    form.value = { question: '', options: ['', '', '', ''], answer: '', difficulty: 'easy', category: 'history' }
  } catch {
    // handled by interceptor
  } finally {
    submitting.value = false
  }
}
</script>

<template>
  <div class="question-management">
    <h2>❓ 题目管理</h2>

    <div class="form-card">
      <ElForm :model="form" label-width="80px" label-position="left">
        <ElFormItem label="题目">
          <ElInput v-model="form.question" type="textarea" :rows="3" placeholder="请输入题目" />
        </ElFormItem>

        <ElFormItem v-for="(_, i) in form.options" :key="i" :label="`选项 ${String.fromCharCode(65 + i)}`">
          <ElInput v-model="form.options[i]" :placeholder="`选项 ${String.fromCharCode(65 + i)}`" />
        </ElFormItem>

        <ElRow :gutter="16">
          <ElCol :span="8">
            <ElFormItem label="正确答案">
              <el-select v-model="form.answer" placeholder="选择正确答案" style="width: 100%">
                <el-option v-for="(opt, i) in form.options.filter((o) => o.trim())" :key="i" :label="`选项 ${String.fromCharCode(65 + i)}`" :value="String.fromCharCode(65 + i)" />
              </el-select>
            </ElFormItem>
          </ElCol>
          <ElCol :span="8">
            <ElFormItem label="难度">
              <el-select v-model="form.difficulty" style="width: 100%">
                <el-option label="简单" value="easy" />
                <el-option label="中等" value="medium" />
                <el-option label="困难" value="hard" />
              </el-select>
            </ElFormItem>
          </ElCol>
          <ElCol :span="8">
            <ElFormItem label="分类">
              <el-select v-model="form.category" style="width: 100%">
                <el-option label="历史" value="history" />
                <el-option label="生态" value="ecology" />
                <el-option label="文化" value="culture" />
                <el-option label="地理" value="geography" />
                <el-option label="水利" value="water_conservancy" />
              </el-select>
            </ElFormItem>
          </ElCol>
        </ElRow>

        <ElFormItem label="解析">
          <ElInput v-model="form.explanation" type="textarea" :rows="2" placeholder="答案解析（可选）" />
        </ElFormItem>

        <ElFormItem>
          <el-button type="primary" :loading="submitting" @click="handleSubmit">添加题目</el-button>
        </ElFormItem>
      </ElForm>
    </div>
  </div>
</template>

<script lang="ts">
import { ElRow, ElCol } from 'element-plus'
</script>

<style scoped>
.question-management {
  max-width: 800px;
}

.question-management h2 {
  font-size: 22px;
  color: #303133;
  margin: 0 0 20px;
}

.form-card {
  background: #fff;
  border-radius: 12px;
  padding: 32px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.06);
}
</style>
