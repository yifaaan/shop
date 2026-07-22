<template>
  <div class="container member-page">
    <el-breadcrumb :separator-icon="ArrowRight" class="crumb">
      <el-breadcrumb-item :to="{ name: 'home' }">首页</el-breadcrumb-item>
      <el-breadcrumb-item>用户中心</el-breadcrumb-item>
    </el-breadcrumb>

    <div class="member-body">
      <aside class="menu">
        <h3 class="menu-title">会员中心</h3>
        <el-menu :default-active="activeMenu" @select="onSelect">
          <el-sub-menu index="order-center">
            <template #title>订单中心</template>
            <el-menu-item index="order">我的订单</el-menu-item>
            <el-menu-item index="receive">收货地址</el-menu-item>
          </el-sub-menu>
          <el-sub-menu index="member-center">
            <template #title>会员中心</template>
            <el-menu-item index="userinfo">用户信息</el-menu-item>
            <el-menu-item index="collection">我的收藏</el-menu-item>
            <el-menu-item index="message">我的留言</el-menu-item>
          </el-sub-menu>
        </el-menu>
      </aside>
      <section class="content">
        <router-view />
      </section>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ArrowRight } from '@element-plus/icons-vue'

const route = useRoute()
const router = useRouter()

const activeMenu = computed(() => (route.name as string) ?? 'userinfo')

function onSelect(index: string) {
  router.push({ name: index })
}
</script>

<style scoped lang="scss">
.member-page {
  padding: 20px 0;
}
.crumb {
  margin-bottom: 16px;
}
.member-body {
  display: flex;
  gap: 16px;
  align-items: flex-start;
}
.menu {
  width: 200px;
  flex-shrink: 0;
  background: #fff;
  border: 1px solid var(--border-color);
  border-radius: 4px;
  overflow: hidden;
}
.menu-title {
  height: 50px;
  line-height: 50px;
  text-align: center;
  font-size: 16px;
  background: #f5f5f5;
  border-bottom: 1px solid var(--border-color);
}
.content {
  flex: 1;
  min-width: 0;
}
</style>
