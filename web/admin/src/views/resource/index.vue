<script setup lang="ts">
defineOptions({ name: 'SystemResource' })

const title = '资源管理'
const permCreate = 'system:resource:create'
const permUpdate = 'system:resource:update'
</script>

<template>
  <div>
    <ACard :title="title">
      <p>业务页占位，路由已可解析。后续对接接口。</p>
      <ASpace>
        <AButton v-privilege="permCreate" type="primary">新建（权限演示）</AButton>
        <AButton v-privilege="permUpdate">编辑（权限演示）</AButton>
      </ASpace>
    </ACard>
  </div>
</template>
