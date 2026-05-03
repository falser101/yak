import api from './index'

// 获取品牌列表
export const getBrands = () => api.get('/brands')

// 获取单个品牌
export const getBrand = (id) => api.get(`/brands/${id}`)

// 创建品牌
export const createBrand = (data) => api.post('/brands', data)

// 更新品牌
export const updateBrand = (id, data) => api.put(`/brands/${id}`, data)

// 删除品牌
export const deleteBrand = (id) => api.delete(`/brands/${id}`)

// 获取品牌车型列表
export const getBrandModels = (brandId) => api.get(`/brands/${brandId}/models`)

// 创建车型
export const createModel = (data) => api.post('/models', data)

// 更新车型
export const updateModel = (id, data) => api.put(`/models/${id}`, data)

// 删除车型
export const deleteModel = (id) => api.delete(`/models/${id}`)
