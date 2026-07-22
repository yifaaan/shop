<template>
  <div class="panel">
    <h2 class="panel-title">我的收藏</h2>
    <el-table v-loading="loading" :data="collections" border stripe>
      <el-table-column label="商品名称" min-width="280">
        <template #default="{ row }">
          <router-link
            :to="{ name: 'productDetail', params: { productId: row.id } }"
            class="link"
            >{{ row.name }}</router-link
          >
        </template>
      </el-table-column>
      <el-table-column label="价格" width="160" align="center">
        <template #default="{ row }">
          <span class="price">￥{{ row.shop_price }}</span>
        </template>
      </el-table-column>
      <el-table-column label="操作" width="120" align="center">
        <template #default="{ row, $index }">
          <el-button link type="danger" @click="remove($index, row.id)">删除</el-button>
        </template>
      </el-table-column>
    </el-table>
    <el-empty v-if="!loading && collections.length === 0" description="还没有收藏任何商品" />
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { getAllFavs, delFav } from '@/api'
import type { Goods } from '@/types'

const loading = ref(false)
const collections = ref<Goods[]>([])

async function load() {
  loading.value = true
  try {
    const res = await getAllFavs()
    collections.value = res.data ?? []
  } finally {
    loading.value = false
  }
}

async function remove(index: number, id: number) {
  try {
    await ElMessageBox.confirm('确定要从收藏夹中删除该商品吗？', '提示', { type: 'warning' })
  } catch {
    return
  }
  try {
    await delFav(id)
    collections.value.splice(index, 1)
    ElMessage.success('已删除')
  } catch {
    /* handled */
  }
}

onMounted(load)
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
  &:hover {
    color: var(--brand-color);
  }
}
</style>
