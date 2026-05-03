<template>
  <div class="users-page">
    <el-card class="list-card">
      <template #header>
        <div class="card-header">
          <span>👥 用户列表</span>
        </div>
      </template>

      <!-- 筛选区 -->
      <div class="filter-row">
        <el-input
          v-model="keyword"
          placeholder="搜索昵称/手机号"
          style="width: 200px"
          clearable
          @clear="loadData"
          @keyup.enter="loadData"
        >
          <template #prefix>
            <el-icon><Search /></el-icon>
          </template>
        </el-input>
        <el-select v-model="membershipLevel" placeholder="会员等级" style="width: 140px" clearable @change="loadData">
          <el-option label="普通会员" :value="0" />
          <el-option label="银卡会员" :value="1" />
          <el-option label="金卡会员" :value="2" />
          <el-option label="钻石会员" :value="3" />
        </el-select>
      </div>

      <el-table :data="users" v-loading="loading" stripe>
        <el-table-column prop="id" label="ID" width="60" />
        <el-table-column label="用户" width="180">
          <template #default="{ row }">
            <div class="user-cell">
              <el-avatar :src="row.avatar" :size="32">{{ row.nickname?.charAt(0) }}</el-avatar>
              <span>{{ row.nickname || '微信用户' }}</span>
            </div>
          </template>
        </el-table-column>
        <el-table-column prop="phone" label="手机号" width="120" />
        <el-table-column prop="membershipLevel" label="会员等级" width="100">
          <template #default="{ row }">
            <el-tag :type="getMembershipType(row.membershipLevel)" size="small">
              {{ getMembershipText(row.membershipLevel) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="totalRides" label="累计里程" width="100">
          <template #default="{ row }">
            {{ row.totalRides || 0 }} km
          </template>
        </el-table-column>
        <el-table-column prop="signupCount" label="报名次数" width="100" />
        <el-table-column prop="status" label="状态" width="80">
          <template #default="{ row }">
            <el-tag :type="row.status === 1 ? 'success' : 'danger'" size="small">
              {{ row.status === 1 ? '正常' : '禁用' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="createdAt" label="注册时间" width="120">
          <template #default="{ row }">
            {{ formatDate(row.createdAt) }}
          </template>
        </el-table-column>
        <el-table-column label="操作" width="150" fixed="right">
          <template #default="{ row }">
            <el-button link type="primary" size="small" @click="editUser(row)">
              编辑
            </el-button>
            <el-button
              link
              :type="row.status === 1 ? 'danger' : 'success'"
              size="small"
              @click="toggleStatus(row)"
            >
              {{ row.status === 1 ? '禁用' : '启用' }}
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

    <!-- 编辑用户对话框 -->
    <el-dialog v-model="dialogVisible" title="编辑用户" width="500px">
      <el-form :model="editForm" label-width="100px">
        <el-form-item label="昵称">
          <el-input v-model="editForm.nickname" />
        </el-form-item>
        <el-form-item label="会员等级">
          <el-select v-model="editForm.membershipLevel" style="width: 100%">
            <el-option label="普通会员" :value="0" />
            <el-option label="银卡会员" :value="1" />
            <el-option label="金卡会员" :value="2" />
            <el-option label="钻石会员" :value="3" />
          </el-select>
        </el-form-item>
        <el-form-item label="状态">
          <el-radio-group v-model="editForm.status">
            <el-radio :label="1">正常</el-radio>
            <el-radio :label="2">禁用</el-radio>
          </el-radio-group>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" @click="saveUser" :loading="saving">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Search } from '@element-plus/icons-vue'
import { getUsers, updateUser, setUserStatus } from '../../api/user'

const loading = ref(false)
const saving = ref(false)
const users = ref([])
const page = ref(1)
const pageSize = ref(10)
const total = ref(0)
const keyword = ref('')
const membershipLevel = ref('')
const dialogVisible = ref(false)

const editForm = reactive({
  id: null,
  nickname: '',
  membershipLevel: 0,
  status: 1
})

const formatDate = (dateStr) => {
  if (!dateStr) return '-'
  const d = new Date(dateStr)
  return `${d.getFullYear()}-${(d.getMonth() + 1).toString().padStart(2, '0')}-${d.getDate().toString().padStart(2, '0')}`
}

const getMembershipType = (level) => {
  const types = ['info', 'silver', 'warning', 'danger']
  return types[level] || 'info'
}

const getMembershipText = (level) => {
  const texts = ['普通会员', '银卡会员', '金卡会员', '钻石会员']
  return texts[level] || '普通会员'
}

const loadData = async () => {
  loading.value = true
  try {
    const params = {
      page: page.value,
      pageSize: pageSize.value
    }
    if (keyword.value) params.keyword = keyword.value
    if (membershipLevel.value !== '') params.membershipLevel = membershipLevel.value

    const res = await getUsers(params)
    users.value = res.data || []
    total.value = res.total || 0
  } catch {
    ElMessage.error('加载用户列表失败')
  } finally {
    loading.value = false
  }
}

const editUser = (row) => {
  editForm.id = row.id
  editForm.nickname = row.nickname
  editForm.membershipLevel = row.membershipLevel
  editForm.status = row.status
  dialogVisible.value = true
}

const saveUser = async () => {
  saving.value = true
  try {
    await updateUser(editForm.id, {
      nickname: editForm.nickname,
      membershipLevel: editForm.membershipLevel,
      status: editForm.status
    })
    ElMessage.success('保存成功')
    dialogVisible.value = false
    loadData()
  } catch {
    ElMessage.error('保存失败')
  } finally {
    saving.value = false
  }
}

const toggleStatus = async (row) => {
  const action = row.status === 1 ? '禁用' : '启用'
  try {
    await ElMessageBox.confirm(`确定要${action}用户「${row.nickname || '微信用户'}」吗？`, '提示', {
      type: 'warning'
    })
    await setUserStatus(row.id, row.status === 1 ? 2 : 1)
    ElMessage.success(`${action}成功`)
    loadData()
  } catch (e) {
    if (e !== 'cancel') ElMessage.error(`${action}失败`)
  }
}

onMounted(() => {
  loadData()
})
</script>

<style scoped>
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

.user-cell {
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
