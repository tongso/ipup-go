<script lang="ts" setup>
import { ref, computed, onMounted } from 'vue'
import { GetLogs as BackendGetLogs, ClearLogs as BackendClearLogs, ExportLogs as BackendExportLogs } from '../../wailsjs/go/app/App'
import { notifySuccess, notifyError } from '../utils/notifications'

interface LogEntry {
  id: number
  timestamp: string
  level: string // 改为 string 以兼容后端返回的类型
  domain: string
  message: string
}

const logs = ref<LogEntry[]>([])
const isLoading = ref(true)

const filterLevel = ref<'all' | 'info' | 'warning' | 'error' | 'success'>('all')
const searchKeyword = ref('')
const autoScroll = ref(true)

// 从数据库加载日志
const loadLogs = async () => {
  try {
    isLoading.value = true
    const result = await BackendGetLogs(filterLevel.value, searchKeyword.value)
    logs.value = result || []
  } catch (error) {
    console.error('加载日志失败:', error)
    notifyError('加载日志失败')
  } finally {
    isLoading.value = false
  }
}

const filteredLogs = computed(() => {
  return logs.value.filter(log => {
    const matchLevel = filterLevel.value === 'all' || log.level === filterLevel.value
    const matchSearch = searchKeyword.value === '' || 
      log.domain.toLowerCase().includes(searchKeyword.value.toLowerCase()) ||
      log.message.toLowerCase().includes(searchKeyword.value.toLowerCase())
    return matchLevel && matchSearch
  })
})

// 清空日志
const clearLogs = async () => {
  if (!confirm('确定要清空所有日志吗？')) {
    return
  }
  
  try {
    await BackendClearLogs()
    logs.value = [] // 清空本地数据
    notifySuccess('日志已清空', 2000)
  } catch (error) {
    console.error('清空日志失败:', error)
    notifyError('清空失败：' + (error as Error).message)
  }
}

// 导出日志
const exportLogs = async () => {
  try {
    const content = await BackendExportLogs()
    
    const blob = new Blob([content], { type: 'text/plain' })
    const url = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = `ddns-logs-${new Date().toISOString().split('T')[0]}.txt`
    a.click()
    URL.revokeObjectURL(url)
    
    notifySuccess('日志已导出')
  } catch (error) {
    console.error('导出日志失败:', error)
    notifyError('导出失败：' + (error as Error).message)
  }
}

const getLevelIcon = (level: string) => {
  const icons: Record<string, string> = {
    info: 'ℹ️',
    warning: '⚠️',
    error: '❌',
    success: '✅'
  }
  return icons[level] || '📝'
}

const getLevelClass = (level: string) => {
  return `level-${level}`
}

// 监听筛选条件变化，自动重新加载
const debouncedLoadLogs = () => {
  setTimeout(() => {
    loadLogs()
  }, 300)
}

// 组件挂载时加载日志
onMounted(() => {
  loadLogs()
})
</script>

<template>
  <div class="logs-container">
    <h2 class="section-title">📝 日志查看</h2>
    
    <!-- Toolbar -->
    <div class="toolbar">
      <div class="filters">
        <input 
          v-model="searchKeyword"
          type="text" 
          placeholder="🔍 搜索日志..."
          class="search-input"
        />
        
        <select v-model="filterLevel" class="filter-select">
          <option value="all">全部级别</option>
          <option value="success">✅ 成功</option>
          <option value="info">ℹ️ 信息</option>
          <option value="warning">⚠️ 警告</option>
          <option value="error">❌ 错误</option>
        </select>
      </div>
      
      <div class="actions">
        <label class="checkbox-label">
          <input type="checkbox" v-model="autoScroll" />
          自动滚动
        </label>
        <button class="btn-secondary" @click="exportLogs">📥 导出</button>
        <button class="btn-danger" @click="clearLogs">🗑️ 清空</button>
      </div>
    </div>
    
    <!-- Stats -->
    <div class="stats">
      <div class="stat-item">
        <span class="stat-label">总日志数</span>
        <span class="stat-value">{{ logs.length }}</span>
      </div>
      <div class="stat-item success">
        <span class="stat-label">成功</span>
        <span class="stat-value">{{ logs.filter(l => l.level === 'success').length }}</span>
      </div>
      <div class="stat-item info">
        <span class="stat-label">信息</span>
        <span class="stat-value">{{ logs.filter(l => l.level === 'info').length }}</span>
      </div>
      <div class="stat-item warning">
        <span class="stat-label">警告</span>
        <span class="stat-value">{{ logs.filter(l => l.level === 'warning').length }}</span>
      </div>
      <div class="stat-item error">
        <span class="stat-label">错误</span>
        <span class="stat-value">{{ logs.filter(l => l.level === 'error').length }}</span>
      </div>
    </div>
    
    <!-- Logs List -->
    <div class="log-list" id="logsList">
      <div 
        v-for="log in filteredLogs" 
        :key="log.id"
        :class="['log-item', getLevelClass(log.level)]"
      >
        <div class="log-header">
          <span class="log-icon">{{ getLevelIcon(log.level) }}</span>
          <span :class="['log-level', log.level]">{{ log.level.toUpperCase() }}</span>
          <span class="log-time">{{ log.timestamp }}</span>
        </div>
        <div class="log-body">
          <span class="log-domain">{{ log.domain }}</span>
          <span class="log-message">{{ log.message }}</span>
        </div>
      </div>
      
      <div v-if="isLoading" class="empty-state">
      <p>正在加载日志...</p>
    </div>
    <div v-else-if="filteredLogs.length === 0" class="empty-state">
      <p>暂无日志记录</p>
      <small>尝试调整筛选条件或等待新日志</small>
    </div>
    </div>
  </div>
</template>

<style scoped>
.logs-container {
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

.toolbar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 24px;
  gap: 15px;
  flex-wrap: wrap;
}

.filters {
  display: flex;
  gap: 10px;
  flex: 1;
  min-width: 300px;
}

.search-input {
  flex: 1;
  padding: 9px 12px;
  background: #f7fafc;
  border: 1px solid #e2e8f0;
  border-radius: 8px;
  color: #2d3748;
  font-size: 13px;
  transition: all 0.3s ease;
}

.search-input:focus {
  outline: none;
  border-color: #9ae6b4;
  background: #ffffff;
  box-shadow: 0 0 0 3px rgba(154, 230, 180, 0.15);
}

.filter-select {
  padding: 9px 12px;
  background: #f7fafc;
  border: 1px solid #e2e8f0;
  border-radius: 8px;
  color: #2d3748;
  font-size: 13px;
  cursor: pointer;
  transition: all 0.3s ease;
}

.filter-select:focus {
  outline: none;
  border-color: #9ae6b4;
  background: #ffffff;
}

.actions {
  display: flex;
  gap: 10px;
  align-items: center;
}

.checkbox-label {
  display: flex;
  align-items: center;
  gap: 6px;
  color: #4a5568;
  font-size: 13px;
  cursor: pointer;
  font-weight: 500;
}

.btn-secondary {
  padding: 9px 16px;
  background: #f7fafc;
  border: 1px solid #e2e8f0;
  border-radius: 8px;
  color: #4a5568;
  cursor: pointer;
  font-size: 13px;
  font-weight: 500;
  transition: all 0.3s ease;
}

.btn-secondary:hover {
  background: #edf2f7;
  border-color: #cbd5e0;
}

.btn-danger {
  padding: 9px 16px;
  background: linear-gradient(135deg, #fc8181 0%, #f56565 100%);
  border: none;
  border-radius: 8px;
  color: white;
  cursor: pointer;
  font-size: 13px;
  font-weight: 500;
  transition: all 0.3s ease;
  box-shadow: 0 2px 4px rgba(245, 101, 101, 0.2);
}

.btn-danger:hover {
  background: linear-gradient(135deg, #f56565 0%, #e53e3e 100%);
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(245, 101, 101, 0.3);
}

.stats {
  display: flex;
  gap: 12px;
  margin-bottom: 20px;
  flex-wrap: wrap;
}

.stat-item {
  flex: 1;
  min-width: 110px;
  background: #ffffff;
  border-radius: 10px;
  padding: 14px;
  text-align: center;
  border: 1px solid #e8f4ec;
  box-shadow: 0 1px 2px rgba(0, 0, 0, 0.03);
  transition: all 0.3s ease;
}

.stat-item:hover {
  transform: translateY(-2px);
  box-shadow: 0 4px 10px rgba(72, 187, 120, 0.08);
}

.stat-label {
  display: block;
  font-size: 11px;
  color: #a0aec0;
  margin-bottom: 6px;
  font-weight: 500;
  text-transform: uppercase;
  letter-spacing: 0.4px;
}

.stat-value {
  display: block;
  font-size: 26px;
  font-weight: 700;
  color: #2d3748;
}

.stat-item.success .stat-value {
  color: #38a169;
}

.stat-item.info .stat-value {
  color: #3182ce;
}

.stat-item.warning .stat-value {
  color: #dd6b20;
}

.stat-item.error .stat-value {
  color: #e53e3e;
}

.log-list {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.log-item {
  background: #ffffff;
  border-radius: 10px;
  padding: 14px;
  border: 1px solid #e8f4ec;
  box-shadow: 0 1px 2px rgba(0, 0, 0, 0.03);
  transition: all 0.3s ease;
  animation: slideIn 0.3s ease;
}

@keyframes slideIn {
  from {
    opacity: 0;
    transform: translateX(-10px);
  }
  to {
    opacity: 1;
    transform: translateX(0);
  }
}

.log-item:hover {
  box-shadow: 0 3px 10px rgba(0, 0, 0, 0.06);
  border-left: 3px solid #48bb78;
}

.log-header {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-bottom: 6px;
}

.log-level {
  padding: 3px 8px;
  border-radius: 5px;
  font-size: 10px;
  font-weight: 700;
  text-transform: uppercase;
  letter-spacing: 0.4px;
}

.log-level.success {
  background: linear-gradient(135deg, rgba(198, 246, 213, 0.4) 0%, rgba(154, 230, 180, 0.4) 100%);
  color: #22543d;
  border: 1px solid rgba(134, 239, 172, 0.3);
}

.log-level.info {
  background: linear-gradient(135deg, rgba(189, 227, 255, 0.4) 0%, rgba(154, 205, 250, 0.4) 100%);
  color: #2c5282;
  border: 1px solid rgba(100, 181, 246, 0.3);
}

.log-level.warning {
  background: linear-gradient(135deg, rgba(254, 226, 191, 0.4) 0%, rgba(253, 207, 159, 0.4) 100%);
  color: #744210;
  border: 1px solid rgba(251, 191, 36, 0.3);
}

.log-level.error {
  background: linear-gradient(135deg, rgba(254, 202, 202, 0.4) 0%, rgba(252, 165, 165, 0.4) 100%);
  color: #742a2a;
  border: 1px solid rgba(248, 113, 113, 0.3);
}

.log-time {
  font-size: 11px;
  color: #a0aec0;
  font-weight: 500;
  font-family: 'Courier New', monospace;
}

.log-body {
  display: flex;
  gap: 10px;
  font-size: 13px;
  line-height: 1.6;
}

.log-domain {
  color: #48bb78;
  font-weight: 600;
  font-family: 'Courier New', monospace;
  min-width: 140px;
}

.log-message {
  color: #4a5568;
  flex: 1;
}

.empty-state {
  text-align: center;
  padding: 50px 20px;
  background: linear-gradient(135deg, rgba(198, 246, 213, 0.1) 0%, rgba(154, 230, 180, 0.1) 100%);
  border-radius: 14px;
  border: 2px dashed #cbd5e0;
}

.empty-state p {
  margin: 0 0 8px 0;
  font-size: 16px;
  color: #4a5568;
  font-weight: 600;
}

.empty-state small {
  font-size: 13px;
  color: #a0aec0;
}
</style>
