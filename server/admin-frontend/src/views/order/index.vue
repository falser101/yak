<template>
  <div class="orders-page">
    <!-- 统计卡片 -->
    <el-row :gutter="20" class="stats-row">
      <el-col :span="8">
        <el-card shadow="hover">
          <div class="stat-item">
            <div class="stat-icon blue">📦</div>
            <div class="stat-info">
              <div class="stat-value">{{ stats.totalOrders }}</div>
              <div class="stat-label">总订单数</div>
            </div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="8">
        <el-card shadow="hover">
          <div class="stat-item">
            <div class="stat-icon orange">⏳</div>
            <div class="stat-info">
              <div class="stat-value">{{ stats.pendingOrders }}</div>
              <div class="stat-label">待支付</div>
            </div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="8">
        <el-card shadow="hover">
          <div class="stat-item">
            <div class="stat-icon green">💰</div>
            <div class="stat-info">
              <div class="stat-value">¥{{ stats.totalRevenue.toFixed(0) }}</div>
              <div class="stat-label">总收入</div>
            </div>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <!-- 订单列表 -->
    <el-card class="list-card">
      <template #header>
        <div class="card-header">
          <span>📦 租车订单</span>
        </div>
      </template>

      <!-- 筛选 -->
      <div class="filter-row">
        <el-select v-model="statusFilter" placeholder="订单状态" style="width: 140px" clearable @change="loadOrders">
          <el-option label="待支付" :value="0" />
          <el-option label="已支付" :value="1" />
          <el-option label="已取消" :value="2" />
          <el-option label="已完成" :value="3" />
          <el-option label="已退款" :value="4" />
        </el-select>
      </div>

      <el-table :data="orders" v-loading="loading" stripe>
        <el-table-column prop="id" label="ID" width="60" />
        <el-table-column prop="orderNo" label="订单号" width="180" show-overflow-tooltip />
        <el-table-column label="用户" width="120">
          <template #default="{ row }">
            {{ row.nickname || '用户' + row.userId }}
          </template>
        </el-table-column>
        <el-table-column label="车辆" min-width="150">
          <template #default="{ row }">
            <div class="bike-cell">
              <el-avatar v-if="row.bikeCover" :src="row.bikeCover" :size="32" shape="square" />
              <span>{{ row.bikeName }}</span>
            </div>
          </template>
        </el-table-column>
        <el-table-column prop="package" label="套餐" width="80">
          <template #default="{ row }">
            {{ getPackageText(row.package) }}
          </template>
        </el-table-column>
        <el-table-column prop="quantity" label="数量" width="60" />
        <el-table-column prop="rentalDate" label="租车日期" width="120" />
        <el-table-column label="金额" width="100">
          <template #default="{ row }">
            <div>
              <div>租金: ¥{{ row.amount }}</div>
              <div style="color:#999;font-size:12px">押金: ¥{{ row.deposit }}</div>
            </div>
          </template>
        </el-table-column>
        <el-table-column prop="status" label="状态" width="90">
          <template #default="{ row }">
            <el-tag :type="getStatusType(row.status)" size="small">
              {{ getStatusText(row.status) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="contactPhone" label="联系电话" width="120" />
        <el-table-column label="操作" width="120" fixed="right">
          <template #default="{ row }">
            <el-button
              v-if="row.status === 0"
              link
              type="success"
              size="small"
              @click="confirmPayment(row)"
            >
              确认收款
            </el-button>
            <span v-else style="color:#999;font-size:12px">
              {{ row.payTime ? formatDate(row.payTime) : '-' }}
            </span>
          </template>
        </el-table-column>
      </el-table>

      <div class="pagination">
        <el-pagination
          v-model:current-page="page"
          :page-size="pageSize"
          :total="total"
          layout="prev, pager, next"
          @current-change="loadOrders"
        />
      </div>
    </el-card>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { getRentalOrders, confirmPayment, getRentalStats } from '../../api/order'

const loading = ref(false)
const orders = ref([])
const page = ref(1)
const pageSize = ref(10)
const total = ref(0)
const statusFilter = ref('')

const stats = reactive({
  totalOrders: 0,
  pendingOrders: 0,
  totalRevenue: 0
})

const getPackageText = (pkg) => {
  const map = { day: '日租', hour: '小时租', team: '团队租', distance: '里程租' }
  return map[pkg] || pkg
}

const getStatusType = (status) => {
  const types = ['warning', 'success', 'info', '', 'danger']
  return types[status] || 'info'
}

const getStatusText = (status) => {
  const texts = ['待支付', '已支付', '已取消', '已完成', '已退款']
  return texts[status] || '未知'
}

const formatDate = (dateStr) => {
  if (!dateStr) return '-'
  const d = new Date(dateStr)
  return `${d.getFullYear()}-${(d.getMonth() + 1).toString().padStart(2, '0')}-${d.getDate().toString().padStart(2, '0')}`
}

const loadStats = async () => {
  try {
    const res = await getRentalStats()
    if (res.data) {
      Object.assign(stats, res.data)
    }
  } catch {
    // ignore
  }
}

const loadOrders = async () => {
  loading.value = true
  try {
    const params = { page: page.value, pageSize: pageSize.value }
    if (statusFilter.value !== '') {
      params.status = statusFilter.value
    }
    const res = await getRentalOrders(params)
    orders.value = res.data || []
    total.value = res.total || 0
  } catch {
    ElMessage.error('加载失败')
  } finally {
    loading.value = false
  }
}

const confirmPayment = async (row) => {
  try {
    await ElMessageBox.confirm(
      `确认收到用户「${row.nickname || '用户' + row.userId}」的付款？\n订单号: ${row.orderNo}\n金额: ¥${row.amount + row.deposit}`,
      '确认收款',
      { type: 'success' }
    )
    await confirmPayment(row.id)
    ElMessage.success('确认成功')
    loadOrders()
    loadStats()
  } catch (e) {
    if (e !== 'cancel') ElMessage.error('操作失败')
  }
}

onMounted(() => {
  loadStats()
  loadOrders()
})
</script>

<style scoped>
.stats-row {
  margin-bottom: 20px;
}

.stat-item {
  display: flex;
  align-items: center;
  gap: 16px;
}

.stat-icon {
  width: 56px;
  height: 56px;
  border-radius: 12px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 28px;
}

.stat-icon.green { background: #e6f7e6; }
.stat-icon.blue { background: #e3f2fd; }
.stat-icon.orange { background: #fff3e0; }

.stat-value {
  font-size: 28px;
  font-weight: 600;
  color: #333;
}

.stat-label {
  font-size: 14px;
  color: #999;
}

.list-card {
  margin-bottom: 20px;
}

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

.bike-cell {
  display: flex;
  align-items: center;
  gap: 8px;
}

.pagination {
  margin-top: 20px;
  display: flex;
  justify-content: flex-end;
}
</style>
