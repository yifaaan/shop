<template>
  <router-link :to="{ name: 'productDetail', params: { productId: goods.id } }" class="product-card">
    <div class="img-box">
      <el-image :src="goods.front_image" fit="cover" lazy />
      <span v-if="goods.is_hot" class="tag tag-hot">热卖</span>
      <span v-else-if="goods.is_new" class="tag tag-new">新品</span>
    </div>
    <h3 class="name" :title="goods.name">{{ goods.name }}</h3>
    <p v-if="goods.goods_brief" class="brief">{{ goods.goods_brief }}</p>
    <div class="card-footer">
      <span class="price">￥{{ goods.shop_price }}</span>
      <span v-if="goods.sold_num != null" class="sold">销量 {{ goods.sold_num }}</span>
    </div>
  </router-link>
</template>

<script setup lang="ts">
import type { Goods } from '@/types'

defineProps<{ goods: Goods }>()
</script>

<style scoped lang="scss">
.product-card {
  display: block;
  background: #fff;
  border: 1px solid #f0f0f0;
  border-radius: 6px;
  overflow: hidden;
  transition: all 0.2s;

  &:hover {
    border-color: var(--brand-color);
    box-shadow: 0 4px 16px rgba(200, 22, 35, 0.12);
    transform: translateY(-2px);
  }
}
.img-box {
  position: relative;
  width: 100%;
  aspect-ratio: 1;
  background: #fafafa;
  :deep(.el-image) {
    width: 100%;
    height: 100%;
  }
  .tag {
    position: absolute;
    top: 8px;
    left: 8px;
    color: #fff;
    font-size: 12px;
    padding: 2px 8px;
    border-radius: 3px;
  }
  .tag-hot {
    background: var(--brand-color);
  }
  .tag-new {
    background: #f9b548;
  }
}
.name {
  font-size: 14px;
  font-weight: normal;
  padding: 10px 10px 4px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.brief {
  font-size: 12px;
  color: #999;
  padding: 0 10px;
  height: 36px;
  line-height: 18px;
  overflow: hidden;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
}
.card-footer {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 8px 10px 12px;
  .price {
    font-size: 18px;
    color: var(--brand-color);
    font-weight: bold;
  }
  .sold {
    font-size: 12px;
    color: #bbb;
  }
}
</style>
