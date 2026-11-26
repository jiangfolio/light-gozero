import { ref } from 'vue'

import type { ResourceNode } from '@/api/system/user'

/** 侧边展开：当前选中的一级菜单。 */
export const selectedTopMenu = ref<ResourceNode | null>(null)
