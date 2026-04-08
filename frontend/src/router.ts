import { createRouter, createWebHistory } from 'vue-router'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    {
      path: '/',
      name: 'lobby',
      component: () => import('./views/LobbyView.vue'),
    },
    {
      path: '/room/:code',
      name: 'room',
      component: () => import('./views/RoomView.vue'),
      props: true,
    },
    {
      path: '/game/:code',
      name: 'game',
      component: () => import('./views/GameView.vue'),
      props: true,
    },
  ],
})

export default router
