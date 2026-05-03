import api from './index'

// 获取租车车辆列表
export const getRentalBikes = () => api.get('/rental-bikes')

// 创建租车车辆
export const createRentalBike = (data) => api.post('/rental-bikes', data)

// 更新租车车辆
export const updateRentalBike = (id, data) => api.put(`/rental-bikes/${id}`, data)

// 删除租车车辆
export const deleteRentalBike = (id) => api.delete(`/rental-bikes/${id}`)

// 获取用户车辆列表
export const getUserBikes = () => api.get('/bikes')
