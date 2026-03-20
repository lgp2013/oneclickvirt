<template>
  <div class="clusters-container">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>{{ $t('admin.clusters.title') }}</span>
          <el-button
            type="primary"
            @click="handleAddCluster"
          >
            {{ $t('admin.clusters.addCluster') }}
          </el-button>
        </div>
      </template>

      <!-- 搜索过滤 -->
      <div class="search-filter">
        <el-row :gutter="20">
          <el-col :span="6">
            <el-input
              v-model="searchForm.name"
              :placeholder="$t('admin.clusters.searchName')"
              clearable
              @keyup.enter="handleSearch"
              @clear="handleSearch"
            >
              <template #prefix>
                <el-icon><Search /></el-icon>
              </template>
            </el-input>
          </el-col>
          <el-col :span="4">
            <el-select
              v-model="searchForm.type"
              :placeholder="$t('admin.clusters.searchType')"
              clearable
              @change="handleSearch"
              @clear="handleSearch"
            >
              <el-option
                label="OpenStack"
                value="openstack"
              />
              <el-option
                label="Kubernetes"
                value="k8s"
              />
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

      <!-- 集群列表表格 -->
      <el-table
        v-loading="loading"
        :data="clusters"
        style="width: 100%; margin-top: 20px;"
        @selection-change="handleSelectionChange"
      >
        <el-table-column type="selection" width="55" />
        
        <el-table-column
          :label="$t('admin.clusters.name')"
          prop="name"
          min-width="120"
          show-overflow-tooltip
        />
        
        <el-table-column
          :label="$t('admin.clusters.type')"
          prop="type"
          width="130"
          align="center"
        >
          <template #default="{ row }">
            <el-tag :type="getTypeTagType(row.type)" effect="light">
              {{ getTypeName(row.type) }}
            </el-tag>
          </template>
        </el-table-column>

        <el-table-column
          :label="$t('admin.clusters.endpoint')"
          prop="endpoint"
          min-width="180"
          show-overflow-tooltip
        />

        <el-table-column
          :label="$t('admin.clusters.status')"
          prop="status"
          width="100"
          align="center"
        >
          <template #default="{ row }">
            <el-tag :type="row.status === 'active' ? 'success' : 'info'" effect="light">
              {{ row.status === 'active' ? $t('common.enabled') : $t('common.disabled') }}
            </el-tag>
          </template>
        </el-table-column>

        <el-table-column
          :label="$t('admin.clusters.region')"
          prop="region"
          width="120"
          align="center"
        />

        <el-table-column
          :label="$t('admin.clusters.instances')"
          prop="instanceCount"
          width="100"
          align="center"
        >
          <template #default="{ row }">
            <el-link type="primary" @click="handleViewInstances(row)">
              {{ row.instanceCount || 0 }}
            </el-link>
          </template>
        </el-table-column>

        <el-table-column
          :label="$t('common.createTime')"
          prop="createdAt"
          width="170"
          align="center"
        >
          <template #default="{ row }">
            {{ formatDate(row.createdAt) }}
          </template>
        </el-table-column>

        <el-table-column
          :label="$t('common.actions')"
          width="200"
          fixed="right"
          align="center"
        >
          <template #default="{ row }">
            <el-button
              type="primary"
              size="small"
              link
              @click="handleEdit(row)"
            >
              <el-icon><Edit /></el-icon>
              {{ $t('common.edit') }}
            </el-button>
            <el-button
              type="success"
              size="small"
              link
              @click="handleConnect(row)"
            >
              <el-icon><Connection /></el-icon>
              {{ $t('admin.clusters.connect') }}
            </el-button>
            <el-button
              type="danger"
              size="small"
              link
              @click="handleDelete(row)"
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

    <!-- 添加/编辑集群对话框 -->
    <ClusterFormDialog
      v-model:visible="showDialog"
      :is-editing="isEditing"
      :cluster-data="clusterForm"
      :loading="formLoading"
      @submit="handleSubmit"
      @cancel="handleCancel"
    />
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { Search, Refresh, Edit, Delete, Connection } from '@element-plus/icons-vue'
import { useI18n } from 'vue-i18n'
import { ElMessage, ElMessageBox } from 'element-plus'
import ClusterFormDialog from './components/ClusterFormDialog.vue'
import { getClusterList, createCluster, updateCluster, deleteCluster, testConnection } from '@/api/admin/clusters'

const { t } = useI18n()

// 状态变量
const loading = ref(false)
const clusters = ref([])
const currentPage = ref(1)
const pageSize = ref(10)
const total = ref(0)
const selectedClusters = ref([])

// 搜索表单
const searchForm = reactive({
  name: '',
  type: ''
})

// 对话框状态
const showDialog = ref(false)
const isEditing = ref(false)
const formLoading = ref(false)
const clusterForm = reactive({
  id: null,
  name: '',
  type: 'openstack',
  endpoint: '',
  port: 22,
  username: '',
  password: '',
  privateKey: '',
  projectId: '',
  domainId: '',
  region: '',
  status: 'active',
  description: ''
})

// 加载数据
const loadClusters = async () => {
  loading.value = true
  try {
    const params = {
      page: currentPage.value,
      pageSize: pageSize.value,
      ...searchForm
    }
    const res = await getClusterList(params)
    clusters.value = res.data?.list || []
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
  loadClusters()
}

// 重置
const handleReset = () => {
  searchForm.name = ''
  searchForm.type = ''
  handleSearch()
}

// 选择变化
const handleSelectionChange = (selection) => {
  selectedClusters.value = selection
}

// 分页大小变化
const handleSizeChange = (val) => {
  pageSize.value = val
  loadClusters()
}

// 页码变化
const handleCurrentChange = (val) => {
  currentPage.value = val
  loadClusters()
}

// 添加集群
const handleAddCluster = () => {
  isEditing.value = false
  resetForm()
  showDialog.value = true
}

// 编辑集群
const handleEdit = (row) => {
  isEditing.value = true
  Object.assign(clusterForm, row)
  showDialog.value = true
}

// 删除集群
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
    
    await deleteCluster({ id: row.id })
    ElMessage.success(t('common.deleteSuccess'))
    loadClusters()
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error(error.message || t('common.deleteFailed'))
    }
  }
}

// 连接测试
const handleConnect = async (row) => {
  try {
    const res = await testConnection({ id: row.id })
    if (res.data?.success) {
      ElMessage.success(t('admin.clusters.connectSuccess'))
    } else {
      ElMessage.warning(t('admin.clusters.connectFailed'))
    }
  } catch (error) {
    ElMessage.error(error.message || t('admin.clusters.connectFailed'))
  }
}

// 查看实例
const handleViewInstances = (row) => {
  // 跳转到实例管理页面，并筛选当前集群的实例
  // TODO: 实现跳转逻辑
}

// 提交表单
const handleSubmit = async (formData) => {
  formLoading.value = true
  try {
    if (isEditing.value) {
      await updateCluster(formData)
      ElMessage.success(t('common.updateSuccess'))
    } else {
      await createCluster(formData)
      ElMessage.success(t('common.createSuccess'))
    }
    showDialog.value = false
    loadClusters()
  } catch (error) {
    ElMessage.error(error.message || (isEditing.value ? t('common.updateFailed') : t('common.createFailed')))
  } finally {
    formLoading.value = false
  }
}

// 取消
const handleCancel = () => {
  showDialog.value = false
  resetForm()
}

// 重置表单
const resetForm = () => {
  clusterForm.id = null
  clusterForm.name = ''
  clusterForm.type = 'openstack'
  clusterForm.endpoint = ''
  clusterForm.port = 22
  clusterForm.username = ''
  clusterForm.password = ''
  clusterForm.privateKey = ''
  clusterForm.projectId = ''
  clusterForm.domainId = ''
  clusterForm.region = ''
  clusterForm.status = 'active'
  clusterForm.description = ''
}

// 获取类型名称
const getTypeName = (type) => {
  const typeMap = {
    openstack: 'OpenStack',
    k8s: 'Kubernetes'
  }
  return typeMap[type] || type
}

// 获取类型标签样式
const getTypeTagType = (type) => {
  const tagMap = {
    openstack: 'warning',
    k8s: 'primary'
  }
  return tagMap[type] || ''
}

// 格式化日期
const formatDate = (date) => {
  if (!date) return '-'
  return new Date(date).toLocaleString()
}

// 初始化
onMounted(() => {
  loadClusters()
})
</script>

<style scoped>
.clusters-container {
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

:deep(.el-table) {
  font-size: 14px;
}

:deep(.el-button + .el-button) {
  margin-left: 8px;
}
</style>
