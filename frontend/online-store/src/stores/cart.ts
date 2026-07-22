import { defineStore } from 'pinia'
import { ref } from 'vue'
import { getShopCarts } from '@/api'
import type { CartItem } from '@/types'

export const useCartStore = defineStore('cart', () => {
  const goodsList = ref<CartItem[]>([])
  const totalPrice = ref(0)

  async function refresh() {
    const token = localStorage.getItem('ignore') // placeholder; token check via cookie elsewhere
    void token
    try {
      const res = await getShopCarts()
      goodsList.value = res.data ?? []
      totalPrice.value = goodsList.value.reduce((sum, item) => sum + item.good_price * item.nums, 0)
    } catch (e) {
      goodsList.value = []
      totalPrice.value = 0
    }
  }

  function clear() {
    goodsList.value = []
    totalPrice.value = 0
  }

  return { goodsList, totalPrice, refresh, clear }
})
