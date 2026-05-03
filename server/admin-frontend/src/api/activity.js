import api from './index'

// 获取活动列表
export const getActivities = (params) => api.get('/activities', { params })

// 获取单个活动
export const getActivity = (id) => api.get(`/activities/${id}`)

// 创建活动
export const createActivity = (data) => api.post('/activities', data)

// 更新活动
export const updateActivity = (id, data) => api.put(`/activities/${id}`, data)

// 删除活动
export const deleteActivity = (id) => api.delete(`/activities/${id}`)

// 获取活动报名列表
export const getActivitySignups = (id) => api.get(`/activities/${id}/signups`)
