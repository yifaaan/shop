<template>
  <div class="panel">
    <h2 class="panel-title">订单详情</h2>
    <div v-loading="loading">
      <el-descriptions :column="2" border title="订单状态">
        <el-descriptions-item label="订单号">{{ orderInfo.order_sn }}</el-descriptions-item>
        <el-descriptions-item label="订单状态">
          <el-tag :type="statusType(orderInfo.status)" size="small">{{ statusText(orderInfo.status) }}</el-tag>
          <el-button
            v-if="![2, 3].includes(Number(orderInfo.status)) && orderInfo.alipay_url"
            type="primary"
            size="small"
            class="pay-btn"
            @click="goPay"
          >
            立即使用支付宝支付
          </el-button>
        </el-descriptions-item>
      </el-descriptions>

      <h3 class="sub-title">商品列表</h3>
      <el-table :data="orderInfo.order_goods ?? []" border stripe>
        <el-table-column label="商品名称" min-width="220">
          <template #default="{ row }">
            <router-link
              :to="{ name: 'productDetail', params: { productId: row.goods_id } }"
              class="link"
              >{{ row.goods_name }}</router-link
            >
          </template>
        </el-table-column>
        <el-table-column label="商品价格" width="140" align="center">
          <template #default="{ row }">￥{{ row.goods_price }}</template>
        </el-table-column>
        <el-table-column label="购买数量" prop="num" width="120" align="center" />
        <el-table-column label="小计" width="140" align="center">
          <template #default="{ row }">
            <span class="price">￥{{ row.goods_price * row.num }}</span>
          </template>
        </el-table-column>
      </el-table>
      <div class="total-row">
        商品总价：<span class="price big">￥{{ totalPrice }}</span>
      </div>

      <h3 class="sub-title">收货人信息</h3>
      <el-descriptions :column="2" border>
        <el-descriptions-item label="收货人姓名">{{ orderInfo.name }}</el-descriptions-item>
        <el-descriptions-item label="收货地址">{{ orderInfo.address }}</el-descriptions-item>
        <el-descriptions-item label="电话">{{ orderInfo.mobile }}</el-descriptions-item>
      </el-descriptions>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { useRoute } from 'vue-router'
import { getOrderDetail } from '@/api'
import type { OrderItem } from '@/types'

const route = useRoute()
const loading = ref(false)
const orderInfo = ref<Partial<OrderItem>>({})

const totalPrice = computed(() =>
  (orderInfo.value.order_goods ?? []).reduce((sum, g) => sum + g.goods_price * g.num, 0),
)

// 后端订单状态为数字：1-待支付 2-已支付 3-已取消
function statusText(status?: number | string) {
  switch (Number(status)) {
    case 2:
      return '已支付'
    case 3:
      return '已取消'
    default:
      return '待支付'
  }
}

function statusType(status?: number | string) {
  switch (Number(status)) {
    case 2:
      return 'success'
    case 3:
      return 'info'
    default:
      return 'warning'
  }
}

function goPay() {
  if (orderInfo.value.alipay_url) window.location.href = orderInfo.value.alipay_url
}

async function loadDetail() {
  loading.value = true
  try {
    orderInfo.value = await getOrderDetail(route.params.orderId as string)
  } finally {
    loading.value = false
  }
}

loadDetail()
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
.sub-title {
  font-size: 15px;
  margin: 24px 0 12px;
}
.pay-btn {
  margin-left: 12px;
}
.total-row {
  text-align: right;
  margin-top: 12px;
  font-size: 14px;
  .big {
    font-size: 20px;
  }
}
.link {
  &:hover {
    color: var(--brand-color);
  }
}
</style>
