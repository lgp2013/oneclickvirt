export default {
  title: '集群管理',
  addCluster: '添加集群',
  editCluster: '编辑集群',
  searchName: '搜索集群名称',
  searchType: '搜索类型',
  name: '集群名称',
  type: '集群类型',
  endpoint: '访问地址',
  port: '端口',
  region: '区域',
  status: '状态',
  instances: '实例数',
  connect: '连接测试',
  connectSuccess: '连接成功',
  connectFailed: '连接失败',
  
  // 步骤
  stepType: '选择类型',
  stepConfig: '配置信息',
  stepConfirm: '确认',
  
  // 集群类型描述
  openstackDesc: 'OpenStack 是一个开源的云计算平台，提供虚拟化计算、存储和网络资源',
  k8sDesc: 'Kubernetes 是一个开源的容器编排平台，用于自动化容器化应用的部署、扩展和管理',
  
  // 表单
  basicInfo: '基本信息',
  namePlaceholder: '请输入集群名称',
  endpointPlaceholder: '请输入访问地址 (如: 192.168.1.100 或 example.com)',
  portPlaceholder: 'SSH端口',
  regionPlaceholder: '请输入区域 (可选)',
  
  // OpenStack 配置
  openstackConfig: 'OpenStack 配置',
  projectId: '项目ID',
  projectIdPlaceholder: '请输入 OpenStack 项目 ID',
  domainId: '域ID',
  domainIdPlaceholder: '请输入 OpenStack 域 ID (可选)',
  
  // K8s 配置
  k8sConfig: 'Kubernetes 配置',
  kubeconfig: 'Kubeconfig',
  kubeconfigPlaceholder: '请输入 kubeconfig 内容',
  namespace: '命名空间',
  
  // 认证信息
  authInfo: '认证信息',
  usernamePlaceholder: '请输入用户名',
  passwordPlaceholder: '请输入密码',
  privateKeyPlaceholder: '请输入 SSH 私钥 (可选)',
  
  // 其他配置
  otherConfig: '其他配置',
  
  // 验证
  nameRequired: '请输入集群名称',
  endpointRequired: '请输入访问地址',
  usernameRequired: '请输入用户名',
}
