<template>
  <div class="panel">
    <h2 class="panel-title">我的订单</h2>
    <el-table v-loading="loading" :data="orders" border stripe>
      <el-table-column label="订单号" min-width="180">
        <template #default="{ row }">
          <a class="link" @click="goDetail(row.id)">{{ row.order_sn }}</a>
        </template>
      </el-table-column>
      <el-table-column label="下单时间" prop="add_time" min-width="160" align="center" />
      <el-table-column label="订单总金额" width="120" align="center">
        <template #default="{ row }">
          <span class="price">￥{{ row.total }}</span>
        </template>
      </el-table-column>
      <el-table-column label="订单状态" width="120" align="center">
        <template #default="{ row }">
          <el-tag :type="statusType(row.status)" size="small">{{ statusText(row.status) }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column label="操作" width="120" align="center">
        <template #default="{ row }">
          <el-button link type="danger" @click="cancelOrder(row.id)">取消订单</el-button>
        </template>
      </el-table-column>
    </el-table>
    <el-empty v-if="!loading && orders.length === 0" description="暂无订单" />
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { getOrders, deleteOrder } from '@/api'
import type { OrderItem } from '@/types'

const router = useRouter()
const loading = ref(false)
const orders = ref<OrderItem[]>([])

function statusText(status?: string) {
  switch (status) {
    case 'TRADE_SUCCESS':
      return '已支付'
    case 'TRADE_CLOSED':
      return '已关闭'
    default:
      return '待支付'
  }
}

function statusType(status?: string) {
  switch (status) {
    case 'TRADE_SUCCESS':
      return 'success'
    case 'TRADE_CLOSED':
      return 'info'
    default:
      return 'warning'
  }
}

function goDetail(id: number) {
  router.push({ name: 'orderDetail', params: { orderId: id } })
}

async function cancelOrder(id: number) {
  try {
    await ElMessageBox.confirm('确认要取消该订单吗？', '提示', { type: 'warning' })
  } catch {
    return
  }
  try {
    await deleteOrder(id)
    ElMessage.success('订单已取消')
    loadOrders()
  } catch {
    /* handled */
  }
}

async function loadOrders() {
  loading.value = true
  try {
    const res = await getOrders()
    orders.value = res.data ?? []
  } finally {
    loading.value = false
  }
}

onMounted(loadOrders)
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
.link {
  color: var(--brand-color);
  cursor: pointer;
  &:hover {
    text-decoration: underline;
  }
}
</style>
