<script setup lang="ts">
import { useRouter } from 'vue-router'
import { ElButton } from 'element-plus'

const router = useRouter()
</script>

<template>
  <div class="not-found">
    <h1>404</h1>
    <p>页面未找到</p>
    <ElButton type="primary" @click="router.push('/')">返回首页</ElButton>
  </div>
</template>

<style scoped>
.not-found {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  min-height: 60vh;
  color: #909399;
}

.not-found h1 {
  font-size: 72px;
  margin: 0;
  color: #dcdfe6;
}

.not-found p {
  font-size: 18px;
  margin: 16px 0 24px;
}
</style>
