<template>
  <footer class="app-footer">
    <div class="ensure-bar">
      <div class="container">
        <ul>
          <li><span class="ensure-icon">正</span><div><h4>正品保障</h4><p>100%正品低价</p></div></li>
          <li><span class="ensure-icon">退</span><div><h4>30天退换货</h4><p>购物有保障</p></div></li>
          <li><span class="ensure-icon">包</span><div><h4>满99就包邮</h4><p>闪电发货</p></div></li>
          <li><span class="ensure-icon">服</span><div><h4>贴心服务</h4><p>售后无忧</p></div></li>
        </ul>
      </div>
    </div>
    <div class="footer-text">
      <p>服务时间：09:00-23:00 &nbsp;&nbsp; 客服热线：<em>010-88888888</em></p>
      <p class="copyright">© {{ year }} 慕学生鲜商城</p>
    </div>
  </footer>
</template>

<script setup lang="ts">
const year = new Date().getFullYear()
</script>

<style scoped lang="scss">
.app-footer {
  border-top: 1px solid var(--border-color);
  background: #fff;
  margin-top: 40px;
}
.ensure-bar {
  background: var(--brand-color);
  height: 90px;
  ul {
    display: flex;
    height: 100%;
    li {
      flex: 1;
      display: flex;
      align-items: center;
      justify-content: center;
      gap: 14px;
      color: #fff;
      .ensure-icon {
        width: 50px;
        height: 50px;
        border-radius: 50%;
        background: rgba(255, 255, 255, 0.2);
        display: flex;
        align-items: center;
        justify-content: center;
        font-size: 22px;
        font-weight: bold;
      }
      h4 {
        font-size: 16px;
      }
      p {
        font-size: 12px;
        opacity: 0.85;
      }
    }
  }
}
.footer-text {
  text-align: center;
  padding: 16px 0;
  color: #666;
  font-size: 12px;
  em {
    color: var(--brand-color);
    font-weight: bold;
    font-size: 20px;
  }
  .copyright {
    margin-top: 6px;
    color: #999;
  }
}
</style>
