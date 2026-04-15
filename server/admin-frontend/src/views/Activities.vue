<template>
  <div class="activities-page">
    <!-- 统计卡片 -->
    <el-row :gutter="20" class="stats-row">
      <el-col :span="8">
        <el-card shadow="hover">
          <div class="stat-item">
            <div class="stat-icon green">📋</div>
            <div class="stat-info">
              <div class="stat-value">{{ stats.totalActivities }}</div>
              <div class="stat-label">总活动数</div>
            </div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="8">
        <el-card shadow="hover">
          <div class="stat-item">
            <div class="stat-icon orange">📢</div>
            <div class="stat-info">
              <div class="stat-value">{{ stats.activeCount }}</div>
              <div class="stat-label">报名中</div>
            </div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="8">
        <el-card shadow="hover">
          <div class="stat-item">
            <div class="stat-icon blue">👥</div>
            <div class="stat-info">
              <div class="stat-value">{{ stats.totalSignups }}</div>
              <div class="stat-label">总报名人数</div>
            </div>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <!-- 活动列表 -->
    <el-card class="list-card">
      <template #header>
        <div class="card-header">
          <span>📝 活动列表</span>
          <el-button type="primary" @click="$router.push('/activities/create')">
            + 发布新活动
          </el-button>
        </div>
      </template>

      <el-table :data="activities" v-loading="loading" stripe>
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
        <el-table-column prop="price" label="费用" width="80">
          <template #default="{ row }">
            ¥{{ row.price }}
          </template>
        </el-table-column>
        <el-table-column prop="status" label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="getStatusType(row.status)">
              {{ getStatusText(row.status) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="150" fixed="right">
          <template #default="{ row }">
            <el-button link type="primary" size="small" @click="showSignups(row)">
              报名
            </el-button>
            <el-button link type="primary" size="small" @click="editActivity(row)">
              编辑
            </el-button>
            <el-button link type="danger" size="small" @click="deleteActivity(row)">
              删除
            </el-button>
          </template>
        </el-table-column>
      </el-table>

      <div class="pagination">
        <el-pagination
          v-model:current-page="page"
          :page-size="pageSize"
          :total="total"
          layout="prev, pager, next"
          @current-change="loadData"
        />
      </div>
    </el-card>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import api from '../api'

const router = useRouter()
const loading = ref(false)
const activities = ref([])
const page = ref(1)
const pageSize = ref(10)
const total = ref(0)

const stats = reactive({
  totalActivities: 0,
  activeCount: 0,
  totalSignups: 0
})

const formatDate = (dateStr) => {
  if (!dateStr) return '-'
  const d = new Date(dateStr)
  return `${d.getMonth() + 1}-${d.getDate()} ${d.getHours().toString().padStart(2, '0')}:${d.getMinutes().toString().padStart(2, '0')}`
}

const getStatusType = (status) => {
  const types = ['', 'warning', 'info']
  return types[status] || 'info'
}

const getStatusText = (status) => {
  const texts = ['', '已满', '已结束']
  return texts[status] || '报名中'
}

const loadData = async () => {
  loading.value = true
  try {
    const res = await api.get('/activities', { params: { page: page.value, pageSize: pageSize.value } })
    activities.value = res.data || []
    total.value = res.total || 0

    stats.totalActivities = res.total || 0
    stats.activeCount = (res.data || []).filter(a => a.status === 0).length
    stats.totalSignups = (res.data || []).reduce((sum, a) => sum + (a.signupCount || 0), 0)
  } catch {
    ElMessage.error('加载失败')
  } finally {
    loading.value = false
  }
}

const showSignups = (row) => {
  router.push(`/activities/${row.id}/signups`)
}

const editActivity = (row) => {
  router.push(`/activities/${row.id}/edit`)
}

const deleteActivity = async (row) => {
  try {
    await ElMessageBox.confirm('确定要删除这个活动吗？删除后无法恢复！', '提示', {
      type: 'warning'
    })
    await api.delete(`/activities/${row.id}`)
    ElMessage.success('删除成功')
    loadData()
  } catch (e) {
    if (e !== 'cancel') ElMessage.error('删除失败')
  }
}

onMounted(() => {
  loadData()
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
.stat-icon.orange { background: #fff3e0; }
.stat-icon.blue { background: #e3f2fd; }

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

.pagination {
  margin-top: 20px;
  display: flex;
  justify-content: flex-end;
}
</style>
