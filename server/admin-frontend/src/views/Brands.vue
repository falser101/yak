<template>
  <div class="brands-page">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>🏷️ 品牌列表</span>
          <el-button type="primary" @click="openBrandForm()">
            + 新增品牌
          </el-button>
        </div>
      </template>

      <el-table :data="brands" v-loading="loading" stripe>
        <el-table-column prop="id" label="ID" width="60" />
        <el-table-column label="Logo" width="80">
          <template #default="{ row }">
            <div class="brand-logo">{{ row.logo }}</div>
          </template>
        </el-table-column>
        <el-table-column prop="name" label="品牌名称" min-width="150" />
        <el-table-column prop="description" label="描述" min-width="300" show-overflow-tooltip />
        <el-table-column prop="modelCount" label="车型数" width="100">
          <template #default="{ row }">
            <el-tag type="success">{{ row.modelCount }} 款</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="200" fixed="right">
          <template #default="{ row }">
            <el-button link type="primary" size="small" @click="openModels(row)">
              车型
            </el-button>
            <el-button link type="primary" size="small" @click="openBrandForm(row)">
              编辑
            </el-button>
            <el-button link type="danger" size="small" @click="deleteBrand(row)">
              删除
            </el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <!-- 品牌表单弹窗 -->
    <el-dialog v-model="brandFormVisible" :title="brandForm.id ? '编辑品牌' : '新增品牌'" width="500px">
      <el-form :model="brandForm" label-width="100px">
        <el-form-item label="品牌名称">
          <el-input v-model="brandForm.name" placeholder="如：Giant" />
        </el-form-item>
        <el-form-item label="Logo">
          <el-input v-model="brandForm.logo" placeholder="如：G" />
        </el-form-item>
        <el-form-item label="品牌描述">
          <el-input v-model="brandForm.description" type="textarea" rows="3" placeholder="品牌介绍..." />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="brandFormVisible = false">取消</el-button>
        <el-button type="primary" @click="saveBrand">保存</el-button>
      </template>
    </el-dialog>

    <!-- 车型管理弹窗 -->
    <el-dialog v-model="modelsVisible" :title="currentBrand?.name + ' - 车型管理'" width="700px">
      <div class="models-header">
        <el-button type="primary" size="small" @click="openModelForm()">
          + 新增车型
        </el-button>
      </div>

      <el-table :data="models" stripe>
        <el-table-column prop="id" label="ID" width="60" />
        <el-table-column prop="name" label="车型名称" min-width="150" />
        <el-table-column prop="bikeType" label="类型" width="100" />
        <el-table-column prop="price" label="参考价格" width="120">
          <template #default="{ row }">
            ¥{{ row.price || 0 }}
          </template>
        </el-table-column>
        <el-table-column label="操作" width="120">
          <template #default="{ row }">
            <el-button link type="primary" size="small" @click="openModelForm(row)">
              编辑
            </el-button>
            <el-button link type="danger" size="small" @click="deleteModel(row)">
              删除
            </el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-dialog>

    <!-- 车型表单弹窗 -->
    <el-dialog v-model="modelFormVisible" :title="modelForm.id ? '编辑车型' : '新增车型'" width="500px">
      <el-form :model="modelForm" label-width="100px">
        <el-form-item label="车型名称">
          <el-input v-model="modelForm.name" placeholder="如：TCR Advanced SL" />
        </el-form-item>
        <el-form-item label="类型">
          <el-input v-model="modelForm.bikeType" placeholder="如：公路、山地" />
        </el-form-item>
        <el-form-item label="参考价格">
          <el-input-number v-model="modelForm.price" :min="0" style="width: 100%" />
        </el-form-item>
        <el-form-item label="图片URL">
          <el-input v-model="modelForm.cover" placeholder="https://..." />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="modelFormVisible = false">取消</el-button>
        <el-button type="primary" @click="saveModel">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import api from '../api'

const loading = ref(false)
const brands = ref([])
const models = ref([])
const currentBrand = ref(null)

const brandFormVisible = ref(false)
const brandForm = reactive({ id: null, name: '', logo: '', description: '' })

const modelsVisible = ref(false)
const modelFormVisible = ref(false)
const modelForm = reactive({ id: null, name: '', bikeType: '', price: 0, cover: '' })

const loadBrands = async () => {
  loading.value = true
  try {
    const res = await api.get('/brands')
    brands.value = res.data || []
  } catch {
    ElMessage.error('加载失败')
  } finally {
    loading.value = false
  }
}

const openBrandForm = (brand = null) => {
  if (brand) {
    Object.assign(brandForm, { id: brand.id, name: brand.name, logo: brand.logo, description: brand.description })
  } else {
    Object.assign(brandForm, { id: null, name: '', logo: '', description: '' })
  }
  brandFormVisible.value = true
}

const saveBrand = async () => {
  if (!brandForm.name.trim()) {
    ElMessage.warning('请输入品牌名称')
    return
  }
  try {
    if (brandForm.id) {
      await api.put(`/brands/${brandForm.id}`, brandForm)
    } else {
      await api.post('/brands', brandForm)
    }
    ElMessage.success('保存成功')
    brandFormVisible.value = false
    loadBrands()
  } catch {
    ElMessage.error('保存失败')
  }
}

const deleteBrand = async (brand) => {
  try {
    await ElMessageBox.confirm('确定要删除这个品牌吗？会同时删除所有车型！', '提示', { type: 'warning' })
    await api.delete(`/brands/${brand.id}`)
    ElMessage.success('删除成功')
    loadBrands()
  } catch (e) {
    if (e !== 'cancel') ElMessage.error('删除失败')
  }
}

const openModels = async (brand) => {
  currentBrand.value = brand
  modelsVisible.value = true
  try {
    const res = await api.get(`/brands/${brand.id}/models`)
    models.value = res.data || []
  } catch {
    ElMessage.error('加载车型失败')
  }
}

const openModelForm = (model = null) => {
  if (model) {
    Object.assign(modelForm, model)
  } else {
    Object.assign(modelForm, { id: null, name: '', bikeType: '', price: 0, cover: '' })
  }
  modelFormVisible.value = true
}

const saveModel = async () => {
  if (!modelForm.name.trim()) {
    ElMessage.warning('请输入车型名称')
    return
  }
  try {
    const data = { ...modelForm, brandId: currentBrand.value.id }
    if (modelForm.id) {
      await api.put(`/models/${modelForm.id}`, data)
    } else {
      await api.post('/models', data)
    }
    ElMessage.success('保存成功')
    modelFormVisible.value = false
    openModels(currentBrand.value)
  } catch {
    ElMessage.error('保存失败')
  }
}

const deleteModel = async (model) => {
  try {
    await ElMessageBox.confirm('确定要删除这个车型吗？', '提示', { type: 'warning' })
    await api.delete(`/models/${model.id}`)
    ElMessage.success('删除成功')
    openModels(currentBrand.value)
  } catch (e) {
    if (e !== 'cancel') ElMessage.error('删除失败')
  }
}

onMounted(() => {
  loadBrands()
})
</script>

<style scoped>
.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.brand-logo {
  width: 40px;
  height: 40px;
  border-radius: 50%;
  background: #f5f5f5;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 18px;
  font-weight: 600;
}

.models-header {
  margin-bottom: 16px;
}
</style>
