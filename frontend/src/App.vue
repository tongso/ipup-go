<script lang="ts" setup>
import { ref, onMounted, provide } from 'vue'
import DomainList from './components/DomainList.vue'
import StatusPanel from './components/StatusPanel.vue'
import Settings from './components/Settings.vue'
import Logs from './components/Logs.vue'
import Notifications from './components/Notifications.vue'

const activeTab = ref('status')

const tabs = [
  { id: 'status', name: '状态监控', icon: '📊' },
  { id: 'domains', name: '域名管理', icon: '🌐' },
  { id: 'logs', name: '日志查看', icon: '📝' },
  { id: 'settings', name: '设置', icon: '⚙️' }
]

// 提供通知方法给所有子组件
const notificationApi = {
  success: (msg: string) => {
    const event = new CustomEvent('show-notification', { 
      detail: { type: 'success', message: msg } 
    })
    window.dispatchEvent(event)
  },
  error: (msg: string) => {
    const event = new CustomEvent('show-notification', { 
      detail: { type: 'error', message: msg } 
    })
    window.dispatchEvent(event)
  },
  info: (msg: string) => {
    const event = new CustomEvent('show-notification', { 
      detail: { type: 'info', message: msg } 
    })
    window.dispatchEvent(event)
  },
  warning: (msg: string) => {
    const event = new CustomEvent('show-notification', { 
      detail: { type: 'warning', message: msg } 
    })
    window.dispatchEvent(event)
  }
}

provide('notify', notificationApi)
</script>

<template>
  <div class="app-container">
    <!-- Header -->
    <header class="header">
      <h1 class="title">🌐 ipup-go - DDNS 动态域名解析</h1>
      <p class="subtitle">Dynamic Domain Name System Manager</p>
    </header>

    <!-- Navigation Tabs -->
    <nav class="nav-tabs">
      <button
        v-for="tab in tabs"
        :key="tab.id"
        :class="['tab-button', { active: activeTab === tab.id }]"
        @click="activeTab = tab.id"
      >
        <span class="tab-icon">{{ tab.icon }}</span>
        <span class="tab-name">{{ tab.name }}</span>
      </button>
    </nav>

    <!-- Main Content -->
    <main class="main-content">
      <StatusPanel v-if="activeTab === 'status'" />
      <DomainList v-else-if="activeTab === 'domains'" />
      <Logs v-else-if="activeTab === 'logs'" />
      <Settings v-else-if="activeTab === 'settings'" />
    </main>

    <!-- Footer -->
    <footer class="footer">
      <p>DDNS Manager © 2026 | Powered by Wails + Vue 3</p>
    </footer>

    <!-- 通知组件 -->
    <Notifications />
  </div>
</template>

<style scoped>
.app-container {
  display: flex;
  flex-direction: column;
  height: 100vh;
  background: transparent;
}

.header {
  padding: 20px;
  background: rgba(255, 255, 255, 0.9);
  border-bottom: 1px solid #e2e8f0;
  backdrop-filter: blur(10px);
}

.title {
  margin: 0;
  font-size: 28px;
  font-weight: 700;
  color: #1a202c;
  letter-spacing: -0.5px;
}

.subtitle {
  margin: 8px 0 0;
  font-size: 14px;
  color: #718096;
  font-weight: 500;
}

.nav-tabs {
  display: flex;
  background: rgba(255, 255, 255, 0.95);
  padding: 12px 20px;
  gap: 8px;
  border-bottom: 1px solid #e2e8f0;
  backdrop-filter: blur(10px);
}

.tab-button {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 8px 16px;
  background: transparent;
  border: none;
  border-radius: 8px;
  color: #718096;
  cursor: pointer;
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  font-size: 14px;
  font-family: inherit;
  font-weight: 500;
}

.tab-button:hover {
  background: rgba(72, 187, 120, 0.08);
  color: #2d3748;
  transform: translateY(-1px);
}

.tab-button.active {
  background: linear-gradient(135deg, rgba(72, 187, 120, 0.1) 0%, rgba(56, 161, 105, 0.1) 100%);
  color: #22543d;
  box-shadow: 0 2px 8px rgba(72, 187, 120, 0.15);
  font-weight: 600;
}

.tab-icon {
  font-size: 16px;
  line-height: 1;
}

.main-content {
  flex: 1;
  overflow-y: auto;
  padding: 24px;
}

.footer {
  padding: 16px 20px;
  background: rgba(255, 255, 255, 0.9);
  border-top: 1px solid #e2e8f0;
  text-align: center;
  font-size: 13px;
  color: #a0aec0;
  font-weight: 500;
}
</style>
