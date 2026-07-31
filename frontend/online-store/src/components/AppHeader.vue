<template>
  <header class="app-header">
    <!-- 顶部欢迎栏 -->
    <div class="top-bar">
      <div class="container top-bar-inner">
        <div class="welcome">
          <template v-if="userStore.isLoggedIn()">
            <router-link :to="{ name: 'userinfo' }" class="nickname">{{ userStore.name }}</router-link>
            <span class="sep">|</span>
            <a class="link" @click="handleLogout">退出</a>
          </template>
          <template v-else>
            <router-link :to="{ name: 'login' }" class="link">请登录</router-link>
            <span class="sep">|</span>
            <router-link :to="{ name: 'register' }" class="link">免费注册</router-link>
          </template>
        </div>
        <el-dropdown trigger="hover" @command="onMenuCommand">
          <span class="link menu-trigger">会员中心<el-icon><ArrowDown /></el-icon></span>
          <template #dropdown>
            <el-dropdown-menu>
              <el-dropdown-item command="order">我的订单</el-dropdown-item>
              <el-dropdown-item command="collection">我的收藏</el-dropdown-item>
              <el-dropdown-item command="receive">修改收货地址</el-dropdown-item>
            </el-dropdown-menu>
          </template>
        </el-dropdown>
      </div>
    </div>

    <!-- 主体：logo + 搜索 -->
    <div class="container main-bar">
      <router-link :to="{ name: 'home' }" class="logo">
        <span class="logo-text">慕学生鲜</span>
      </router-link>
      <div class="search-box">
        <el-input
          v-model="searchWord"
          placeholder="搜索商品"
          class="search-input"
          @keyup.enter="searchSubmit"
        >
          <template #append>
            <el-button type="primary" :icon="Search" @click="searchSubmit">搜索</el-button>
          </template>
        </el-input>
        <div v-if="hotSearch.length" class="hot-search">
          <span>热搜榜：</span>
          <router-link
            v-for="item in hotSearch"
            :key="item.keywords"
            :to="{ name: 'search', params: { keyword: item.keywords } }"
          >
            {{ item.keywords }}
          </router-link>
        </div>
      </div>
      <!-- 购物车 -->
      <el-dropdown trigger="hover" placement="bottom-end" @visible-change="onCartVisible">
        <div class="cart-entry">
          <el-badge :value="cartStore.goodsList.length" :hidden="cartStore.goodsList.length === 0">
            <el-button type="primary" :icon="ShoppingCart">购物车</el-button>
          </el-badge>
        </div>
        <template #dropdown>
          <div class="cart-dropdown">
            <div v-if="cartStore.goodsList.length === 0" class="cart-empty">购物车是空的</div>
            <template v-else>
              <div
                v-for="(item, index) in cartStore.goodsList"
                :key="item.goods_id"
                class="cart-item"
              >
                <router-link :to="{ name: 'productDetail', params: { productId: item.goods_id } }">
                  <el-image :src="item.goods_image" class="cart-thumb" fit="cover" />
                </router-link>
                <div class="cart-info">
                  <router-link
                    :to="{ name: 'productDetail', params: { productId: item.goods_id } }"
                    class="cart-name"
                    >{{ item.goods_name }}</router-link
                  >
                  <div class="cart-meta">
                    <span class="price">￥{{ item.goods_price }}</span>
                    <span class="nums">×{{ item.num }}</span>
                    <el-button link type="danger" size="small" @click="removeCart(index, item.id)"
                      >删除</el-button
                    >
                  </div>
                </div>
              </div>
              <div class="cart-footer">
                <span
                  >共 <b class="price">{{ cartStore.goodsList.length }}</b> 件 总价：
                  <b class="price">￥{{ cartStore.totalPrice.toFixed(2) }}</b></span
                >
                <el-button type="primary" size="small" @click="router.push({ name: 'cart' })"
                  >去结算</el-button
                >
              </div>
            </template>
          </div>
        </template>
      </el-dropdown>
    </div>

    <!-- 分类导航 -->
    <nav class="nav-bar">
      <div class="container nav-inner">
        <el-dropdown
          v-for="item in allMenu"
          :key="item.id"
          trigger="hover"
          placement="bottom-start"
          @visible-change="(v: boolean) => v && loadSubMenu(item.id)"
        >
          <router-link :to="{ name: 'list', params: { id: item.id } }" class="nav-item">
            {{ item.name }}
          </router-link>
          <template #dropdown>
            <div class="sub-menu">
              <template v-if="subMenuCache[item.id]?.length">
                <div
                  v-for="sub in subMenuCache[item.id]"
                  :key="sub.id"
                  class="sub-group"
                >
                  <router-link
                    :to="{ name: 'list', params: { id: sub.id } }"
                    class="sub-title"
                    >{{ sub.name }}</router-link
                  >
                  <div class="sub-children">
                    <router-link
                      v-for="child in sub.sub_category"
                      :key="child.id"
                      :to="{ name: 'list', params: { id: child.id } }"
                      >{{ child.name }}</router-link
                    >
                  </div>
                </div>
              </template>
              <el-skeleton v-else :rows="3" animated style="width: 300px" />
            </div>
          </template>
        </el-dropdown>
        <router-link :to="{ name: 'home' }" class="nav-item home-link">首页</router-link>
      </div>
    </nav>
  </header>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { Search, ShoppingCart, ArrowDown } from '@element-plus/icons-vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { useUserStore, useCartStore } from '@/stores'
import { getCategoryList, getCategory, getHotSearch, deleteShopCart } from '@/api'
import type { Category } from '@/types'

const router = useRouter()
const userStore = useUserStore()
const cartStore = useCartStore()

const searchWord = ref('')
const hotSearch = ref<{ keywords: string }[]>([])
const allMenu = ref<Category[]>([])
const subMenuCache = reactive<Record<number, Category[]>>({})

function searchSubmit() {
  if (!searchWord.value.trim()) return
  router.push({ name: 'search', params: { keyword: searchWord.value.trim() } })
}

function handleLogout() {
  userStore.logout()
  cartStore.clear()
  ElMessage.success('已退出登录')
  router.push({ name: 'home' })
}

function onMenuCommand(cmd: string) {
  router.push({ name: cmd })
}

async function loadSubMenu(id: number) {
  if (subMenuCache[id]) return
  try {
    const res = await getCategory(id)
    subMenuCache[id] = res.sub_categories ?? []
  } catch {
    subMenuCache[id] = []
  }
}

function onCartVisible(v: boolean) {
  if (v && userStore.isLoggedIn()) cartStore.refresh()
}

async function removeCart(index: number, id: number) {
  try {
    await ElMessageBox.confirm('确定将该商品移出购物车吗？', '提示', { type: 'warning' })
  } catch {
    return
  }
  try {
    await deleteShopCart(id)
    cartStore.goodsList.splice(index, 1)
    ElMessage.success('已删除')
  } catch {
    /* handled by interceptor */
  }
}

onMounted(async () => {
  try {
    allMenu.value = await getCategoryList()
  } catch {
    allMenu.value = []
  }
  try {
    hotSearch.value = await getHotSearch()
  } catch {
    hotSearch.value = []
  }
  if (userStore.isLoggedIn()) cartStore.refresh()
})
</script>

<style scoped lang="scss">
.app-header {
  background: #fff;
  position: relative;
  z-index: 100;
}

.top-bar {
  height: 34px;
  line-height: 34px;
  border-bottom: 1px solid #e5e5e5;
  background: #f5f5f5;
  font-size: 12px;

  .top-bar-inner {
    display: flex;
    justify-content: space-between;
    align-items: center;
  }
  .nickname {
    color: var(--brand-color);
    font-weight: bold;
    margin-right: 4px;
  }
  .sep {
    color: #ccc;
    margin: 0 8px;
  }
  .link {
    cursor: pointer;
    &:hover {
      color: var(--brand-color);
    }
  }
  .menu-trigger {
    display: inline-flex;
    align-items: center;
    gap: 2px;
  }
}

.main-bar {
  display: flex;
  align-items: center;
  padding: 20px 0;
  gap: 40px;

  .logo {
    flex-shrink: 0;
    .logo-text {
      font-size: 28px;
      font-weight: bold;
      color: var(--brand-color);
      letter-spacing: 2px;
    }
  }
  .search-box {
    flex: 1;
    max-width: 560px;
    .search-input {
      :deep(.el-input__wrapper) {
        border: 2px solid var(--brand-color);
        box-shadow: none;
      }
    }
    .hot-search {
      margin-top: 6px;
      font-size: 12px;
      color: #999;
      a {
        margin: 0 6px;
        color: #666;
        &:hover {
          color: var(--brand-color);
        }
      }
    }
  }
  .cart-entry {
    flex-shrink: 0;
  }
}

.nav-bar {
  background: var(--brand-color);
  height: 40px;

  .nav-inner {
    display: flex;
    align-items: center;
    height: 100%;
  }
  .nav-item {
    display: inline-flex;
    align-items: center;
    height: 100%;
    padding: 0 24px;
    color: #fff;
    font-size: 14px;
    &:hover {
      background: var(--brand-color-dark);
      color: #fff;
    }
  }
  .home-link {
    margin-left: auto;
  }
}

// 购物车下拉
.cart-dropdown {
  width: 320px;
  padding: 8px;
  .cart-empty {
    text-align: center;
    color: #999;
    padding: 30px 0;
  }
  .cart-item {
    display: flex;
    gap: 10px;
    padding: 8px 0;
    border-bottom: 1px dashed #eee;
    .cart-thumb {
      width: 50px;
      height: 50px;
      border-radius: 4px;
      flex-shrink: 0;
    }
    .cart-info {
      flex: 1;
      min-width: 0;
    }
    .cart-name {
      display: block;
      font-size: 12px;
      overflow: hidden;
      text-overflow: ellipsis;
      white-space: nowrap;
    }
    .cart-meta {
      display: flex;
      align-items: center;
      gap: 8px;
      margin-top: 4px;
      font-size: 12px;
      .nums {
        color: #999;
      }
    }
  }
  .cart-footer {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding-top: 10px;
    font-size: 12px;
  }
}

// 分类子菜单
.sub-menu {
  max-width: 520px;
  padding: 10px;
  .sub-group {
    padding: 6px 0;
    border-bottom: 1px dashed #eee;
    &:last-child {
      border-bottom: none;
    }
  }
  .sub-title {
    font-weight: bold;
    font-size: 13px;
    color: var(--brand-color);
  }
  .sub-children {
    margin-top: 4px;
    a {
      display: inline-block;
      margin: 2px 10px 2px 0;
      font-size: 12px;
      color: #666;
      &:hover {
        color: var(--brand-color);
      }
    }
  }
}

@media (max-width: 767px) {
  .main-bar {
    flex-wrap: wrap;
    gap: 12px;
    padding: 12px 0;

    .logo {
      flex: 1;
      min-width: 0;

      .logo-text {
        font-size: 22px;
        letter-spacing: 0;
      }
    }

    .search-box {
      order: 3;
      flex: 0 0 100%;
      max-width: none;
      min-width: 0;

      .hot-search {
        display: none;
      }
    }
  }

  .nav-bar {
    .nav-inner {
      overflow-x: auto;
      scrollbar-width: none;

      &::-webkit-scrollbar {
        display: none;
      }
    }

    .nav-item {
      flex-shrink: 0;
      padding: 0 14px;
    }

    .home-link {
      margin-left: 0;
    }
  }
}
</style>
