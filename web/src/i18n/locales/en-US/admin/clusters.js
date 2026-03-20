export default {
  title: 'Cluster Management',
  addCluster: 'Add Cluster',
  editCluster: 'Edit Cluster',
  searchName: 'Search by name',
  searchType: 'Search by type',
  name: 'Cluster Name',
  type: 'Cluster Type',
  endpoint: 'Endpoint',
  port: 'Port',
  region: 'Region',
  status: 'Status',
  instances: 'Instances',
  connect: 'Test Connection',
  connectSuccess: 'Connection successful',
  connectFailed: 'Connection failed',
  
  // Steps
  stepType: 'Select Type',
  stepConfig: 'Configuration',
  stepConfirm: 'Confirm',
  
  // Cluster type descriptions
  openstackDesc: 'OpenStack is an open source cloud computing platform that provides virtualization compute, storage and network resources',
  k8sDesc: 'Kubernetes is an open source container orchestration platform for automating deployment, scaling, and management of containerized applications',
  
  // Form
  basicInfo: 'Basic Info',
  namePlaceholder: 'Enter cluster name',
  endpointPlaceholder: 'Enter endpoint (e.g., 192.168.1.100 or example.com)',
  portPlaceholder: 'SSH port',
  regionPlaceholder: 'Enter region (optional)',
  
  // OpenStack config
  openstackConfig: 'OpenStack Config',
  projectId: 'Project ID',
  projectIdPlaceholder: 'Enter OpenStack Project ID',
  domainId: 'Domain ID',
  domainIdPlaceholder: 'Enter OpenStack Domain ID (optional)',
  
  // K8s config
  k8sConfig: 'Kubernetes Config',
  kubeconfig: 'Kubeconfig',
  kubeconfigPlaceholder: 'Enter kubeconfig content',
  namespace: 'Namespace',
  
  // Auth info
  authInfo: 'Authentication',
  usernamePlaceholder: 'Enter username',
  passwordPlaceholder: 'Enter password',
  privateKeyPlaceholder: 'Enter SSH private key (optional)',
  
  // Other config
  otherConfig: 'Other Config',
  
  // Validation
  nameRequired: 'Please enter cluster name',
  endpointRequired: 'Please enter endpoint',
  usernameRequired: 'Please enter username',
}
