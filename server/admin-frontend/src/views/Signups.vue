<template>
  <div class="signups-page">
    <el-card>
      <template #header>
        <div class="card-header">
          <div class="activity-info">
            <span class="activity-title">{{ activity.title }}</span>
            <span class="activity-meta">
              📅 {{ activity.date }} &nbsp;&nbsp;
              📍 {{ activity.location }} &nbsp;&nbsp;
              👥 {{ signups.length }} 人已报名
            </span>
          </div>
          <el-button @click="$router.back()">返回活动列表</el-button>
        </div>
      </template>

      <el-table :data="signups" v-loading="loading" stripe>
        <el-table-column prop="id" label="ID" width="60" />
        <el-table-column prop="name" label="姓名" width="120" />
        <el-table-column prop="phone" label="电话" width="130" />
        <el-table-column prop="emergencyContact" label="紧急联系人" width="120" />
        <el-table-column prop="emergencyPhone" label="紧急联系电话" width="130" />
        <el-table-column prop="remark" label="备注" min-width="150" show-overflow-tooltip />
        <el-table-column prop="createdAt" label="报名时间" width="160">
          <template #default="{ row }">
            {{ formatDate(row.createdAt) }}
          </template>
        </el-table-column>
        <el-table-column prop="status" label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="row.status === 1 ? 'success' : 'info'">
              {{ row.status === 1 ? '已报名' : '已取消' }}
            </el-tag>
          </template>
        </el-table-column>
      </el-table>
    </el-card>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { ElMessage } from 'element-plus'
import api from '../api'

const route = useRoute()
const loading = ref(false)
const activity = ref({ title: '', date: '', location: '' })
const signups = ref([])

const formatDate = (dateStr) => {
  if (!dateStr) return '-'
  const d = new Date(dateStr)
  return `${d.getFullYear()}-${(d.getMonth() + 1).toString().padStart(2, '0')}-${d.getDate().toString().padStart(2, '0')} ${d.getHours().toString().padStart(2, '0')}:${d.getMinutes().toString().padStart(2, '0')}`
}

const loadData = async () => {
  const activityId = route.params.id
  loading.value = true
  try {
    const [activityRes, signupsRes] = await Promise.all([
      api.get(`/activities/${activityId}`),
      api.get(`/activities/${activityId}/signups`)
    ])
    activity.value = activityRes.data || {}
    signups.value = signupsRes.data || []
  } catch {
    ElMessage.error('加载失败')
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  loadData()
})
</script>

<style scoped>
.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.activity-info {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.activity-title {
  font-size: 16px;
  font-weight: 600;
}

.activity-meta {
  font-size: 14px;
  color: #666;
}
</style>
