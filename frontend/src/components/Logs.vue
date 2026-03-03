<script lang="ts" setup>
import { ref, computed, onMounted } from 'vue'
import { GetLogs as BackendGetLogs, ClearLogs as BackendClearLogs, ExportLogs as BackendExportLogs } from '../../wailsjs/go/app/App'

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
    alert('加载日志失败')
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
    alert('日志已清空')
  } catch (error) {
    console.error('清空日志失败:', error)
    alert('清空失败：' + (error as Error).message)
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
  } catch (error) {
    console.error('导出日志失败:', error)
    alert('导出失败：' + (error as Error).message)
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
    <div class="logs-list" id="logsList">
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
}

.section-title {
  font-size: 24px;
  margin-bottom: 20px;
  color: #fff;
}

.toolbar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
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
  padding: 10px 12px;
  background: rgba(0, 0, 0, 0.3);
  border: 1px solid rgba(255, 255, 255, 0.2);
  border-radius: 6px;
  color: #fff;
  font-size: 14px;
}

.search-input:focus {
  outline: none;
  border-color: #667eea;
}

.filter-select {
  padding: 10px 12px;
  background: rgba(0, 0, 0, 0.3);
  border: 1px solid rgba(255, 255, 255, 0.2);
  border-radius: 6px;
  color: #fff;
  font-size: 14px;
  cursor: pointer;
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
  color: rgba(255, 255, 255, 0.7);
  font-size: 14px;
  cursor: pointer;
}

.btn-secondary {
  padding: 10px 16px;
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

.btn-danger {
  padding: 10px 16px;
  background: rgba(245, 101, 101, 0.2);
  border: 1px solid rgba(245, 101, 101, 0.3);
  border-radius: 6px;
  color: #f56565;
  cursor: pointer;
  font-size: 14px;
  transition: all 0.3s ease;
}

.btn-danger:hover {
  background: rgba(245, 101, 101, 0.3);
}

.stats {
  display: flex;
  gap: 15px;
  margin-bottom: 20px;
  flex-wrap: wrap;
}

.stat-item {
  flex: 1;
  min-width: 100px;
  background: rgba(255, 255, 255, 0.05);
  border-radius: 8px;
  padding: 15px;
  text-align: center;
  border: 1px solid rgba(255, 255, 255, 0.1);
}

.stat-label {
  display: block;
  font-size: 12px;
  color: rgba(255, 255, 255, 0.5);
  margin-bottom: 8px;
}

.stat-value {
  display: block;
  font-size: 24px;
  font-weight: bold;
  color: #fff;
}

.stat-item.success .stat-value {
  color: #48bb78;
}

.stat-item.info .stat-value {
  color: #667eea;
}

.stat-item.warning .stat-value {
  color: #ed8936;
}

.stat-item.error .stat-value {
  color: #f56565;
}

.logs-list {
  display: flex;
  flex-direction: column;
  gap: 10px;
  max-height: calc(100vh - 400px);
  overflow-y: auto;
}

.log-item {
  background: rgba(0, 0, 0, 0.2);
  border-radius: 8px;
  padding: 15px;
  border-left: 4px solid transparent;
  transition: all 0.3s ease;
}

.log-item:hover {
  background: rgba(0, 0, 0, 0.3);
  transform: translateX(4px);
}

.log-item.level-success {
  border-left-color: #48bb78;
}

.log-item.level-info {
  border-left-color: #667eea;
}

.log-item.level-warning {
  border-left-color: #ed8936;
}

.log-item.level-error {
  border-left-color: #f56565;
}

.log-header {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-bottom: 8px;
}

.log-icon {
  font-size: 16px;
}

.log-level {
  font-size: 12px;
  font-weight: 600;
  padding: 2px 8px;
  border-radius: 4px;
}

.log-level.success {
  background: rgba(72, 187, 120, 0.2);
  color: #48bb78;
}

.log-level.info {
  background: rgba(102, 126, 234, 0.2);
  color: #667eea;
}

.log-level.warning {
  background: rgba(237, 137, 54, 0.2);
  color: #ed8936;
}

.log-level.error {
  background: rgba(245, 101, 101, 0.2);
  color: #f56565;
}

.log-time {
  font-size: 12px;
  color: rgba(255, 255, 255, 0.4);
  margin-left: auto;
}

.log-body {
  display: flex;
  gap: 10px;
  align-items: center;
}

.log-domain {
  font-weight: 600;
  color: #a78bfa;
  font-size: 14px;
}

.log-message {
  color: rgba(255, 255, 255, 0.8);
  font-size: 14px;
  flex: 1;
}

.empty-state {
  text-align: center;
  padding: 60px 20px;
  color: rgba(255, 255, 255, 0.5);
}

/* 滚动条样式 */
.logs-list::-webkit-scrollbar {
  width: 8px;
}

.logs-list::-webkit-scrollbar-track {
  background: rgba(0, 0, 0, 0.2);
  border-radius: 4px;
}

.logs-list::-webkit-scrollbar-thumb {
  background: rgba(255, 255, 255, 0.2);
  border-radius: 4px;
}

.logs-list::-webkit-scrollbar-thumb:hover {
  background: rgba(255, 255, 255, 0.3);
}
</style>
