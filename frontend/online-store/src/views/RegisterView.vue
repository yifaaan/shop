<template>
  <div class="register-page">
    <div class="register-box">
      <h2>手机注册</h2>
      <el-form ref="formRef" :model="form" :rules="rules" label-position="top" @submit.prevent>
        <el-form-item label="手机号" prop="mobile">
          <el-input v-model="form.mobile" placeholder="请输入手机号码" :prefix-icon="Iphone" />
        </el-form-item>
        <el-form-item label="短信验证码" prop="code">
          <div class="code-row">
            <el-input v-model="form.code" placeholder="输入手机验证码" />
            <el-button :disabled="countdown > 0" :loading="sending" @click="sendMessage">
              {{ countdown > 0 ? `重新发送(${countdown})` : '免费获取验证码' }}
            </el-button>
          </div>
        </el-form-item>
        <el-form-item label="密码" prop="password">
          <el-input
            v-model="form.password"
            type="password"
            placeholder="请输入 6-15 位密码"
            show-password
            :prefix-icon="Lock"
          />
        </el-form-item>
        <el-button
          type="primary"
          class="submit-btn"
          :loading="loading"
          size="large"
          @click="handleRegister"
          >注册并登录</el-button
        >
      </el-form>
      <p class="form-footer">
        已有账号？<router-link :to="{ name: 'login' }">[立即登录]</router-link>
      </p>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onUnmounted } from 'vue'
import { useRouter } from 'vue-router'
import { Iphone, Lock } from '@element-plus/icons-vue'
import { ElMessage, type FormInstance, type FormRules } from 'element-plus'
import { register, sendSms } from '@/api'
import { parseError } from '@/api'
import { useUserStore } from '@/stores'
import type { RegisterParams } from '@/types'

const router = useRouter()
const userStore = useUserStore()

const formRef = ref<FormInstance>()
const loading = ref(false)
const sending = ref(false)
const countdown = ref(0)
let timer: ReturnType<typeof setInterval> | null = null

const form = reactive<RegisterParams>({
  mobile: '',
  password: '',
  code: '',
})

const rules: FormRules<RegisterParams> = {
  mobile: [
    {
      required: true,
      validator: (_r, value, cb) => {
        if (!value) return cb(new Error('请输入手机号'))
        if (!/^1[3-9]\d{9}$/.test(value)) return cb(new Error('手机号格式不正确'))
        cb()
      },
      trigger: 'blur',
    },
  ],
  code: [{ required: true, message: '请输入短信验证码', trigger: 'blur' }],
  password: [{ required: true, min: 6, max: 15, message: '请输入 6-15 位密码', trigger: 'blur' }],
}

function startCountdown() {
  countdown.value = 60
  timer = setInterval(() => {
    countdown.value--
    if (countdown.value <= 0 && timer) {
      clearInterval(timer)
      timer = null
    }
  }, 1000)
}

async function sendMessage() {
  if (!/^1[3-9]\d{9}$/.test(form.mobile)) {
    ElMessage.warning('请先输入正确的手机号')
    return
  }
  sending.value = true
  try {
    await sendSms({ mobile: form.mobile, type: 1 })
    ElMessage.success('短信发送成功')
    startCountdown()
  } catch {
    ElMessage.error('短信发送失败')
  } finally {
    sending.value = false
  }
}

async function handleRegister() {
  if (!formRef.value) return
  await formRef.value.validate(async (valid) => {
    if (!valid) return
    loading.value = true
    try {
      const res = await register(form)
      userStore.login(res.nickname, res.token)
      ElMessage.success('注册成功')
      router.push({ name: 'home' })
    } catch (e: any) {
      const err = parseError(e)
      const msg = err.msg
      if (msg && typeof msg === 'object') {
        ElMessage.error(Object.values(msg).join('；'))
      } else {
        ElMessage.error((msg as string) || '注册失败')
      }
    } finally {
      loading.value = false
    }
  })
}

onUnmounted(() => {
  if (timer) clearInterval(timer)
})
</script>

<style scoped lang="scss">
.register-page {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  background: linear-gradient(135deg, #c81623 0%, #f3646a 100%);
}
.register-box {
  width: 380px;
  background: #fff;
  border-radius: 8px;
  padding: 30px 36px;
  box-shadow: 0 8px 30px rgba(0, 0, 0, 0.15);

  h2 {
    text-align: center;
    font-size: 20px;
    margin-bottom: 24px;
  }
  .code-row {
    display: flex;
    gap: 10px;
    width: 100%;
    .el-button {
      flex-shrink: 0;
      width: 140px;
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
