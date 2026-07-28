import request from '@/utils/request'
import type {
  AddCartParams,
  AddMessageParams,
  Address,
  ApiError,
  AuthResult,
  Banner,
  CaptchaResult,
  Category,
  CreateOrderParams,
  FavItem,
  FavParams,
  Goods,
  GoodsListParams,
  LoginParams,
  MessageItem,
  OrderItem,
  PagedResult,
  RegisterParams,
  SendSmsParams,
  UpdateCartParams,
  UserDetail,
} from '@/types'

// 基础路径：对齐四个 web 后端
const USER = '/u/v1' // user_web
const GOODS = '/g/v1' // goods_web
const ORDER = '/o/v1' // order_web
const USER_OP = '/uo/v1' // userop_web

function parseError(e: any): ApiError {
  return e ?? {}
}

export { parseError }

// ===== 用户认证 =====
export const getCaptcha = (): Promise<CaptchaResult> => request.get(`${USER}/base/captcha`)

export const sendSms = (params: SendSmsParams): Promise<{ msg: string }> =>
  request.post(`${USER}/base/send_sms`, params)

export const login = (params: LoginParams): Promise<AuthResult> =>
  request.post(`${USER}/user/pwd_login`, params)

export const register = (params: RegisterParams): Promise<AuthResult> =>
  request.post(`${USER}/user/register`, params)

// 个人资料（user_web GET /user/detail，按 token 中的 userId 查询）
export const getUserDetail = (): Promise<UserDetail> => request.get(`${USER}/user/detail`)

// 更新个人资料（user_web PATCH /user/update；后端不支持改 mobile，忽略入参 mobile）
export const updateUserInfo = (params: Partial<UserDetail>): Promise<any> =>
  request.patch(`${USER}/user/update`, params)

// ===== 商品 =====
export const getGoodsList = (params: GoodsListParams): Promise<PagedResult<Goods>> =>
  request.get(`${GOODS}/goods`, { params })

export const getGoodsDetail = (id: number | string): Promise<Goods> => request.get(`${GOODS}/goods/${id}`)

// 分类：GET /category 返回 { total, data: <tree> }，这里取 data 作为顶层分类数组
export const getCategoryList = (): Promise<Category[]> =>
  request.get(`${GOODS}/category`).then((r: any) => r?.data ?? [])

export const getCategory = (id: number | string): Promise<{
  total: number
  info: Category
  sub_categories: Category[]
}> => request.get(`${GOODS}/category/${id}`)

// 轮播图：GET /banner 返回 { total, data: [...] }（proto camelCase），取 data
export const getBanners = (): Promise<Banner[]> =>
  request.get(`${GOODS}/banner`).then((r: any) => r?.data ?? [])

// 首页按分类分组的商品（外部 legacy 接口）
export const queryCategoryGoods = (): Promise<any[]> => request.get('/ext/indexgoods/')

export const getHotSearch = (): Promise<{ keywords: string }[]> => request.get('/ext/hotsearchs')

// ===== 购物车（order_web /cart） =====
export const getShopCarts = (): Promise<{ total?: number; data: any[] }> =>
  request.get(`${ORDER}/cart`)

export const addShopCart = (params: AddCartParams): Promise<any> =>
  request.post(`${ORDER}/cart`, params)

// 注意：路径 id 为购物车记录ID（非 goods_id）；更新用 PUT
export const updateShopCart = (id: number | string, params: UpdateCartParams): Promise<any> =>
  request.put(`${ORDER}/cart/${id}`, params)

export const deleteShopCart = (id: number | string): Promise<any> =>
  request.delete(`${ORDER}/cart/${id}`)

// ===== 订单 =====
export const getOrders = (): Promise<{ total?: number; data: OrderItem[] }> =>
  request.get(`${ORDER}/orders`)

export const getOrderDetail = (orderId: number | string): Promise<OrderItem> =>
  request.get(`${ORDER}/orders/${orderId}`)

export const createOrder = (params: CreateOrderParams): Promise<{ id: number; alipay_url: string }> =>
  request.post(`${ORDER}/orders`, params)

// 更新订单状态（order_web PUT /orders/status）：1-待支付 2-已支付 3-已取消
export const updateOrderStatus = (params: { order_sn: string; status: number }): Promise<any> =>
  request.put(`${ORDER}/orders/status`, params)

// 删除订单（order_web DELETE /orders/{id}，带归属校验）
export const deleteOrder = (orderId: number | string): Promise<any> =>
  request.delete(`${ORDER}/orders/${orderId}`)

// ===== 收藏（userop_web /favs） =====
export const getAllFavs = (): Promise<{ total?: number; data: FavItem[] }> =>
  request.get(`${USER_OP}/favs`)

export const getFav = (goodsId: number | string): Promise<any> =>
  request.get(`${USER_OP}/favs/${goodsId}`)

export const addFav = (params: FavParams): Promise<any> => request.post(`${USER_OP}/favs`, params)

export const delFav = (goodsId: number | string): Promise<any> =>
  request.delete(`${USER_OP}/favs/${goodsId}`)

// ===== 收货地址（userop_web /addresses） =====
export const getAddress = (): Promise<{ total?: number; data: Address[] }> =>
  request.get(`${USER_OP}/addresses`)

export const addAddress = (params: Address): Promise<any> => request.post(`${USER_OP}/addresses`, params)

export const updateAddress = (addressId: number | string, params: Address): Promise<any> =>
  request.put(`${USER_OP}/addresses/${addressId}`, params)

export const delAddress = (addressId: number | string): Promise<any> =>
  request.delete(`${USER_OP}/addresses/${addressId}`)

// ===== 留言（userop_web /messages） =====
export const getMessages = (): Promise<{ total?: number; data: MessageItem[] }> =>
  request.get(`${USER_OP}/messages`)

export const addMessage = (params: AddMessageParams): Promise<any> =>
  request.post(`${USER_OP}/messages`, params)

export const delMessage = (messageId: number | string): Promise<any> =>
  request.delete(`${USER_OP}/messages/${messageId}`)