<script setup lang="ts">
import { message } from 'antdv-next'
import { onMounted, reactive, ref } from 'vue'

import { pageSecurityConfig, updateSecurityConfig } from '@/api/system/security-config'
import { fetchAllPages } from '@/lib/tree'

defineOptions({ name: 'SecurityConfig' })

/** 仅映射系统已落地能力的配置键；会话分钟保存时换算为秒 */
const KEY = {
  maxFail: 'auth.max_fail_attempts',
  lockMinutes: 'auth.lock_minutes',
  tokenTtlSeconds: 'auth.token_ttl_seconds',
  passwordComplexity: 'auth.password_complexity_enabled',
  passwordMinLen: 'auth.password_min_len',
  passwordExpireDays: 'auth.password_expire_days',
  passwordHistory: 'auth.password_history_count',
  captchaTtl: 'captcha.ttl_seconds',
  captchaLen: 'captcha.length',
} as const

const activeTab = ref('login')
const loading = ref(false)
const submitting = ref(false)
const formRef = ref()

const formState = reactive({
  maxFail: 5,
  lockMinutes: 30,
  tokenMinutes: 1440,
  passwordComplexity: true,
  passwordMinLen: 8,
  passwordExpireDays: 90,
  passwordHistory: 5,
  captchaTtl: 120,
  captchaLen: 4,
})

function parseBool(v: string | undefined, def = false) {
  if (v == null || v === '') return def
  return v === 'true' || v === '1'
}

function parseIntSafe(v: string | undefined, def: number) {
  const n = Number(v)
  return Number.isFinite(n) ? n : def
}

async function loadData() {
  loading.value = true
  try {
    const items = await fetchAllPages((page, pageSize) =>
      pageSecurityConfig({ page: String(page), page_size: String(pageSize) }),
    )
    const map: Record<string, string> = {}
    for (const it of items) {
      if (it?.config_key) map[it.config_key] = String(it.config_value ?? '')
    }
    formState.maxFail = parseIntSafe(map[KEY.maxFail], 5)
    formState.lockMinutes = parseIntSafe(map[KEY.lockMinutes], 30)
    const ttlSec = parseIntSafe(map[KEY.tokenTtlSeconds], 86400)
    formState.tokenMinutes = Math.max(1, Math.round(ttlSec / 60))
    formState.passwordComplexity = parseBool(map[KEY.passwordComplexity], true)
    formState.passwordMinLen = parseIntSafe(map[KEY.passwordMinLen], 8)
    formState.passwordExpireDays = parseIntSafe(map[KEY.passwordExpireDays], 90)
    formState.passwordHistory = parseIntSafe(map[KEY.passwordHistory], 5)
    formState.captchaTtl = parseIntSafe(map[KEY.captchaTtl], 120)
    formState.captchaLen = parseIntSafe(map[KEY.captchaLen], 4)
  } catch (e: unknown) {
    const err = e as { message?: string }
    message.error(err.message || '加载安全配置失败')
  } finally {
    loading.value = false
  }
}

async function onSave() {
  try {
    await formRef.value?.validate()
  } catch {
    return
  }
  submitting.value = true
  try {
    // 1. 会话分钟换算为秒写入 token_ttl；其余字段原样提交
    // 2. 布尔写 true/false，与库中 value_type=bool 一致
    const ttlSeconds = Math.round(Number(formState.tokenMinutes) * 60)
    await updateSecurityConfig([
      { config_key: KEY.maxFail, config_value: String(formState.maxFail) },
      { config_key: KEY.lockMinutes, config_value: String(formState.lockMinutes) },
      { config_key: KEY.tokenTtlSeconds, config_value: String(ttlSeconds) },
      {
        config_key: KEY.passwordComplexity,
        config_value: formState.passwordComplexity ? 'true' : 'false',
      },
      { config_key: KEY.passwordMinLen, config_value: String(formState.passwordMinLen) },
      { config_key: KEY.passwordExpireDays, config_value: String(formState.passwordExpireDays) },
      { config_key: KEY.passwordHistory, config_value: String(formState.passwordHistory) },
      { config_key: KEY.captchaTtl, config_value: String(formState.captchaTtl) },
      { config_key: KEY.captchaLen, config_value: String(formState.captchaLen) },
    ])
    message.success('保存成功')
    await loadData()
  } catch (e: unknown) {
    const err = e as { message?: string }
    message.error(err.message || '保存失败')
  } finally {
    submitting.value = false
  }
}

onMounted(() => {
  loadData()
})
</script>

<template>
  <div class="page">
    <ASpin :spinning="loading" class="page-spin">
      <ACard variant="borderless" class="form-card" :bordered="false">
        <AForm ref="formRef" :model="formState" layout="vertical" class="sec-form">
          <div class="sec-body">
            <ATabs v-model:active-key="activeTab" class="top-tabs" tab-position="top">
              <ATabPane key="login" tab="登录配置" force-render>
                <div class="panel">
                  <div class="section-title">登录配置</div>
                  <ARow :gutter="[24, 8]">
                    <ACol :xs="24" :md="12">
                      <AFormItem
                        label="最大连续登录失败次数"
                        name="maxFail"
                        required
                        :rules="[{ required: true, message: '请输入失败次数' }]"
                      >
                        <AInputNumber
                          v-model:value="formState.maxFail"
                          :min="1"
                          :max="20"
                          :precision="0"
                          class="field-control"
                          addon-after="次"
                        />
                        <div class="hint">连续登录失败超过该次数后锁定账号；默认 5 次</div>
                      </AFormItem>
                    </ACol>
                    <ACol :xs="24" :md="12">
                      <AFormItem
                        label="连续登录失败锁定分钟"
                        name="lockMinutes"
                        required
                        :rules="[{ required: true, message: '请输入锁定分钟' }]"
                      >
                        <AInputNumber
                          v-model:value="formState.lockMinutes"
                          :min="1"
                          :max="1440"
                          :precision="0"
                          class="field-control"
                          addon-after="分钟"
                        />
                        <div class="hint">连续登录失败锁定时长；默认 30 分钟</div>
                      </AFormItem>
                    </ACol>
                    <ACol :xs="24" :md="12">
                      <AFormItem
                        label="会话 Token 有效期"
                        name="tokenMinutes"
                        required
                        :rules="[{ required: true, message: '请输入会话有效期' }]"
                      >
                        <AInputNumber
                          v-model:value="formState.tokenMinutes"
                          :min="1"
                          :max="10080"
                          :precision="0"
                          class="field-control"
                          addon-after="分钟"
                        />
                        <div class="hint">
                          对应 auth.token_ttl_seconds；例如 60 表示 1 小时后需重新登录；默认 1440
                          分钟
                        </div>
                      </AFormItem>
                    </ACol>
                  </ARow>
                </div>
              </ATabPane>

              <ATabPane key="password" tab="密码配置" force-render>
                <div class="panel">
                  <div class="section-title">密码配置</div>
                  <ARow :gutter="[24, 8]">
                    <ACol :xs="24" :md="12">
                      <AFormItem label="开启密码复杂度" name="passwordComplexity">
                        <ASwitch
                          v-model:checked="formState.passwordComplexity"
                          checked-children="开启"
                          un-checked-children="关闭"
                        />
                        <div class="hint">开启后密码需同时包含大小写字母、数字和特殊字符</div>
                      </AFormItem>
                    </ACol>
                    <ACol :xs="24" :md="12">
                      <AFormItem
                        label="密码最小长度"
                        name="passwordMinLen"
                        required
                        :rules="[{ required: true, message: '请输入密码最小长度' }]"
                      >
                        <AInputNumber
                          v-model:value="formState.passwordMinLen"
                          :min="6"
                          :max="64"
                          :precision="0"
                          class="field-control"
                          addon-after="位"
                        />
                        <div class="hint">新建/修改密码时的最小长度；默认 8 位</div>
                      </AFormItem>
                    </ACol>
                    <ACol :xs="24" :md="12">
                      <AFormItem
                        label="密码过期天数"
                        name="passwordExpireDays"
                        required
                        :rules="[{ required: true, message: '请输入密码过期天数' }]"
                      >
                        <AInputNumber
                          v-model:value="formState.passwordExpireDays"
                          :min="0"
                          :max="3650"
                          :precision="0"
                          class="field-control"
                          addon-after="天"
                        />
                        <div class="hint">超过该天数需强制修改；0 表示不过期；默认 90 天</div>
                      </AFormItem>
                    </ACol>
                    <ACol :xs="24" :md="12">
                      <AFormItem
                        label="禁止复用历史密码条数"
                        name="passwordHistory"
                        required
                        :rules="[{ required: true, message: '请输入历史密码条数' }]"
                      >
                        <AInputNumber
                          v-model:value="formState.passwordHistory"
                          :min="0"
                          :max="24"
                          :precision="0"
                          class="field-control"
                          addon-after="条"
                        />
                        <div class="hint">禁止与最近 N 次历史密码相同；0 表示不限制</div>
                      </AFormItem>
                    </ACol>
                  </ARow>
                </div>
              </ATabPane>

              <ATabPane key="captcha" tab="验证码配置" force-render>
                <div class="panel">
                  <div class="section-title">验证码配置</div>
                  <ARow :gutter="[24, 8]">
                    <ACol :xs="24" :md="12">
                      <AFormItem
                        label="验证码有效期"
                        name="captchaTtl"
                        required
                        :rules="[{ required: true, message: '请输入验证码有效期' }]"
                      >
                        <AInputNumber
                          v-model:value="formState.captchaTtl"
                          :min="30"
                          :max="600"
                          :precision="0"
                          class="field-control"
                          addon-after="秒"
                        />
                        <div class="hint">图形验证码有效时间；默认 120 秒</div>
                      </AFormItem>
                    </ACol>
                    <ACol :xs="24" :md="12">
                      <AFormItem
                        label="验证码位数"
                        name="captchaLen"
                        required
                        :rules="[{ required: true, message: '请输入验证码位数' }]"
                      >
                        <AInputNumber
                          v-model:value="formState.captchaLen"
                          :min="4"
                          :max="8"
                          :precision="0"
                          class="field-control"
                          addon-after="位"
                        />
                        <div class="hint">图形验证码字符数；默认 4 位</div>
                      </AFormItem>
                    </ACol>
                  </ARow>
                </div>
              </ATabPane>
            </ATabs>
          </div>

          <div class="sec-footer">
            <ASpace>
              <AButton
                v-privilege="'system:security_config:update'"
                type="primary"
                :loading="submitting"
                @click="onSave"
              >
                保存
              </AButton>
              <AButton :loading="loading" @click="loadData">刷新</AButton>
            </ASpace>
          </div>
        </AForm>
      </ACard>
    </ASpin>
  </div>
</template>

<style scoped>
.page {
  display: flex;
  flex-direction: column;
  height: 100%;
  min-height: 0;
  overflow: hidden;
  padding-bottom: 10px;
}

.page-spin {
  display: flex;
  flex: 1;
  flex-direction: column;
  min-height: 0;
}

.page :deep(.ant-spin-nested-loading),
.page :deep(.ant-spin-container) {
  display: flex;
  flex: 1;
  flex-direction: column;
  min-height: 0;
  height: 100%;
}

.form-card {
  display: flex;
  flex: 1;
  flex-direction: column;
  min-height: 0;
}

.form-card :deep(.ant-card-body) {
  display: flex;
  flex: 1;
  flex-direction: column;
  min-height: 0;
  padding: 0 16px;
}

.sec-form {
  display: flex;
  flex: 1;
  flex-direction: column;
  min-height: 0;
}

/* 内容区滚动；Tab 栏吸顶在滚动容器顶部 */
.sec-body {
  flex: 1;
  min-height: 0;
  overflow: auto;
  margin: 0 -16px;
  padding: 0 16px;
}

.top-tabs :deep(.ant-tabs-nav) {
  position: sticky;
  top: 0;
  z-index: 2;
  margin-bottom: 0;
  background: var(--ant-color-bg-container, #fff);
}

.panel {
  padding-top: 20px;
  padding-bottom: 8px;
}

.section-title {
  display: flex;
  align-items: center;
  margin-bottom: 20px;
  color: var(--ant-color-text);
  font-size: 16px;
  font-weight: 600;
  line-height: 1.4;
}

.section-title::before {
  content: '';
  display: inline-block;
  width: 4px;
  height: 16px;
  margin-right: 8px;
  border-radius: var(--app-border-radius, 2px);
  background: var(--ant-color-primary);
}

.sec-form :deep(.ant-form-item) {
  margin-bottom: 16px;
}

.field-control {
  width: 100%;
}

.hint {
  margin-top: 6px;
  color: var(--ant-color-text-secondary);
  font-size: 12px;
  line-height: 1.5;
}

/* 底部操作栏固定，对齐 PageTable 分页脚部 */
.sec-footer {
  display: flex;
  flex-shrink: 0;
  align-items: center;
  margin-top: 12px;
  padding: 12px 0;
  border-top: 1px solid var(--ant-color-split);
}
</style>
