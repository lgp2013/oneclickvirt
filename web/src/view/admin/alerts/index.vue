<template>
  <div class="alerts-container">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>{{ $t('admin.alerts.title') }}</span>
          <div class="header-actions">
            <el-badge :value="unreadCount" :hidden="unreadCount === 0" :max="99">
              <el-button @click="handleMarkAllRead">
                {{ $t('admin.alerts.markAllRead') }}
              </el-button>
            </el-badge>
          </div>
        </div>
      </template>

      <!-- 搜索过滤 -->
      <div class="search-filter">
        <el-row :gutter="20">
          <el-col :span="4">
            <el-select
              v-model="searchForm.level"
              :placeholder="$t('admin.alerts.searchLevel')"
              clearable
              @change="handleSearch"
            >
              <el-option :label="$t('admin.alerts.levelInfo')" value="info" />
              <el-option :label="$t('admin.alerts.levelWarning')" value="warning" />
              <el-option :label="$t('admin.alerts.levelError')" value="error" />
              <el-option :label="$t('admin.alerts.levelCritical')" value="critical" />
            </el-select>
          </el-col>
          <el-col :span="4">
            <el-select
              v-model="searchForm.status"
              :placeholder="$t('admin.alerts.searchStatus')"
              clearable
              @change="handleSearch"
            >
              <el-option :label="$t('admin.alerts.statusUnread')" value="unread" />
              <el-option :label="$t('admin.alerts.statusRead')" value="read" />
              <el-option :label="$t('admin.alerts.statusResolved')" value="resolved" />
            </el-select>
          </el-col>
          <el-col :span="4">
            <el-select
              v-model="searchForm.type"
              :placeholder="$t('admin.alerts.searchType')"
              clearable
              @change="handleSearch"
            >
              <el-option :label="$t('admin.alerts.typeSystem')" value="system" />
              <el-option :label="$t('admin.alerts.typeInstance')" value="instance" />
              <el-option :label="$t('admin.alerts.typeCluster')" value="cluster" />
              <el-option :label="$t('admin.alerts.typeUser')" value="user" />
            </el-select>
          </el-col>
          <el-col :span="4">
            <el-button type="primary" @click="handleSearch">
              <el-icon><Search /></el-icon>
              {{ $t('common.search') }}
            </el-button>
            <el-button @click="handleReset">
              <el-icon><Refresh /></el-icon>
              {{ $t('common.reset') }}
            </el-button>
          </el-col>
        </el-row>
      </div>

      <!-- 告警列表 -->
      <el-table
        v-loading="loading"
        :data="alerts"
        style="width: 100%; margin-top: 20px;"
        @row-click="handleRowClick"
        :row-class-name="getRowClassName"
      >
        <el-table-column
          :label="$t('admin.alerts.level')"
          prop="level"
          width="100"
          align="center"
        >
          <template #default="{ row }">
            <el-tag :type="getLevelTagType(row.level)" effect="dark">
              {{ getLevelName(row.level) }}
            </el-tag>
          </template>
        </el-table-column>

        <el-table-column
          :label="$t('admin.alerts.title')"
          prop="title"
          min-width="200"
          show-overflow-tooltip
        />

        <el-table-column
          :label="$t('admin.alerts.content')"
          prop="content"
          min-width="250"
          show-overflow-tooltip
        />

        <el-table-column
          :label="$t('admin.alerts.type')"
          prop="type"
          width="100"
          align="center"
        >
          <template #default="{ row }">
            <el-tag effect="light">{{ getTypeName(row.type) }}</el-tag>
          </template>
        </el-table-column>

        <el-table-column
          :label="$t('admin.alerts.status')"
          prop="status"
          width="100"
          align="center"
        >
          <template #default="{ row }">
            <el-tag :type="row.isRead ? 'info' : 'danger'" effect="light">
              {{ row.isRead ? $t('admin.alerts.statusRead') : $t('admin.alerts.statusUnread') }}
            </el-tag>
          </template>
        </el-table-column>

        <el-table-column
          :label="$t('admin.alerts.time')"
          prop="createdAt"
          width="180"
          align="center"
        >
          <template #default="{ row }">
            {{ formatDate(row.createdAt) }}
          </template>
        </el-table-column>

        <el-table-column
          :label="$t('common.actions')"
          width="150"
          align="center"
          fixed="right"
        >
          <template #default="{ row }">
            <el-button
              v-if="!row.isRead"
              type="primary"
              size="small"
              link
              @click.stop="handleMarkRead(row)"
            >
              <el-icon><Check /></el-icon>
              {{ $t('admin.alerts.markRead') }}
            </el-button>
            <el-button
              v-if="row.status !== 'resolved'"
              type="success"
              size="small"
              link
              @click.stop="handleResolve(row)"
            >
              <el-icon><CircleCheck /></el-icon>
              {{ $t('admin.alerts.resolve') }}
            </el-button>
            <el-button
              type="danger"
              size="small"
              link
              @click.stop="handleDelete(row)"
            >
              <el-icon><Delete /></el-icon>
            </el-button>
          </template>
        </el-table-column>
      </el-table>

      <!-- 分页 -->
      <div class="pagination-container">
        <el-pagination
          v-model:current-page="currentPage"
          v-model:page-size="pageSize"
          :page-sizes="[10, 20, 50, 100]"
          :total="total"
          layout="total, sizes, prev, pager, next, jumper"
          background
          @size-change="handleSizeChange"
          @current-change="handleCurrentChange"
        />
      </div>
    </el-card>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted, onUnmounted } from 'vue'
import { Search, Refresh, Check, CircleCheck, Delete } from '@element-plus/icons-vue'
import { useI18n } from 'vue-i18n'
import { ElMessage, ElMessageBox } from 'element-plus'
import { getAlerts, markAsRead, markAllAsRead, resolveAlert, deleteAlert, getUnreadCount } from '@/api/admin/alerts'

const { t } = useI18n()

// 状态变量
const loading = ref(false)
const alerts = ref([])
const currentPage = ref(1)
const pageSize = ref(10)
const total = ref(0)
const unreadCount = ref(0)
let refreshTimer = null

// 搜索表单
const searchForm = reactive({
  level: '',
  status: '',
  type: ''
})

// 加载告警列表
const loadAlerts = async () => {
  loading.value = true
  try {
    const params = {
      page: currentPage.value,
      pageSize: pageSize.value,
      ...searchForm
    }
    const res = await getAlerts(params)
    alerts.value = res.data?.list || []
    total.value = res.data?.total || 0
  } catch (error) {
    ElMessage.error(error.message || t('common.loadFailed'))
  } finally {
    loading.value = false
  }
}

// 加载未读数量
const loadUnreadCount = async () => {
  try {
    const res = await getUnreadCount()
    unreadCount.value = res.data?.count || 0
  } catch (error) {
    console.error('Failed to load unread count')
  }
}

// 搜索
const handleSearch = () => {
  currentPage.value = 1
  loadAlerts()
}

// 重置
const handleReset = () => {
  searchForm.level = ''
  searchForm.status = ''
  searchForm.type = ''
  handleSearch()
}

// 分页大小变化
const handleSizeChange = (val) => {
  pageSize.value = val
  loadAlerts()
}

// 页码变化
const handleCurrentChange = (val) => {
  currentPage.value = val
  loadAlerts()
}

// 标记已读
const handleMarkRead = async (row) => {
  try {
    await markAsRead({ id: row.id })
    ElMessage.success(t('admin.alerts.markReadSuccess'))
    loadAlerts()
    loadUnreadCount()
  } catch (error) {
    ElMessage.error(error.message || t('common.operationFailed'))
  }
}

// 标记全部已读
const handleMarkAllRead = async () => {
  try {
    await markAllAsRead()
    ElMessage.success(t('admin.alerts.markAllReadSuccess'))
    loadAlerts()
    loadUnreadCount()
  } catch (error) {
    ElMessage.error(error.message || t('common.operationFailed'))
  }
}

// 解决告警
const handleResolve = async (row) => {
  try {
    await resolveAlert({ id: row.id })
    ElMessage.success(t('admin.alerts.resolveSuccess'))
    loadAlerts()
  } catch (error) {
    ElMessage.error(error.message || t('common.operationFailed'))
  }
}

// 删除告警
const handleDelete = async (row) => {
  try {
    await ElMessageBox.confirm(
      t('common.deleteConfirm'),
      t('common.warning'),
      {
        confirmButtonText: t('common.confirm'),
        cancelButtonText: t('common.cancel'),
        type: 'warning'
      }
    )
    await deleteAlert({ id: row.id })
    ElMessage.success(t('common.deleteSuccess'))
    loadAlerts()
    loadUnreadCount()
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error(error.message || t('common.deleteFailed'))
    }
  }
}

// 行点击
const handleRowClick = (row) => {
  if (!row.isRead) {
    handleMarkRead(row)
  }
}

// 获取行样式
const getRowClassName = ({ row }) => {
  return row.isRead ? '' : 'unread-row'
}

// 获取级别标签类型
const getLevelTagType = (level) => {
  const tagMap = {
    info: 'info',
    warning: 'warning',
    error: 'danger',
    critical: 'danger'
  }
  return tagMap[level] || 'info'
}

// 获取级别名称
const getLevelName = (level) => {
  const nameMap = {
    info: t('admin.alerts.levelInfo'),
    warning: t('admin.alerts.levelWarning'),
    error: t('admin.alerts.levelError'),
    critical: t('admin.alerts.levelCritical')
  }
  return nameMap[level] || level
}

// 获取类型名称
const getTypeName = (type) => {
  const nameMap = {
    system: t('admin.alerts.typeSystem'),
    instance: t('admin.alerts.typeInstance'),
    cluster: t('admin.alerts.typeCluster'),
    user: t('admin.alerts.typeUser')
  }
  return nameMap[type] || type
}

// 格式化日期
const formatDate = (date) => {
  if (!date) return '-'
  return new Date(date).toLocaleString()
}

// 初始化
onMounted(() => {
  loadAlerts()
  loadUnreadCount()
  
  // 每30秒刷新一次
  refreshTimer = setInterval(() => {
    loadUnreadCount()
  }, 30000)
})

onUnmounted(() => {
  if (refreshTimer) {
    clearInterval(refreshTimer)
  }
})
</script>

<style scoped>
.alerts-container {
  padding: 20px;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.header-actions {
  display: flex;
  gap: 10px;
}

.search-filter {
  margin-top: 10px;
}

.pagination-container {
  margin-top: 20px;
  display: flex;
  justify-content: flex-end;
}

:deep(.el-table) {
  cursor: pointer;
}

:deep(.unread-row) {
  background-color: #fff5f5;
}

:deep(.el-table tr.unread-row:hover > td) {
  background-color: #ffeaea !important;
}
</style>
