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
  captchaId: string
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
export interface CartItem {
  goods_id: number
  good_name?: string
  good_image?: string
  good_price: number
  goods_price?: number
  nums: number
  checked?: boolean
}

export interface CartListResult {
  total?: number
  data: CartItem[]
}

export interface AddCartParams {
  goods: number
  nums: number
}

export interface UpdateCartParams {
  nums?: number
  checked?: boolean
}

// ===== 订单相关 =====
export interface CreateOrderParams {
  post: string
  address: string
  name: string
  mobile: string
  order_mount: number | string
}

export interface OrderItem {
  id: number
  order_sn?: string
  add_time?: string
  total?: number | string
  status?: string // paying | TRADE_SUCCESS | TRADE_CLOSED | ''
  alipay_url?: string
  name?: string
  address?: string
  mobile?: string
  goods?: OrderGoods[]
}

export interface OrderGoods {
  id: number
  name: string
  price: number
  nums: number
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
export interface FavParams {
  goods: number
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
