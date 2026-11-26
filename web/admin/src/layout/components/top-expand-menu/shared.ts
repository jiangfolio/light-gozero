import { ref } from 'vue'

import type { ResourceNode } from '@/api/system/user'

/** 顶部展开：当前选中的一级菜单。 */
export const selectedTopMenu = ref<ResourceNode | null>(null)

/** 是否存在二级侧栏，用于计算顶栏覆盖层宽度。 */
export const hasSideMenu = ref(false)
