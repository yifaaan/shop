<template>
  <div class="panel">
    <h2 class="panel-title">我的留言</h2>

    <!-- 已有留言 -->
    <ul v-loading="loading" class="msg-list">
      <li v-for="(item, index) in messageAll" :key="item.id" class="msg-item">
        <div class="msg-head">
          <el-tag size="small" type="info">{{ typeText(item.type) }}</el-tag>
          <span class="subject">{{ item.subject }}</span>
        </div>
        <div class="msg-body">内容：{{ item.message }}</div>
        <div class="msg-actions">
          <el-button link type="danger" @click="remove(index, item.id)">删除</el-button>
          <el-button v-if="item.file" link type="primary">
            <a :href="item.file" target="_blank">查看上传的文件</a>
          </el-button>
        </div>
      </li>
      <el-empty v-if="!loading && messageAll.length === 0" description="还没有留言" />
    </ul>

    <!-- 新增留言 -->
    <h3 class="sub-title">发表留言</h3>
    <el-form label-width="100px" class="msg-form">
      <el-form-item label="留言类型：">
        <el-radio-group v-model="form.type">
          <el-radio :value="1">留言</el-radio>
          <el-radio :value="2">投诉</el-radio>
          <el-radio :value="3">询问</el-radio>
          <el-radio :value="4">售后</el-radio>
          <el-radio :value="5">求购</el-radio>
        </el-radio-group>
      </el-form-item>
      <el-form-item label="主题：">
        <el-input v-model="form.subject" placeholder="留言主题" style="width: 320px" />
      </el-form-item>
      <el-form-item label="留言内容：">
        <el-input v-model="form.message" type="textarea" :rows="4" style="width: 480px" />
      </el-form-item>
      <el-form-item label="上传文件：">
        <el-input v-model="form.file" placeholder="文件链接（可选）" style="width: 480px" />
      </el-form-item>
      <el-form-item>
        <el-button type="primary" @click="submit">提交</el-button>
      </el-form-item>
    </el-form>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { getMessages, addMessage, delMessage } from '@/api'
import type { MessageItem } from '@/types'

const loading = ref(false)
const messageAll = ref<MessageItem[]>([])
const form = reactive({ type: 1, subject: '', message: '', file: '' })

function typeText(type: number) {
  const map: Record<number, string> = { 1: '留言', 2: '投诉', 3: '询问', 4: '售后', 5: '求购' }
  return map[type] ?? '留言'
}

async function load() {
  loading.value = true
  try {
    const res = await getMessages()
    messageAll.value = res.data ?? []
  } finally {
    loading.value = false
  }
}

async function submit() {
  if (!form.subject.trim() || !form.message.trim()) {
    ElMessage.warning('请填写主题和内容')
    return
  }
  try {
    await addMessage({ ...form })
    ElMessage.success('留言成功')
    form.subject = ''
    form.message = ''
    form.file = ''
    form.type = 1
    load()
  } catch {
    /* handled */
  }
}

async function remove(index: number, id: number) {
  try {
    await delMessage(id)
    messageAll.value.splice(index, 1)
    ElMessage.success('删除成功')
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
  margin-bottom: 16px;
  padding-bottom: 12px;
  border-bottom: 1px solid #f0f0f0;
}
.msg-list {
  margin-bottom: 24px;
}
.msg-item {
  padding: 12px 0;
  border-bottom: 1px dashed #eee;
  .msg-head {
    display: flex;
    align-items: center;
    gap: 10px;
    margin-bottom: 6px;
    .subject {
      font-weight: bold;
    }
  }
  .msg-body {
    color: #666;
    margin-bottom: 6px;
  }
  .msg-actions {
    display: flex;
    gap: 12px;
  }
}
.sub-title {
  font-size: 15px;
  margin-bottom: 16px;
  padding-top: 16px;
  border-top: 1px solid #f0f0f0;
}
</style>
