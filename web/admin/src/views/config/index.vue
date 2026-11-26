<script setup lang="ts">
defineOptions({ name: 'SystemConfig' })

const title = '系统配置'
const permCreate = 'system:config:update'
const permUpdate = 'system:config:detail'
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
