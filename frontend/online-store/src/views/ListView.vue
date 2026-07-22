<template>
  <div class="container list-page">
    <!-- 面包屑 -->
    <el-breadcrumb :separator-icon="ArrowRight" class="crumb">
      <el-breadcrumb-item :to="{ name: 'home' }">首页</el-breadcrumb-item>
      <el-breadcrumb-item v-if="currentCategoryName">{{ currentCategoryName }}</el-breadcrumb-item>
      <el-breadcrumb-item v-else>搜索 "{{ keyword }}"</el-breadcrumb-item>
    </el-breadcrumb>

    <div class="list-body">
      <!-- 左侧筛选 -->
      <aside class="sidebar">
        <div class="side-box">
          <h3 class="side-title">相关分类</h3>
          <ul class="cate-list">
            <li
              v-for="c in cateMenu"
              :key="c.id"
              :class="{ active: activeCate === c.id }"
              @click="changeMenu(c.id)"
            >
              {{ c.name }}
            </li>
          </ul>
        </div>
      </aside>

      <!-- 右侧商品 -->
      <div class="maincon">
        <!-- 排序 + 价格区间 -->
        <div class="toolbar">
          <div class="sort-group">
            <span>排序：</span>
            <el-radio-group v-model="ordering" size="small" @change="loadGoods">
              <el-radio-button label="-add_time">最新</el-radio-button>
              <el-radio-button label="shop_price">价格升序</el-radio-button>
              <el-radio-button label="-shop_price">价格降序</el-radio-button>
              <el-radio-button label="-sold_num">销量</el-radio-button>
            </el-radio-group>
          </div>
          <div class="price-filter">
            <span>价格：</span>
            <el-input-number v-model="priceMin" :min="0" :controls="false" size="small" placeholder="￥" />
            <span class="dash">-</span>
            <el-input-number v-model="priceMax" :min="0" :controls="false" size="small" placeholder="￥" />
            <el-button type="primary" size="small" @click="changePrice">确定</el-button>
          </div>
          <div class="total">共 <b class="price">{{ total }}</b> 件商品</div>
        </div>

        <!-- 商品列表 -->
        <div v-loading="loading" class="goods-grid">
          <ProductCard v-for="item in listData" :key="item.id" :goods="item" />
          <el-empty v-if="!loading && listData.length === 0" description="暂无商品" />
        </div>

        <!-- 分页 -->
        <div v-if="total > 0" class="pager">
          <el-pagination
            v-model:current-page="curPage"
            :page-size="pageSize"
            :total="total"
            layout="prev, pager, next, jumper"
            background
            @current-change="loadGoods"
          />
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, watch, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { ArrowRight } from '@element-plus/icons-vue'
import ProductCard from '@/components/ProductCard.vue'
import { getGoodsList, getCategory } from '@/api'
import type { Category, Goods } from '@/types'

const route = useRoute()

const listData = ref<Goods[]>([])
const loading = ref(false)
const total = ref(0)
const curPage = ref(1)
const pageSize = 12
const ordering = ref('-add_time')
const priceMin = ref<number | undefined>(undefined)
const priceMax = ref<number | undefined>(undefined)

const cateMenu = ref<Category[]>([])
const activeCate = ref<number | undefined>(undefined)
const currentCategoryName = ref('')
const keyword = ref('')

const isSearch = () => route.name === 'search'

async function loadMenu(id?: number | string) {
  if (id == null) {
    cateMenu.value = []
    return
  }
  try {
    const res = await getCategory(id)
    cateMenu.value = res.sub_categories ?? []
    currentCategoryName.value = res.info?.name ?? ''
  } catch {
    cateMenu.value = []
  }
}

async function loadGoods() {
  loading.value = true
  try {
    let res
    if (isSearch()) {
      res = await getGoodsList({ q: keyword.value, pnum: pageSize })
    } else {
      res = await getGoodsList({
        p: curPage.value,
        c: activeCate.value,
        pnum: pageSize,
        pmin: priceMin.value ?? '',
        pmax: priceMax.value ?? '',
      })
    }
    listData.value = res.data ?? []
    total.value = res.total ?? 0
  } finally {
    loading.value = false
  }
}

function changeMenu(id: number) {
  activeCate.value = id
  curPage.value = 1
  loadGoods()
}

function changePrice() {
  curPage.value = 1
  loadGoods()
}

function initFromRoute() {
  curPage.value = 1
  if (isSearch()) {
    keyword.value = route.params.keyword as string
    currentCategoryName.value = ''
    activeCate.value = undefined
    cateMenu.value = []
  } else {
    keyword.value = ''
    const id = route.params.id as string | undefined
    if (id) {
      activeCate.value = Number(id)
      loadMenu(id)
    } else {
      activeCate.value = undefined
      cateMenu.value = []
      currentCategoryName.value = '全部商品'
    }
  }
  loadGoods()
}

watch(() => route.fullPath, initFromRoute)
onMounted(initFromRoute)
</script>

<style scoped lang="scss">
.list-page {
  padding: 20px 0;
}
.crumb {
  margin-bottom: 16px;
}
.list-body {
  display: flex;
  gap: 16px;
  align-items: flex-start;
}
.sidebar {
  width: 200px;
  flex-shrink: 0;
}
.side-box {
  background: #fff;
  border: 1px solid var(--border-color);
  border-radius: 4px;
  overflow: hidden;
}
.side-title {
  padding: 12px 16px;
  font-size: 14px;
  background: #f5f5f5;
  border-bottom: 1px solid var(--border-color);
}
.cate-list {
  li {
    padding: 10px 16px;
    cursor: pointer;
    border-bottom: 1px solid #f5f5f5;
    &:last-child {
      border-bottom: none;
    }
    &:hover,
    &.active {
      color: var(--brand-color);
      background: #fff5f5;
    }
  }
}
.maincon {
  flex: 1;
  min-width: 0;
}
.toolbar {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 20px;
  background: #fff;
  border: 1px solid var(--border-color);
  padding: 12px 16px;
  margin-bottom: 16px;
  border-radius: 4px;
  .sort-group,
  .price-filter {
    display: flex;
    align-items: center;
    gap: 6px;
    font-size: 13px;
    .el-input-number {
      width: 90px;
    }
    .dash {
      color: #999;
    }
  }
  .total {
    margin-left: auto;
    font-size: 13px;
    color: #666;
  }
}
.goods-grid {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 14px;
  min-height: 200px;
}
.pager {
  display: flex;
  justify-content: center;
  margin-top: 24px;
}
</style>
