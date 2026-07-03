<script setup lang="ts">
import { useAuthStore } from '@/stores/auth'

const auth = useAuthStore()
auth.restoreSession()
</script>

<template>
  <router-view />
</template>
