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

        <el-form-item label="活动分类" prop="category">
          <el-select v-model="form.category" placeholder="选择活动分类" style="width: 100%">
            <el-option label="赛事" value="activity" />
            <el-option label="公益" value="charity" />
            <el-option label="俱乐部" value="club" />
          </el-select>
        </el-form-item>

        <el-form-item label="活动封面">
          <div class="cover-upload">
            <el-upload
              class="cover-uploader"
              action="/api/upload"
              :headers="{ Authorization: 'Bearer ' + token }"
              :show-file-list="false"
              :on-success="handleUploadSuccess"
              :on-error="handleUploadError"
              :before-upload="beforeUpload"
            >
              <img v-if="form.cover" :src="form.cover" class="cover-preview" />
              <el-icon v-else class="cover-uploader-icon"><Plus /></el-icon>
            </el-upload>
            <div class="cover-tip">支持 JPG、PNG 格式，建议尺寸 750x400</div>
          </div>
        </el-form-item>

        <el-row :gutter="20">
          <el-col :span="12">
            <el-form-item label="活动时间" prop="date">
              <el-date-picker
                v-model="form.date"
                type="datetime"
                placeholder="选择活动时间"
                style="width: 100%"
              />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="报名截止">
              <el-date-picker
                v-model="form.signupEndTime"
                type="datetime"
                placeholder="选择报名截止时间"
                style="width: 100%"
              />
            </el-form-item>
          </el-col>
        </el-row>

        <el-form-item label="活动地点" prop="location">
          <el-input v-model="form.location" placeholder="请输入活动地点" />
        </el-form-item>

        <el-row :gutter="20">
          <el-col :span="8">
            <el-form-item label="人数上限" prop="maxParticipants">
              <el-input-number v-model="form.maxParticipants" :min="1" style="width: 100%" />
            </el-form-item>
          </el-col>
          <el-col :span="8">
            <el-form-item label="活动费用" prop="price">
              <el-input-number v-model="form.price" :min="0" :precision="2" style="width: 100%" />
            </el-form-item>
          </el-col>
          <el-col :span="8">
            <el-form-item label="活动状态">
              <el-select v-model="form.status" style="width: 100%">
                <el-option label="草稿" :value="0" />
                <el-option label="报名中" :value="1" />
                <el-option label="进行中" :value="2" />
                <el-option label="已结束" :value="3" />
              </el-select>
            </el-form-item>
          </el-col>
        </el-row>

        <el-form-item label="活动规程">
          <el-input v-model="form.rules" type="textarea" rows="3" placeholder="请输入活动规程和规则" />
        </el-form-item>

        <el-form-item label="路线信息">
          <el-input v-model="form.route" type="textarea" rows="2" placeholder="请输入骑行路线信息" />
        </el-form-item>

        <el-form-item label="奖项设置">
          <el-input v-model="form.awards" type="textarea" rows="2" placeholder="请输入奖项设置" />
        </el-form-item>

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
import { Plus } from '@element-plus/icons-vue'
import { getActivity, createActivity, updateActivity } from '../../api/activity'

const router = useRouter()
const route = useRoute()
const formRef = ref()
const saving = ref(false)
const token = localStorage.getItem('admin_token') || ''

const isEdit = computed(() => !!route.params.id)

const form = reactive({
  title: '',
  category: 'activity',
  cover: '',
  date: '',
  signupEndTime: '',
  location: '',
  maxParticipants: 50,
  price: 0,
  status: 0,
  rules: '',
  route: '',
  awards: '',
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
    const res = await getActivity(route.params.id)
    const data = res.data || {}
    form.title = data.title || ''
    form.category = data.category || 'activity'
    form.cover = data.cover || ''
    form.date = data.date ? new Date(data.date) : ''
    form.signupEndTime = data.signupEndTime ? new Date(data.signupEndTime) : ''
    form.location = data.location || ''
    form.maxParticipants = data.maxParticipants || 50
    form.price = data.price || 0
    form.status = data.status || 0
    form.rules = data.rules || ''
    form.route = data.route || ''
    form.awards = data.awards || ''
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
      await updateActivity(route.params.id, data)
      ElMessage.success('保存成功')
    } else {
      await createActivity(data)
      ElMessage.success('发布成功')
    }
    router.push('/activities')
  } catch {
    ElMessage.error(isEdit.value ? '保存失败' : '发布失败')
  } finally {
    saving.value = false
  }
}

const handleUploadSuccess = (res) => {
  if (res.url) {
    form.cover = res.url
    ElMessage.success('上传成功')
  } else {
    ElMessage.error('上传失败')
  }
}

const handleUploadError = () => {
  ElMessage.error('上传失败')
}

const beforeUpload = (file) => {
  const isImage = file.type.startsWith('image/')
  const isLt5M = file.size / 1024 / 1024 < 5
  if (!isImage) {
    ElMessage.error('只能上传图片文件')
    return false
  }
  if (!isLt5M) {
    ElMessage.error('图片大小不能超过 5MB')
    return false
  }
  return true
}

onMounted(() => {
  loadData()
})
</script>

<style scoped>
.activity-form-page {
  width: 100%;
}

.cover-upload {
  display: flex;
  align-items: flex-start;
  gap: 16px;
}

.cover-uploader {
  border: 1px dashed #d9d9d9;
  border-radius: 8px;
  cursor: pointer;
  position: relative;
  overflow: hidden;
  transition: border-color 0.2s;
}

.cover-uploader:hover {
  border-color: #07c160;
}

.cover-uploader :deep(.el-upload) {
  width: 200px;
  height: 120px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.cover-preview {
  width: 200px;
  height: 120px;
  object-fit: cover;
  display: block;
}

.cover-uploader-icon {
  font-size: 28px;
  color: #8c939d;
}

.cover-tip {
  color: #999;
  font-size: 12px;
  line-height: 1.4;
  padding-top: 4px;
}
</style>
