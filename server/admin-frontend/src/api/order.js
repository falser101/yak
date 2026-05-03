import api from './index'

// 获取租车订单列表
export const getRentalOrders = (params) => api.get('/rental-orders', { params })

// 确认收款
export const confirmPayment = (id) => api.put(`/rental-orders/${id}/confirm`, { status: 1 })

// 获取租车统计
export const getRentalStats = () => api.get('/rental-stats')
