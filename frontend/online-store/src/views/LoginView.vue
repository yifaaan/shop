<template>
  <div class="login-page">
    <div class="login-box">
      <h2>帐号登录</h2>
      <el-form ref="formRef" :model="form" :rules="rules" label-position="top" @submit.prevent>
        <el-form-item label="用户名（手机号）" prop="mobile">
          <el-input v-model="form.mobile" placeholder="请输入手机号" :prefix-icon="User" />
        </el-form-item>
        <el-form-item label="密码" prop="password">
          <el-input
            v-model="form.password"
            type="password"
            placeholder="请输入密码"
            show-password
            :prefix-icon="Lock"
          />
        </el-form-item>
        <el-form-item label="验证码" prop="captcha">
          <div class="captcha-row">
            <el-input v-model="form.captcha" placeholder="请输入验证码" />
            <img
              v-if="captchaPic"
              :src="captchaPic"
              alt="验证码"
              class="captcha-img"
              title="点击刷新"
              @click="loadCaptcha"
            />
            <div v-else class="captcha-placeholder" @click="loadCaptcha">点击获取</div>
          </div>
        </el-form-item>
        <el-button
          type="primary"
          class="submit-btn"
          :loading="loading"
          size="large"
          @click="handleLogin"
          >立即登录</el-button
        >
      </el-form>
      <p class="form-footer">
        没有帐号？<router-link :to="{ name: 'register' }">[立即注册]</router-link>
      </p>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { User, Lock } from '@element-plus/icons-vue'
import { ElMessage, type FormInstance, type FormRules } from 'element-plus'
import { login, getCaptcha } from '@/api'
import { useUserStore } from '@/stores'
import { parseError } from '@/api'
import type { LoginParams } from '@/types'

const route = useRoute()
const router = useRouter()
const userStore = useUserStore()

const formRef = ref<FormInstance>()
const loading = ref(false)
const captchaPic = ref('')

const form = reactive<LoginParams>({
  mobile: '',
  password: '',
  captcha: '',
  captcha_id: '',
})

const rules: FormRules<LoginParams> = {
  mobile: [{ required: true, message: '请输入手机号', trigger: 'blur' }],
  password: [{ required: true, min: 6, max: 15, message: '请输入 6-15 位密码', trigger: 'blur' }],
  captcha: [{ required: true, message: '请输入验证码', trigger: 'blur' }],
}

async function loadCaptcha() {
  try {
    const res = await getCaptcha()
    captchaPic.value = res.picPath
    form.captcha_id = res.captcha_id
  } catch {
    /* handled */
  }
}

async function handleLogin() {
  if (!formRef.value) return
  await formRef.value.validate(async (valid) => {
    if (!valid) return
    loading.value = true
    try {
      const res = await login(form)
      userStore.login(res.nickname, res.token)
      ElMessage.success('登录成功')
      const redirect = (route.query.redirect as string) || '/'
      router.push(redirect)
    } catch (e: any) {
      const err = parseError(e)
      const msg = err.msg
      if (msg && typeof msg === 'object') {
        ElMessage.error(Object.values(msg).join('；'))
      } else {
        ElMessage.error((msg as string) || '登录失败')
      }
      loadCaptcha()
    } finally {
      loading.value = false
    }
  })
}

onMounted(() => {
  // 进入登录页清除旧登录态
  userStore.logout()
  loadCaptcha()
})
</script>

<style scoped lang="scss">
.login-page {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  background: linear-gradient(135deg, #c81623 0%, #f3646a 100%);
}
.login-box {
  width: 380px;
  background: #fff;
  border-radius: 8px;
  padding: 30px 36px;
  box-shadow: 0 8px 30px rgba(0, 0, 0, 0.15);

  h2 {
    text-align: center;
    font-size: 20px;
    margin-bottom: 24px;
    color: #333;
  }
  .captcha-row {
    display: flex;
    gap: 10px;
    width: 100%;
    .captcha-img,
    .captcha-placeholder {
      width: 110px;
      height: 32px;
      flex-shrink: 0;
      cursor: pointer;
      border: 1px solid #dcdfe6;
      border-radius: 4px;
    }
    .captcha-img {
      object-fit: cover;
      border: none;
    }
    .captcha-placeholder {
      display: flex;
      align-items: center;
      justify-content: center;
      font-size: 12px;
      color: #999;
      background: #f5f5f5;
    }
  }
  .submit-btn {
    width: 100%;
    margin-top: 8px;
  }
  .form-footer {
    text-align: center;
    margin-top: 16px;
    font-size: 13px;
    a {
      color: var(--brand-color);
    }
  }
}
</style>
