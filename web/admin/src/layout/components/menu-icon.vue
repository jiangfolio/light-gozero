<script setup lang="ts">
/**
 * 将种子里的 Ant Design 图标名映射到 IconPark 离线图标。
 */
defineProps<{ name?: string; size?: number | string }>()

const alias: Record<string, string> = {
  SettingOutlined: 'setting',
  ApartmentOutlined: 'apartment',
  UserOutlined: 'user',
  TeamOutlined: 'team',
  FileSearchOutlined: 'search',
  ToolOutlined: 'tool',
  BookOutlined: 'book',
  Home: 'home',
}
</script>

<template>
  <span class="menu-icon">
    <IconParkSettingTwo v-if="(alias[name || ''] || name) === 'setting'" :size="size || 16" />
    <IconParkBuildingOne
      v-else-if="(alias[name || ''] || name) === 'apartment'"
      :size="size || 16"
    />
    <IconParkUser v-else-if="(alias[name || ''] || name) === 'user'" :size="size || 16" />
    <IconParkPeoples v-else-if="(alias[name || ''] || name) === 'team'" :size="size || 16" />
    <IconParkFileSearch
      v-else-if="(alias[name || ''] || name) === 'search'"
      :size="size || 16"
    />
    <IconParkTool v-else-if="(alias[name || ''] || name) === 'tool'" :size="size || 16" />
    <IconParkBookOpen v-else-if="(alias[name || ''] || name) === 'book'" :size="size || 16" />
    <IconParkHome v-else-if="(alias[name || ''] || name) === 'home'" :size="size || 16" />
    <IconParkApplicationOne v-else :size="size || 16" />
  </span>
</template>

<style scoped>
.menu-icon {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  vertical-align: middle;
}
</style>
