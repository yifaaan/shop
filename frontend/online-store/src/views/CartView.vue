<template>
  <div class="container cart-page">
    <h1 class="page-title">我的购物车</h1>

    <div v-loading="loading" class="cart-table">
      <el-table :data="cartList" border row-key="id" @selection-change="onSelectionChange">
        <el-table-column type="selection" width="50" reserve-selection />
        <el-table-column label="商品" min-width="320">
          <template #default="{ row }">
            <div class="goods-cell">
              <router-link :to="{ name: 'productDetail', params: { productId: row.goods_id } }">
                <el-image :src="row.goods_image" fit="cover" class="thumb" />
              </router-link>
              <router-link
                :to="{ name: 'productDetail', params: { productId: row.goods_id } }"
                class="g-name"
                >{{ row.goods_name }}</router-link
              >
            </div>
          </template>
        </el-table-column>
        <el-table-column label="单价" width="120" align="center">
          <template #default="{ row }">
            <span class="price">￥{{ row.goods_price }}</span>
          </template>
        </el-table-column>
        <el-table-column label="数量" width="180" align="center">
          <template #default="{ row }">
            <el-input-number
              v-model="row.num"
              :min="1"
              :max="999"
              size="small"
              @change="(val) => updateQty(row as CartItem, val)"
            />
          </template>
        </el-table-column>
        <el-table-column label="小计" width="120" align="center">
          <template #default="{ row }">
            <span class="price">￥{{ (row.goods_price * row.num).toFixed(2) }}</span>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="100" align="center">
          <template #default="{ row, $index }">
            <el-button link type="danger" @click="removeGoods($index, row.id)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>

      <el-empty v-if="!loading && cartList.length === 0" description="购物车是空的">
        <el-button type="primary" @click="router.push({ name: 'home' })">去逛逛</el-button>
      </el-empty>
    </div>

    <!-- 结算栏 -->
    <div v-if="cartList.length" class="checkout-bar">
      <div class="left">
        <el-button @click="router.push({ name: 'home' })">继续购物</el-button>
      </div>
      <div class="addr-pay">
        <div class="addr-box">
          <span class="lbl">配送地址：</span>
          <el-select
            v-model="selectedAddrId"
            placeholder="选择收货地址"
            style="width: 360px"
            @change="onAddrChange"
          >
            <el-option
              v-for="a in addrList"
              :key="a.id"
              :label="`${a.signer_name} ${a.signer_mobile} ${a.province}${a.city}${a.district}${a.address}`"
              :value="a.id as number"
            />
          </el-select>
          <router-link :to="{ name: 'receive' }" class="add-addr">+ 添加地址</router-link>
        </div>
        <div class="remark">
          <span class="lbl">留言：</span>
          <el-input v-model="postScript" placeholder="选填：给卖家留言" />
        </div>
      </div>
      <div class="summary">
        <p>已选 <b class="price">{{ selectedCount }}</b> 件，总价：<b class="price big">￥{{ selectedTotal.toFixed(2) }}</b></p>
        <el-button type="primary" size="large" :disabled="selectedTotal <= 0" @click="balance">
          去结算
        </el-button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  getShopCarts,
  updateShopCart,
  deleteShopCart,
  getAddress,
  createOrder,
} from '@/api'
import { useCartStore } from '@/stores'
import type { CartItem, Address } from '@/types'

const router = useRouter()
const cartStore = useCartStore()

const loading = ref(false)
const cartList = ref<CartItem[]>([])
const selected = ref<CartItem[]>([])
const addrList = ref<Address[]>([])
const selectedAddrId = ref<number>()
const selectedAddr = ref<Address | null>(null)
const postScript = ref('')

const selectedCount = computed(() => selected.value.reduce((s, i) => s + i.num, 0))
const selectedTotal = computed(() =>
  selected.value.reduce((s, i) => s + i.goods_price * i.num, 0),
)

function onSelectionChange(rows: CartItem[]) {
  selected.value = rows
}

async function loadCart() {
  loading.value = true
  try {
    const res = await getShopCarts()
    cartList.value = res.data ?? []
    cartStore.goodsList = cartList.value
  } finally {
    loading.value = false
  }
}

async function loadAddr() {
  try {
    const res = await getAddress()
    addrList.value = res.data ?? []
  } catch {
    addrList.value = []
  }
}

function onAddrChange(id: number) {
  selectedAddr.value = addrList.value.find((a) => a.id === id) ?? null
}

async function updateQty(row: CartItem, val: number | undefined) {
  try {
    await updateShopCart(row.id, { num: val ?? 1, checked: row.checked })
    cartStore.refresh()
  } catch {
    loadCart()
  }
}

async function removeGoods(index: number, id: number) {
  try {
    await ElMessageBox.confirm('确定把该商品移除购物车吗？', '提示', { type: 'warning' })
  } catch {
    return
  }
  try {
    await deleteShopCart(id)
    cartList.value.splice(index, 1)
    cartStore.refresh()
    ElMessage.success('已删除')
  } catch {
    /* handled */
  }
}

async function balance() {
  if (selectedTotal.value <= 0) {
    ElMessage.warning('请选择要结算的商品')
    return
  }
  if (!selectedAddr.value) {
    ElMessage.warning('请选择收货地址')
    return
  }
  const a = selectedAddr.value
  try {
    const res = await createOrder({
      post: postScript.value,
      address: `${a.province}${a.city}${a.district}${a.address}`,
      name: a.signer_name,
      mobile: a.signer_mobile,
      pay_type: 2, // 支付宝
    })
    ElMessage.success('订单创建成功')
    if (res.alipay_url) {
      window.location.href = res.alipay_url
    } else {
      router.push({ name: 'order' })
    }
  } catch {
    /* handled */
  }
}

onMounted(() => {
  loadCart()
  loadAddr()
})
</script>

<style scoped lang="scss">
.cart-page {
  padding: 20px 0;
}
.page-title {
  font-size: 22px;
  margin-bottom: 16px;
}
.goods-cell {
  display: flex;
  align-items: center;
  gap: 12px;
  .thumb {
    width: 60px;
    height: 60px;
    border-radius: 4px;
    flex-shrink: 0;
    border: 1px solid #f0f0f0;
  }
  .g-name {
    flex: 1;
    min-width: 0;
    &:hover {
      color: var(--brand-color);
    }
  }
}
.checkout-bar {
  margin-top: 20px;
  background: #fff;
  border: 1px solid var(--border-color);
  border-radius: 4px;
  padding: 16px 20px;
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  gap: 20px;
  flex-wrap: wrap;
}
.addr-pay {
  flex: 1;
  min-width: 300px;
  .lbl {
    color: #999;
  }
  .addr-box,
  .remark {
    display: flex;
    align-items: center;
    gap: 8px;
    margin-bottom: 10px;
  }
  .add-addr {
    color: var(--brand-color);
    font-size: 13px;
  }
}
.summary {
  text-align: right;
  p {
    margin-bottom: 10px;
    font-size: 14px;
  }
  .big {
    font-size: 22px;
  }
}
</style>
