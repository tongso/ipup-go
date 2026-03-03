<script lang="ts" setup>
import { ref, onMounted } from 'vue'
import { GetPublicIP, GetDomainStatus, RefreshStatus as BackendRefreshStatus, GetSettings } from '../../wailsjs/go/app/App'

interface IPInfo {
  publicIP: string
  ipv4?: string
  ipv6?: string
  location: string
  isp: string
}

interface DomainStatus {
  domain: string
  currentIP: string
  lastUpdate: string
  status: string
  message: string
}

interface Settings {
  autoStart: boolean
  checkInterval: number
  retryCount: number
  retryDelay: number
  logLevel: string
  notifySuccess: boolean
  notifyError: boolean
  proxy: string
  apiEndpoint: string
}

const ipInfo = ref<IPInfo>({
  publicIP: '获取中...',
  ipv4: '',
  ipv6: '',
  location: '',
  isp: ''
})

const domainStatuses = ref<DomainStatus[]>([])

const isLoading = ref(false)

// 刷新状态
const refreshStatus = async () => {
  isLoading.value = true
  try {
    // 获取公网 IP
    const ipResult = await GetPublicIP()
    if (ipResult) {
      ipInfo.value = ipResult
    }
    
    // 获取域名状态 - 每次都从数据库获取最新数据
    const statusResult = await GetDomainStatus()
    domainStatuses.value = statusResult || []
  } catch (error) {
    console.error('刷新状态失败:', error)
    alert('刷新状态失败：' + (error as Error).message)
  } finally {
    isLoading.value = false
  }
}

onMounted(() => {
  // 初始加载
  refreshStatus()
  
  // 从设置中读取检查间隔（默认 5 分钟）
  let checkInterval = 5 * 60 * 1000 // 默认值
  
  // 异步获取设置
  GetSettings().then((settings: Settings) => {
    if (settings && settings.checkInterval) {
      checkInterval = settings.checkInterval * 60 * 1000 // 转换为毫秒
      console.log(`📡 状态监控刷新间隔：${settings.checkInterval}分钟`)
    }
  }).catch((err: Error) => {
    console.warn('获取设置失败，使用默认值:', err)
  })
  
  // 定时刷新状态监控面板
  const interval = setInterval(refreshStatus, checkInterval)
  
  // 监听域名更新事件，立即刷新状态
  window.addEventListener('domains-updated', () => {
    console.log('📡 检测到域名更新，立即刷新状态...')
    refreshStatus()
  })
  
  return () => {
    clearInterval(interval)
    window.removeEventListener('domains-updated', () => {})
  }
})
</script>

<template>
  <div class="status-panel">
    <h2 class="section-title">📊 状态监控</h2>
    
    <!-- Public IP Card -->
    <div class="card ip-card">
      <div class="card-header">
        <h3>🌍 公网 IP 地址</h3>
        <button 
          :class="['refresh-btn', { loading: isLoading }]" 
          @click="refreshStatus"
          :disabled="isLoading"
        >
          {{ isLoading ? '刷新中...' : '🔄 刷新' }}
        </button>
      </div>
      <div class="ip-info">
        <div class="ip-address">{{ ipInfo.publicIP }}</div>
        <div class="ip-details" v-if="ipInfo.location || ipInfo.isp">
          <span v-if="ipInfo.location">📍 {{ ipInfo.location }}</span>
          <span v-if="ipInfo.isp">🌐 {{ ipInfo.isp }}</span>
        </div>
        <!-- 显示 IPv4 和 IPv6 -->
        <div class="dual-stack-ips" v-if="ipInfo.ipv4 || ipInfo.ipv6">
          <div class="ip-item" v-if="ipInfo.ipv4" title="IPv4 地址">
            <span class="ip-label">IPv4:</span>
            <span class="ip-value">{{ ipInfo.ipv4 }}</span>
          </div>
          <div class="ip-item" v-if="ipInfo.ipv6" title="IPv6 地址">
            <span class="ip-label">IPv6:</span>
            <span class="ip-value">{{ ipInfo.ipv6 }}</span>
          </div>
        </div>
      </div>
    </div>

    <!-- Domain Status List -->
    <div class="card">
      <div class="card-header">
        <h3>🔗 域名状态</h3>
        <span class="badge">{{ domainStatuses.length }} 个域名</span>
      </div>
      
      <div class="domain-list">
        <div 
          v-for="(domain, index) in domainStatuses" 
          :key="index"
          :class="['domain-item', `status-${domain.status}`]"
        >
          <div class="domain-info">
            <div class="domain-name">{{ domain.domain }}</div>
            <div class="domain-ip">
              <span class="label">当前 IP:</span>
              <span v-if="domain.currentIP" :class="['ip', 'ip-success']">{{ domain.currentIP }}</span>
              <span v-else :class="['ip', 'ip-error']">
                <template v-if="domain.status === 'pending'">⏳ 等待 DNS 解析</template>
                <template v-else-if="domain.status === 'error'">❌ {{ domain.message }}</template>
                <template v-else>⚠️ 无法解析</template>
              </span>
            </div>
          </div>
          <div class="domain-status">
            <div class="status-indicator">
              <span :class="['status-dot', domain.status]"></span>
              <span class="status-text">{{ domain.message }}</span>
            </div>
            <div class="update-time">
              <small>更新于：{{ domain.lastUpdate }}</small>
            </div>
          </div>
        </div>
        
        <div v-if="domainStatuses.length === 0" class="empty-state">
          <p>暂无域名配置</p>
          <small>请前往域名管理添加域名</small>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.status-panel {
  max-width: 1200px;
  margin: 0 auto;
}

.section-title {
  font-size: 24px;
  margin-bottom: 20px;
  color: #fff;
}

.card {
  background: rgba(255, 255, 255, 0.05);
  border-radius: 12px;
  padding: 20px;
  margin-bottom: 20px;
  border: 1px solid rgba(255, 255, 255, 0.1);
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 15px;
}

.card-header h3 {
  margin: 0;
  font-size: 18px;
  color: #fff;
}

.badge {
  background: rgba(102, 126, 234, 0.3);
  color: #a78bfa;
  padding: 4px 12px;
  border-radius: 12px;
  font-size: 12px;
}

.refresh-btn {
  padding: 8px 16px;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  border: none;
  border-radius: 6px;
  color: white;
  cursor: pointer;
  font-size: 14px;
  transition: all 0.3s ease;
}

.refresh-btn:hover:not(:disabled) {
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(102, 126, 234, 0.4);
}

.refresh-btn:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.ip-card {
  text-align: center;
}

.ip-info {
  padding: 20px;
}

.ip-address {
  font-size: 36px;
  font-weight: bold;
  color: #667eea;
  margin-bottom: 10px;
  font-family: 'Courier New', monospace;
}

.ip-details {
  display: flex;
  justify-content: center;
  gap: 20px;
  color: rgba(255, 255, 255, 0.7);
  font-size: 14px;
}

/* IPv4/IPv6 双栈显示 */
.dual-stack-ips {
  margin-top: 15px;
  padding: 12px;
  background: rgba(102, 126, 234, 0.1);
  border-radius: 8px;
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.ip-item {
  display: flex;
  align-items: center;
  gap: 10px;
  font-size: 14px;
}

.ip-label {
  color: rgba(255, 255, 255, 0.6);
  font-weight: 500;
  min-width: 50px;
}

.ip-value {
  color: #667eea;
  font-family: 'Courier New', monospace;
  font-size: 14px;
  word-break: break-all;
}

.domain-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.domain-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 15px;
  background: rgba(0, 0, 0, 0.2);
  border-radius: 8px;
  border-left: 4px solid transparent;
  transition: all 0.3s ease;
}

.domain-item:hover {
  background: rgba(0, 0, 0, 0.3);
  transform: translateX(4px);
}

.domain-item.status-success {
  border-left-color: #48bb78;
}

.domain-item.status-pending {
  border-left-color: #ed8936;
}

.domain-item.status-error {
  border-left-color: #f56565;
}

.domain-info {
  flex: 1;
}

.domain-name {
  font-size: 16px;
  font-weight: 600;
  color: #fff;
  margin-bottom: 6px;
}

.domain-ip {
  font-size: 14px;
  color: rgba(255, 255, 255, 0.6);
}

.domain-ip .label {
  margin-right: 6px;
}

.domain-ip .ip {
  color: #667eea;
  font-family: 'Courier New', monospace;
}

.domain-ip .ip-success {
  color: #48bb78;
  font-weight: 500;
}

.domain-ip .ip-error {
  color: #f56565;
  font-style: italic;
  font-size: 13px;
}

.domain-status {
  text-align: right;
}

.status-indicator {
  display: flex;
  align-items: center;
  justify-content: flex-end;
  gap: 8px;
  margin-bottom: 4px;
}

.status-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  animation: pulse 2s infinite;
}

.status-dot.success {
  background-color: #48bb78;
}

.status-dot.pending {
  background-color: #ed8936;
}

.status-dot.error {
  background-color: #f56565;
}

.status-text {
  font-size: 13px;
  color: rgba(255, 255, 255, 0.8);
}

.update-time {
  color: rgba(255, 255, 255, 0.4);
  font-size: 12px;
}

.empty-state {
  text-align: center;
  padding: 40px 20px;
  color: rgba(255, 255, 255, 0.5);
}

.empty-state p {
  font-size: 16px;
  margin-bottom: 8px;
}

@keyframes pulse {
  0%, 100% {
    opacity: 1;
  }
  50% {
    opacity: 0.5;
  }
}
</style>
