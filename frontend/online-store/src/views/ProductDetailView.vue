<template>
  <div class="container detail-page">
    <el-breadcrumb :separator-icon="ArrowRight" class="crumb">
      <el-breadcrumb-item :to="{ name: 'home' }">首页</el-breadcrumb-item>
      <el-breadcrumb-item v-if="proDetail.name">{{ proDetail.name }}</el-breadcrumb-item>
    </el-breadcrumb>

    <div v-loading="loading" class="detail-box">
      <!-- 上方：图片 + 信息 -->
      <div class="top">
        <div class="gallery">
          <div class="main-pic">
            <el-image :src="curShow" fit="contain" hide-on-click-modal :preview-src-list="images" />
          </div>
          <ul v-if="images.length" class="thumbs">
            <li
              v-for="(img, i) in images"
              :key="i"
              :class="{ active: curIndex === i }"
              @click="replaceShow(i)"
            >
              <el-image :src="img" fit="cover" />
            </li>
          </ul>
        </div>

        <div class="info">
          <h1 class="title">{{ proDetail.name }}</h1>
          <p v-if="proDetail.goods_brief" class="brief">{{ proDetail.goods_brief }}</p>

          <div class="price-panel">
            <div class="row">
              <span class="lbl">促销价</span>
              <strong class="sale-price">￥{{ proDetail.shop_price }}</strong>
            </div>
            <div v-if="proDetail.ship_free" class="row ship">
              <el-tag type="success" size="small">免运费</el-tag>
            </div>
            <div v-if="proDetail.sold_num != null" class="row">
              <span class="lbl">销量</span>
              <span>最近售出 <b class="price">{{ proDetail.sold_num }}</b> 件</span>
            </div>
          </div>

          <div class="buy-box">
            <div class="qty">
              <span class="lbl">数量</span>
              <el-input-number v-model="buyNum" :min="1" :max="999" />
            </div>
            <div class="actions">
              <el-button type="primary" size="large" :icon="ShoppingCart" @click="addToCart">
                加入购物车
              </el-button>
              <el-button size="large" :type="hasFav ? 'info' : 'default'" @click="toggleFav">
                {{ hasFav ? '已收藏' : '收藏' }}
              </el-button>
            </div>
          </div>
        </div>

        <!-- 热卖商品 -->
        <aside class="hot-sales">
          <h3 class="hot-title">热卖商品</h3>
          <div v-loading="loadingHot" class="hot-list">
            <router-link
              v-for="item in hotGoods"
              :key="item.id"
              :to="{ name: 'productDetail', params: { productId: item.id } }"
              class="hot-item"
            >
              <el-image :src="item.front_image" fit="cover" />
              <p class="hot-brief">{{ item.goods_brief }}</p>
              <p class="price">￥{{ item.shop_price }}</p>
            </router-link>
          </div>
        </aside>
      </div>

      <!-- 下方：详情图 -->
      <div v-if="descImages.length" class="desc-section">
        <h2 class="section-title">商品详情</h2>
        <div class="desc-images">
          <img v-for="(img, i) in descImages" :key="i" :src="img" alt="详情图" />
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { ArrowRight, ShoppingCart } from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'
import { getGoodsDetail, getGoodsList, addShopCart, getFav, addFav, delFav } from '@/api'
import { useUserStore, useCartStore } from '@/stores'
import type { Goods } from '@/types'

const route = useRoute()
const userStore = useUserStore()
const cartStore = useCartStore()

const loading = ref(false)
const loadingHot = ref(false)
const proDetail = ref<Partial<Goods>>({})
const hotGoods = ref<Goods[]>([])
const buyNum = ref(1)
const curIndex = ref(0)
const hasFav = ref(false)

const images = computed(() => proDetail.value.images ?? [])
const descImages = computed(() => proDetail.value.desc_images ?? [])
const curShow = computed(() => images.value[curIndex.value] ?? proDetail.value.front_image ?? '')

function replaceShow(i: number) {
  curIndex.value = i
}

async function loadDetail() {
  loading.value = true
  const id = route.params.productId as string
  try {
    const res = await getGoodsDetail(id)
    proDetail.value = res
    curIndex.value = 0
  } finally {
    loading.value = false
  }
}

async function loadHot() {
  loadingHot.value = true
  try {
    const res = await getGoodsList({ ih: '1', pnum: 4 })
    hotGoods.value = res.data ?? []
  } finally {
    loadingHot.value = false
  }
}

async function checkFav() {
  if (!userStore.isLoggedIn()) {
    hasFav.value = false
    return
  }
  try {
    await getFav(route.params.productId as string)
    hasFav.value = true
  } catch {
    hasFav.value = false
  }
}

async function addToCart() {
  try {
    await addShopCart({ goods: Number(route.params.productId), nums: buyNum.value })
    ElMessage.success('已加入购物车')
    cartStore.refresh()
  } catch {
    /* handled */
  }
}

async function toggleFav() {
  const id = route.params.productId as string
  if (hasFav.value) {
    try {
      await delFav(id)
      hasFav.value = false
      ElMessage.success('已取消收藏')
    } catch {
      /* handled */
    }
  } else {
    try {
      await addFav({ goods: Number(id) })
      hasFav.value = true
      ElMessage.success('已加入收藏')
    } catch {
      /* handled */
    }
  }
}

function init() {
  loadDetail()
  loadHot()
  checkFav()
}

watch(() => route.params.productId, init)
onMounted(init)
</script>

<style scoped lang="scss">
.detail-page {
  padding: 20px 0;
}
.crumb {
  margin-bottom: 16px;
}
.top {
  display: flex;
  gap: 20px;
  background: #fff;
  padding: 24px;
  border: 1px solid var(--border-color);
  border-radius: 4px;
}
.gallery {
  width: 400px;
  flex-shrink: 0;
  .main-pic {
    width: 400px;
    height: 400px;
    border: 1px solid #f0f0f0;
    border-radius: 4px;
    overflow: hidden;
    :deep(.el-image) {
      width: 100%;
      height: 100%;
    }
  }
  .thumbs {
    display: flex;
    gap: 8px;
    margin-top: 12px;
    li {
      width: 60px;
      height: 60px;
      border: 2px solid transparent;
      border-radius: 4px;
      overflow: hidden;
      cursor: pointer;
      &.active {
        border-color: var(--brand-color);
      }
      :deep(.el-image) {
        width: 100%;
        height: 100%;
      }
    }
  }
}
.info {
  flex: 1;
  min-width: 0;
}
.title {
  font-size: 22px;
  font-weight: bold;
  margin-bottom: 8px;
}
.brief {
  color: #999;
  margin-bottom: 20px;
  font-size: 13px;
}
.price-panel {
  background: #fff9f9;
  border: 1px solid #ffe6e6;
  border-radius: 4px;
  padding: 16px 20px;
  margin-bottom: 24px;
  .row {
    display: flex;
    align-items: center;
    gap: 12px;
    margin-bottom: 10px;
    &:last-child {
      margin-bottom: 0;
    }
    .lbl {
      color: #999;
      width: 50px;
    }
  }
  .sale-price {
    font-size: 32px;
    color: var(--brand-color);
  }
}
.buy-box {
  .qty {
    display: flex;
    align-items: center;
    gap: 12px;
    margin-bottom: 20px;
    .lbl {
      color: #999;
    }
  }
  .actions {
    display: flex;
    gap: 16px;
  }
}
.hot-sales {
  width: 200px;
  flex-shrink: 0;
  border-left: 1px solid #f0f0f0;
  padding-left: 16px;
}
.hot-title {
  font-size: 16px;
  text-align: center;
  padding-bottom: 10px;
  border-bottom: 1px solid #f0f0f0;
  margin-bottom: 10px;
}
.hot-item {
  display: block;
  margin-bottom: 14px;
  text-align: center;
  :deep(.el-image) {
    width: 100%;
    height: 160px;
    border-radius: 4px;
  }
  .hot-brief {
    font-size: 12px;
    color: #666;
    margin: 6px 0 2px;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
}
.desc-section {
  margin-top: 24px;
  background: #fff;
  border: 1px solid var(--border-color);
  border-radius: 4px;
  padding: 24px;
}
.section-title {
  font-size: 18px;
  border-bottom: 2px solid var(--brand-color);
  padding-bottom: 10px;
  margin-bottom: 20px;
}
.desc-images {
  img {
    display: block;
    max-width: 100%;
    margin: 0 auto 10px;
  }
}
</style>
