<script setup lang="ts">
/**
 * IconPark 图标选择器：点击展开面板检索并点选，写入图标名供菜单等复用。
 */
import { computed, ref, watch } from 'vue'

import MenuIcon from '@/layout/components/menu-icon.vue'
import { filterIconParkNames } from '@/lib/icon-park'

const props = withDefaults(
  defineProps<{
    value?: string
    placeholder?: string
    disabled?: boolean
    allowClear?: boolean
  }>(),
  {
    value: '',
    placeholder: '请选择图标',
    disabled: false,
    allowClear: true,
  },
)

const emit = defineEmits<{
  'update:value': [string]
}>()

const open = ref(false)
const keyword = ref('')

const displayValue = computed(() => (props.value || '').trim())

const iconNames = computed(() => filterIconParkNames(keyword.value, 160))

watch(open, (v) => {
  if (!v) keyword.value = ''
})

/** 选中图标并关闭面板 */
function selectIcon(name: string) {
  // 1. 写回 v-model
  // 2. 关闭浮层，避免残留搜索态
  emit('update:value', name)
  open.value = false
}

/** 清空已选 */
function clearIcon(e: Event) {
  e.stopPropagation()
  emit('update:value', '')
}
</script>

<template>
  <APopover
    v-model:open="open"
    placement="bottomLeft"
    :trigger="disabled ? [] : ['click']"
  >
    <template #content>
      <div class="icon-picker-panel">
        <AInput
          v-model:value="keyword"
          allow-clear
          placeholder="搜索图标名称，如 user、setting"
        />
        <div class="icon-picker-hint">
          {{
            keyword.trim()
              ? `匹配 ${iconNames.length} 个（最多展示 160）`
              : '输入关键词可筛选全部图标'
          }}
        </div>
        <div class="icon-picker-grid">
          <button
            v-for="name in iconNames"
            :key="name"
            type="button"
            class="icon-picker-item"
            :class="{ active: name === displayValue }"
            :title="name"
            @click="selectIcon(name)"
          >
            <MenuIcon :name="name" :size="20" />
            <span class="icon-picker-name">{{ name }}</span>
          </button>
          <div v-if="!iconNames.length" class="icon-picker-empty">无匹配图标</div>
        </div>
      </div>
    </template>

    <div
      class="icon-picker-trigger"
      :class="{ disabled, open }"
      role="button"
      :tabindex="disabled ? -1 : 0"
    >
      <span class="icon-picker-prefix">
        <MenuIcon v-if="displayValue" :name="displayValue" :size="16" />
        <span v-else class="icon-picker-placeholder-dot" />
      </span>
      <span class="icon-picker-text" :class="{ placeholder: !displayValue }">
        {{ displayValue || placeholder }}
      </span>
      <button
        v-if="allowClear && displayValue && !disabled"
        type="button"
        class="icon-picker-clear"
        aria-label="清除"
        @click="clearIcon"
      >
        ×
      </button>
      <span class="icon-picker-arrow">▾</span>
    </div>
  </APopover>
</template>

<style scoped>
.icon-picker-trigger {
  display: flex;
  width: 100%;
  min-height: 32px;
  align-items: center;
  gap: 8px;
  padding: 4px 11px;
  border: 1px solid var(--ant-color-border, #d9d9d9);
  border-radius: var(--app-border-radius, 6px);
  background: var(--ant-color-bg-container, #fff);
  cursor: pointer;
  transition: border-color 0.2s, box-shadow 0.2s;
}

.icon-picker-trigger:hover,
.icon-picker-trigger.open {
  border-color: var(--ant-color-primary, #1677ff);
}

.icon-picker-trigger.open {
  box-shadow: 0 0 0 2px color-mix(in srgb, var(--ant-color-primary, #1677ff) 20%, transparent);
}

.icon-picker-trigger.disabled {
  cursor: not-allowed;
  background: var(--ant-color-bg-container-disabled, #f5f5f5);
  opacity: 0.7;
}

.icon-picker-prefix {
  display: inline-flex;
  width: 16px;
  flex-shrink: 0;
  align-items: center;
  justify-content: center;
  color: var(--ant-color-text, rgba(0, 0, 0, 0.88));
}

.icon-picker-placeholder-dot {
  display: inline-block;
  width: 10px;
  height: 10px;
  border-radius: 50%;
  background: var(--ant-color-fill-secondary, #f0f0f0);
}

.icon-picker-text {
  flex: 1;
  min-width: 0;
  overflow: hidden;
  color: var(--ant-color-text, rgba(0, 0, 0, 0.88));
  font-size: 14px;
  line-height: 1.5;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.icon-picker-text.placeholder {
  color: var(--ant-color-text-placeholder, rgba(0, 0, 0, 0.25));
}

.icon-picker-clear {
  display: inline-flex;
  width: 16px;
  height: 16px;
  flex-shrink: 0;
  align-items: center;
  justify-content: center;
  border: 0;
  border-radius: 50%;
  background: var(--ant-color-text-quaternary, rgba(0, 0, 0, 0.25));
  color: #fff;
  cursor: pointer;
  font-size: 12px;
  line-height: 1;
  padding: 0;
}

.icon-picker-clear:hover {
  background: var(--ant-color-text-tertiary, rgba(0, 0, 0, 0.45));
}

.icon-picker-arrow {
  flex-shrink: 0;
  color: var(--ant-color-text-quaternary, rgba(0, 0, 0, 0.25));
  font-size: 12px;
}

.icon-picker-panel {
  width: 380px;
  max-width: 70vw;
}

.icon-picker-hint {
  margin: 8px 0;
  color: var(--ant-color-text-secondary, rgba(0, 0, 0, 0.45));
  font-size: 12px;
}

.icon-picker-grid {
  display: grid;
  max-height: 280px;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: 6px;
  overflow: auto;
  padding-right: 2px;
}

.icon-picker-item {
  display: flex;
  min-width: 0;
  flex-direction: column;
  align-items: center;
  gap: 4px;
  padding: 8px 4px;
  border: 1px solid transparent;
  border-radius: var(--app-border-radius, 6px);
  background: var(--ant-color-fill-alter, #fafafa);
  color: var(--ant-color-text, rgba(0, 0, 0, 0.88));
  cursor: pointer;
  transition: border-color 0.15s, background 0.15s, color 0.15s;
}

.icon-picker-item:hover {
  border-color: var(--ant-color-primary-border, #91caff);
  background: var(--ant-color-primary-bg, #e6f4ff);
  color: var(--ant-color-primary, #1677ff);
}

.icon-picker-item.active {
  border-color: var(--ant-color-primary, #1677ff);
  background: var(--ant-color-primary-bg, #e6f4ff);
  color: var(--ant-color-primary, #1677ff);
}

.icon-picker-name {
  width: 100%;
  overflow: hidden;
  font-size: 11px;
  line-height: 1.2;
  text-align: center;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.icon-picker-empty {
  grid-column: 1 / -1;
  padding: 24px 0;
  color: var(--ant-color-text-secondary, rgba(0, 0, 0, 0.45));
  font-size: 13px;
  text-align: center;
}
</style>
