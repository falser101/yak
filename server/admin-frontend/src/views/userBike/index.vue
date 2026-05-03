<template>
  <div class="bikes-page">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>🚴 用户车辆列表</span>
          <span class="total-count">共 {{ bikes.length }} 辆</span>
        </div>
      </template>

      <el-table :data="bikes" v-loading="loading" stripe>
        <el-table-column prop="id" label="ID" width="60" />
        <el-table-column label="自行车" min-width="200">
          <template #default="{ row }">
            <div class="bike-info">
              <div class="bike-name">{{ row.name }}</div>
              <el-tag type="success" size="small">{{ row.bikeType || '未分类' }}</el-tag>
            </div>
          </template>
        </el-table-column>
        <el-table-column prop="userName" label="用户" width="120" />
        <el-table-column label="品牌/车型" min-width="150">
          <template #default="{ row }">
            <div>{{ row.brandName || '-' }}</div>
            <div class="model-name">{{ row.modelName || '' }}</div>
          </template>
        </el-table-column>
        <el-table-column prop="purchaseDate" label="购买日期" width="120" />
        <el-table-column prop="cost" label="价格" width="100">
          <template #default="{ row }">
            <span class="price">¥{{ row.cost || 0 }}</span>
          </template>
        </el-table-column>
      </el-table>
    </el-card>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import api from '../../api'

const loading = ref(false)
const bikes = ref([])

const loadBikes = async () => {
  loading.value = true
  try {
    const res = await api.get('/bikes')
    bikes.value = res.data || []
  } catch {
    ElMessage.error('加载失败')
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  loadBikes()
})
</script>

<style scoped>
.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.total-count {
  color: #999;
  font-size: 14px;
}

.bike-info {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.bike-name {
  font-weight: 500;
}

.model-name {
  color: #999;
  font-size: 12px;
}

.price {
  color: #ff4d4f;
  font-weight: 600;
}
</style>
