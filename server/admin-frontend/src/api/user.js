import api from './index'

// 获取用户列表
export const getUsers = (params) => api.get('/users', { params })

// 获取单个用户
export const getUser = (id) => api.get(`/users/${id}`)

// 更新用户
export const updateUser = (id, data) => api.put(`/users/${id}`, data)

// 禁用/启用用户
export const setUserStatus = (id, status) => api.put(`/users/${id}/status`, { status })
