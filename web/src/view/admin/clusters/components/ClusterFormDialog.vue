<template>
  <el-dialog
    v-model="dialogVisible"
    :title="isEditing ? $t('admin.clusters.editCluster') : $t('admin.clusters.addCluster')"
    width="800px"
    :close-on-click-modal="false"
    @close="handleClose"
  >
    <!-- 步骤条 -->
    <el-steps
      v-if="!isEditing"
      :active="currentStep"
      finish-status="success"
      style="margin-bottom: 30px;"
    >
      <el-step :title="$t('admin.clusters.stepType')" />
      <el-step :title="$t('admin.clusters.stepConfig')" />
      <el-step :title="$t('admin.clusters.stepConfirm')" />
    </el-steps>

    <!-- 第一步：选择集群类型 -->
    <div v-show="!isEditing && currentStep === 0">
      <div class="cluster-type-selection">
        <el-row :gutter="20">
          <el-col :span="12">
            <el-card
              shadow="hover"
              :class="{ selected: formData.type === 'openstack' }"
              @click="selectType('openstack')"
            >
              <div class="cluster-type-card">
                <el-icon size="48" color="#409EFF">
                  <Cloud />
                </el-icon>
                <h3>OpenStack</h3>
                <p>{{ $t('admin.clusters.openstackDesc') }}</p>
              </div>
            </el-card>
          </el-col>
          <el-col :span="12">
            <el-card
              shadow="hover"
              :class="{ selected: formData.type === 'k8s' }"
              @click="selectType('k8s')"
            >
              <div class="cluster-type-card">
                <el-icon size="48" color="#67C23A">
                  <Grid />
                </el-icon>
                <h3>Kubernetes</h3>
                <p>{{ $t('admin.clusters.k8sDesc') }}</p>
              </div>
            </el-card>
          </el-col>
        </el-row>
      </div>
    </div>

    <!-- 第二步：配置信息 -->
    <div v-show="(isEditing && showConfig) || (!isEditing && currentStep === 1)">
      <el-form
        ref="formRef"
        :model="formData"
        :rules="rules"
        label-width="140px"
      >
        <!-- 基本信息 -->
        <el-divider content-position="left">
          {{ $t('admin.clusters.basicInfo') }}
        </el-divider>
        
        <el-form-item
          :label="$t('admin.clusters.name')"
          prop="name"
        >
          <el-input
            v-model="formData.name"
            :placeholder="$t('admin.clusters.namePlaceholder')"
            maxlength="32"
          />
        </el-form-item>

        <el-form-item
          :label="$t('admin.clusters.endpoint')"
          prop="endpoint"
        >
          <el-input
            v-model="formData.endpoint"
            :placeholder="$t('admin.clusters.endpointPlaceholder')"
          />
        </el-form-item>

        <el-form-item
          :label="$t('admin.clusters.port')"
          prop="port"
        >
          <el-input-number
            v-model="formData.port"
            :min="1"
            :max="65535"
          />
        </el-form-item>

        <el-form-item
          :label="$t('admin.clusters.region')"
        >
          <el-input
            v-model="formData.region"
            :placeholder="$t('admin.clusters.regionPlaceholder')"
          />
        </el-form-item>

        <!-- OpenStack 专用配置 -->
        <template v-if="formData.type === 'openstack'">
          <el-divider content-position="left">
            {{ $t('admin.clusters.openstackConfig') }}
          </el-divider>

          <el-form-item
            :label="$t('admin.clusters.projectId')"
            prop="projectId"
          >
            <el-input
              v-model="formData.projectId"
              :placeholder="$t('admin.clusters.projectIdPlaceholder')"
            />
          </el-form-item>

          <el-form-item
            :label="$t('admin.clusters.domainId')"
          >
            <el-input
              v-model="formData.domainId"
              :placeholder="$t('admin.clusters.domainIdPlaceholder')"
            />
          </el-form-item>
        </template>

        <!-- K8s 专用配置 -->
        <template v-if="formData.type === 'k8s'">
          <el-divider content-position="left">
            {{ $t('admin.clusters.k8sConfig') }}
          </el-divider>

          <el-form-item
            :label="$t('admin.clusters.kubeconfig')"
          >
            <el-input
              v-model="formData.kubeconfig"
              type="textarea"
              :rows="5"
              :placeholder="$t('admin.clusters.kubeconfigPlaceholder')"
            />
          </el-form-item>

          <el-form-item
            :label="$t('admin.clusters.namespace')"
          >
            <el-input
              v-model="formData.namespace"
              placeholder="default"
            />
          </el-form-item>
        </template>

        <!-- 认证信息 -->
        <el-divider content-position="left">
          {{ $t('admin.clusters.authInfo') }}
        </el-divider>

        <el-form-item
          :label="$t('admin.clusters.username')"
          prop="username"
        >
          <el-input
            v-model="formData.username"
            :placeholder="$t('admin.clusters.usernamePlaceholder')"
          />
        </el-form-item>

        <el-form-item
          :label="$t('admin.clusters.password')"
          prop="password"
        >
          <el-input
            v-model="formData.password"
            type="password"
            show-password
            :placeholder="$t('admin.clusters.passwordPlaceholder')"
          />
        </el-form-item>

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

        <!-- 其他配置 -->
        <el-divider content-position="left">
          {{ $t('admin.clusters.otherConfig') }}
        </el-divider>

        <el-form-item
          :label="$t('common.status')"
        >
          <el-radio-group v-model="formData.status">
            <el-radio value="active">{{ $t('common.enabled') }}</el-radio>
            <el-radio value="inactive">{{ $t('common.disabled') }}</el-radio>
          </el-radio-group>
        </el-form-item>

        <el-form-item
          :label="$t('common.description')"
        >
          <el-input
            v-model="formData.description"
            type="textarea"
            :rows="3"
          />
        </el-form-item>
      </el-form>
    </div>

    <!-- 第三步：确认 -->
    <div v-show="!isEditing && currentStep === 2">
      <el-descriptions :column="1" border>
        <el-descriptions-item :label="$t('admin.clusters.name')">
          {{ formData.name }}
        </el-descriptions-item>
        <el-descriptions-item :label="$t('admin.clusters.type')">
          <el-tag :type="formData.type === 'openstack' ? 'warning' : 'primary'">
            {{ formData.type === 'openstack' ? 'OpenStack' : 'Kubernetes' }}
          </el-tag>
        </el-descriptions-item>
        <el-descriptions-item :label="$t('admin.clusters.endpoint')">
          {{ formData.endpoint }}:{{ formData.port }}
        </el-descriptions-item>
        <el-descriptions-item :label="$t('admin.clusters.region')">
          {{ formData.region || '-' }}
        </el-descriptions-item>
        <el-descriptions-item :label="$t('admin.clusters.username')">
          {{ formData.username }}
        </el-descriptions-item>
        <el-descriptions-item :label="$t('admin.clusters.status')">
          {{ formData.status === 'active' ? $t('common.enabled') : $t('common.disabled') }}
        </el-descriptions-item>
      </el-descriptions>
    </div>

    <!-- 底部按钮 -->
    <template #footer>
      <div class="dialog-footer">
        <el-button @click="handleClose">
          {{ isEditing ? $t('common.cancel') : $t('common.cancel') }}
        </el-button>
        
        <template v-if="!isEditing">
          <el-button
            v-if="currentStep > 0"
            @click="handlePrevStep"
          >
            {{ $t('common.prevStep') }}
          </el-button>
          
          <el-button
            v-if="currentStep < 2"
            type="primary"
            @click="handleNextStep"
          >
            {{ $t('common.nextStep') }}
          </el-button>
          
          <el-button
            v-if="currentStep === 2"
            type="primary"
            :loading="loading"
            @click="handleSubmit"
          >
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
import { Cloud, Grid } from '@element-plus/icons-vue'

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
const handleSubmit = () => {
  emit('submit', { ...formData })
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
.cluster-type-selection {
  padding: 20px;
}

.cluster-type-card {
  text-align: center;
  padding: 20px;
  cursor: pointer;
  transition: all 0.3s;
}

.cluster-type-card:hover {
  transform: translateY(-5px);
}

.cluster-type-card h3 {
  margin: 15px 0 10px;
}

.cluster-type-card p {
  color: #666;
  font-size: 14px;
}

.el-card.selected {
  border-color: #409EFF;
  background-color: #ecf5ff;
}

.dialog-footer {
  text-align: right;
}
</style>
