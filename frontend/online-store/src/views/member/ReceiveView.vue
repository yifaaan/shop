<template>
  <div class="panel">
    <h2 class="panel-title">收货地址</h2>
    <div v-loading="loading">
      <!-- 已有地址 -->
      <div v-for="(item, index) in list" :key="item.id" class="addr-item">
        <el-form label-width="100px" class="addr-form">
          <div class="form-row">
            <el-form-item label="所在地区：">
              <div class="region">
                <el-input v-model="item.province" placeholder="省" />
                <el-input v-model="item.city" placeholder="市" />
                <el-input v-model="item.district" placeholder="区/县" />
              </div>
            </el-form-item>
          </div>
          <div class="form-row">
            <el-form-item label="收货人：">
              <el-input v-model="item.signer_name" placeholder="收货人姓名" class="mid-input" />
            </el-form-item>
          </div>
          <div class="form-row">
            <el-form-item label="详细地址：">
              <el-input v-model="item.address" placeholder="详细地址" class="long-input" />
            </el-form-item>
          </div>
          <div class="form-row">
            <el-form-item label="手机：">
              <el-input v-model="item.signer_mobile" placeholder="手机号" class="mid-input" />
            </el-form-item>
          </div>
          <div class="form-actions">
            <el-button type="primary" @click="update(item, index)">确定修改</el-button>
            <el-button @click="remove(item.id)">删除</el-button>
          </div>
        </el-form>
      </div>

      <el-empty v-if="!loading && list.length === 0" description="还没有收货地址" />

      <!-- 新增地址 -->
      <div class="addr-item new-addr">
        <h3 class="sub-title">新增收货地址</h3>
        <el-form label-width="100px">
          <div class="form-row">
            <el-form-item label="所在地区：">
              <div class="region">
                <el-input v-model="form.province" placeholder="省" />
                <el-input v-model="form.city" placeholder="市" />
                <el-input v-model="form.district" placeholder="区/县" />
              </div>
            </el-form-item>
          </div>
          <div class="form-row">
            <el-form-item label="收货人：">
              <el-input v-model="form.signer_name" placeholder="收货人姓名" class="mid-input" />
            </el-form-item>
          </div>
          <div class="form-row">
            <el-form-item label="详细地址：">
              <el-input v-model="form.address" placeholder="详细地址" class="long-input" />
            </el-form-item>
          </div>
          <div class="form-row">
            <el-form-item label="手机：">
              <el-input v-model="form.signer_mobile" placeholder="手机号" class="mid-input" />
            </el-form-item>
          </div>
          <div class="form-actions">
            <el-button type="primary" @click="add">新增收货地址</el-button>
          </div>
        </el-form>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { getAddress, addAddress, updateAddress, delAddress } from '@/api'
import type { Address } from '@/types'

const loading = ref(false)
const list = ref<Address[]>([])
const form = reactive<Address>({
  province: '',
  city: '',
  district: '',
  address: '',
  signer_name: '',
  signer_mobile: '',
})

function resetForm() {
  form.province = ''
  form.city = ''
  form.district = ''
  form.address = ''
  form.signer_name = ''
  form.signer_mobile = ''
}

async function load() {
  loading.value = true
  try {
    const res = await getAddress()
    list.value = res.data ?? []
  } finally {
    loading.value = false
  }
}

async function add() {
  try {
    await addAddress({ ...form })
    ElMessage.success('添加成功')
    resetForm()
    load()
  } catch {
    /* handled */
  }
}

async function update(item: Address, _index: number) {
  if (!item.id) return
  try {
    await updateAddress(item.id, item)
    ElMessage.success('修改成功')
    load()
  } catch {
    /* handled */
  }
}

async function remove(id?: number) {
  if (!id) return
  try {
    await ElMessageBox.confirm('确认删除该收货地址吗？', '提示', { type: 'warning' })
  } catch {
    return
  }
  try {
    await delAddress(id)
    ElMessage.success('删除成功')
    load()
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
.addr-item {
  padding: 16px;
  border: 1px solid #f0f0f0;
  border-radius: 4px;
  margin-bottom: 16px;
}
.new-addr {
  border-style: dashed;
}
.sub-title {
  font-size: 15px;
  margin-bottom: 16px;
  color: #666;
}
.region {
  display: flex;
  gap: 8px;
  .el-input {
    width: 130px;
  }
}
.mid-input {
  width: 240px;
}
.long-input {
  width: 420px;
}
.form-row {
  margin-bottom: 12px;
}
.form-actions {
  padding-left: 100px;
  display: flex;
  gap: 12px;
}
</style>
