<template>
  <div class="container home">
    <!-- 轮播图 -->
    <el-carousel v-if="banners.length" height="400px" class="banner" :interval="3000" arrow="hover">
      <el-carousel-item v-for="item in banners" :key="item.id">
        <router-link
          :to="{ name: 'productDetail', params: { productId: item.url } }"
          class="banner-link"
        >
          <el-image :src="item.image" fit="cover" />
        </router-link>
      </el-carousel-item>
    </el-carousel>

    <!-- 新品 -->
    <section class="section">
      <div class="section-head">
        <h2>刚出炉新品</h2>
        <router-link :to="{ name: 'list' }" class="more">更多 &gt;&gt;</router-link>
      </div>
      <div v-loading="loadingNew" class="goods-grid">
        <ProductCard v-for="item in newGoods" :key="item.id" :goods="item" />
      </div>
    </section>

    <!-- 按分类分组 -->
    <section v-for="(group, idx) in categoryGroups" :key="idx" class="section">
      <div class="section-head">
        <h2>{{ group.name }}</h2>
      </div>
      <div v-loading="loadingGroup" class="goods-grid">
        <ProductCard v-for="item in group.goods" :key="item.id" :goods="item" />
      </div>
    </section>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import ProductCard from '@/components/ProductCard.vue'
import { getBanners, getCategoryList, getGoodsDetail, getGoodsList } from '@/api'
import type { Banner, Goods } from '@/types'

const banners = ref<Banner[]>([])
const newGoods = ref<Goods[]>([])
const loadingNew = ref(false)
const categoryGroups = ref<{ name: string; goods: Goods[] }[]>([])
const loadingGroup = ref(false)

async function loadBanners() {
  try {
    const items = await getBanners()
    banners.value = await Promise.all(
      items.map(async (item) => {
        if (!item.image.includes('shop.projectsedu.com')) return item
        try {
          const goods = await getGoodsDetail(item.url)
          return { ...item, image: goods.front_image || item.image }
        } catch {
          return item
        }
      }),
    )
  } catch {
    banners.value = []
  }
}

async function loadNewGoods() {
  loadingNew.value = true
  try {
    const res = await getGoodsList({ in: '1', pnum: 8 })
    newGoods.value = res.data ?? []
  } catch {
    newGoods.value = []
  } finally {
    loadingNew.value = false
  }
}

async function loadCategoryGroups() {
  loadingGroup.value = true
  try {
    const categories = await getCategoryList()
    const groups = await Promise.all(
      categories.map(async (category) => {
        try {
          const res = await getGoodsList({ c: category.id, pnum: 8 })
          const goods = res.data ?? []
          return goods.length ? { name: category.name, goods } : null
        } catch {
          return null
        }
      }),
    )
    categoryGroups.value = groups.filter(
      (group): group is { name: string; goods: Goods[] } => group !== null,
    )
  } catch {
    categoryGroups.value = []
  } finally {
    loadingGroup.value = false
  }
}

onMounted(() => {
  loadBanners()
  loadNewGoods()
  loadCategoryGroups()
})
</script>

<style scoped lang="scss">
.home {
  padding: 20px 0;
}
.banner {
  border-radius: 8px;
  overflow: hidden;
  margin-bottom: 30px;
  .banner-link {
    display: block;
    width: 100%;
    height: 100%;
    :deep(.el-image) {
      width: 100%;
      height: 100%;
    }
  }
}
.section {
  margin-bottom: 36px;
}
.section-head {
  display: flex;
  justify-content: space-between;
  align-items: center;
  border-bottom: 2px solid var(--brand-color);
  margin-bottom: 20px;
  padding-bottom: 10px;
  h2 {
    font-size: 20px;
    font-weight: bold;
  }
  .more {
    font-size: 13px;
    color: #666;
    &:hover {
      color: var(--brand-color);
    }
  }
}
.goods-grid {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 16px;
  min-height: 100px;
}
</style>
