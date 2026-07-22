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

// 基础路径
const USER = '/u/v1'
const GOODS = '/g/v1'
const ORDER = '/o/v1'
const USER_OP = '/up/v1'

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

export const getUserDetail = (): Promise<UserDetail> => request.get(`${USER}/user/detail`)

export const updateUserInfo = (params: Partial<UserDetail>): Promise<any> =>
  request.patch(`${USER}/user/update`, params)

// ===== 商品 =====
export const getGoodsList = (params: GoodsListParams): Promise<PagedResult<Goods>> =>
  request.get(`${GOODS}/goods`, { params })

export const getGoodsDetail = (id: number | string): Promise<Goods> => request.get(`${GOODS}/goods/${id}`)

// 分类
export const getCategoryList = (): Promise<Category[]> => request.get(`${GOODS}/category`)

export const getCategory = (id: number | string): Promise<{
  total: number
  info: Category
  sub_categories: Category[]
}> => request.get(`${GOODS}/category/${id}`)

// 轮播图
export const getBanners = (): Promise<Banner[]> => request.get(`${GOODS}/banner`)

// 首页按分类分组的商品（外部 legacy 接口）
export const queryCategoryGoods = (): Promise<any[]> => request.get('/ext/indexgoods/')

export const getHotSearch = (): Promise<{ keywords: string }[]> => request.get('/ext/hotsearchs')

// ===== 购物车 =====
export const getShopCarts = (): Promise<{ total?: number; data: any[] }> =>
  request.get(`${ORDER}/shopcarts`)

export const addShopCart = (params: AddCartParams): Promise<any> =>
  request.post(`${ORDER}/shopcarts`, params)

export const updateShopCart = (goodsId: number | string, params: UpdateCartParams): Promise<any> =>
  request.patch(`${ORDER}/shopcarts/${goodsId}`, params)

export const deleteShopCart = (goodsId: number | string): Promise<any> =>
  request.delete(`${ORDER}/shopcarts/${goodsId}`)

// ===== 订单 =====
export const getOrders = (): Promise<{ data: OrderItem[] }> => request.get(`${ORDER}/orders`)

export const getOrderDetail = (orderId: number | string): Promise<OrderItem> =>
  request.get(`${ORDER}/orders/${orderId}`)

export const createOrder = (params: CreateOrderParams): Promise<{ alipay_url: string }> =>
  request.post(`${ORDER}/orders`, params)

export const deleteOrder = (orderId: number | string): Promise<any> =>
  request.delete(`${ORDER}/orders/${orderId}`)

// ===== 收藏 =====
export const getAllFavs = (): Promise<{ data: Goods[] }> => request.get(`${USER_OP}/userfavs`)

export const getFav = (goodsId: number | string): Promise<any> =>
  request.get(`${USER_OP}/userfavs/${goodsId}`)

export const addFav = (params: FavParams): Promise<any> => request.post(`${USER_OP}/userfavs`, params)

export const delFav = (goodsId: number | string): Promise<any> =>
  request.delete(`${USER_OP}/userfavs/${goodsId}`)

// ===== 收货地址 =====
export const getAddress = (): Promise<{ data: Address[] }> => request.get(`${USER_OP}/address`)

export const addAddress = (params: Address): Promise<any> => request.post(`${USER_OP}/address`, params)

export const updateAddress = (addressId: number | string, params: Address): Promise<any> =>
  request.patch(`${USER_OP}/address/${addressId}`, params)

export const delAddress = (addressId: number | string): Promise<any> =>
  request.delete(`${USER_OP}/address/${addressId}`)

// ===== 留言 =====
export const getMessages = (): Promise<{ data: MessageItem[] }> => request.get(`${USER_OP}/message`)

export const addMessage = (params: AddMessageParams): Promise<any> =>
  request.post(`${USER_OP}/message`, params)

export const delMessage = (messageId: number | string): Promise<any> =>
  request.delete(`${USER_OP}/message/${messageId}`)
