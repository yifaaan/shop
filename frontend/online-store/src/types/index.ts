// ===== 用户相关 =====
export interface UserInfo {
  name: string
  token: string
}

export interface LoginParams {
  mobile: string
  password: string
  captcha: string
  captcha_id: string
}

export interface RegisterParams {
  mobile: string
  password: string
  code: string
}

export interface AuthResult {
  id: number
  nickname: string
  token: string
}

export interface CaptchaResult {
  captcha_id: string
  picPath: string
}

export interface SendSmsParams {
  mobile: string
  type: 1 | 2 // 1=注册 2=登录
}

export interface UserDetail {
  id?: number
  name?: string
  nickname?: string
  mobile?: string
  birthday?: string
  gender?: 'male' | 'female' | string
}

// ===== 商品相关 =====
export interface Brand {
  id: number
  title?: string
  name?: string
  logo?: string
  image?: string
  description?: string
}

export interface Category {
  id: number
  name: string
  title?: string
  parent?: number
  level?: number
  is_tab?: boolean
  sub_category?: Category[]
}

export interface Goods {
  id: number
  name: string
  title?: string
  goods_brief?: string
  description?: string
  desc?: string
  shop_price: number
  ship_free?: boolean
  images?: string[]
  desc_images?: string[]
  front_image?: string
  goods_front_image?: string
  is_hot?: boolean
  is_new?: boolean
  is_tab?: boolean
  sold_num?: number
  goods_num?: number
  category?: Category | { id: number; name?: string }
  brand?: Brand
}

export interface GoodsListParams {
  p?: number // 页码
  pnum?: number // 每页数量
  c?: number // 分类 id
  b?: number // 品牌 id
  pmin?: number | string
  pmax?: number | string
  q?: string // 搜索关键词
  ih?: string // 是否热销 "1"
  in?: string // 是否新品 "1"
  it?: string // 是否 tab "1"
}

export interface PagedResult<T> {
  total: number
  data: T[]
}

export interface Banner {
  id: number
  index?: number
  image: string
  url: string | number // goods id
}

// ===== 购物车相关 =====
// 与 order_web cartItemResponse 对齐：id 为购物车记录ID（更新/删除按此id）
export interface CartItem {
  id: number
  goods_id: number
  goods_name?: string
  goods_image?: string
  goods_price: number
  num: number
  checked?: boolean
}

export interface CartListResult {
  total?: number
  data: CartItem[]
}

export interface AddCartParams {
  goods_id: number
  num: number
  checked?: boolean
}

export interface UpdateCartParams {
  num?: number
  checked?: boolean
}

// ===== 订单相关 =====
// 与 order_web CreateOrderForm 对齐：pay_type 必填（1微信 2支付宝）
export interface CreateOrderParams {
  address: string
  name: string
  mobile: string
  post?: string
  pay_type: 1 | 2
  post_fee?: number | string
}

export interface OrderItem {
  id: number
  order_sn?: string
  add_time?: number | string
  total?: number | string
  status?: number | string // 1待支付 2已支付 3已取消
  pay_type?: number
  post_fee?: number | string
  alipay_url?: string
  name?: string
  address?: string
  mobile?: string
  order_goods?: OrderGoods[]
}

export interface OrderGoods {
  id: number
  order_id?: number
  goods_id: number
  goods_name: string
  goods_image?: string
  goods_price: number
  num: number
}

// ===== 收货地址相关 =====
export interface Address {
  id?: number
  province: string
  city: string
  district: string
  address: string
  signer_name: string
  signer_mobile: string
}

// ===== 收藏 =====
// userop_web UserFavForm 以 goods_id 标识
export interface FavParams {
  goods_id: number
}

// userop_web 收藏列表项：id 为收藏记录ID（取消收藏按此），goods_id 指向商品；
// goods_name/image/price 由 userop_web 调 goods_srv 补全。
export interface FavItem {
  id: number
  user_id?: number
  goods_id: number
  goods_name?: string
  goods_image?: string
  goods_price?: number
}

// ===== 留言 =====
export interface MessageItem {
  id: number
  subject: string
  message: string
  type: number // 1留言 2投诉 3询问 4售后 5求购
  file?: string
}

export interface AddMessageParams {
  file?: string
  subject: string
  message: string
  type: number
}

// ===== 后端统一错误 =====
export interface ApiError {
  msg?: string | Record<string, string>
  [key: string]: any
}
