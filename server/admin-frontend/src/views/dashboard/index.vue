<template>
  <div class="dashboard-page">
    <!-- 统计卡片 -->
    <el-row :gutter="20" class="stats-row">
      <el-col :span="6">
        <el-card shadow="hover">
          <div class="stat-item">
            <div class="stat-icon green">🏔️</div>
            <div class="stat-info">
              <div class="stat-value">{{ stats.totalActivities }}</div>
              <div class="stat-label">总活动数</div>
            </div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card shadow="hover">
          <div class="stat-item">
            <div class="stat-icon blue">📢</div>
            <div class="stat-info">
              <div class="stat-value">{{ stats.activeActivities }}</div>
              <div class="stat-label">正在进行</div>
            </div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card shadow="hover">
          <div class="stat-item">
            <div class="stat-icon orange">👥</div>
            <div class="stat-info">
              <div class="stat-value">{{ stats.totalSignups }}</div>
              <div class="stat-label">总报名人数</div>
            </div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card shadow="hover">
          <div class="stat-item">
            <div class="stat-icon purple">💰</div>
            <div class="stat-info">
              <div class="stat-value">¥{{ stats.totalRevenue.toFixed(0) }}</div>
              <div class="stat-label">总收入</div>
            </div>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <el-row :gutter="20" class="stats-row">
      <el-col :span="6">
        <el-card shadow="hover">
          <div class="stat-item">
            <div class="stat-icon cyan">👤</div>
            <div class="stat-info">
              <div class="stat-value">{{ stats.totalUsers }}</div>
              <div class="stat-label">总用户数</div>
            </div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card shadow="hover">
          <div class="stat-item">
            <div class="stat-icon teal">✅</div>
            <div class="stat-info">
              <div class="stat-value">{{ stats.activeUsers }}</div>
              <div class="stat-label">活跃用户</div>
            </div>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <!-- 最近活动 -->
    <el-card class="list-card">
      <template #header>
        <div class="card-header">
          <span>📅 最近活动</span>
          <el-button type="primary" size="small" @click="$router.push('/activities')">
            查看全部
          </el-button>
        </div>
      </template>

      <el-table :data="recentActivities" v-loading="loading" stripe>
        <el-table-column prop="id" label="ID" width="60" />
        <el-table-column prop="title" label="活动标题" min-width="200" show-overflow-tooltip />
        <el-table-column prop="date" label="时间" width="120">
          <template #default="{ row }">
            {{ formatDate(row.date) }}
          </template>
        </el-table-column>
        <el-table-column prop="location" label="地点" width="120" show-overflow-tooltip />
        <el-table-column label="人数" width="100">
          <template #default="{ row }">
            {{ row.signupCount }}/{{ row.maxParticipants }}
          </template>
        </el-table-column>
        <el-table-column prop="status" label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="getStatusType(row.status)">
              {{ getStatusText(row.status) }}
            </el-tag>
          </template>
        </el-table-column>
      </el-table>
    </el-card>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { getStats } from '../../api/stats'
import { getActivities } from '../../api/activity'

const loading = ref(false)
const recentActivities = ref([])

const stats = reactive({
  totalActivities: 0,
  activeActivities: 0,
  totalSignups: 0,
  pendingSignups: 0,
  totalUsers: 0,
  activeUsers: 0,
  totalRevenue: 0
})

const formatDate = (dateStr) => {
  if (!dateStr) return '-'
  const d = new Date(dateStr)
  return `${d.getMonth() + 1}-${d.getDate()} ${d.getHours().toString().padStart(2, '0')}:${d.getMinutes().toString().padStart(2, '0')}`
}

const getStatusType = (status) => {
  const types = ['info', 'warning', 'success', 'danger']
  return types[status] || 'info'
}

const getStatusText = (status) => {
  const texts = ['草稿', '报名中', '进行中', '已结束']
  return texts[status] || '未知'
}

const loadStats = async () => {
  try {
    const res = await getStats()
    if (res.data) {
      Object.assign(stats, res.data)
    }
  } catch {
    ElMessage.error('加载统计数据失败')
  }
}

const loadRecentActivities = async () => {
  loading.value = true
  try {
    const res = await getActivities({ page: 1, pageSize: 5 })
    recentActivities.value = res.data || []
  } catch {
    ElMessage.error('加载活动列表失败')
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  loadStats()
  loadRecentActivities()
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
.stat-icon.purple { background: #f3e8ff; }
.stat-icon.cyan { background: #e0f7fa; }
.stat-icon.teal { background: #e6fffa; }

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
</style>
