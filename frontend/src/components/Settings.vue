<script lang="ts" setup>
import { ref, reactive, onMounted } from 'vue'
import { GetSettings as BackendGetSettings, SaveSettings as BackendSaveSettings, ResetSettings as BackendResetSettings } from '../../wailsjs/go/app/App'
import { notifySuccess, notifyError } from '../utils/notifications'

const settings = ref({
  autoStart: true,
  checkInterval: 300,
  retryCount: 3,
  retryDelay: 10,
  logLevel: 'info',
  notifySuccess: false,
  notifyError: true,
  proxy: '',
  apiEndpoint: 'https://api.ipify.org'
})

const isModified = ref(false)
const isLoading = ref(true)

// 从数据库加载设置
const loadSettings = async () => {
  try {
    isLoading.value = true
    const result = await BackendGetSettings()
    settings.value = result || settings.value
  } catch (error) {
    console.error('加载设置失败:', error)
    notifyError('加载设置失败')
  } finally {
    isLoading.value = false
  }
}

// 保存设置到数据库
const saveSettings = async () => {
  try {
    await BackendSaveSettings(settings.value)
    isModified.value = false
    notifySuccess('设置已保存！')
  } catch (error) {
    console.error('保存设置失败:', error)
    notifyError('保存失败：' + (error as Error).message)
  }
}

// 重置为默认设置
const resetSettings = async () => {
  if (confirm('确定要重置为默认设置吗？')) {
    try {
      await BackendResetSettings()
      await loadSettings() // 重新加载默认设置
      notifySuccess('设置已重置为默认值')
    } catch (error) {
      console.error('重置设置失败:', error)
      notifyError('重置失败：' + (error as Error).message)
    }
  }
}

const handleSettingChange = () => {
  isModified.value = true
}

// 组件挂载时加载设置
onMounted(() => {
  loadSettings()
})
</script>

<template>
  <div class="settings-container">
    <h2 class="section-title">⚙️ 设置</h2>
    
    <div class="settings-card">
      <h3 class="card-title">🚀 基本设置</h3>
      
      <div class="setting-item">
        <div class="setting-info">
          <label class="setting-label">自动启动</label>
          <p class="setting-desc">应用启动后自动开始 DDNS 服务</p>
        </div>
        <label class="switch">
          <input type="checkbox" v-model="settings.autoStart" @change="handleSettingChange" />
          <span class="slider"></span>
        </label>
      </div>
      
      <div class="setting-item">
        <div class="setting-info">
          <label class="setting-label">检查间隔（秒）</label>
          <p class="setting-desc">每隔多久检查一次 IP 地址变化</p>
        </div>
        <input 
          type="number" 
          v-model.number="settings.checkInterval"
          @input="handleSettingChange"
          min="60"
          step="60"
          class="setting-input"
        />
      </div>
    </div>
    
    <div class="settings-card">
      <h3 class="card-title">🔄 重试设置</h3>
      
      <div class="setting-item">
        <div class="setting-info">
          <label class="setting-label">最大重试次数</label>
          <p class="setting-desc">更新失败时的最大重试次数</p>
        </div>
        <input 
          type="number" 
          v-model.number="settings.retryCount"
          @input="handleSettingChange"
          min="0"
          max="10"
          class="setting-input"
        />
      </div>
      
      <div class="setting-item">
        <div class="setting-info">
          <label class="setting-label">重试延迟（秒）</label>
          <p class="setting-desc">每次重试之间的延迟时间</p>
        </div>
        <input 
          type="number" 
          v-model.number="settings.retryDelay"
          @input="handleSettingChange"
          min="5"
          step="5"
          class="setting-input"
        />
      </div>
    </div>
    
    <div class="settings-card">
      <h3 class="card-title">📢 通知设置</h3>
      
      <div class="setting-item">
        <div class="setting-info">
          <label class="setting-label">成功时通知</label>
          <p class="setting-desc">IP 更新成功时显示通知</p>
        </div>
        <label class="switch">
          <input type="checkbox" v-model="settings.notifySuccess" @change="handleSettingChange" />
          <span class="slider"></span>
        </label>
      </div>
      
      <div class="setting-item">
        <div class="setting-info">
          <label class="setting-label">失败时通知</label>
          <p class="setting-desc">IP 更新失败时显示通知</p>
        </div>
        <label class="switch">
          <input type="checkbox" v-model="settings.notifyError" @change="handleSettingChange" />
          <span class="slider"></span>
        </label>
      </div>
    </div>
    
    <div class="settings-card">
      <h3 class="card-title">🌐 网络设置</h3>
      
      <div class="setting-item">
        <div class="setting-info">
          <label class="setting-label">日志级别</label>
          <p class="setting-desc">选择日志记录的详细程度</p>
        </div>
        <select 
          v-model="settings.logLevel"
          @change="handleSettingChange"
          class="setting-select"
        >
          <option value="debug">Debug - 调试信息</option>
          <option value="info">Info - 普通信息</option>
          <option value="warning">Warning - 警告信息</option>
          <option value="error">Error - 错误信息</option>
        </select>
      </div>
      
      <div class="setting-item">
        <div class="setting-info">
          <label class="setting-label">代理服务器</label>
          <p class="setting-desc">可选的 HTTP/HTTPS 代理地址</p>
        </div>
        <input 
          type="text" 
          v-model="settings.proxy"
          @input="handleSettingChange"
          placeholder="http://proxy.example.com:8080"
          class="setting-input"
        />
      </div>
      
      <div class="setting-item">
        <div class="setting-info">
          <label class="setting-label">API 端点</label>
          <p class="setting-desc">用于获取公网 IP 的 API 地址</p>
        </div>
        <input 
          type="text" 
          v-model="settings.apiEndpoint"
          @input="handleSettingChange"
          class="setting-input"
        />
      </div>
    </div>
    
    <!-- Action Buttons -->
    <div class="action-buttons">
      <button class="btn-secondary" @click="resetSettings">
        🔄 重置为默认
      </button>
      <button 
        :class="['btn-primary', isModified ? '' : 'disabled']"
        @click="saveSettings"
        :disabled="!isModified"
      >
        💾 保存设置
      </button>
    </div>
  </div>
</template>

<style scoped>
.settings-container {
  max-width: 800px;
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

.settings-card {
  background: #ffffff;
  border-radius: 12px;
  padding: 20px;
  margin-bottom: 24px;
  border: 1px solid #e8f4ec;
  box-shadow: 0 1px 2px rgba(0, 0, 0, 0.03), 0 4px 8px rgba(72, 187, 120, 0.05);
}

.card-title {
  font-size: 18px;
  margin: 0 0 18px;
  color: #1a202c;
  font-weight: 600;
  padding-bottom: 10px;
  border-bottom: 1px solid #f0fdf4;
  letter-spacing: -0.2px;
}

.setting-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 14px 0;
  border-bottom: 1px solid #fafcfb;
}

.setting-item:last-child {
  border-bottom: none;
}

.setting-info {
  flex: 1;
}

.setting-label {
  display: block;
  font-size: 14px;
  color: #2d3748;
  margin-bottom: 4px;
  font-weight: 600;
}

.setting-desc {
  font-size: 12px;
  color: #a0aec0;
  margin: 0;
  font-weight: 400;
}

.setting-input,
.setting-select {
  width: 260px;
  padding: 9px 12px;
  background: #f7fafc;
  border: 1px solid #e2e8f0;
  border-radius: 8px;
  color: #2d3748;
  font-size: 14px;
  font-family: inherit;
  transition: all 0.3s ease;
}

.setting-input:focus,
.setting-select:focus {
  outline: none;
  border-color: #9ae6b4;
  background: #ffffff;
  box-shadow: 0 0 0 3px rgba(154, 230, 180, 0.15);
}

.setting-select {
  cursor: pointer;
}

/* Toggle Switch */
.switch {
  position: relative;
  display: inline-block;
  width: 48px;
  height: 24px;
}

.switch input {
  opacity: 0;
  width: 0;
  height: 0;
}

.slider {
  position: absolute;
  cursor: pointer;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background-color: #e2e8f0;
  transition: .3s;
  border-radius: 24px;
}

.slider:before {
  position: absolute;
  content: "";
  height: 18px;
  width: 18px;
  left: 3px;
  bottom: 3px;
  background-color: white;
  transition: .3s;
  border-radius: 50%;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
}

input:checked + .slider {
  background: linear-gradient(135deg, #9ae6b4 0%, #68d391 100%);
}

input:checked + .slider:before {
  transform: translateX(24px);
}

.action-buttons {
  display: flex;
  justify-content: flex-end;
  gap: 10px;
  padding: 16px 20px;
  margin-top: 20px;
}

.btn-secondary {
  padding: 10px 20px;
  background: #f7fafc;
  border: 1px solid #e2e8f0;
  border-radius: 8px;
  color: #4a5568;
  cursor: pointer;
  font-size: 14px;
  font-weight: 500;
  transition: all 0.3s ease;
}

.btn-secondary:hover {
  background: #edf2f7;
  border-color: #cbd5e0;
}

.btn-primary {
  padding: 10px 20px;
  background: linear-gradient(135deg, #48bb78 0%, #38a169 100%);
  border: none;
  border-radius: 8px;
  color: white;
  cursor: pointer;
  font-size: 14px;
  font-weight: 500;
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  box-shadow: 0 2px 4px rgba(72, 187, 120, 0.2);
}

.btn-primary:hover:not(.disabled) {
  background: linear-gradient(135deg, #38a169 0%, #2f855a 100%);
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(72, 187, 120, 0.3);
}

.btn-primary.disabled {
  opacity: 0.5;
  cursor: not-allowed;
}
</style>
