import { createRouter, createWebHistory } from 'vue-router'

const routes = [
  {
    path: '/login',
    name: 'Login',
    component: () => import('../views/login/index.vue')
  },
  {
    path: '/',
    component: () => import('../views/layout/index.vue'),
    children: [
      {
        path: '',
        redirect: '/dashboard'
      },
      {
        path: '/dashboard',
        name: 'Dashboard',
        component: () => import('../views/dashboard/index.vue')
      },
      {
        path: '/activities',
        name: 'Activities',
        component: () => import('../views/activity/index.vue')
      },
      {
        path: '/activities/create',
        name: 'ActivityCreate',
        component: () => import('../views/activity/Form.vue')
      },
      {
        path: '/activities/:id/edit',
        name: 'ActivityEdit',
        component: () => import('../views/activity/Form.vue')
      },
      {
        path: '/activities/:id/signups',
        name: 'Signups',
        component: () => import('../views/activity/Signups.vue')
      },
      {
        path: '/orders',
        name: 'Orders',
        component: () => import('../views/order/index.vue')
      },
      {
        path: '/bikes',
        name: 'Bikes',
        component: () => import('../views/bike/index.vue')
      },
      {
        path: '/users',
        name: 'Users',
        component: () => import('../views/user/index.vue')
      }
    ]
  }
]

const router = createRouter({
  history: createWebHistory(),
  routes
})

export default router
