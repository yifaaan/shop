<template>
  <div class="panel">
    <header class="panel-header">
      <div class="title-row">
        <el-tooltip content="返回订单列表" placement="top">
          <el-button
            text
            circle
            :icon="ArrowLeft"
            aria-label="返回订单列表"
            @click="router.push({ name: 'order' })"
          />
        </el-tooltip>
        <h2 class="panel-title">订单详情</h2>
      </div>
      <el-tooltip content="刷新订单" placement="top">
        <el-button
          circle
          :icon="Refresh"
          :loading="refreshing"
          aria-label="刷新订单"
          @click="loadDetail(true)"
        />
      </el-tooltip>
    </header>

    <div v-loading="loading" class="detail-content">
      <el-result v-if="loadFailed && !orderInfo.id" icon="error" title="订单详情加载失败">
        <template #extra>
          <el-button type="primary" @click="loadDetail()">重新加载</el-button>
        </template>
      </el-result>

      <template v-else-if="orderInfo.id">
        <section :class="['status-strip', `status-${Number(orderInfo.status)}`]">
          <div class="status-main">
            <el-tag :type="orderStatusType(orderInfo.status)">
              {{ orderStatusText(orderInfo.status) }}
            </el-tag>
            <strong>{{ statusMessage }}</strong>
          </div>
          <div v-if="isPending && !isExpired" class="status-actions">
            <el-button
              v-if="orderInfo.alipay_url"
              type="primary"
              :icon="Wallet"
              @click="goPay"
            >
              立即支付
            </el-button>
            <el-button :loading="canceling" @click="cancelOrder">取消订单</el-button>
          </div>
        </section>

        <section class="detail-section">
          <h3 class="section-title">订单信息</h3>
          <dl class="info-grid">
            <div>
              <dt>订单号</dt>
              <dd>{{ orderInfo.order_sn }}</dd>
            </div>
            <div>
              <dt>下单时间</dt>
              <dd>{{ formatOrderTime(orderInfo.add_time) }}</dd>
            </div>
            <div>
              <dt>支付方式</dt>
              <dd>{{ paymentTypeText }}</dd>
            </div>
            <div>
              <dt>订单状态</dt>
              <dd>{{ orderStatusText(orderInfo.status) }}</dd>
            </div>
          </dl>
        </section>

        <section class="detail-section">
          <h3 class="section-title">商品列表</h3>
          <el-table :data="orderInfo.order_goods ?? []" border stripe class="desktop-goods-table">
            <el-table-column label="商品名称" min-width="240">
              <template #default="{ row }">
                <div class="goods-cell">
                  <el-image v-if="row.goods_image" :src="row.goods_image" fit="cover" class="goods-image" />
                  <router-link
                    :to="{ name: 'productDetail', params: { productId: row.goods_id } }"
                    class="goods-name"
                  >
                    {{ row.goods_name }}
                  </router-link>
                </div>
              </template>
            </el-table-column>
            <el-table-column label="单价" width="130" align="right">
              <template #default="{ row }">{{ formatMoney(row.goods_price) }}</template>
            </el-table-column>
            <el-table-column label="数量" prop="num" width="90" align="center" />
            <el-table-column label="小计" width="130" align="right">
              <template #default="{ row }">
                <span class="price">{{ formatMoney(row.goods_price * row.num) }}</span>
              </template>
            </el-table-column>
          </el-table>

          <div class="mobile-goods-list">
            <article v-for="goods in orderInfo.order_goods ?? []" :key="goods.id" class="mobile-goods-item">
              <el-image v-if="goods.goods_image" :src="goods.goods_image" fit="cover" class="goods-image" />
              <div class="mobile-goods-info">
                <router-link
                  :to="{ name: 'productDetail', params: { productId: goods.goods_id } }"
                  class="goods-name"
                >
                  {{ goods.goods_name }}
                </router-link>
                <div class="mobile-goods-meta">
                  <span>{{ formatMoney(goods.goods_price) }} × {{ goods.num }}</span>
                  <strong class="price">{{ formatMoney(goods.goods_price * goods.num) }}</strong>
                </div>
              </div>
            </article>
          </div>

          <dl class="amount-summary">
            <div>
              <dt>商品合计</dt>
              <dd>{{ formatMoney(goodsTotal) }}</dd>
            </div>
            <div>
              <dt>运费</dt>
              <dd>{{ formatMoney(orderInfo.post_fee) }}</dd>
            </div>
            <div class="amount-total">
              <dt>订单金额</dt>
              <dd class="price">{{ formatMoney(orderInfo.total) }}</dd>
            </div>
          </dl>
        </section>

        <section class="detail-section">
          <h3 class="section-title">收货信息</h3>
          <dl class="info-grid receiver-grid">
            <div>
              <dt>收货人</dt>
              <dd>{{ orderInfo.name || '--' }}</dd>
            </div>
            <div>
              <dt>联系电话</dt>
              <dd>{{ orderInfo.mobile || '--' }}</dd>
            </div>
            <div class="full-row">
              <dt>收货地址</dt>
              <dd>{{ orderInfo.address || '--' }}</dd>
            </div>
          </dl>
        </section>
      </template>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ArrowLeft, Refresh, Wallet } from '@element-plus/icons-vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { getOrderDetail, updateOrderStatus } from '@/api'
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

const route = useRoute()
const router = useRouter()
const loading = ref(false)
const refreshing = ref(false)
const loadFailed = ref(false)
const canceling = ref(false)
const orderInfo = ref<Partial<OrderItem>>({})
const now = ref(Date.now())

let requestInProgress = false
let clockTimer: number | undefined
let timeoutRefreshCount = 0
let lastTimeoutRefreshAt = 0

const goodsTotal = computed(() =>
  (orderInfo.value.order_goods ?? []).reduce((sum, goods) => sum + goods.goods_price * goods.num, 0),
)
const isPending = computed(() => isPendingOrder(orderInfo.value.status))
const remaining = computed(() => orderRemainingMs(orderInfo.value.add_time, now.value))
const isExpired = computed(() => remaining.value === 0)
const paymentTypeText = computed(() => {
  switch (Number(orderInfo.value.pay_type)) {
    case 1:
      return '微信支付'
    case 2:
      return '支付宝'
    default:
      return '--'
  }
})
const statusMessage = computed(() => {
  if (isPending.value) {
    if (remaining.value === null) return '等待支付'
    if (isExpired.value) return '支付时间已结束，正在确认取消结果'
    return `请在 ${formatCountdown(remaining.value)} 内完成支付`
  }
  if (Number(orderInfo.value.status) === 2) return '订单已支付'
  if (Number(orderInfo.value.status) === 3) return '订单已取消'
  return '订单状态暂不可用'
})

function goPay() {
  if (orderInfo.value.alipay_url && !isExpired.value) {
    window.location.assign(orderInfo.value.alipay_url)
  }
}

async function cancelOrder() {
  if (!orderInfo.value.order_sn) return
  try {
    await ElMessageBox.confirm('取消后该订单将无法继续支付，确认取消吗？', '取消订单', {
      type: 'warning',
      confirmButtonText: '确认取消',
      cancelButtonText: '暂不取消',
    })
  } catch {
    return
  }

  canceling.value = true
  try {
    await updateOrderStatus({ order_sn: orderInfo.value.order_sn, status: 3 })
    ElMessage.success('订单已取消')
    await loadDetail(true)
  } finally {
    canceling.value = false
  }
}

async function loadDetail(silent = false) {
  if (requestInProgress) return
  requestInProgress = true
  refreshing.value = true
  if (!silent) loading.value = true
  loadFailed.value = false

  try {
    orderInfo.value = await getOrderDetail(route.params.orderId as string)
  } catch {
    loadFailed.value = true
    if (orderInfo.value.id) ElMessage.error('订单刷新失败，请稍后重试')
  } finally {
    loading.value = false
    refreshing.value = false
    requestInProgress = false
  }
}

function tickClock() {
  now.value = Date.now()
  if (!isPending.value || !isExpired.value || timeoutRefreshCount >= 3) return
  if (now.value - lastTimeoutRefreshAt < 2500) return

  timeoutRefreshCount += 1
  lastTimeoutRefreshAt = now.value
  loadDetail(true)
}

onMounted(() => {
  loadDetail()
  clockTimer = window.setInterval(tickClock, 1000)
})

onBeforeUnmount(() => {
  if (clockTimer !== undefined) window.clearInterval(clockTimer)
})
</script>

<style scoped lang="scss">
.panel {
  min-height: 520px;
  background: #fff;
  border: 1px solid var(--border-color);
  border-radius: 4px;
  padding: 20px;
}

.panel-header,
.title-row,
.status-strip,
.status-main,
.status-actions,
.goods-cell,
.mobile-goods-item,
.mobile-goods-meta {
  display: flex;
  align-items: center;
}

.panel-header {
  justify-content: space-between;
  min-height: 41px;
  padding-bottom: 12px;
  border-bottom: 1px solid #f0f0f0;
}

.title-row {
  gap: 4px;
}

.panel-title {
  font-size: 16px;
  font-weight: bold;
}

.detail-content {
  min-height: 400px;
}

.status-strip {
  justify-content: space-between;
  gap: 20px;
  min-height: 68px;
  margin-top: 16px;
  padding: 12px 16px;
  background: #fff8e8;
  border-left: 3px solid #d99017;

  &.status-2 {
    background: #f0f9eb;
    border-left-color: #67c23a;
  }

  &.status-3 {
    background: #f5f7fa;
    border-left-color: #909399;
  }
}

.status-main {
  gap: 12px;
  min-width: 0;

  strong {
    font-size: 14px;
    font-variant-numeric: tabular-nums;
  }
}

.status-actions {
  flex-shrink: 0;
}

.detail-section {
  margin-top: 24px;
}

.section-title {
  margin-bottom: 12px;
  font-size: 15px;
}

.info-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  border-top: 1px solid var(--border-color);
  border-left: 1px solid var(--border-color);

  > div {
    display: grid;
    grid-template-columns: 96px minmax(0, 1fr);
    min-height: 44px;
    border-right: 1px solid var(--border-color);
    border-bottom: 1px solid var(--border-color);
  }

  dt,
  dd {
    display: flex;
    align-items: center;
    padding: 10px 12px;
    overflow-wrap: anywhere;
  }

  dt {
    color: #606266;
    background: #fafafa;
    border-right: 1px solid var(--border-color);
  }

  .full-row {
    grid-column: 1 / -1;
  }
}

.goods-cell {
  gap: 10px;
}

.goods-image {
  width: 52px;
  height: 52px;
  flex-shrink: 0;
  border: 1px solid var(--border-color);
  border-radius: 4px;
}

.goods-name {
  min-width: 0;
  overflow-wrap: anywhere;

  &:hover {
    color: var(--brand-color);
  }
}

.mobile-goods-list {
  display: none;
}

.amount-summary {
  width: 300px;
  margin: 12px 0 0 auto;

  > div {
    display: flex;
    justify-content: space-between;
    gap: 20px;
    padding: 5px 0;
  }

  dt {
    color: var(--text-light);
  }

  .amount-total {
    margin-top: 4px;
    padding-top: 10px;
    border-top: 1px solid var(--border-color);

    dd {
      font-size: 20px;
    }
  }
}

@media (max-width: 767px) {
  .panel {
    min-height: 420px;
    padding: 14px;
  }

  .status-strip {
    align-items: flex-start;
    flex-direction: column;
    gap: 12px;
  }

  .status-main {
    align-items: flex-start;

    strong {
      line-height: 24px;
    }
  }

  .status-actions {
    width: 100%;

    .el-button {
      flex: 1;
    }
  }

  .info-grid {
    grid-template-columns: 1fr;

    > div,
    .full-row {
      grid-column: auto;
      grid-template-columns: 88px minmax(0, 1fr);
    }
  }

  .desktop-goods-table {
    display: none;
  }

  .mobile-goods-list {
    display: grid;
    border-top: 1px solid var(--border-color);
  }

  .mobile-goods-item {
    gap: 10px;
    padding: 12px 0;
    border-bottom: 1px solid var(--border-color);
  }

  .mobile-goods-info {
    flex: 1;
    min-width: 0;
  }

  .mobile-goods-meta {
    justify-content: space-between;
    gap: 12px;
    margin-top: 8px;
    font-size: 13px;
    color: var(--text-light);
  }

  .amount-summary {
    width: 100%;
  }
}
</style>
