<script lang="ts" setup>
import { ref, onMounted } from 'vue'

interface Notification {
  id: number
  type: 'success' | 'error' | 'info' | 'warning'
  message: string
  duration?: number
}

const notifications = ref<Notification[]>([])
const notificationId = ref(0)

// 添加通知
const addNotification = (type: 'success' | 'error' | 'info' | 'warning', message: string, duration: number = 3000) => {
  const id = ++notificationId.value
  notifications.value.push({ id, type, message, duration })
  
  // 自动移除
  if (duration > 0) {
    setTimeout(() => {
      removeNotification(id)
    }, duration)
  }
}

// 移除通知
const removeNotification = (id: number) => {
  const index = notifications.value.findIndex(n => n.id === id)
  if (index !== -1) {
    notifications.value.splice(index, 1)
  }
}

// 获取图标
const getIcon = (type: string) => {
  const icons: Record<string, string> = {
    success: '✓',
    error: '✗',
    info: 'ℹ',
    warning: '⚠'
  }
  return icons[type] || 'ℹ'
}

// 获取样式类
const getTypeClass = (type: string) => {
  return `notification-${type}`
}

// 监听全局通知事件
onMounted(() => {
  window.addEventListener('show-notification', (event: any) => {
    const { type, message, duration } = event.detail
    addNotification(type, message, duration)
  })
})
</script>

<template>
  <div class="notification-container">
    <transition-group name="notification-list">
      <div 
        v-for="notification in notifications" 
        :key="notification.id"
        :class="['notification', getTypeClass(notification.type)]"
      >
        <span class="notification-icon">{{ getIcon(notification.type) }}</span>
        <span class="notification-message">{{ notification.message }}</span>
        <button class="notification-close" @click="removeNotification(notification.id)">×</button>
      </div>
    </transition-group>
  </div>
</template>

<style scoped>
.notification-container {
  position: fixed;
  top: 80px;
  right: 24px;
  z-index: 9999;
  display: flex;
  flex-direction: column;
  gap: 12px;
  max-width: 420px;
  pointer-events: none;
}

.notification {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 14px 16px;
  background: #ffffff;
  border-radius: 10px;
  box-shadow: 0 4px 16px rgba(0, 0, 0, 0.1), 0 2px 8px rgba(0, 0, 0, 0.06);
  border-left: 4px solid;
  pointer-events: auto;
  animation: slideInRight 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  backdrop-filter: blur(10px);
  min-width: 320px;
}

@keyframes slideInRight {
  from {
    opacity: 0;
    transform: translateX(100%);
  }
  to {
    opacity: 1;
    transform: translateX(0);
  }
}

.notification-success {
  border-left-color: #48bb78;
}

.notification-success .notification-icon {
  color: #48bb78;
  background: linear-gradient(135deg, rgba(72, 187, 120, 0.1) 0%, rgba(56, 161, 105, 0.1) 100%);
}

.notification-error {
  border-left-color: #f56565;
}

.notification-error .notification-icon {
  color: #f56565;
  background: linear-gradient(135deg, rgba(245, 101, 101, 0.1) 0%, rgba(229, 62, 62, 0.1) 100%);
}

.notification-info {
  border-left-color: #4299e1;
}

.notification-info .notification-icon {
  color: #4299e1;
  background: linear-gradient(135deg, rgba(66, 153, 225, 0.1) 0%, rgba(49, 130, 206, 0.1) 100%);
}

.notification-warning {
  border-left-color: #ed8936;
}

.notification-warning .notification-icon {
  color: #ed8936;
  background: linear-gradient(135deg, rgba(237, 137, 54, 0.1) 0%, rgba(221, 107, 32, 0.1) 100%);
}

.notification-icon {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 28px;
  height: 28px;
  border-radius: 50%;
  font-size: 16px;
  font-weight: 700;
  flex-shrink: 0;
}

.notification-message {
  flex: 1;
  color: #2d3748;
  font-size: 14px;
  line-height: 1.5;
  font-weight: 500;
}

.notification-close {
  width: 24px;
  height: 24px;
  border: none;
  background: transparent;
  color: #a0aec0;
  cursor: pointer;
  border-radius: 6px;
  font-size: 20px;
  line-height: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all 0.2s ease;
  flex-shrink: 0;
}

.notification-close:hover {
  background: #f7fafc;
  color: #4a5568;
}

/* 列表动画 */
.notification-list-enter-active,
.notification-list-leave-active {
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
}

.notification-list-enter-from {
  opacity: 0;
  transform: translateX(100%);
}

.notification-list-leave-to {
  opacity: 0;
  transform: translateX(100%);
  margin-top: -60px;
}

/* 响应式 */
@media (max-width: 768px) {
  .notification-container {
    top: 70px;
    right: 16px;
    left: 16px;
    max-width: none;
  }
  
  .notification {
    min-width: auto;
  }
}
</style>
