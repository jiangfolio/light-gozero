<script setup lang="ts">
/**
 * 个人中心占位面板：用于我的消息/通知公告等尚未接入后端的能力。
 * 明确展示演示占位，避免空白或假数据误导。
 */
defineProps<{ title: string; description: string }>()
</script>

<template>
  <ACard size="small" :bordered="false" class="panel-card" :title="title">
    <div class="panel-scroll">
      <AEmpty :description="description" class="placeholder-empty">
        <template #image>
          <IconParkInbox :size="56" style="color: #d9d9d9" />
        </template>
      </AEmpty>
      <div class="placeholder-tip">该能力暂未开放，接入真实接口后在此替换本占位</div>
    </div>
  </ACard>
</template>

<style scoped lang="less">
.panel-card {
  display: flex;
  flex: 1;
  flex-direction: column;
  min-height: 0;

  :deep(.ant-card-body) {
    display: flex;
    flex: 1;
    flex-direction: column;
    min-height: 0;
    padding: 12px;
  }
}

.panel-scroll {
  display: flex;
  flex: 1;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  min-height: 0;
  overflow-x: hidden;
  overflow-y: auto;
}

.placeholder-empty {
  margin: 0;
}

.placeholder-tip {
  margin-top: 12px;
  font-size: 12px;
  color: #bbb;
}
</style>
