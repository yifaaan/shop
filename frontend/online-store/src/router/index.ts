import { createRouter, createWebHistory, type RouteRecordRaw } from 'vue-router'
import { cookie } from '@/utils/cookie'

// 路由组件（懒加载）
const DefaultLayout = () => import('@/layouts/DefaultLayout.vue')
const LoginLayout = () => import('@/layouts/LoginLayout.vue')

const HomeView = () => import('@/views/HomeView.vue')
const LoginView = () => import('@/views/LoginView.vue')
const RegisterView = () => import('@/views/RegisterView.vue')
const ListView = () => import('@/views/ListView.vue')
const ProductDetailView = () => import('@/views/ProductDetailView.vue')
const CartView = () => import('@/views/CartView.vue')

const MemberView = () => import('@/views/member/MemberView.vue')
const OrderView = () => import('@/views/member/OrderView.vue')
const OrderDetailView = () => import('@/views/member/OrderDetailView.vue')
const ReceiveView = () => import('@/views/member/ReceiveView.vue')
const CollectionView = () => import('@/views/member/CollectionView.vue')
const MessageView = () => import('@/views/member/MessageView.vue')
const UserInfoView = () => import('@/views/member/UserInfoView.vue')
const NotFoundView = () => import('@/views/NotFoundView.vue')

const routes: RouteRecordRaw[] = [
  {
    path: '/',
    component: DefaultLayout,
    children: [
      { path: '', redirect: { name: 'home' } },
      {
        path: 'home',
        name: 'home',
        component: HomeView,
        meta: { title: '慕学生鲜-首页' },
      },
      {
        path: 'list/:id?',
        name: 'list',
        component: ListView,
        meta: { title: '慕学生鲜-商品列表' },
      },
      {
        path: 'search/:keyword',
        name: 'search',
        component: ListView,
        meta: { title: '搜索结果' },
      },
      {
        path: 'product/:productId',
        name: 'productDetail',
        component: ProductDetailView,
        meta: { title: '慕学生鲜-商品详情' },
      },
      {
        path: 'cart',
        name: 'cart',
        component: CartView,
        meta: { title: '慕学生鲜-购物车', requireAuth: true },
      },
      {
        path: 'member',
        component: MemberView,
        meta: { requireAuth: true },
        children: [
          { path: '', redirect: { name: 'userinfo' } },
          { path: 'order', name: 'order', component: OrderView, meta: { title: '我的订单' } },
          {
            path: 'order/:orderId',
            name: 'orderDetail',
            component: OrderDetailView,
            meta: { title: '订单详情' },
          },
          { path: 'receive', name: 'receive', component: ReceiveView, meta: { title: '收货地址' } },
          { path: 'userinfo', name: 'userinfo', component: UserInfoView, meta: { title: '用户信息' } },
          {
            path: 'collection',
            name: 'collection',
            component: CollectionView,
            meta: { title: '我的收藏' },
          },
          { path: 'message', name: 'message', component: MessageView, meta: { title: '我的留言' } },
        ],
      },
    ],
  },
  {
    path: '/',
    component: LoginLayout,
    children: [
      { path: 'login', name: 'login', component: LoginView, meta: { title: '慕学生鲜-登录' } },
      { path: 'register', name: 'register', component: RegisterView, meta: { title: '慕学生鲜-注册' } },
    ],
  },
  { path: '/:pathMatch(.*)*', name: 'notFound', component: NotFoundView },
]

const router = createRouter({
  history: createWebHistory(),
  routes,
  scrollBehavior() {
    return { top: 0 }
  },
})

// 全局前置守卫：鉴权 + 标题
router.beforeEach((to, _from, next) => {
  const title = to.meta.title as string | undefined
  if (title) document.title = title

  if (to.meta.requireAuth && !cookie.get('token')) {
    next({ name: 'login', query: { redirect: to.fullPath } })
  } else {
    next()
  }
})

export default router
