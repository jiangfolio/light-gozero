<script setup lang="ts">
import { message } from 'antdv-next'
import { onMounted, reactive, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'

import { fetchCaptcha } from '@/api/system/auth'
import {
  getRememberUsername,
  setRememberUsername,
} from '@/lib/auth-token'
import { useUserStore } from '@/stores/user'

const router = useRouter()
const route = useRoute()
const userStore = useUserStore()

const loading = ref(false)
const remember = ref(Boolean(getRememberUsername()))
const captchaId = ref('')
const captchaImage = ref('')

const form = reactive({
  username: getRememberUsername() || '',
  password: '',
  captcha_code: '',
})

/** 刷新图形验证码。 */
async function loadCaptcha() {
  try {
    const data = await fetchCaptcha()
    captchaId.value = data.captcha_id
    const img = data.captcha_image || ''
    captchaImage.value = img.startsWith('data:') ? img : `data:image/png;base64,${img}`
  } catch (e: unknown) {
    const err = e as { message?: string }
    message.error(err.message || '验证码加载失败')
  }
}

/** 提交登录并跳转。 */
async function onSubmit() {
  if (loading.value) return
  if (!form.username || !form.password || !form.captcha_code) {
    message.warning('请填写完整登录信息')
    return
  }
  loading.value = true
  try {
    // 1. SM2 登录拿 token
    await userStore.login({
      username: form.username,
      password: form.password,
      captcha_id: captchaId.value,
      captcha_code: form.captcha_code,
    })
    // 2. 记住用户名
    setRememberUsername(remember.value ? form.username : '')
    message.success('登录成功')
    // 3. 跳转目标页（守卫会拉 resources）
    const redirect = (route.query.redirect as string) || '/'
    await router.replace(redirect)
  } catch (e: unknown) {
    const err = e as { message?: string }
    message.error(err.message || '登录失败')
    form.captcha_code = ''
    await loadCaptcha()
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  loadCaptcha()
})
</script>

<template>
  <div class="login-bg">
    <div class="login-card">
      <div class="login-welcome">
        <div class="welcome-bg">
          <p class="welcome-title">欢迎登录 light-gozero</p>
          <p class="welcome-sub">轻量 · 简洁 · 安全的管理后台</p>
        </div>
      </div>
      <div class="login-form-pane">
        <h2 class="login-form-title">账号登录</h2>
        <AForm layout="vertical" @finish="onSubmit" @submit.prevent="onSubmit">
          <AFormItem label="用户名">
            <AInput
              v-model:value="form.username"
              size="large"
              placeholder="请输入用户名"
              allow-clear
            />
          </AFormItem>
          <AFormItem label="密码">
            <AInputPassword
              v-model:value="form.password"
              size="large"
              placeholder="请输入密码"
              autocomplete="current-password"
            />
          </AFormItem>
          <AFormItem label="验证码">
            <div class="captcha-row">
              <AInput
                v-model:value="form.captcha_code"
                size="large"
                placeholder="请输入验证码"
                @press-enter="onSubmit"
              />
              <img
                class="captcha-img"
                :src="captchaImage"
                alt="验证码"
                title="点击刷新"
                @click="loadCaptcha"
              />
            </div>
          </AFormItem>
          <AFormItem>
            <ACheckbox v-model:checked="remember">记住用户名</ACheckbox>
          </AFormItem>
          <!-- antdv-next 下仅 html-type=submit 不一定触发 finish，点击需显式提交 -->
          <AButton
            type="primary"
            html-type="button"
            size="large"
            block
            :loading="loading"
            @click="onSubmit"
          >
            登录
          </AButton>
        </AForm>
      </div>
    </div>
  </div>
</template>

<style scoped lang="less">
.login-bg {
  box-sizing: border-box;
  display: flex;
  min-height: 100vh;
  min-height: 100dvh;
  align-items: center;
  justify-content: center;
  padding: 24px;
  background-image: url('@/assets/images/login/login-bg.png');
  background-size: cover;
  background-position: center;
}

.login-card {
  display: flex;
  width: 100%;
  max-width: 888px;
  min-height: 600px;
  overflow: hidden;
  border-radius: 12px;
  background: #fff;
  box-shadow: 0 16px 73px 8px rgba(203, 203, 203, 0.2);

  @media (max-width: 860px) {
    min-height: 0;
    flex-direction: column;
  }
}

.login-welcome {
  width: 444px;
  padding: 23px 25px;
  background: #fff;

  @media (max-width: 860px) {
    display: none;
  }
}

.welcome-bg {
  height: 100%;
  padding: 35px 24px 0;
  border-radius: 8px;
  background-image: url('@/assets/images/login/left-bg1.png');
  background-size: cover;
  background-position: center;
  text-align: center;
  box-shadow: 0 6px 20px 0 rgba(33, 47, 117, 0.1);
}

.welcome-title {
  margin: 0;
  font-size: 26px;
  font-weight: 700;
  line-height: 25px;
  letter-spacing: 0.26px;
  color: #333;
}

.welcome-sub {
  margin-top: 12px;
  font-size: 20px;
  font-weight: 700;
  line-height: 25px;
  letter-spacing: 0.26px;
  color: #333;
}

.login-form-pane {
  display: flex;
  width: 444px;
  flex-direction: column;
  justify-content: center;
  padding: 34px 42px;

  @media (max-width: 860px) {
    width: 100%;
  }
}

.login-form-title {
  margin-bottom: 28px;
  font-size: 20px;
  font-weight: 600;
  color: #1e1e1e;
  text-align: center;
}

.captcha-row {
  display: flex;
  width: 100%;
  align-items: center;
  gap: 12px;
}

.captcha-img {
  width: 120px;
  height: 40px;
  flex-shrink: 0;
  cursor: pointer;
  border: 1px solid #e5e7eb;
  border-radius: 6px;
  object-fit: contain;
  background: #f8fafc;
}
</style>
