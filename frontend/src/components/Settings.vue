<script lang="ts" setup>
import { ref, onMounted } from 'vue'
import { LoadSettings, SaveSettings as BackendSaveSettings, ResetSettings as BackendResetSettings } from '../../wailsjs/go/app/App'

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
    const result = await LoadSettings()
    settings.value = result || settings.value
  } catch (error) {
    console.error('加载设置失败:', error)
    alert('加载设置失败')
  } finally {
    isLoading.value = false
  }
}

// 保存设置到数据库
const saveSettings = async () => {
  try {
    await BackendSaveSettings(settings.value)
    isModified.value = false
    alert('设置已保存！')
  } catch (error) {
    console.error('保存设置失败:', error)
    alert('保存失败：' + (error as Error).message)
  }
}

// 重置为默认设置
const resetSettings = async () => {
  if (confirm('确定要重置为默认设置吗？')) {
    try {
      await BackendResetSettings()
      await loadSettings() // 重新加载默认设置
      alert('设置已重置为默认值')
    } catch (error) {
      console.error('重置设置失败:', error)
      alert('重置失败：' + (error as Error).message)
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
}

.section-title {
  font-size: 24px;
  margin-bottom: 20px;
  color: #fff;
}

.settings-card {
  background: rgba(255, 255, 255, 0.05);
  border-radius: 12px;
  padding: 20px;
  margin-bottom: 20px;
  border: 1px solid rgba(255, 255, 255, 0.1);
}

.card-title {
  font-size: 18px;
  margin: 0 0 20px;
  color: #fff;
  padding-bottom: 15px;
  border-bottom: 1px solid rgba(255, 255, 255, 0.1);
}

.setting-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 15px 0;
  border-bottom: 1px solid rgba(255, 255, 255, 0.05);
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
  color: #fff;
  margin-bottom: 6px;
  font-weight: 500;
}

.setting-desc {
  font-size: 12px;
  color: rgba(255, 255, 255, 0.5);
  margin: 0;
}

.setting-input,
.setting-select {
  width: 250px;
  padding: 10px 12px;
  background: rgba(0, 0, 0, 0.3);
  border: 1px solid rgba(255, 255, 255, 0.2);
  border-radius: 6px;
  color: #fff;
  font-size: 14px;
  font-family: inherit;
}

.setting-input:focus,
.setting-select:focus {
  outline: none;
  border-color: #667eea;
}

.setting-select {
  cursor: pointer;
}

/* Toggle Switch */
.switch {
  position: relative;
  display: inline-block;
  width: 50px;
  height: 26px;
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
  background-color: rgba(255, 255, 255, 0.2);
  transition: .4s;
  border-radius: 26px;
}

.slider:before {
  position: absolute;
  content: "";
  height: 20px;
  width: 20px;
  left: 3px;
  bottom: 3px;
  background-color: white;
  transition: .4s;
  border-radius: 50%;
}

input:checked + .slider {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
}

input:checked + .slider:before {
  transform: translateX(24px);
}

/* Action Buttons */
.action-buttons {
  display: flex;
  justify-content: space-between;
  gap: 15px;
  margin-top: 30px;
}

.btn-secondary {
  padding: 12px 24px;
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

.btn-primary {
  padding: 12px 24px;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  border: none;
  border-radius: 6px;
  color: white;
  cursor: pointer;
  font-size: 14px;
  transition: all 0.3s ease;
}

.btn-primary:hover:not(.disabled) {
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(102, 126, 234, 0.4);
}

.btn-primary.disabled {
  opacity: 0.5;
  cursor: not-allowed;
}
</style>
