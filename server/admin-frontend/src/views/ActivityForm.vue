<template>
  <div class="activity-form-page">
    <el-card>
      <template #header>
        <span>{{ isEdit ? '✏️ 编辑活动' : '🎉 发布新活动' }}</span>
      </template>

      <el-form :model="form" :rules="rules" ref="formRef" label-width="100px">
        <el-form-item label="活动标题" prop="title">
          <el-input v-model="form.title" placeholder="请输入活动标题" />
        </el-form-item>

        <el-form-item label="活动封面">
          <el-input v-model="form.cover" placeholder="封面图片URL" />
        </el-form-item>

        <el-form-item label="活动时间" prop="date">
          <el-date-picker
            v-model="form.date"
            type="datetime"
            placeholder="选择活动时间"
            style="width: 100%"
          />
        </el-form-item>

        <el-form-item label="活动地点" prop="location">
          <el-input v-model="form.location" placeholder="请输入活动地点" />
        </el-form-item>

        <el-row :gutter="20">
          <el-col :span="12">
            <el-form-item label="人数上限" prop="maxParticipants">
              <el-input-number v-model="form.maxParticipants" :min="1" style="width: 100%" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="活动费用" prop="price">
              <el-input-number v-model="form.price" :min="0" :precision="2" style="width: 100%" />
            </el-form-item>
          </el-col>
        </el-row>

        <el-form-item label="活动详情" prop="description">
          <el-input v-model="form.description" type="textarea" rows="4" placeholder="请描述活动详情、行程安排、注意事项等" />
        </el-form-item>

        <el-form-item>
          <el-button @click="$router.back()">取消</el-button>
          <el-button type="primary" :loading="saving" @click="handleSubmit">
            {{ isEdit ? '保存' : '发布' }}
          </el-button>
        </el-form-item>
      </el-form>
    </el-card>
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { ElMessage } from 'element-plus'
import api from '../api'

const router = useRouter()
const route = useRoute()
const formRef = ref()
const saving = ref(false)

const isEdit = computed(() => !!route.params.id)

const form = reactive({
  title: '',
  cover: '',
  date: '',
  location: '',
  maxParticipants: 50,
  price: 0,
  description: ''
})

const rules = {
  title: [{ required: true, message: '请输入活动标题', trigger: 'blur' }],
  date: [{ required: true, message: '请选择活动时间', trigger: 'change' }],
  location: [{ required: true, message: '请输入活动地点', trigger: 'blur' }],
  maxParticipants: [{ required: true, message: '请输入人数上限', trigger: 'blur' }],
  price: [{ required: true, message: '请输入活动费用', trigger: 'blur' }]
}

const loadData = async () => {
  if (!isEdit.value) return
  try {
    const res = await api.get(`/activities/${route.params.id}`)
    const data = res.data || {}
    form.title = data.title || ''
    form.cover = data.cover || ''
    form.date = data.date ? new Date(data.date) : ''
    form.location = data.location || ''
    form.maxParticipants = data.maxParticipants || 50
    form.price = data.price || 0
    form.description = data.description || ''
  } catch {
    ElMessage.error('加载活动失败')
  }
}

const handleSubmit = async () => {
  const valid = await formRef.value.validate().catch(() => false)
  if (!valid) return

  saving.value = true
  try {
    const data = {
      ...form,
      date: form.date ? new Date(form.date).toISOString() : ''
    }
    if (isEdit.value) {
      await api.put(`/activities/${route.params.id}`, data)
      ElMessage.success('保存成功')
    } else {
      await api.post('/activities', data)
      ElMessage.success('发布成功')
    }
    router.push('/activities')
  } catch {
    ElMessage.error(isEdit.value ? '保存失败' : '发布失败')
  } finally {
    saving.value = false
  }
}

onMounted(() => {
  loadData()
})
</script>

<style scoped>
.activity-form-page {
  max-width: 800px;
}
</style>
