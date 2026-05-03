import api from './index'

// 获取统计数据
export const getStats = () => api.get('/stats')
