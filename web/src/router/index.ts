import { createRouter, createWebHistory } from 'vue-router'
import type { RouteRecordRaw } from 'vue-router'

const routes: RouteRecordRaw[] = [
  {
    path: '/login',
    name: 'Login',
    component: () => import('@/views/auth/Login.vue'),
    meta: { requiresAuth: false }
  },
  {
    path: '/',
    component: () => import('@/layouts/BasicLayout.vue'),
    redirect: '/dashboard',
    children: [
      {
        path: 'dashboard',
        name: 'Dashboard',
        component: () => import('@/views/Dashboard.vue'),
        meta: { title: 'menu.dashboard' }
      },
      {
        path: 'system/user',
        name: 'SystemUser',
        component: () => import('@/views/system/User.vue'),
        meta: { title: 'menu.system.user', permission: 'system:user:query' }
      },
      {
        path: 'system/role',
        name: 'SystemRole',
        component: () => import('@/views/system/Role.vue'),
        meta: { title: 'menu.system.role', permission: 'system:role:query' }
      },
      {
        path: 'system/menu',
        name: 'SystemMenu',
        component: () => import('@/views/system/Menu.vue'),
        meta: { title: 'menu.system.menu', permission: 'system:menu:query' }
      },
      {
        path: 'base-data/factory',
        name: 'Factory',
        component: () => import('@/views/base-data/Factory.vue'),
        meta: { title: 'menu.baseData.factory', permission: 'base:factory:query' }
      },
      {
        path: 'base-data/workshop',
        name: 'Workshop',
        component: () => import('@/views/base-data/Workshop.vue'),
        meta: { title: 'menu.baseData.workshop', permission: 'base:workshop:query' }
      },
      {
        path: 'base-data/line',
        name: 'ProductionLine',
        component: () => import('@/views/base-data/ProductionLine.vue'),
        meta: { title: 'menu.baseData.line', permission: 'base:line:query' }
      },
      {
        path: 'base-data/product',
        name: 'Product',
        component: () => import('@/views/base-data/Product.vue'),
        meta: { title: 'menu.baseData.product', permission: 'base:product:query' }
      },
      {
        path: 'base-data/route',
        name: 'ProcessRoute',
        component: () => import('@/views/base-data/ProcessRoute.vue'),
        meta: { title: 'menu.baseData.route', permission: 'base:route:query' }
      },
      {
        path: 'base-data/operation',
        name: 'Operation',
        component: () => import('@/views/base-data/Operation.vue'),
        meta: { title: 'menu.baseData.operation', permission: 'base:operation:query' }
      },
      {
        path: 'base-data/recipe',
        name: 'Recipe',
        component: () => import('@/views/base-data/Recipe.vue'),
        meta: { title: 'menu.baseData.recipe', permission: 'base:recipe:query' }
      }
    ]
  }
]

const router = createRouter({
  history: createWebHistory(),
  routes
})

router.beforeEach((to, from, next) => {
  const token = localStorage.getItem('token')
  
  if (to.path === '/login') {
    if (token) {
      next('/')
    } else {
      next()
    }
  } else {
    if (token) {
      next()
    } else {
      next('/login')
    }
  }
})

export default router
