import { createRouter, createWebHistory } from 'vue-router'

const routes = [
  {
    path: '/login',
    name: 'Login',
    component: () => import('../views/Login.vue')
  },
  {
    path: '/',
    component: () => import('../views/Layout.vue'),
    children: [
      {
        path: '',
        redirect: '/activities'
      },
      {
        path: '/activities',
        name: 'Activities',
        component: () => import('../views/Activities.vue')
      },
      {
        path: '/activities/create',
        name: 'ActivityCreate',
        component: () => import('../views/ActivityForm.vue')
      },
      {
        path: '/activities/:id/edit',
        name: 'ActivityEdit',
        component: () => import('../views/ActivityForm.vue')
      },
      {
        path: '/activities/:id/signups',
        name: 'Signups',
        component: () => import('../views/Signups.vue')
      },
      {
        path: '/brands',
        name: 'Brands',
        component: () => import('../views/Brands.vue')
      },
      {
        path: '/bikes',
        name: 'Bikes',
        component: () => import('../views/Bikes.vue')
      }
    ]
  }
]

const router = createRouter({
  history: createWebHistory(),
  routes
})

export default router
