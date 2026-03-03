<script lang="ts" setup>
import { ref, onMounted } from 'vue'
import DomainList from './components/DomainList.vue'
import StatusPanel from './components/StatusPanel.vue'
import Settings from './components/Settings.vue'
import Logs from './components/Logs.vue'

const activeTab = ref('status')

const tabs = [
  { id: 'status', name: '状态监控', icon: '📊' },
  { id: 'domains', name: '域名管理', icon: '🌐' },
  { id: 'logs', name: '日志查看', icon: '📝' },
  { id: 'settings', name: '设置', icon: '⚙️' }
]
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
  </div>
</template>

<style scoped>
.app-container {
  display: flex;
  flex-direction: column;
  height: 100vh;
  background: linear-gradient(135deg, #1a1f3a 0%, #2d3748 100%);
}

.header {
  padding: 20px;
  background: rgba(0, 0, 0, 0.3);
  border-bottom: 1px solid rgba(255, 255, 255, 0.1);
}

.title {
  margin: 0;
  font-size: 28px;
  font-weight: 600;
  color: #fff;
  text-shadow: 0 2px 4px rgba(0, 0, 0, 0.3);
}

.subtitle {
  margin: 8px 0 0;
  font-size: 14px;
  color: rgba(255, 255, 255, 0.6);
}

.nav-tabs {
  display: flex;
  background: rgba(0, 0, 0, 0.2);
  padding: 10px 20px;
  gap: 10px;
  border-bottom: 1px solid rgba(255, 255, 255, 0.1);
}

.tab-button {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 12px 24px;
  background: rgba(255, 255, 255, 0.05);
  border: 1px solid rgba(255, 255, 255, 0.1);
  border-radius: 8px;
  color: rgba(255, 255, 255, 0.7);
  cursor: pointer;
  transition: all 0.3s ease;
  font-size: 14px;
  font-family: inherit;
}

.tab-button:hover {
  background: rgba(255, 255, 255, 0.1);
  color: #fff;
  transform: translateY(-2px);
}

.tab-button.active {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: #fff;
  border-color: transparent;
  box-shadow: 0 4px 12px rgba(102, 126, 234, 0.4);
}

.tab-icon {
  font-size: 18px;
}

.main-content {
  flex: 1;
  overflow-y: auto;
  padding: 20px;
}

.footer {
  padding: 15px;
  text-align: center;
  background: rgba(0, 0, 0, 0.3);
  border-top: 1px solid rgba(255, 255, 255, 0.1);
  font-size: 12px;
  color: rgba(255, 255, 255, 0.5);
}
</style>
