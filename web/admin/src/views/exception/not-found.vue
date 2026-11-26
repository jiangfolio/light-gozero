<script setup lang="ts">
import { useRouter } from 'vue-router'

const router = useRouter()
</script>

<template>
  <div class="not-found">
    <AResult status="404" title="404" sub-title="页面不存在或无权访问">
      <template #extra>
        <AButton type="primary" @click="router.push('/home')">回首页</AButton>
      </template>
    </AResult>
  </div>
</template>

<style scoped lang="less">
.not-found {
  min-height: 60vh;
  display: flex;
  align-items: center;
  justify-content: center;
}
</style>
