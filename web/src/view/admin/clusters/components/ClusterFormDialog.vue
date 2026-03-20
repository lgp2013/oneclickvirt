<template>
  <el-dialog
    v-model="dialogVisible"
    :title="isEditing ? $t('admin.clusters.editCluster') : $t('admin.clusters.addCluster')"
    width="750px"
    :close-on-click-modal="false"
    destroy-on-close
    @close="handleClose"
  >
    <!-- 步骤条 -->
    <el-steps
      v-if="!isEditing"
      :active="currentStep"
      finish-status="success"
      align-center
      style="margin-bottom: 30px;"
    >
      <el-step :title="$t('admin.clusters.stepType')" :icon="Pointer" />
      <el-step :title="$t('admin.clusters.stepConfig')" :icon="Setting" />
      <el-step :title="$t('admin.clusters.stepConfirm')" :icon="Check" />
    </el-steps>

    <!-- 第一步：选择集群类型 -->
    <div v-show="!isEditing && currentStep === 0" class="step-content">
      <div class="cluster-type-selection">
        <el-row :gutter="30">
          <el-col :span="12">
            <el-card
              shadow="hover"
              :class="['cluster-type-card', { selected: formData.type === 'openstack' }]"
              @click="selectType('openstack')"
            >
              <div class="cluster-type-card-content">
                <div class="cluster-type-icon openstack">
                  <el-icon size="42"><Cloud /></el-icon>
                </div>
                <h3>OpenStack</h3>
                <p>{{ $t('admin.clusters.openstackDesc') }}</p>
              </div>
            </el-card>
          </el-col>
          <el-col :span="12">
            <el-card
              shadow="hover"
              :class="['cluster-type-card', { selected: formData.type === 'k8s' }]"
              @click="selectType('k8s')"
            >
              <div class="cluster-type-card-content">
                <div class="cluster-type-icon k8s">
                  <el-icon size="42"><Grid /></el-icon>
                </div>
                <h3>Kubernetes</h3>
                <p>{{ $t('admin.clusters.k8sDesc') }}</p>
              </div>
            </el-card>
          </el-col>
        </el-row>
      </div>
    </div>

    <!-- 第二步：配置信息 -->
    <div v-show="(isEditing && showConfig) || (!isEditing && currentStep === 1)" class="step-content">
      <el-scrollbar max-height="450px">
        <el-form
          ref="formRef"
          :model="formData"
          :rules="rules"
          label-width="130px"
          class="cluster-form"
        >
          <!-- 基本信息 -->
          <el-divider content-position="left">
            <span class="divider-title">{{ $t('admin.clusters.basicInfo') }}</span>
          </el-divider>
          
          <el-row :gutter="20">
            <el-col :span="12">
              <el-form-item
                :label="$t('admin.clusters.name')"
                prop="name"
              >
                <el-input
                  v-model="formData.name"
                  :placeholder="$t('admin.clusters.namePlaceholder')"
                  maxlength="32"
                  clearable
                />
              </el-form-item>
            </el-col>
            <el-col :span="12">
              <el-form-item
                :label="$t('admin.clusters.status')"
              >
                <el-radio-group v-model="formData.status">
                  <el-radio value="active">{{ $t('common.enabled') }}</el-radio>
                  <el-radio value="inactive">{{ $t('common.disabled') }}</el-radio>
                </el-radio-group>
              </el-form-item>
            </el-col>
          </el-row>

          <el-row :gutter="20">
            <el-col :span="16">
              <el-form-item
                :label="$t('admin.clusters.endpoint')"
                prop="endpoint"
              >
                <el-input
                  v-model="formData.endpoint"
                  :placeholder="$t('admin.clusters.endpointPlaceholder')"
                  clearable
                >
                  <template #prepend>HTTP</template>
                </el-input>
              </el-form-item>
            </el-col>
            <el-col :span="8">
              <el-form-item
                :label="$t('admin.clusters.port')"
              >
                <el-input-number
                  v-model="formData.port"
                  :min="1"
                  :max="65535"
                  style="width: 100%;"
                />
              </el-form-item>
            </el-col>
          </el-row>

          <el-form-item
            :label="$t('admin.clusters.region')"
          >
            <el-input
              v-model="formData.region"
              :placeholder="$t('admin.clusters.regionPlaceholder')"
              clearable
            />
          </el-form-item>

          <!-- OpenStack 专用配置 -->
          <template v-if="formData.type === 'openstack'">
            <el-divider content-position="left">
              <span class="divider-title">{{ $t('admin.clusters.openstackConfig') }}</span>
            </el-divider>

            <el-row :gutter="20">
              <el-col :span="12">
                <el-form-item
                  :label="$t('admin.clusters.projectId')"
                >
                  <el-input
                    v-model="formData.projectId"
                    :placeholder="$t('admin.clusters.projectIdPlaceholder')"
                    clearable
                  />
                </el-form-item>
              </el-col>
              <el-col :span="12">
                <el-form-item
                  :label="$t('admin.clusters.domainId')"
                >
                  <el-input
                    v-model="formData.domainId"
                    :placeholder="$t('admin.clusters.domainIdPlaceholder')"
                    clearable
                  />
                </el-form-item>
              </el-col>
            </el-row>
          </template>

          <!-- K8s 专用配置 -->
          <template v-if="formData.type === 'k8s'">
            <el-divider content-position="left">
              <span class="divider-title">{{ $t('admin.clusters.k8sConfig') }}</span>
            </el-divider>

            <el-form-item
              :label="$t('admin.clusters.namespace')"
            >
              <el-input
                v-model="formData.namespace"
                placeholder="default"
                clearable
              />
            </el-form-item>

            <el-form-item
              :label="$t('admin.clusters.kubeconfig')"
            >
              <el-input
                v-model="formData.kubeconfig"
                type="textarea"
                :rows="4"
                :placeholder="$t('admin.clusters.kubeconfigPlaceholder')"
              />
            </el-form-item>
          </template>

          <!-- 认证信息 -->
          <el-divider content-position="left">
            <span class="divider-title">{{ $t('admin.clusters.authInfo') }}</span>
          </el-divider>

          <el-row :gutter="20">
            <el-col :span="12">
              <el-form-item
                :label="$t('admin.clusters.username')"
                prop="username"
              >
                <el-input
                  v-model="formData.username"
                  :placeholder="$t('admin.clusters.usernamePlaceholder')"
                  clearable
                />
              </el-form-item>
            </el-col>
            <el-col :span="12">
              <el-form-item
                :label="$t('admin.clusters.password')"
              >
                <el-input
                  v-model="formData.password"
                  type="password"
                  show-password
                  :placeholder="$t('admin.clusters.passwordPlaceholder')"
                />
              </el-form-item>
            </el-col>
          </el-row>

          <el-form-item
            :label="$t('admin.clusters.privateKey')"
          >
            <el-input
              v-model="formData.privateKey"
              type="textarea"
              :rows="3"
              :placeholder="$t('admin.clusters.privateKeyPlaceholder')"
            />
          </el-form-item>

          <!-- 描述 -->
          <el-divider content-position="left">
            <span class="divider-title">{{ $t('common.description') }}</span>
          </el-divider>

          <el-form-item
            :label="$t('common.description')"
          >
            <el-input
              v-model="formData.description"
              type="textarea"
              :rows="2"
              maxlength="200"
              show-word-limit
            />
          </el-form-item>
        </el-form>
      </el-scrollbar>
    </div>

    <!-- 第三步：确认 -->
    <div v-show="!isEditing && currentStep === 2" class="step-content">
      <el-result
        icon="success"
        :title="formData.name"
      >
        <template #sub-title>
          <el-descriptions :column="2" border style="margin-top: 20px;">
            <el-descriptions-item :label="$t('admin.clusters.type')">
              <el-tag :type="formData.type === 'openstack' ? 'warning' : 'primary'" effect="light">
                {{ formData.type === 'openstack' ? 'OpenStack' : 'Kubernetes' }}
              </el-tag>
            </el-descriptions-item>
            <el-descriptions-item :label="$t('admin.clusters.status')">
              <el-tag :type="formData.status === 'active' ? 'success' : 'info'" effect="light">
                {{ formData.status === 'active' ? $t('common.enabled') : $t('common.disabled') }}
              </el-tag>
            </el-descriptions-item>
            <el-descriptions-item :label="$t('admin.clusters.endpoint')" :span="2">
              {{ formData.endpoint }}:{{ formData.port }}
            </el-descriptions-item>
            <el-descriptions-item :label="$t('admin.clusters.region')">
              {{ formData.region || '-' }}
            </el-descriptions-item>
            <el-descriptions-item :label="$t('admin.clusters.username')">
              {{ formData.username }}
            </el-descriptions-item>
            <el-descriptions-item v-if="formData.type === 'openstack'" :label="$t('admin.clusters.projectId')" :span="2">
              {{ formData.projectId || '-' }}
            </el-descriptions-item>
            <el-descriptions-item v-if="formData.type === 'k8s'" :label="$t('admin.clusters.namespace')">
              {{ formData.namespace || 'default' }}
            </el-descriptions-item>
          </el-descriptions>
        </template>
      </el-result>
    </div>

    <!-- 底部按钮 -->
    <template #footer>
      <div class="dialog-footer">
        <el-button @click="handleClose">
          {{ $t('common.cancel') }}
        </el-button>
        
        <template v-if="!isEditing">
          <el-button
            v-if="currentStep > 0"
            @click="handlePrevStep"
          >
            <el-icon><ArrowLeft /></el-icon>
            {{ $t('common.prevStep') }}
          </el-button>
          
          <el-button
            v-if="currentStep < 2"
            type="primary"
            @click="handleNextStep"
          >
            {{ $t('common.nextStep') }}
            <el-icon><ArrowRight /></el-icon>
          </el-button>
          
          <el-button
            v-if="currentStep === 2"
            type="primary"
            :loading="loading"
            @click="handleSubmit"
          >
            <el-icon><Check /></el-icon>
            {{ $t('common.confirm') }}
          </el-button>
        </template>
        
        <template v-else>
          <el-button
            type="primary"
            :loading="loading"
            @click="handleSubmit"
          >
            {{ $t('common.confirm') }}
          </el-button>
        </template>
      </div>
    </template>
  </el-dialog>
</template>

<script setup>
import { ref, reactive, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { Cloud, Grid, ArrowLeft, ArrowRight, Check, Pointer, Setting } from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'

const props = defineProps({
  visible: {
    type: Boolean,
    default: false
  },
  isEditing: {
    type: Boolean,
    default: false
  },
  clusterData: {
    type: Object,
    default: () => ({})
  },
  loading: {
    type: Boolean,
    default: false
  }
})

const emit = defineEmits(['update:visible', 'submit', 'cancel'])

const { t } = useI18n()

// 表单引用
const formRef = ref()

// 当前步骤
const currentStep = ref(0)
const showConfig = ref(true)

// 表单数据
const formData = reactive({
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
  description: '',
  kubeconfig: '',
  namespace: 'default'
})

// 表单验证规则
const rules = {
  name: [
    { required: true, message: t('admin.clusters.nameRequired'), trigger: 'blur' }
  ],
  endpoint: [
    { required: true, message: t('admin.clusters.endpointRequired'), trigger: 'blur' }
  ],
  username: [
    { required: true, message: t('admin.clusters.usernameRequired'), trigger: 'blur' }
  ]
}

// 对话框可见性
const dialogVisible = ref(false)

watch(() => props.visible, (val) => {
  dialogVisible.value = val
  if (val) {
    if (props.isEditing && props.clusterData) {
      Object.assign(formData, props.clusterData)
      currentStep.value = 1
      showConfig.value = true
    } else {
      resetForm()
    }
  }
})

watch(dialogVisible, (val) => {
  emit('update:visible', val)
})

// 选择集群类型
const selectType = (type) => {
  formData.type = type
}

// 下一步
const handleNextStep = async () => {
  if (currentStep.value === 0) {
    if (!formData.type) {
      ElMessage.warning('请选择集群类型')
      return
    }
    currentStep.value = 1
  } else if (currentStep.value === 1) {
    try {
      await formRef.value.validate()
      currentStep.value = 2
    } catch (error) {
      return
    }
  }
}

// 上一步
const handlePrevStep = () => {
  if (currentStep.value > 0) {
    currentStep.value--
  }
}

// 提交
const handleSubmit = async () => {
  try {
    if (currentStep.value === 1) {
      await formRef.value.validate()
    }
    emit('submit', { ...formData })
  } catch (error) {
    // 表单验证失败
  }
}

// 关闭
const handleClose = () => {
  dialogVisible.value = false
  emit('cancel')
  resetForm()
}

// 重置表单
const resetForm = () => {
  currentStep.value = 0
  showConfig.value = false
  Object.assign(formData, {
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
    description: '',
    kubeconfig: '',
    namespace: 'default'
  })
}
</script>

<style scoped>
.step-content {
  min-height: 300px;
}

.cluster-type-selection {
  padding: 20px 10px;
}

.cluster-type-card {
  cursor: pointer;
  transition: all 0.3s;
  border: 2px solid transparent;
}

.cluster-type-card:hover {
  transform: translateY(-5px);
}

.cluster-type-card.selected {
  border-color: #409EFF;
  background-color: #ecf5ff;
}

.cluster-type-card-content {
  text-align: center;
  padding: 20px 10px;
}

.cluster-type-icon {
  width: 80px;
  height: 80px;
  border-radius: 50%;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  margin-bottom: 15px;
}

.cluster-type-icon.openstack {
  background: linear-gradient(135deg, #409EFF 0%, #67C23A 100%);
  color: white;
}

.cluster-type-icon.k8s {
  background: linear-gradient(135deg, #67C23A 0%, #409EFF 100%);
  color: white;
}

.cluster-type-card h3 {
  margin: 10px 0;
  font-size: 18px;
  color: #303133;
}

.cluster-type-card p {
  color: #909399;
  font-size: 13px;
  line-height: 1.5;
  min-height: 40px;
}

.cluster-form {
  padding: 10px;
}

.divider-title {
  font-size: 14px;
  font-weight: 600;
  color: #303133;
}

.dialog-footer {
  display: flex;
  justify-content: flex-end;
  gap: 10px;
}

:deep(.el-form-item) {
  margin-bottom: 18px;
}

:deep(.el-divider--horizontal) {
  margin: 20px 0 15px;
}
</style>
