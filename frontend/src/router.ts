import { createRouter, createWebHistory } from 'vue-router'
import { useRoomStore } from './stores/room'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    {
      path: '/',
      name: 'lobby',
      component: () => import('./views/LobbyView.vue'),
    },
    {
      path: '/room',
      redirect: '/',
    },
    {
      path: '/room/:code',
      name: 'room',
      component: () => import('./views/RoomView.vue'),
      props: true,
      beforeEnter: (to) => {
        const roomStore = useRoomStore()
        if (roomStore.status !== 'waiting' && roomStore.status !== 'playing') {
          return { path: '/', query: { code: to.params.code as string } }
        }
      },
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
