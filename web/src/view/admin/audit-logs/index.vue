<template>
  <div class="audit-logs-container">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>{{ $t('admin.auditLogs.title') }}</span>
        </div>
      </template>

      <!-- 搜索过滤 -->
      <div class="search-filter">
        <el-row :gutter="20">
          <el-col :span="4">
            <el-input
              v-model="searchForm.username"
              :placeholder="$t('admin.auditLogs.searchUsername')"
              clearable
              @keyup.enter="handleSearch"
              @clear="handleSearch"
            >
              <template #prefix>
                <el-icon><User /></el-icon>
              </template>
            </el-input>
          </el-col>
          <el-col :span="4">
            <el-select
              v-model="searchForm.method"
              :placeholder="$t('admin.auditLogs.searchMethod')"
              clearable
              @change="handleSearch"
            >
              <el-option label="GET" value="GET" />
              <el-option label="POST" value="POST" />
              <el-option label="PUT" value="PUT" />
              <el-option label="DELETE" value="DELETE" />
            </el-select>
          </el-col>
          <el-col :span="6">
            <el-date-picker
              v-model="searchForm.dateRange"
              type="daterange"
              range-separator="-"
              :start-placeholder="$t('common.startDate')"
              :end-placeholder="$t('common.endDate')"
              @change="handleSearch"
            />
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

      <!-- 日志列表表格 -->
      <el-table
        v-loading="loading"
        :data="logs"
        style="width: 100%; margin-top: 20px;"
        stripe
        border
      >
        <el-table-column
          :label="$t('admin.auditLogs.time')"
          prop="createdAt"
          width="180"
          align="center"
        >
          <template #default="{ row }">
            {{ formatDate(row.createdAt) }}
          </template>
        </el-table-column>

        <el-table-column
          :label="$t('admin.auditLogs.username')"
          prop="username"
          width="120"
          align="center"
        />

        <el-table-column
          :label="$t('admin.auditLogs.method')"
          prop="method"
          width="100"
          align="center"
        >
          <template #default="{ row }">
            <el-tag :type="getMethodTagType(row.method)" effect="light">
              {{ row.method }}
            </el-tag>
          </template>
        </el-table-column>

        <el-table-column
          :label="$t('admin.auditLogs.path')"
          prop="path"
          min-width="200"
          show-overflow-tooltip
        />

        <el-table-column
          :label="$t('admin.auditLogs.statusCode')"
          prop="statusCode"
          width="100"
          align="center"
        >
          <template #default="{ row }">
            <el-tag :type="row.statusCode >= 400 ? 'danger' : 'success'" effect="light">
              {{ row.statusCode }}
            </el-tag>
          </template>
        </el-table-column>

        <el-table-column
          :label="$t('admin.auditLogs.latency')"
          prop="latency"
          width="100"
          align="center"
        >
          <template #default="{ row }">
            {{ row.latency }}ms
          </template>
        </el-table-column>

        <el-table-column
          :label="$t('admin.auditLogs.clientIP')"
          prop="clientIP"
          width="140"
          align="center"
        />

        <el-table-column
          :label="$t('common.actions')"
          width="100"
          align="center"
          fixed="right"
        >
          <template #default="{ row }">
            <el-button
              type="primary"
              size="small"
              link
              @click="handleViewDetail(row)"
            >
              <el-icon><View /></el-icon>
            </el-button>
          </template>
        </el-table-column>
      </el-table>

      <!-- 分页 -->
      <div class="pagination-container">
        <el-pagination
          v-model:current-page="currentPage"
          v-model:page-size="pageSize"
          :page-sizes="[20, 50, 100, 200]"
          :total="total"
          layout="total, sizes, prev, pager, next, jumper"
          background
          @size-change="handleSizeChange"
          @current-change="handleCurrentChange"
        />
      </div>
    </el-card>

    <!-- 日志详情对话框 -->
    <el-dialog
      v-model="detailDialogVisible"
      :title="$t('admin.auditLogs.detail')"
      width="800px"
    >
      <el-descriptions :column="2" border>
        <el-descriptions-item :label="$t('admin.auditLogs.time')">
          {{ formatDate(currentLog?.createdAt) }}
        </el-descriptions-item>
        <el-descriptions-item :label="$t('admin.auditLogs.username')">
          {{ currentLog?.username }}
        </el-descriptions-item>
        <el-descriptions-item :label="$t('admin.auditLogs.method')">
          <el-tag :type="getMethodTagType(currentLog?.method)">
            {{ currentLog?.method }}
          </el-tag>
        </el-descriptions-item>
        <el-descriptions-item :label="$t('admin.auditLogs.statusCode')">
          <el-tag :type="(currentLog?.statusCode || 0) >= 400 ? 'danger' : 'success'">
            {{ currentLog?.statusCode }}
          </el-tag>
        </el-descriptions-item>
        <el-descriptions-item :label="$t('admin.auditLogs.path')" :span="2">
          {{ currentLog?.path }}
        </el-descriptions-item>
        <el-descriptions-item :label="$t('admin.auditLogs.clientIP')">
          {{ currentLog?.clientIP }}
        </el-descriptions-item>
        <el-descriptions-item :label="$t('admin.auditLogs.latency')">
          {{ currentLog?.latency }}ms
        </el-descriptions-item>
        <el-descriptions-item :label="$t('admin.auditLogs.userAgent')" :span="2">
          {{ currentLog?.userAgent }}
        </el-descriptions-item>
        <el-descriptions-item :label="$t('admin.auditLogs.request')" :span="2">
          <pre class="json-content">{{ currentLog?.request }}</pre>
        </el-descriptions-item>
        <el-descriptions-item :label="$t('admin.auditLogs.response')" :span="2">
          <pre class="json-content">{{ currentLog?.response }}</pre>
        </el-descriptions-item>
      </el-descriptions>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { User, Search, Refresh, View } from '@element-plus/icons-vue'
import { useI18n } from 'vue-i18n'
import { ElMessage } from 'element-plus'
import { getAuditLogs } from '@/api/admin/audit'

const { t } = useI18n()

// 状态变量
const loading = ref(false)
const logs = ref([])
const currentPage = ref(1)
const pageSize = ref(20)
const total = ref(0)
const detailDialogVisible = ref(false)
const currentLog = ref(null)

// 搜索表单
const searchForm = reactive({
  username: '',
  method: '',
  dateRange: []
})

// 获取日志列表
const loadLogs = async () => {
  loading.value = true
  try {
    const params = {
      page: currentPage.value,
      pageSize: pageSize.value,
      ...searchForm,
      startTime: searchForm.dateRange?.[0],
      endTime: searchForm.dateRange?.[1]
    }
    const res = await getAuditLogs(params)
    logs.value = res.data?.list || []
    total.value = res.data?.total || 0
  } catch (error) {
    ElMessage.error(error.message || t('common.loadFailed'))
  } finally {
    loading.value = false
  }
}

// 搜索
const handleSearch = () => {
  currentPage.value = 1
  loadLogs()
}

// 重置
const handleReset = () => {
  searchForm.username = ''
  searchForm.method = ''
  searchForm.dateRange = []
  handleSearch()
}

// 分页大小变化
const handleSizeChange = (val) => {
  pageSize.value = val
  loadLogs()
}

// 页码变化
const handleCurrentChange = (val) => {
  currentPage.value = val
  loadLogs()
}

// 查看详情
const handleViewDetail = (row) => {
  currentLog.value = row
  detailDialogVisible.value = true
}

// 获取方法标签类型
const getMethodTagType = (method) => {
  const tagMap = {
    GET: 'success',
    POST: 'primary',
    PUT: 'warning',
    DELETE: 'danger'
  }
  return tagMap[method] || ''
}

// 格式化日期
const formatDate = (date) => {
  if (!date) return '-'
  return new Date(date).toLocaleString()
}

// 初始化
onMounted(() => {
  loadLogs()
})
</script>

<style scoped>
.audit-logs-container {
  padding: 20px;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.search-filter {
  margin-top: 10px;
}

.pagination-container {
  margin-top: 20px;
  display: flex;
  justify-content: flex-end;
}

.json-content {
  background: #f5f7fa;
  padding: 10px;
  border-radius: 4px;
  max-height: 200px;
  overflow: auto;
  white-space: pre-wrap;
  word-break: break-all;
  font-size: 12px;
}

:deep(.el-table) {
  font-size: 13px;
}
</style>
