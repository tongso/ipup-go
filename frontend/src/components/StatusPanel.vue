<script lang="ts" setup>
import { ref, onMounted } from 'vue'
import { GetPublicIP, GetDomainStatus, GetSettings, UpdateDomainDNS } from '../../wailsjs/go/app/App'
import { notifyError, notifyInfo } from '../utils/notifications'

interface IPInfo {
  publicIP: string
  ipv4?: string
  ipv6?: string
  location: string
  isp: string
}

interface DomainStatus {
  id: number
  domain: string
  currentIP: string
  lastUpdate: string
  status: string
  message: string
  provider: string
  lastApiCall: string
  apiStatus: string
  apiMessage: string
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

const updatingDomains = ref<Map<number, boolean>>(new Map())

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
    notifyError('刷新状态失败：' + (error as Error).message)
  } finally {
    isLoading.value = false
  }
}

// 手动更新域名的 DNS 解析
const updateDomainDNS = async (domainId: number, domainName: string) => {
  // 设置该域名正在更新
  updatingDomains.value.set(domainId, true)
  
  try {
    const result = await UpdateDomainDNS(domainId)
    console.log(`✅ ${domainName} DNS 更新成功：`, result)
    notifyInfo(`✅ ${domainName}: DNS 解析更新成功！`)
    
    // 立即刷新状态列表
    await refreshStatus()
    
    // 触发自定义事件，通知其他组件
    window.dispatchEvent(new CustomEvent('domain-dns-updated', {
      detail: { domainId, domainName, success: true }
    }))
  } catch (error) {
    console.error(`❌ ${domainName} DNS 更新失败:`, error)
    notifyError(`❌ ${domainName}: ${(error as Error).message}`, 5000)
    
    // 触发失败事件
    window.dispatchEvent(new CustomEvent('domain-dns-updated', {
      detail: { domainId, domainName, success: false, error }
    }))
  } finally {
    updatingDomains.value.delete(domainId)
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
          :class="['domain-item', `status-${domain.status}`, `api-${domain.apiStatus}`]"
        >
          <!-- 左侧域名信息 -->
          <div class="domain-main-info">
            <div class="domain-name">
              {{ domain.domain }}
              <span class="provider-badge">{{ domain.provider }}</span>
            </div>
            
            <!-- 当前IP显示 -->
            <div class="domain-ip">
              <span class="label">当前 IP:</span>
              <span v-if="domain.currentIP" :class="['ip', 'ip-success']">{{ domain.currentIP }}</span>
              <span v-else :class="['ip', 'ip-error']">
                <template v-if="domain.status === 'pending'">⏳ 等待 DNS 解析</template>
                <template v-else-if="domain.status === 'error'">❌ {{ domain.message }}</template>
                <template v-else>⚠️ 无法解析</template>
              </span>
            </div>
            
            <!-- API 调用状态 -->
            <div class="api-status" v-if="domain.lastApiCall || domain.apiMessage">
              <span :class="['api-indicator', domain.apiStatus]">
                <span class="api-dot"></span>
                <span class="api-text">
                  <template v-if="domain.apiStatus === 'success'">✅</template>
                  <template v-else-if="domain.apiStatus === 'error'">❌</template>
                  <template v-else>⏳</template>
                  {{ domain.apiMessage }}
                </span>
              </span>
              <small class="api-time" v-if="domain.lastApiCall">{{ domain.lastApiCall }}</small>
            </div>
          </div>

          <!-- 中间状态信息 -->
          <div class="domain-status-info">
            <div class="status-indicator">
              <span :class="['status-dot', domain.status]"></span>
              <span class="status-text">{{ domain.message }}</span>
            </div>
            <div class="update-time">
              <small>更新于：{{ domain.lastUpdate }}</small>
            </div>
          </div>

          <!-- 右侧操作按钮 -->
          <div class="domain-actions">
            <button 
              :class="['update-dns-btn', { updating: updatingDomains.get(domain.id) }]"
              @click="updateDomainDNS(domain.id, domain.domain)"
              :disabled="updatingDomains.get(domain.id)"
              :title="`手动更新 ${domain.domain} 的 DNS 解析`"
            >
              <span v-if="updatingDomains.get(domain.id)" class="spinner"></span>
              <span v-else>🔄</span>
              <span class="btn-text">{{ updatingDomains.get(domain.id) ? '更新中...' : '更新 DNS' }}</span>
            </button>
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
  padding: 24px;
}

.section-title {
  font-size: 28px;
  margin-bottom: 24px;
  color: #1a202c;
  font-weight: 700;
  letter-spacing: -0.5px;
}

.card {
  background: #ffffff;
  border-radius: 12px;
  padding: 20px;
  margin-bottom: 24px;
  border: 1px solid #e8f4ec;
  box-shadow: 0 1px 2px rgba(0, 0, 0, 0.03), 0 4px 8px rgba(72, 187, 120, 0.05);
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
  padding-bottom: 14px;
  border-bottom: 1px solid #f0fdf4;
}

.card-header h3 {
  margin: 0;
  font-size: 18px;
  color: #1a202c;
  font-weight: 600;
}

.badge {
  background: linear-gradient(135deg, rgba(198, 246, 213, 0.5) 0%, rgba(154, 230, 180, 0.5) 100%);
  color: #22543d;
  padding: 5px 12px;
  border-radius: 12px;
  font-size: 12px;
  font-weight: 500;
  letter-spacing: 0.3px;
  border: 1px solid rgba(134, 239, 172, 0.3);
}

.refresh-btn {
  padding: 8px 16px;
  background: linear-gradient(135deg, #48bb78 0%, #38a169 100%);
  border: none;
  border-radius: 8px;
  color: white;
  cursor: pointer;
  font-size: 13px;
  font-weight: 500;
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  box-shadow: 0 2px 4px rgba(72, 187, 120, 0.2);
  letter-spacing: 0.3px;
}

.refresh-btn:hover:not(:disabled) {
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(72, 187, 120, 0.3);
  background: linear-gradient(135deg, #38a169 0%, #2f855a 100%);
}

.refresh-btn:active:not(:disabled) {
  transform: translateY(0);
}

.refresh-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.ip-card {
  text-align: center;
}

.ip-info {
  padding: 20px;
}

.ip-address {
  font-size: 40px;
  font-weight: 700;
  background: linear-gradient(135deg, #48bb78 0%, #38a169 100%);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
  margin-bottom: 10px;
  font-family: 'Courier New', monospace;
  letter-spacing: -0.5px;
}

.ip-details {
  display: flex;
  justify-content: center;
  gap: 20px;
  color: #718096;
  font-size: 13px;
  font-weight: 500;
}

/* IPv4/IPv6 双栈显示 */
.dual-stack-ips {
  margin-top: 16px;
  padding: 14px;
  background: linear-gradient(135deg, rgba(198, 246, 213, 0.25) 0%, rgba(154, 230, 180, 0.25) 100%);
  border-radius: 10px;
  border: 1px solid rgba(134, 239, 172, 0.25);
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.ip-item {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  font-size: 13px;
}

.ip-label {
  font-weight: 600;
  color: #22543d;
  font-size: 12px;
  text-transform: uppercase;
  letter-spacing: 0.5px;
}

.ip-value {
  font-family: 'Courier New', monospace;
  color: #276749;
  font-weight: 600;
  font-size: 14px;
}

.domain-list {
  display: flex;
  flex-direction: column;
  gap: 14px;
}

.domain-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 18px;
  background: #ffffff;
  border-radius: 10px;
  border: 1px solid #f0fdf4;
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  position: relative;
  overflow: hidden;
}

/* 左侧主要信息区域 */
.domain-main-info {
  flex: 1;
  min-width: 0; /* 允许内容溢出时正确截断 */
}

/* 中间状态信息区域 */
.domain-status-info {
  width: 160px;
  text-align: right;
  margin-left: 20px;
  margin-right: 20px;
  flex-shrink: 0;
}

/* 右侧操作区域 */
.domain-actions {
  width: 120px;
  flex-shrink: 0;
  display: flex;
  justify-content: flex-end;
}

.domain-item::before {
  content: '';
  position: absolute;
  left: 0;
  top: 0;
  bottom: 0;
  width: 3px;
  background: linear-gradient(180deg, #48bb78 0%, #38a169 100%);
  transition: all 0.3s ease;
}

.domain-item:hover {
  background: #fafcfb;
  border-color: #e8f4ec;
  transform: translateX(3px);
  box-shadow: 0 3px 10px rgba(72, 187, 120, 0.08);
}

.update-dns-btn {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 8px 14px;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  border: none;
  border-radius: 8px;
  color: white;
  cursor: pointer;
  font-size: 12px;
  font-weight: 600;
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  box-shadow: 0 2px 4px rgba(102, 126, 234, 0.3);
  letter-spacing: 0.3px;
}

.update-dns-btn:hover:not(:disabled) {
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(102, 126, 234, 0.4);
}

.update-dns-btn:active:not(:disabled) {
  transform: translateY(0);
}

.update-dns-btn:disabled {
  opacity: 0.6;
  cursor: not-allowed;
  transform: none;
}

.update-dns-btn.updating {
  background: linear-gradient(135deg, #a0aec0 0%, #718096 100%);
}

.spinner {
  display: inline-block;
  width: 14px;
  height: 14px;
  border: 2px solid rgba(255, 255, 255, 0.3);
  border-top-color: white;
  border-radius: 50%;
  animation: spin 0.8s linear infinite;
}

@keyframes spin {
  to {
    transform: rotate(360deg);
  }
}

.btn-text {
  font-size: 11px;
  text-transform: uppercase;
  letter-spacing: 0.5px;
  font-weight: 700;
}

.domain-item::before {
  content: '';
  position: absolute;
  left: 0;
  top: 0;
  bottom: 0;
  width: 3px;
  background: linear-gradient(180deg, #48bb78 0%, #38a169 100%);
  transition: all 0.3s ease;
}

.domain-item:hover {
  background: #fafcfb;
  border-color: #e8f4ec;
  transform: translateX(3px);
  box-shadow: 0 3px 10px rgba(72, 187, 120, 0.08);
}

.domain-item.status-success::before {
  background: linear-gradient(180deg, #48bb78 0%, #38a169 100%);
}

.domain-item.status-pending::before {
  background: linear-gradient(180deg, #ed8936 0%, #dd6b20 100%);
}

.domain-item.status-error::before {
  background: linear-gradient(180deg, #fc8181 0%, #f56565 100%);
}

.domain-info {
  flex: 1;
}

.domain-name {
  font-size: 16px;
  font-weight: 600;
  color: #1a202c;
  margin-bottom: 6px;
  letter-spacing: -0.2px;
}

.domain-name .provider-badge {
  background: linear-gradient(135deg, rgba(198, 246, 213, 0.5) 0%, rgba(154, 230, 180, 0.5) 100%);
  color: #22543d;
  padding: 4px 8px;
  border-radius: 12px;
  font-size: 12px;
  font-weight: 500;
  letter-spacing: 0.3px;
  border: 1px solid rgba(134, 239, 172, 0.3);
  margin-left: 8px;
}

.domain-ip {
  font-size: 13px;
  color: #718096;
}

.domain-ip .label {
  margin-right: 6px;
  font-weight: 500;
  color: #a0aec0;
  font-size: 12px;
}

.domain-ip .ip {
  color: #48bb78;
  font-family: 'Courier New', monospace;
  font-weight: 600;
  font-size: 13px;
}

.domain-ip .ip-success {
  color: #38a169;
  font-weight: 600;
}

.domain-ip .ip-error {
  color: #fc8181;
  font-style: italic;
  font-size: 12px;
}

.api-status {
  margin-top: 8px;
  padding: 8px 10px;
  background: #f7fafc;
  border-radius: 6px;
  border-left: 3px solid #cbd5e0;
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 12px;
}

.api-indicator {
  display: flex;
  align-items: center;
  gap: 6px;
  flex: 1;
}

.api-dot {
  width: 6px;
  height: 6px;
  border-radius: 50%;
  display: inline-block;
}

.api-success .api-dot {
  background-color: #48bb78;
  box-shadow: 0 0 4px rgba(72, 187, 120, 0.4);
}

.api-error .api-dot {
  background-color: #fc8181;
  box-shadow: 0 0 4px rgba(252, 129, 129, 0.4);
}

.api-pending .api-dot {
  background-color: #ed8936;
  box-shadow: 0 0 4px rgba(237, 137, 54, 0.4);
}

.api-text {
  color: #4a5568;
  font-size: 11px;
  line-height: 1.4;
}

.api-time {
  color: #a0aec0;
  font-size: 10px;
  white-space: nowrap;
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
  box-shadow: 0 0 6px currentColor;
}

.status-dot.success {
  background-color: #48bb78;
  color: #48bb78;
}

.status-dot.pending {
  background-color: #ed8936;
  color: #ed8936;
}

.status-dot.error {
  background-color: #fc8181;
  color: #fc8181;
}

@keyframes pulse {
  0%, 100% {
    opacity: 1;
    transform: scale(1);
  }
  50% {
    opacity: 0.7;
    transform: scale(1.05);
  }
}

/* 状态项的额外样式 */
.domain-item.api-success {
  border-left: 3px solid #48bb78;
}

.domain-item.api-error {
  border-left: 3px solid #fc8181;
}

.domain-item.api-pending {
  border-left: 3px solid #ed8936;
}

.status-text {
  font-size: 13px;
  font-weight: 500;
  color: #4a5568;
}

.status-success .status-text {
  color: #22543d;
}

.status-pending .status-text {
  color: #744210;
}

.status-error .status-text {
  color: #742a2a;
}

.update-time {
  color: #a0aec0;
  font-size: 11px;
  font-weight: 500;
}

.empty-state {
  text-align: center;
  padding: 50px 20px;
  color: #a0aec0;
}

.empty-state p {
  margin: 0 0 8px 0;
  font-size: 16px;
  color: #718096;
  font-weight: 600;
}

.empty-state small {
  font-size: 13px;
  color: #a0aec0;
}
</style>
