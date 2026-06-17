<template>
  <div class="post-create">
    <h2>发布帖子</h2>
    <el-form :model="form" :rules="rules" ref="formRef" label-width="0">
      <el-form-item prop="title">
        <el-input v-model="form.title" placeholder="标题" size="large" />
      </el-form-item>
      <el-form-item prop="content">
        <el-input v-model="form.content" type="textarea" :rows="8" placeholder="分享你的运河故事..." />
      </el-form-item>
      <el-form-item>
        <el-select v-model="form.topic" placeholder="选择话题">
          <el-option label="分享" value="share" />
          <el-option label="生态" value="ecology" />
          <el-option label="文化" value="culture" />
          <el-option label="问答" value="question" />
        </el-select>
      </el-form-item>
      <el-form-item>
        <el-button type="primary" @click="publish" :loading="publishing">发布</el-button>
        <el-button @click="$router.back()">取消</el-button>
      </el-form-item>
    </el-form>
  </div>
</template>

<script setup lang="ts">
import { reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import { postsApi } from '@/api/posts'
import { ElMessage } from 'element-plus'

const router = useRouter()
const formRef = ref()
const publishing = ref(false)

const form = reactive({ title: '', content: '', topic: 'share' })
const rules = {
  title: [{ required: true, message: '请输入标题', trigger: 'blur' }],
  content: [{ required: true, message: '请输入内容', trigger: 'blur' }],
}

async function publish() {
  await formRef.value?.validate()
  publishing.value = true
  try {
    const res = await postsApi.create(form)
    ElMessage.success('发布成功')
    router.push(`/posts/${res.data.data!.id}`)
  } catch { ElMessage.error('发布失败') }
  finally { publishing.value = false }
}
</script>

<style scoped>
.post-create { max-width: 700px; margin: 0 auto; }
h2 { margin-bottom: 20px; }
</style>
