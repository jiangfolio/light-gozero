<script setup lang="ts">
defineOptions({ name: 'NotFoundComponent' })
</script>

<template>
  <AResult status="warning" title="组件未找到" sub-title="请检查资源 component_path 是否对应 views 下页面。" />
</template>
