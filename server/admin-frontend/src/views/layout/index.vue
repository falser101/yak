<template>
  <el-container class="layout-container">
    <!-- 左侧菜单 -->
    <el-aside class="aside" :width="isCollapse ? '64px' : '220px'">
      <div class="logo-area">
        <span v-if="!isCollapse" class="logo-text">🚴 管理后台</span>
        <span v-else class="logo-text">🚴</span>
      </div>
      <el-menu
        :default-active="activeMenu"
        class="aside-menu"
        :collapse="isCollapse"
        @select="handleMenuSelect"
      >
        <el-menu-item index="/dashboard">
          <el-icon><DataBoard /></el-icon>
          <template #title>控制台</template>
        </el-menu-item>
        <el-menu-item index="/activities">
          <el-icon><Calendar /></el-icon>
          <template #title>活动管理</template>
        </el-menu-item>
        <el-menu-item index="/orders">
          <el-icon><List /></el-icon>
          <template #title>订单管理</template>
        </el-menu-item>
        <el-menu-item index="/bikes">
          <el-icon><Box /></el-icon>
          <template #title>装备租赁</template>
        </el-menu-item>
        <el-menu-item index="/users">
          <el-icon><User /></el-icon>
          <template #title>用户管理</template>
        </el-menu-item>
      </el-menu>
      <div class="collapse-btn" @click="isCollapse = !isCollapse">
        <el-icon v-if="isCollapse"><DArrowRight /></el-icon>
        <el-icon v-else><DArrowLeft /></el-icon>
      </div>
    </el-aside>

    <el-container class="main-container">
      <!-- 顶部导航 -->
      <el-header class="header">
        <div class="header-left">
          <span class="page-title">{{ pageTitle }}</span>
        </div>
        <div class="header-right">
          <el-dropdown @command="handleCommand">
            <span class="user-info">
              <el-icon><User /></el-icon>
              {{ username }}
              <el-icon><ArrowDown /></el-icon>
            </span>
            <template #dropdown>
              <el-dropdown-menu>
                <el-dropdown-item command="logout">退出登录</el-dropdown-item>
              </el-dropdown-menu>
            </template>
          </el-dropdown>
        </div>
      </el-header>

      <!-- 内容区 -->
      <el-main class="main-content">
        <router-view />
      </el-main>
    </el-container>
  </el-container>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { Calendar, User, ArrowDown, DArrowLeft, DArrowRight, DataBoard, List, Box } from '@element-plus/icons-vue'

const router = useRouter()
const route = useRoute()
const username = ref(localStorage.getItem('admin_username') || 'Admin')
const isCollapse = ref(false)

const activeMenu = computed(() => route.path)

const pageTitle = computed(() => {
  const map = {
    '/dashboard': '控制台',
    '/activities': '活动管理',
    '/orders': '订单管理',
    '/bikes': '装备租赁',
    '/users': '用户管理',
    '/activities/create': '创建活动',
    '/activities/:id/edit': '编辑活动',
    '/activities/:id/signups': '报名管理'
  }
  return map[route.path] || '管理后台'
})

const handleMenuSelect = (index) => {
  router.push(index)
}

const handleCommand = (command) => {
  if (command === 'logout') {
    localStorage.removeItem('admin_token')
    localStorage.removeItem('admin_username')
    router.push('/login')
  }
}
</script>

<style scoped>
.layout-container {
  height: 100vh;
  overflow: hidden;
}

.aside {
  background: #304156;
  display: flex;
  flex-direction: column;
  transition: width 0.3s;
  position: relative;
}

.logo-area {
  height: 60px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: #263445;
  border-bottom: 1px solid #3d4a5c;
}

.logo-text {
  color: #fff;
  font-size: 16px;
  font-weight: 600;
  white-space: nowrap;
}

.aside-menu {
  flex: 1;
  border-right: none;
  background: #304156;
}

.aside-menu:not(.el-menu--collapse) {
  width: 220px;
}

.aside-menu .el-menu-item {
  color: #bfcbd9;
}

.aside-menu .el-menu-item:hover,
.aside-menu .el-menu-item.is-active {
  background: #263445;
  color: #07c160;
}

.aside-menu .el-menu-item .el-icon {
  color: inherit;
}

.collapse-btn {
  position: absolute;
  bottom: 0;
  left: 0;
  right: 0;
  height: 40px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: #263445;
  color: #bfcbd9;
  cursor: pointer;
  border-top: 1px solid #3d4a5c;
}

.collapse-btn:hover {
  color: #07c160;
}

.main-container {
  display: flex;
  flex-direction: column;
}

.header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  background: #fff;
  padding: 0 20px;
  box-shadow: 0 1px 4px rgba(0, 0, 0, 0.08);
  height: 60px;
}

.header-left {
  display: flex;
  align-items: center;
}

.page-title {
  font-size: 16px;
  font-weight: 500;
  color: #333;
}

.header-right {
  display: flex;
  align-items: center;
}

.user-info {
  color: #333;
  cursor: pointer;
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 8px 12px;
  border-radius: 4px;
}

.user-info:hover {
  background: #f5f5f5;
}

.main-content {
  flex: 1;
  padding: 20px;
  background: #f5f5f5;
  overflow-y: auto;
}
</style>
