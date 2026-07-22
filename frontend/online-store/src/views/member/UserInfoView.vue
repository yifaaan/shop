<template>
  <div class="panel">
    <h2 class="panel-title">个人资料</h2>
    <div v-loading="loading">
      <el-form label-width="120px" class="info-form">
        <el-form-item label="姓名：">
          <el-input v-model="userInfo.name" placeholder="请输入姓名" style="width: 240px" />
        </el-form-item>
        <el-form-item label="出生日期：">
          <el-date-picker
            v-model="userInfo.birthday"
            type="date"
            value-format="YYYY-MM-DD"
            placeholder="选择出生日期"
            style="width: 240px"
          />
        </el-form-item>
        <el-form-item label="性别：">
          <el-radio-group v-model="userInfo.gender">
            <el-radio value="male">男</el-radio>
            <el-radio value="female">女</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item label="手机：">
          <el-input v-model="userInfo.mobile" disabled style="width: 240px" />
          <span class="star">*</span>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="save">确认修改</el-button>
        </el-form-item>
      </el-form>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { getUserDetail, updateUserInfo } from '@/api'
import type { UserDetail } from '@/types'

const loading = ref(false)
const userInfo = reactive<UserDetail>({
  name: '',
  birthday: '',
  gender: 'male',
  mobile: '',
})

async function load() {
  loading.value = true
  try {
    const res = await getUserDetail()
    Object.assign(userInfo, res)
  } finally {
    loading.value = false
  }
}

async function save() {
  try {
    await updateUserInfo({ ...userInfo })
    ElMessage.success('修改成功')
  } catch {
    /* handled */
  }
}

onMounted(load)
</script>

<style scoped lang="scss">
.panel {
  background: #fff;
  border: 1px solid var(--border-color);
  border-radius: 4px;
  padding: 20px;
}
.panel-title {
  font-size: 16px;
  font-weight: bold;
  margin-bottom: 20px;
  padding-bottom: 12px;
  border-bottom: 1px solid #f0f0f0;
}
.info-form {
  max-width: 480px;
}
.star {
  color: var(--brand-color);
  margin-left: 8px;
}
</style>
