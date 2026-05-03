<template>
  <div class="bikes-page">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>🚴 租车车辆管理</span>
          <el-button type="primary" @click="openDialog('create')">
            + 添加车辆
          </el-button>
        </div>
      </template>

      <!-- 筛选 -->
      <div class="filter-row">
        <el-select v-model="bikeTypeFilter" placeholder="车型分类" style="width: 140px" clearable @change="loadBikes">
          <el-option label="公路车" value="公路车" />
          <el-option label="山地车" value="山地车" />
          <el-option label="平把公路" value="平把公路" />
          <el-option label="电动车" value="电动车" />
        </el-select>
      </div>

      <el-table :data="bikes" v-loading="loading" stripe>
        <el-table-column prop="id" label="ID" width="60" />
        <el-table-column prop="name" label="车辆名称" min-width="160" />
        <el-table-column prop="bikeType" label="车型" width="100">
          <template #default="{ row }">
            <el-tag size="small" :type="getTypeTagType(row.bikeType)">{{ row.bikeType }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="brandName" label="品牌" width="120" />
        <el-table-column label="标签" width="80">
          <template #default="{ row }">
            <el-tag v-if="row.tag" size="small" :type="getTagType(row.tag)">{{ row.tag }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="价格体系" min-width="180">
          <template #default="{ row }">
            <div class="price-cell">
              <span>日租 <b>¥{{ row.priceDay }}</b></span>
              <span>时租 <b>¥{{ row.priceHour }}</b></span>
              <span>团队 <b>¥{{ row.priceTeam }}</b></span>
            </div>
          </template>
        </el-table-column>
        <el-table-column prop="deposit" label="押金" width="80">
          <template #default="{ row }">
            <span class="price">¥{{ row.deposit }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="status" label="状态" width="80">
          <template #default="{ row }">
            <el-tag :type="row.status === 1 ? 'success' : 'danger'" size="small">
              {{ row.status === 1 ? '上架' : '下架' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="150" fixed="right">
          <template #default="{ row }">
            <el-button link type="primary" size="small" @click="openDialog('edit', row)">编辑</el-button>
            <el-button link type="danger" size="small" @click="deleteBike(row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <!-- 创建/编辑对话框 -->
    <el-dialog v-model="dialogVisible" :title="dialogTitle" width="600px">
      <el-form :model="form" label-width="100px" ref="formRef">
        <el-form-item label="车辆名称" prop="name">
          <el-input v-model="form.name" placeholder="如: TCR Advanced SL 公路车" />
        </el-form-item>
        <el-form-item label="品牌">
          <el-select v-model="form.brandId" placeholder="选择品牌" style="width: 100%">
            <el-option v-for="b in brands" :key="b.id" :label="b.name" :value="b.id" />
          </el-select>
        </el-form-item>
        <el-form-item label="车型分类">
          <el-select v-model="form.bikeType" placeholder="选择分类" style="width: 100%">
            <el-option label="公路车" value="公路车" />
            <el-option label="山地车" value="山地车" />
            <el-option label="平把公路" value="平把公路" />
            <el-option label="电动车" value="电动车" />
          </el-select>
        </el-form-item>
        <el-form-item label="标签">
          <el-select v-model="form.tag" placeholder="选择标签" style="width: 100%">
            <el-option label="热门" value="热门" />
            <el-option label="推荐" value="推荐" />
            <el-option label="新品" value="新品" />
            <el-option label="无" value="" />
          </el-select>
        </el-form-item>
        <el-row :gutter="20">
          <el-col :span="12">
            <el-form-item label="日租价格">
              <el-input-number v-model="form.priceDay" :min="0" :precision="2" style="width: 100%" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="时租价格">
              <el-input-number v-model="form.priceHour" :min="0" :precision="2" style="width: 100%" />
            </el-form-item>
          </el-col>
        </el-row>
        <el-row :gutter="20">
          <el-col :span="12">
            <el-form-item label="团队租价格">
              <el-input-number v-model="form.priceTeam" :min="0" :precision="2" style="width: 100%" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="押金">
              <el-input-number v-model="form.deposit" :min="0" :precision="2" style="width: 100%" />
            </el-form-item>
          </el-col>
        </el-row>
        <el-form-item label="规格参数">
          <el-input v-model="specsText" type="textarea" rows="2" placeholder="如: 车架M码,变速器Shimano 105" />
        </el-form-item>
        <el-form-item label="租车说明">
          <el-input v-model="form.notes" type="textarea" rows="2" placeholder="租车注意事项、取车还车规则" />
        </el-form-item>
        <el-form-item label="状态">
          <el-radio-group v-model="form.status">
            <el-radio :label="1">上架</el-radio>
            <el-radio :label="2">下架</el-radio>
          </el-radio-group>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" @click="saveBike" :loading="saving">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { getRentalBikes, createRentalBike, updateRentalBike, deleteRentalBike } from '../../api/bike'
import { getBrands } from '../../api/brand'

const loading = ref(false)
const saving = ref(false)
const bikes = ref([])
const brands = ref([])
const bikeTypeFilter = ref('')
const dialogVisible = ref(false)
const dialogTitle = ref('')
const editId = ref(null)
const formRef = ref()

const form = reactive({
  name: '',
  brandId: null,
  bikeType: '',
  tag: '',
  priceDay: 0,
  priceHour: 0,
  priceTeam: 0,
  deposit: 0,
  notes: '',
  status: 1
})

const specsText = ref('')

const getTypeTagType = (type) => {
  const map = { '公路车': '', '山地车': 'warning', '平把公路': 'success', '电动车': 'info' }
  return map[type] || ''
}

const getTagType = (tag) => {
  const map = { '热门': 'danger', '推荐': 'warning', '新品': 'success' }
  return map[tag] || ''
}

const loadBikes = async () => {
  loading.value = true
  try {
    const res = await getRentalBikes()
    let data = res.data || []
    if (bikeTypeFilter.value) {
      data = data.filter(b => b.bikeType === bikeTypeFilter.value)
    }
    bikes.value = data
  } catch {
    ElMessage.error('加载失败')
  } finally {
    loading.value = false
  }
}

const loadBrands = async () => {
  try {
    const res = await getBrands()
    brands.value = res.data || []
  } catch {}
}

const openDialog = (mode, row) => {
  if (mode === 'create') {
    dialogTitle.value = '添加车辆'
    editId.value = null
    Object.assign(form, {
      name: '', brandId: null, bikeType: '', tag: '',
      priceDay: 0, priceHour: 0, priceTeam: 0, deposit: 0,
      notes: '', status: 1
    })
    specsText.value = ''
  } else {
    dialogTitle.value = '编辑车辆'
    editId.value = row.id
    form.name = row.name
    form.brandId = row.brandId
    form.bikeType = row.bikeType
    form.tag = row.tag
    form.priceDay = row.priceDay
    form.priceHour = row.priceHour
    form.priceTeam = row.priceTeam
    form.deposit = row.deposit
    form.notes = row.notes
    form.status = row.status
    // specs to text
    const specs = row.specs || {}
    const parts = []
    if (specs.frame) parts.push(`车架${specs.frame}`)
    if (specs.derailleur) parts.push(`变速${specs.derailleur}`)
    if (specs.brake) parts.push(`刹车${specs.brake}`)
    if (specs.wheel) parts.push(`轮径${specs.wheel}`)
    if (specs.height) parts.push(`身高${specs.height}`)
    specsText.value = parts.join(', ')
  }
  dialogVisible.value = true
}

const saveBike = async () => {
  if (!form.name) {
    ElMessage.warning('请填写车辆名称')
    return
  }
  saving.value = true
  try {
    // specs text to object
    const specs = {}
    const parts = specsText.value.split(',').map(s => s.trim()).filter(Boolean)
    parts.forEach(p => {
      if (p.startsWith('车架')) specs.frame = p.replace('车架', '')
      else if (p.startsWith('变速')) specs.derailleur = p.replace('变速', '')
      else if (p.startsWith('刹车')) specs.brake = p.replace('刹车', '')
      else if (p.startsWith('轮径')) specs.wheel = p.replace('轮径', '')
      else if (p.startsWith('身高')) specs.height = p.replace('身高', '')
    })

    const data = { ...form, specs }
    if (editId.value) {
      await updateRentalBike(editId.value, data)
      ElMessage.success('更新成功')
    } else {
      await createRentalBike(data)
      ElMessage.success('添加成功')
    }
    dialogVisible.value = false
    loadBikes()
  } catch {
    ElMessage.error('保存失败')
  } finally {
    saving.value = false
  }
}

const deleteBike = async (row) => {
  try {
    await ElMessageBox.confirm(`确定删除「${row.name}」吗？`, '提示', { type: 'warning' })
    await deleteRentalBike(row.id)
    ElMessage.success('删除成功')
    loadBikes()
  } catch (e) {
    if (e !== 'cancel') ElMessage.error('删除失败')
  }
}

onMounted(() => {
  loadBikes()
  loadBrands()
})
</script>

<style scoped>
.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.filter-row {
  display: flex;
  gap: 12px;
  margin-bottom: 16px;
}

.bike-info {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.bike-name {
  font-weight: 500;
}

.price-cell {
  display: flex;
  flex-direction: column;
  gap: 2px;
  font-size: 13px;
  color: #666;
}

.price-cell b {
  color: #ff4d4f;
}

.price {
  color: #ff4d4f;
  font-weight: 600;
}
</style>
