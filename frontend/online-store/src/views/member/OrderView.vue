<template>
  <div class="panel">
    <header class="panel-header">
      <h2 class="panel-title">我的订单</h2>
      <el-tooltip content="刷新订单" placement="top">
        <el-button
          circle
          :icon="Refresh"
          :loading="refreshing"
          aria-label="刷新订单"
          @click="loadOrders(true)"
        />
      </el-tooltip>
    </header>

    <div class="filter-bar">
      <el-radio-group v-model="activeStatus" size="small" @change="onStatusChange">
        <el-radio-button :value="0">全部</el-radio-button>
        <el-radio-button :value="1">待支付</el-radio-button>
        <el-radio-button :value="2">已支付</el-radio-button>
        <el-radio-button :value="3">已取消</el-radio-button>
      </el-radio-group>
    </div>

    <div v-loading="loading" class="order-content">
      <el-result v-if="loadFailed && orders.length === 0" icon="error" title="订单加载失败">
        <template #extra>
          <el-button type="primary" @click="loadOrders()">重新加载</el-button>
        </template>
      </el-result>

      <el-empty
        v-else-if="!loading && orders.length === 0"
        :description="activeStatus === 0 ? '暂无订单' : '暂无此状态的订单'"
      />

      <template v-else>
        <el-table :data="orders" border stripe class="desktop-order-table">
          <el-table-column label="订单号" min-width="180">
            <template #default="{ row }">
              <button class="order-link" type="button" @click="goDetail(row.id)">
                {{ row.order_sn }}
              </button>
            </template>
          </el-table-column>
          <el-table-column label="下单时间" min-width="168" align="center">
            <template #default="{ row }">{{ formatOrderTime(row.add_time) }}</template>
          </el-table-column>
          <el-table-column label="订单金额" width="124" align="right">
            <template #default="{ row }">
              <span class="price">{{ formatMoney(row.total) }}</span>
            </template>
          </el-table-column>
          <el-table-column label="订单状态" width="112" align="center">
            <template #default="{ row }">
              <el-tag :type="orderStatusType(row.status)" size="small">
                {{ orderStatusText(row.status) }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column label="剩余支付时间" width="132" align="center">
            <template #default="{ row }">
              <span v-if="isPendingOrder(row.status)" :class="['countdown', { expired: isExpired(row) }]">
                {{ countdownText(row) }}
              </span>
              <span v-else class="muted">--</span>
            </template>
          </el-table-column>
          <el-table-column label="操作" width="188" align="center" fixed="right">
            <template #default="{ row }">
              <el-button link type="primary" @click="goDetail(row.id)">
                {{ actionText(row) }}
              </el-button>
              <el-button
                v-if="canCancel(row)"
                link
                type="danger"
                :loading="cancelingOrderSn === row.order_sn"
                @click="cancelOrder(row)"
              >
                取消订单
              </el-button>
            </template>
          </el-table-column>
        </el-table>

        <div class="mobile-order-list">
          <article v-for="row in orders" :key="row.id" class="order-card">
            <div class="order-card-header">
              <button class="order-link" type="button" @click="goDetail(row.id)">
                {{ row.order_sn }}
              </button>
              <el-tag :type="orderStatusType(row.status)" size="small">
                {{ orderStatusText(row.status) }}
              </el-tag>
            </div>
            <dl class="order-meta">
              <div>
                <dt>下单时间</dt>
                <dd>{{ formatOrderTime(row.add_time) }}</dd>
              </div>
              <div>
                <dt>订单金额</dt>
                <dd class="price">{{ formatMoney(row.total) }}</dd>
              </div>
              <div v-if="isPendingOrder(row.status)">
                <dt>剩余时间</dt>
                <dd :class="['countdown', { expired: isExpired(row) }]">{{ countdownText(row) }}</dd>
              </div>
            </dl>
            <div class="order-actions">
              <el-button type="primary" size="small" @click="goDetail(row.id)">
                {{ actionText(row) }}
              </el-button>
              <el-button
                v-if="canCancel(row)"
                size="small"
                :loading="cancelingOrderSn === row.order_sn"
                @click="cancelOrder(row)"
              >
                取消订单
              </el-button>
            </div>
          </article>
        </div>
      </template>
    </div>
  </div>
</template>

<script setup lang="ts">
import { onBeforeUnmount, onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { Refresh } from '@element-plus/icons-vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { getOrders, updateOrderStatus } from '@/api'
import type { OrderItem } from '@/types'
import {
  formatCountdown,
  formatMoney,
  formatOrderTime,
  isPendingOrder,
  orderRemainingMs,
  orderStatusText,
  orderStatusType,
} from '@/utils/order'

const router = useRouter()
const loading = ref(false)
const refreshing = ref(false)
const loadFailed = ref(false)
const activeStatus = ref(0)
const orders = ref<OrderItem[]>([])
const now = ref(Date.now())
const cancelingOrderSn = ref<string | null>(null)

let requestInProgress = false
let clockTimer: number | undefined
const expiryRefreshes = new Map<string, { count: number; lastAt: number }>()

function orderKey(row: Partial<OrderItem>): string {
  return row.order_sn || String(row.id)
}

function remaining(row: Partial<OrderItem>): number | null {
  return orderRemainingMs(row.add_time, now.value)
}

function isExpired(row: Partial<OrderItem>): boolean {
  return remaining(row) === 0
}

function countdownText(row: Partial<OrderItem>): string {
  const value = remaining(row)
  return value === 0 ? '状态更新中' : formatCountdown(value)
}

function canCancel(row: Partial<OrderItem>): boolean {
  return isPendingOrder(row.status) && !isExpired(row)
}

function actionText(row: Partial<OrderItem>): string {
  return canCancel(row) ? '查看并支付' : '查看详情'
}

function goDetail(id: number) {
  router.push({ name: 'orderDetail', params: { orderId: id } })
}

async function cancelOrder(row: Partial<OrderItem>) {
  if (!row.order_sn) return
  try {
    await ElMessageBox.confirm('取消后该订单将无法继续支付，确认取消吗？', '取消订单', {
      type: 'warning',
      confirmButtonText: '确认取消',
      cancelButtonText: '暂不取消',
    })
  } catch {
    return
  }

  cancelingOrderSn.value = row.order_sn
  try {
    await updateOrderStatus({ order_sn: row.order_sn, status: 3 })
    ElMessage.success('订单已取消')
    await loadOrders(true)
  } finally {
    cancelingOrderSn.value = null
  }
}

async function loadOrders(silent = false) {
  if (requestInProgress) return
  requestInProgress = true
  refreshing.value = true
  if (!silent) loading.value = true
  loadFailed.value = false

  try {
    const res = await getOrders({
      status: activeStatus.value || undefined,
      p: 1,
      pnum: 50,
    })
    orders.value = res.data ?? []
  } catch {
    loadFailed.value = true
    if (orders.value.length > 0) ElMessage.error('订单刷新失败，请稍后重试')
  } finally {
    loading.value = false
    refreshing.value = false
    requestInProgress = false
  }
}

function onStatusChange() {
  orders.value = []
  expiryRefreshes.clear()
  loadOrders()
}

function tickClock() {
  now.value = Date.now()
  let shouldRefresh = false

  for (const row of orders.value) {
    if (!isPendingOrder(row.status) || !isExpired(row)) continue

    const key = orderKey(row)
    const state = expiryRefreshes.get(key) ?? { count: 0, lastAt: 0 }
    if (state.count < 3 && now.value - state.lastAt >= 2500) {
      expiryRefreshes.set(key, { count: state.count + 1, lastAt: now.value })
      shouldRefresh = true
    }
  }

  if (shouldRefresh) loadOrders(true)
}

onMounted(() => {
  loadOrders()
  clockTimer = window.setInterval(tickClock, 1000)
})

onBeforeUnmount(() => {
  if (clockTimer !== undefined) window.clearInterval(clockTimer)
})
</script>

<style scoped lang="scss">
.panel {
  min-height: 420px;
  background: #fff;
  border: 1px solid var(--border-color);
  border-radius: 4px;
  padding: 20px;
}

.panel-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  min-height: 41px;
  padding-bottom: 12px;
  border-bottom: 1px solid #f0f0f0;
}

.panel-title {
  font-size: 16px;
  font-weight: bold;
}

.filter-bar {
  display: flex;
  align-items: center;
  min-height: 56px;
}

.order-content {
  min-height: 300px;
}

.order-link {
  max-width: 100%;
  padding: 0;
  overflow: hidden;
  color: var(--brand-color);
  font: inherit;
  text-overflow: ellipsis;
  white-space: nowrap;
  background: none;
  border: 0;

  &:hover {
    text-decoration: underline;
  }
}

.countdown {
  color: #b26a00;
  font-variant-numeric: tabular-nums;

  &.expired {
    color: var(--text-light);
  }
}

.muted {
  color: var(--text-light);
}

.mobile-order-list {
  display: none;
}

@media (max-width: 767px) {
  .panel {
    min-height: 360px;
    padding: 14px;
  }

  .filter-bar {
    min-height: 54px;
    overflow-x: auto;
  }

  .desktop-order-table {
    display: none;
  }

  .mobile-order-list {
    display: grid;
    gap: 10px;
  }

  .order-card {
    border: 1px solid var(--border-color);
    border-radius: 4px;
  }

  .order-card-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 12px;
    min-height: 42px;
    padding: 8px 12px;
    background: #fafafa;
    border-bottom: 1px solid var(--border-color);

    .order-link {
      flex: 1;
      min-width: 0;
      text-align: left;
    }
  }

  .order-meta {
    display: grid;
    gap: 8px;
    padding: 12px;

    > div {
      display: flex;
      align-items: baseline;
      justify-content: space-between;
      gap: 16px;
    }

    dt {
      flex-shrink: 0;
      color: var(--text-light);
    }

    dd {
      min-width: 0;
      text-align: right;
      overflow-wrap: anywhere;
    }
  }

  .order-actions {
    display: flex;
    justify-content: flex-end;
    min-height: 48px;
    padding: 8px 12px;
    border-top: 1px solid var(--border-color);
  }
}
</style>
