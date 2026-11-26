<script setup lang="ts">
/**
 * 个人中心-基本资料：展示账号信息，昵称可编辑保存（复用 user/update 接口）。
 * 用户模型暂无手机/邮箱/头像字段，不展示对应表单项。
 */
import { message } from 'antdv-next'
import { computed, onMounted, reactive, ref } from 'vue'

import { treeOrg } from '@/api/system/org'
import { getUserDetail, updateUser } from '@/api/system/user'
import { setUserInfo } from '@/lib/auth-token'
import { useUserStore } from '@/stores/user'

const userStore = useUserStore()

const loading = ref(false)
const saving = ref(false)
const formRef = ref()

const formState = reactive({
  username: '',
  nickname: '',
  orgName: '',
  status: '',
  createdAt: '',
  updatedAt: '',
})

const avatarText = computed(() => (formState.nickname || formState.username || '用').slice(0, 1))

/** 在组织树里按 id 递归找名称 */
function findOrgName(nodes: any[], id: string): string {
  for (const n of nodes || []) {
    if (n?.id === id) return n?.name || ''
    const hit = findOrgName(n?.children, id)
    if (hit) return hit
  }
  return ''
}

async function loadData() {
  const uid = userStore.userInfo?.id
  if (!uid) return
  loading.value = true
  try {
    // 1. 拉用户详情与组织树（并行）
    const [detail, orgTree] = await Promise.all([getUserDetail(uid), treeOrg()])
    // 2. 回填表单；组织只展示名称，不可改
    formState.username = detail?.username || ''
    formState.nickname = detail?.nickname || ''
    formState.status = detail?.status || ''
    formState.createdAt = detail?.created_at || ''
    formState.updatedAt = detail?.updated_at || ''
    formState.orgName = findOrgName(orgTree?.items ?? orgTree ?? [], detail?.org_id) || '-'
  } catch (e: unknown) {
    const err = e as { message?: string }
    message.error(err.message || '加载个人信息失败')
  } finally {
    loading.value = false
  }
}

async function onSave() {
  const uid = userStore.userInfo?.id
  if (!uid) return
  await formRef.value?.validate()
  saving.value = true
  try {
    // 1. 仅提交可编辑字段，其余由后端保持
    await updateUser({ id: uid, nickname: formState.nickname.trim() })
    // 2. 同步本地会话信息，头部昵称立即生效
    if (userStore.userInfo) {
      userStore.userInfo.nickname = formState.nickname.trim()
      setUserInfo(userStore.userInfo)
    }
    message.success('已保存')
  } finally {
    saving.value = false
  }
}

onMounted(loadData)
</script>

<template>
  <ACard size="small" :bordered="false" class="panel-card" title="基本资料">
    <ASpin :spinning="loading">
      <div class="panel-scroll">
        <div class="profile-body">
          <AForm
            ref="formRef"
            :model="formState"
            :label-col="{ style: { width: '90px' } }"
            class="profile-form"
          >
            <AFormItem label="登录账号">
              <AInput v-model:value="formState.username" disabled />
            </AFormItem>
            <AFormItem label="所属组织">
              <AInput v-model:value="formState.orgName" disabled />
            </AFormItem>
            <AFormItem
              label="昵称"
              name="nickname"
              :rules="[{ required: true, message: '请输入昵称' }]"
            >
              <AInput
                v-model:value="formState.nickname"
                placeholder="请输入昵称"
                :maxlength="32"
                allow-clear
              />
            </AFormItem>
            <AFormItem label="注册时间">
              <AInput v-model:value="formState.createdAt" disabled />
            </AFormItem>
            <AFormItem label="最近更新">
              <AInput v-model:value="formState.updatedAt" disabled />
            </AFormItem>
            <AFormItem :wrapper-col="{ offset: 0 }" class="form-actions">
              <AButton type="primary" :loading="saving" @click="onSave">保存</AButton>
            </AFormItem>
          </AForm>

          <div class="profile-avatar">
            <AAvatar :size="88" style="background: #1677ff; font-size: 32px">
              {{ avatarText }}
            </AAvatar>
            <div class="avatar-tip">头像暂不支持上传</div>
          </div>
        </div>
      </div>
    </ASpin>
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

/* 内容超出时仅本区域滚动，侧边栏保持不动 */
.panel-scroll {
  flex: 1;
  min-height: 0;
  overflow-x: hidden;
  overflow-y: auto;
}

.profile-body {
  display: flex;
  gap: 32px;
  align-items: flex-start;
}

.profile-form {
  flex: 1;
  max-width: 460px;
}

.form-actions {
  margin-bottom: 0;
}

.profile-avatar {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 8px;
  padding-top: 8px;
}

.avatar-tip {
  font-size: 12px;
  color: #aaa;
}
</style>
