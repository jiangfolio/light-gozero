<script setup lang="ts">
/**
 * 个人中心-修改密码：旧密 + 新密 + 确认，提交前经公钥 SM2 加密。
 * 复杂度提示取自安全配置接口；改密成功后后端吊销全部会话，前端直接回登录页。
 */
import { message } from 'antdv-next'
import { computed, onMounted, reactive, ref } from 'vue'
import { useRouter } from 'vue-router'

import { changePassword, fetchPublicKey } from '@/api/system/auth'
import { pageSecurityConfig } from '@/api/system/security-config'
import { encryptPassword } from '@/lib/crypto/sm2'
import { fetchAllPages } from '@/lib/tree'
import { resetDynamicRoutes } from '@/router/guard'
import router from '@/router'
import { useUserStore } from '@/stores/user'

const vueRouter = useRouter()
const userStore = useUserStore()

const submitting = ref(false)
const formRef = ref()

const formState = reactive({
  old_password: '',
  new_password: '',
  confirm_password: '',
})

/** 密码策略提示：默认值与后端 appcfg 默认一致，接口拿到后覆盖 */
const policy = reactive({ minLen: 8, complexity: true })

const policyHint = computed(() => {
  const base = `密码长度至少 ${policy.minLen} 位`
  return policy.complexity
    ? `${base}，需同时包含大小写字母、数字和特殊字符`
    : base
})

async function loadPolicy() {
  try {
    const items = await fetchAllPages((page, pageSize) =>
      pageSecurityConfig({ page: String(page), page_size: String(pageSize) }),
    )
    const map: Record<string, string> = {}
    for (const it of items) {
      if (it?.config_key) map[it.config_key] = String(it.config_value ?? '')
    }
    const minLen = Number(map['auth.password_min_len'])
    if (Number.isFinite(minLen) && minLen > 0) policy.minLen = minLen
    const complexity = map['auth.password_complexity_enabled']
    if (complexity != null && complexity !== '') {
      policy.complexity = complexity === 'true' || complexity === '1'
    }
  } catch {
    // 策略拉取失败不阻断改密，按默认提示展示
  }
}

async function onSubmit() {
  await formRef.value?.validate()
  submitting.value = true
  try {
    // 1. 取公钥，新旧密码分别加密
    const { public_key } = await fetchPublicKey()
    // 2. 提交改密；成功后服务端会吊销该用户全部会话
    await changePassword({
      old_password: encryptPassword(formState.old_password, public_key),
      new_password: encryptPassword(formState.new_password, public_key),
    })
    // 3. 清本地会话与动态路由，回登录页重新登录
    message.success('密码已修改，请重新登录')
    resetDynamicRoutes(router)
    await userStore.logout(false)
    await vueRouter.replace('/login')
  } finally {
    submitting.value = false
  }
}

const rules = {
  old_password: [{ required: true, message: '请输入原密码' }],
  new_password: [
    { required: true, message: '请输入新密码' },
    {
      validator: (_rule: unknown, value: string) => {
        if (!value) return Promise.resolve()
        if (value.length < policy.minLen) {
          return Promise.reject(`密码长度至少 ${policy.minLen} 位`)
        }
        if (value === formState.old_password) {
          return Promise.reject('新密码不能与原密码相同')
        }
        return Promise.resolve()
      },
    },
  ],
  confirm_password: [
    { required: true, message: '请再次输入新密码' },
    {
      validator: (_rule: unknown, value: string) => {
        if (!value || value === formState.new_password) return Promise.resolve()
        return Promise.reject('两次输入的密码不一致')
      },
    },
  ],
}

onMounted(loadPolicy)
</script>

<template>
  <ACard size="small" :bordered="false" class="panel-card" title="修改密码">
    <div class="panel-scroll">
      <AForm
        ref="formRef"
        :model="formState"
        :rules="rules"
        :label-col="{ style: { width: '90px' } }"
        class="password-form"
      >
        <AFormItem label="原密码" name="old_password">
          <AInputPassword
            v-model:value="formState.old_password"
            placeholder="请输入原密码"
            autocomplete="current-password"
          />
        </AFormItem>
        <AFormItem label="新密码" name="new_password">
          <AInputPassword
            v-model:value="formState.new_password"
            placeholder="请输入新密码"
            autocomplete="new-password"
          />
          <div class="hint">{{ policyHint }}</div>
        </AFormItem>
        <AFormItem label="确认密码" name="confirm_password">
          <AInputPassword
            v-model:value="formState.confirm_password"
            placeholder="请再次输入新密码"
            autocomplete="new-password"
          />
        </AFormItem>
        <AFormItem class="form-actions">
          <AButton type="primary" :loading="submitting" @click="onSubmit">修改密码</AButton>
        </AFormItem>
      </AForm>
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

/* 内容超出时仅本区域滚动，侧边栏保持不动 */
.panel-scroll {
  flex: 1;
  min-height: 0;
  overflow-x: hidden;
  overflow-y: auto;
}

.password-form {
  max-width: 460px;
}

.hint {
  margin-top: 4px;
  font-size: 12px;
  line-height: 1.5;
  color: #999;
}

.form-actions {
  margin-bottom: 0;
}
</style>
