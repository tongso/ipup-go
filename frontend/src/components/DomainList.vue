<script lang="ts" setup>
import { ref, reactive, watch, onMounted, onUnmounted } from 'vue'
import { ListDomains, AddDomain as BackendAddDomain, UpdateDomain as BackendUpdateDomain, DeleteDomain as BackendDeleteDomain, ToggleDomain as BackendToggleDomain, GetDomainStatus } from '../../wailsjs/go/app/App'

interface DomainConfig {
  id: number
  domain: string
  provider: string
  token: string
  interval: number
  enabled: boolean
  currentIP?: string
  lastUpdate?: string
  createdAt?: string
  updatedAt?: string
}

// 从数据库加载域名列表
const domains = ref<DomainConfig[]>([])
const isLoading = ref(true)

// 存储每个域名的刷新定时器
const refreshTimers = new Map<number, number>()

const showAddModal = ref(false)
const editingDomain = ref<DomainConfig | null>(null)

// 统一的表单数据
const form = reactive<DomainConfig>({
  id: 0,
  domain: '',
  provider: 'Cloudflare',
  token: '',
  interval: 300,
  enabled: true
})

const providers = [
  { value: 'Cloudflare', label: '☁️ Cloudflare' },
  { value: 'Aliyun', label: '🔶 阿里云' },
  { value: 'Tencent', label: '🐧 腾讯云' },
  { value: 'DNSPod', label: '🎯 DNSPod' },
  { value: 'GoDaddy', label: '🌟 GoDaddy' }
]

// 从数据库加载域名列表
const loadDomains = async () => {
  try {
    isLoading.value = true
    const result = await ListDomains()
    domains.value = result || []
    
    // 为每个启用的域名创建独立刷新定时器
    updateAllRefreshTimers()
  } catch (error) {
    console.error('加载域名列表失败:', error)
    alert('加载域名列表失败')
  } finally {
    isLoading.value = false
  }
}

// 为单个域名创建刷新定时器
const createRefreshTimer = (domainId: number, intervalSeconds: number) => {
  // 清除旧的定时器
  const oldTimer = refreshTimers.get(domainId)
  if (oldTimer) {
    clearInterval(oldTimer)
  }
  
  // 创建新的定时器（转换为毫秒）
  const intervalMs = intervalSeconds * 1000
  const timer = window.setInterval(() => {
    console.log(`⏰ 自动刷新域名 ID ${domainId} 的状态...`)
    // 只刷新状态，不重新加载整个列表
    refreshSingleDomainStatus(domainId)
  }, intervalMs)
  
  refreshTimers.set(domainId, timer)
  console.log(`✅ 域名 ID ${domainId} 的刷新定时器已创建，间隔：${intervalSeconds}秒`)
}

// 刷新单个域名的状态（仅更新状态显示，不重新加载列表）
const refreshSingleDomainStatus = async (domainId: number) => {
  try {
    const statuses = await GetDomainStatus()
    if (statuses) {
      const status = statuses.find(s => {
        // 通过域名匹配（因为 GetDomainStatus 返回的是 DomainStatus，不是完整 Domain）
        const targetDomain = domains.value.find(d => d.id === domainId)
        return targetDomain && s.domain === targetDomain.domain
      })
      
      if (status) {
        // 更新对应域名的状态信息
        const target = domains.value.find(d => d.id === domainId)
        if (target) {
          target.currentIP = status.currentIP
          target.lastUpdate = status.lastUpdate
        }
      }
    }
  } catch (error) {
    console.error(`刷新域名 ID ${domainId} 状态失败:`, error)
  }
}

// 更新所有域名刷新定时器
const updateAllRefreshTimers = () => {
  // 清除所有旧定时器
  refreshTimers.forEach((timer, id) => {
    clearInterval(timer)
  })
  refreshTimers.clear()
  
  // 为每个启用的域名创建新定时器
  domains.value.forEach(domain => {
    if (domain.enabled && domain.interval > 0) {
      createRefreshTimer(domain.id, domain.interval)
    }
  })
  
  console.log(`📊 已创建 ${refreshTimers.size} 个域名刷新定时器`)
}

// 清除指定域名的定时器
const clearDomainTimer = (domainId: number) => {
  const timer = refreshTimers.get(domainId)
  if (timer) {
    clearInterval(timer)
    refreshTimers.delete(domainId)
    console.log(`🗑️ 已清除域名 ID ${domainId} 的刷新定时器`)
  }
}

// 添加域名
const addDomain = async () => {
  if (!form.domain || !form.token) {
    alert('请填写完整信息')
    return
  }
  
  try {
    await BackendAddDomain({
      id: 0,
      domain: form.domain,
      provider: form.provider,
      token: form.token,
      interval: form.interval,
      enabled: form.enabled,
      currentIP: '',
      lastUpdate: '',
      createdAt: '',
      updatedAt: ''
    })
    
    alert('添加成功')
    resetForm()
    showAddModal.value = false
    
    // 立即重新加载列表
    await loadDomains()
    
    // 通知其他组件数据已更新
    window.dispatchEvent(new CustomEvent('domains-updated'))
  } catch (error) {
    console.error('添加域名失败:', error)
    alert('添加失败：' + (error as Error).message)
  }
}

// 编辑域名
const editDomain = (domain: DomainConfig) => {
  editingDomain.value = domain
  // 复制数据到表单
  form.id = domain.id
  form.domain = domain.domain
  form.provider = domain.provider
  form.token = domain.token
  form.interval = domain.interval
  form.enabled = domain.enabled
  showAddModal.value = true
}

// 保存编辑
const saveEdit = async () => {
  if (!editingDomain.value) return
  
  try {
    await BackendUpdateDomain({
      id: form.id,
      domain: form.domain,
      provider: form.provider,
      token: form.token,
      interval: form.interval,
      enabled: form.enabled,
      currentIP: editingDomain.value?.currentIP || '',
      lastUpdate: editingDomain.value?.lastUpdate || '',
      createdAt: editingDomain.value?.createdAt || '',
      updatedAt: ''
    })
    
    alert('更新成功')
    resetForm()
    showAddModal.value = false
    editingDomain.value = null
    
    // 立即重新加载列表，确保数据最新
    await loadDomains()
    
    // 通知其他组件数据已更新（可选）
    window.dispatchEvent(new CustomEvent('domains-updated'))
  } catch (error) {
    console.error('更新域名失败:', error)
    alert('更新失败：' + (error as Error).message)
  }
}

// 删除域名
const deleteDomain = async (id: number) => {
  if (!confirm('确定要删除这个域名配置吗？')) {
    return
  }
  
  try {
    await BackendDeleteDomain(id)
    alert('删除成功')
    
    // 立即重新加载列表
    await loadDomains()
    
    // 通知其他组件数据已更新
    window.dispatchEvent(new CustomEvent('domains-updated'))
  } catch (error) {
    console.error('删除域名失败:', error)
    alert('删除失败：' + (error as Error).message)
  }
}

// 切换域名状态
const toggleEnabled = async (domain: DomainConfig) => {
  try {
    await BackendToggleDomain(domain.id)
    domain.enabled = !domain.enabled // 立即更新 UI
    await loadDomains() // 重新加载列表以确保一致性
  } catch (error) {
    console.error('切换域名状态失败:', error)
    alert('操作失败：' + (error as Error).message)
  }
}

const resetForm = () => {
  form.id = 0
  form.domain = ''
  form.provider = 'Cloudflare'
  form.token = ''
  form.interval = 300
  form.enabled = true
  editingDomain.value = null
}

const closeModal = () => {
  showAddModal.value = false
  resetForm()
}

const isEditing = ref(false)

// 监听编辑状态变化
watch(editingDomain, (newVal) => {
  isEditing.value = !!newVal
})

// 组件挂载时加载数据
onMounted(() => {
  loadDomains()
})

// 组件卸载时清除所有定时器
onUnmounted(() => {
  refreshTimers.forEach((timer, id) => {
    clearInterval(timer)
  })
  refreshTimers.clear()
  console.log('🗑️ 已清除所有域名刷新定时器')
})
</script>

<template>
<div class="domain-list">
  <h2 class="section-title">🌐 域名管理</h2>
  
  <!-- Toolbar -->
  <div class="toolbar">
    <button class="btn-primary" @click="showAddModal = true">
      ➕ 添加域名
    </button>
    <span class="count">共 {{ domains.length }} 个域名</span>
  </div>
  
  <!-- Domain Cards -->
  <div class="cards-container">
    <div 
      v-for="domain in domains" 
      :key="domain.id"
      :class="['card', { disabled: !domain.enabled }]"
    >
      <div class="card-header">
        <div class="domain-info">
          <h3 class="domain-name">{{ domain.domain }}</h3>
          <span class="provider-badge">{{ domain.provider }}</span>
        </div>
        <div class="actions">
          <button 
            :class="['toggle-btn', { active: domain.enabled }]"
            @click="toggleEnabled(domain)"
            :title="domain.enabled ? '已启用' : '已禁用'"
          >
            {{ domain.enabled ? '✓' : '✗' }}
          </button>
          <button class="action-btn" @click="editDomain(domain)">✏️</button>
          <button class="action-btn delete" @click="deleteDomain(domain.id)">🗑️</button>
        </div>
      </div>
      
      <div class="card-body">
        <div class="info-row">
          <span class="label">Token:</span>
          <span class="value token">{{ domain.token.replace(/^(.{4}).*(.{4})$/, '$1****$2') }}</span>
        </div>
        <div class="info-row">
          <span class="label">更新间隔:</span>
          <span class="value">{{ domain.interval / 60 }} 分钟</span>
        </div>
        <div class="info-row">
          <span class="label">状态:</span>
          <span :class="['status-badge', domain.enabled ? 'enabled' : 'disabled']">
            {{ domain.enabled ? '🟢 运行中' : '⚪ 已禁用' }}
          </span>
        </div>
      </div>
    </div>
    
    <div v-if="domains.length === 0" class="empty-state">
      <p>暂无域名配置</p>
      <small>点击右上角按钮添加第一个域名</small>
    </div>
  </div>
  
  <!-- Add/Edit Modal -->
  <div v-if="showAddModal" class="modal-overlay" @click="closeModal">
    <div class="modal" @click.stop>
      <div class="modal-header">
        <h3>{{ isEditing ? '编辑域名' : '添加域名' }}</h3>
        <button class="close-btn" @click="closeModal">×</button>
      </div>
      
      <div class="modal-body">
        <div class="form-group">
          <label>域名</label>
          <input 
            v-model="form.domain"
            type="text" 
            placeholder="例如：example.com 或 www.example.com"
            class="form-control"
          />
        </div>
        
        <div class="form-group">
          <label>服务提供商</label>
          <select 
            v-model="form.provider"
            class="form-control"
          >
            <option 
              v-for="provider in providers" 
              :key="provider.value"
              :value="provider.value"
            >
              {{ provider.label }}
            </option>
          </select>
        </div>
        
        <div class="form-group">
          <label>API Token / Key</label>
          <input 
            v-model="form.token"
            type="password" 
            placeholder="输入您的 API Token 或密钥"
            class="form-control"
          />
        </div>
        
        <div class="form-group">
          <label>更新间隔（秒）</label>
          <input 
            v-model.number="form.interval"
            type="number" 
            min="60"
            step="60"
            class="form-control"
          />
          <small class="help-text">建议设置不少于 300 秒（5 分钟）</small>
        </div>
        
        <div class="form-group checkbox-group">
          <label>
            <input 
              type="checkbox"
              v-model="form.enabled"
            />
            立即启用此域名
          </label>
        </div>
      </div>
      
      <div class="modal-footer">
        <button class="btn-secondary" @click="closeModal">取消</button>
        <button 
          :class="['btn-primary', isEditing ? '' : 'btn-success']"
          @click="isEditing ? saveEdit() : addDomain()"
        >
          {{ isEditing ? '保存修改' : '添加' }}
        </button>
      </div>
    </div>
  </div>
</div>
</template>

<style scoped>
.domain-list {
  max-width: 1200px;
  margin: 0 auto;
}

.section-title {
  font-size: 24px;
  margin-bottom: 20px;
  color: #fff;
}

.toolbar {
  display: flex;
  align-items: center;
  gap: 15px;
  margin-bottom: 20px;
}

.btn-primary {
  padding: 10px 20px;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  border: none;
  border-radius: 6px;
  color: white;
  cursor: pointer;
  font-size: 14px;
  transition: all 0.3s ease;
}

.btn-primary:hover {
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(102, 126, 234, 0.4);
}

.count {
  color: rgba(255, 255, 255, 0.6);
  font-size: 14px;
}

.cards-container {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(350px, 1fr));
  gap: 20px;
}

.card {
  background: rgba(255, 255, 255, 0.05);
  border-radius: 12px;
  padding: 20px;
  border: 1px solid rgba(255, 255, 255, 0.1);
  transition: all 0.3s ease;
}

.card:hover:not(.disabled) {
  transform: translateY(-4px);
  box-shadow: 0 8px 24px rgba(0, 0, 0, 0.3);
}

.card.disabled {
  opacity: 0.6;
  filter: grayscale(0.5);
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 15px;
  padding-bottom: 15px;
  border-bottom: 1px solid rgba(255, 255, 255, 0.1);
}

.domain-name {
  margin: 0;
  font-size: 18px;
  color: #fff;
  word-break: break-all;
}

.provider-badge {
  display: inline-block;
  background: rgba(102, 126, 234, 0.2);
  color: #a78bfa;
  padding: 4px 8px;
  border-radius: 4px;
  font-size: 12px;
  margin-top: 6px;
}

.actions {
  display: flex;
  gap: 8px;
}

.toggle-btn {
  width: 32px;
  height: 32px;
  border-radius: 6px;
  border: 1px solid rgba(255, 255, 255, 0.2);
  background: rgba(255, 255, 255, 0.05);
  color: rgba(255, 255, 255, 0.6);
  cursor: pointer;
  font-size: 16px;
  transition: all 0.3s ease;
}

.toggle-btn.active {
  background: linear-gradient(135deg, #48bb78 0%, #38a169 100%);
  border-color: transparent;
  color: white;
}

.action-btn {
  width: 32px;
  height: 32px;
  border-radius: 6px;
  border: 1px solid rgba(255, 255, 255, 0.2);
  background: rgba(255, 255, 255, 0.05);
  color: rgba(255, 255, 255, 0.8);
  cursor: pointer;
  font-size: 16px;
  transition: all 0.3s ease;
}

.action-btn:hover {
  background: rgba(255, 255, 255, 0.1);
}

.action-btn.delete:hover {
  background: rgba(245, 101, 101, 0.2);
  border-color: #f56565;
}

.card-body {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.info-row {
  display: flex;
  align-items: center;
  gap: 10px;
  font-size: 14px;
}

.label {
  color: rgba(255, 255, 255, 0.5);
  min-width: 80px;
}

.value {
  color: rgba(255, 255, 255, 0.9);
}

.token {
  font-family: 'Courier New', monospace;
}

.status-badge {
  display: inline-block;
  padding: 4px 12px;
  border-radius: 12px;
  font-size: 12px;
}

.status-badge.enabled {
  background: rgba(72, 187, 120, 0.2);
  color: #48bb78;
}

.status-badge.disabled {
  background: rgba(255, 255, 255, 0.1);
  color: rgba(255, 255, 255, 0.5);
}

.empty-state {
  grid-column: 1 / -1;
  text-align: center;
  padding: 60px 20px;
  color: rgba(255, 255, 255, 0.5);
}

.modal-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.7);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
}

.modal {
  background: #1a1f3a;
  border-radius: 12px;
  width: 90%;
  max-width: 500px;
  max-height: 90vh;
  overflow-y: auto;
  border: 1px solid rgba(255, 255, 255, 0.1);
}

.modal-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 20px;
  border-bottom: 1px solid rgba(255, 255, 255, 0.1);
}

.modal-header h3 {
  margin: 0;
  color: #fff;
  font-size: 20px;
}

.close-btn {
  background: none;
  border: none;
  color: rgba(255, 255, 255, 0.6);
  font-size: 28px;
  cursor: pointer;
  line-height: 1;
}

.close-btn:hover {
  color: #fff;
}

.modal-body {
  padding: 20px;
}

.form-group {
  margin-bottom: 20px;
}

.form-group label {
  display: block;
  margin-bottom: 8px;
  color: rgba(255, 255, 255, 0.8);
  font-size: 14px;
}

.form-control {
  width: 100%;
  padding: 10px 12px;
  background: rgba(0, 0, 0, 0.3);
  border: 1px solid rgba(255, 255, 255, 0.2);
  border-radius: 6px;
  color: #fff;
  font-size: 14px;
  font-family: inherit;
}

.form-control:focus {
  outline: none;
  border-color: #667eea;
}

.checkbox-group label {
  display: flex;
  align-items: center;
  gap: 8px;
  cursor: pointer;
  user-select: none;
}

.checkbox-group input[type="checkbox"] {
  width: 16px;
  height: 16px;
  cursor: pointer;
}

.help-text {
  display: block;
  margin-top: 6px;
  color: rgba(255, 255, 255, 0.4);
  font-size: 12px;
}

.modal-footer {
  display: flex;
  justify-content: flex-end;
  gap: 10px;
  padding: 20px;
  border-top: 1px solid rgba(255, 255, 255, 0.1);
}

.btn-secondary {
  padding: 10px 20px;
  background: rgba(255, 255, 255, 0.1);
  border: 1px solid rgba(255, 255, 255, 0.2);
  border-radius: 6px;
  color: #fff;
  cursor: pointer;
  font-size: 14px;
  transition: all 0.3s ease;
}

.btn-secondary:hover {
  background: rgba(255, 255, 255, 0.15);
}

.btn-success {
  background: linear-gradient(135deg, #48bb78 0%, #38a169 100%);
}
</style>
