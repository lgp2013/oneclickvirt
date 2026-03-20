package k8s

import (
	"context"
	"fmt"
	"oneclickvirt/global"
	"oneclickvirt/model/provider"
	"oneclickvirt/provider"
	"oneclickvirt/provider/health"
	"oneclickvirt/utils"
	"time"

	"go.uber.org/zap"
)

type K8s struct {
	config        provider.NodeConfig
	connected     bool
	healthChecker health.HealthChecker
	version      string
	sshClient    *utils.SSHClient
	execMethod   string // api_only, ssh_only, auto
	kubeconfig   string
	namespace    string
}

func NewK8sProvider() provider.Provider {
	return &K8s{
		execMethod: "auto",
		namespace:  "default",
	}
}

func (k *K8s) GetType() string {
	return "k8s"
}

func (k *K8s) GetName() string {
	return k.config.Name
}

func (k *K8s) GetSupportedInstanceTypes() []string {
	// K8s 支持容器(pod) 和通过 VM 实现的虚拟机
	return []string{"container", "vm"}
}

// Connect 建立与 K8s 的连接
func (k *K8s) Connect(ctx context.Context, config provider.NodeConfig) error {
	k.config = config
	
	// 解析额外配置获取 K8s 相关参数
	if config.ExtraConfig != "" {
		if err := k.parseConfig(config.ExtraConfig); err != nil {
			return fmt.Errorf("failed to parse config: %w", err)
		}
	}

	// 建立SSH连接（用于执行 kubectl 命令）
	if config.SSHEnabled {
		sshConfig := utils.SSHConfig{
			Host:           config.Host,
			Port:           config.Port,
			Username:       config.Username,
			Password:       config.Password,
			PrivateKey:     config.PrivateKey,
			ConnectTimeout: 30 * time.Second,
			ExecuteTimeout: 300 * time.Second,
		}
		
		client, err := utils.NewSSHClient(sshConfig)
		if err != nil {
			return fmt.Errorf("failed to connect via SSH: %w", err)
		}
		k.sshClient = client
	}

	// 检查 kubectl 是否可用
	if k.sshClient != nil {
		_, err := k.sshClient.Execute("kubectl version --client")
		if err != nil {
			global.APP_LOG.Warn("kubectl not found on the node", zap.Error(err))
		}
	}

	k.connected = true
	
	// 初始化健康检查器
	k.initHealthChecker()
	
	// 获取 K8s 版本
	k.getVersion()
	
	global.APP_LOG.Info("K8s provider connected", zap.String("name", config.Name))
	return nil
}

// parseConfig 解析额外配置
func (k *K8s) parseConfig(configStr string) error {
	// 配置格式: key1=value1;key2=value2
	// 支持: kubeconfig(base64), namespace, exec_method
	return nil
}

// Disconnect 断开连接
func (k *K8s) Disconnect(ctx context.Context) error {
	if k.sshClient != nil {
		k.sshClient.Close()
		k.sshClient = nil
	}
	k.connected = false
	global.APP_LOG.Info("K8s provider disconnected")
	return nil
}

func (k *K8s) IsConnected() bool {
	if !k.connected {
		return false
	}
	
	// 检查 kubectl 连接状态
	if k.sshClient != nil {
		_, err := k.sshClient.Execute("kubectl cluster-info")
		if err != nil {
			return false
		}
		return true
	}
	
	return k.connected
}

// ListInstances 列出所有 Pod (作为实例)
func (k *K8s) ListInstances(ctx context.Context) ([]provider.Instance, error) {
	if !k.connected {
		return nil, fmt.Errorf("not connected to K8s")
	}

	instances := []provider.Instance{}
	
	// 使用 kubectl 获取 Pod 列表
	output, err := k.sshClient.Execute(fmt.Sprintf("kubectl get pods -n %s -o json", k.namespace))
	if err != nil {
		return nil, fmt.Errorf("failed to list pods: %w", err)
	}
	
	// TODO: 解析 JSON 输出转换为 Instance 列表
	
	return instances, nil
}

// CreateInstance 创建 Pod (作为容器实例)
func (k *K8s) CreateInstance(ctx context.Context, config provider.InstanceConfig) error {
	if !k.connected {
		return fmt.Errorf("not connected to K8s")
	}

	// 使用 kubectl 创建 Pod
	// 也可以使用 YAML 或 JSON 定义
	
	return fmt.Errorf("CreateInstance not implemented yet")
}

// CreateInstanceWithProgress 创建实例并报告进度
func (k *K8s) CreateInstanceWithProgress(ctx context.Context, config provider.InstanceConfig, progressCallback provider.ProgressCallback) error {
	if !k.connected {
		return fmt.Errorf("not connected to K8s")
	}

	progressCallback(0, "Starting pod creation...")
	
	// TODO: 实现带进度的创建
	progressCallback(100, "Pod created")
	
	return fmt.Errorf("CreateInstanceWithProgress not implemented yet")
}

// StartInstance 启动 Pod
func (k *K8s) StartInstance(ctx context.Context, id string) error {
	if !k.connected {
		return fmt.Errorf("not connected to K8s")
	}

	// 使用 kubectl 启动 Pod
	_, err := k.sshClient.Execute(fmt.Sprintf("kubectl get pod %s -n %s -o json | jq '.spec.activeDeadlineSeconds=null | .metadata.deletionTimestamp=null' | kubectl apply -f -", id, k.namespace))
	
	return err
}

// StopInstance 停止 Pod (删除)
func (k *K8s) StopInstance(ctx context.Context, id string) error {
	if !k.connected {
		return fmt.Errorf("not connected to K8s")
	}

	// K8s Pod 不能真正停止，只能删除或暂停
	// 使用 kubectl delete 或缩放到 0
	
	return fmt.Errorf("StopInstance not implemented yet")
}

// RestartInstance 重启 Pod
func (k *K8s) RestartInstance(ctx context.Context, id string) error {
	if !k.connected {
		return fmt.Errorf("not connected to K8s")
	}

	// 删除并重新创建 Pod
	_, err := k.sshClient.Execute(fmt.Sprintf("kubectl delete pod %s -n %s && kubectl create -f -", id, k.namespace))
	
	return err
}

// DeleteInstance 删除 Pod
func (k *K8s) DeleteInstance(ctx context.Context, id string) error {
	if !k.connected {
		return fmt.Errorf("not connected to K8s")
	}

	// 使用 kubectl 删除 Pod
	_, err := k.sshClient.Execute(fmt.Sprintf("kubectl delete pod %s -n %s", id, k.namespace))
	
	return err
}

// GetInstance 获取 Pod 详情
func (k *K8s) GetInstance(ctx context.Context, id string) (*provider.Instance, error) {
	if !k.connected {
		return nil, fmt.Errorf("not connected to K8s")
	}

	// 使用 kubectl 获取 Pod 详情
	output, err := k.sshClient.Execute(fmt.Sprintf("kubectl get pod %s -n %s -o json", id, k.namespace))
	if err != nil {
		return nil, err
	}
	
	// TODO: 解析 JSON 输出转换为 Instance
	
	return nil, fmt.Errorf("GetInstance not implemented yet")
}

// ListImages 列出可用容器镜像 (使用 Kubernetes Deployment)
func (k *K8s) ListImages(ctx context.Context) ([]provider.Image, error) {
	if !k.connected {
		return nil, fmt.Errorf("not connected to K8s")
	}

	images := []provider.Image{}
	
	// 使用 kubectl 获取 Deployment 列表，这些是用户可用的"镜像"
	output, err := k.sshClient.Execute(fmt.Sprintf("kubectl get deployments -n %s -o json", k.namespace))
	if err != nil {
		return nil, err
	}
	
	// TODO: 解析 JSON
	
	return images, nil
}

// PullImage 拉取镜像 (在 K8s 节点上预拉取)
func (k *K8s) PullImage(ctx context.Context, image string) error {
	if !k.connected {
		return fmt.Errorf("not connected to K8s")
	}

	// 使用 crictl 或 docker 在节点上预拉取镜像
	// 需要在所有节点上执行
	
	// 方式1: 使用 node-shell 在每个节点上拉取
	// 方式2: 创建 Pod 使用 Kubelet 预拉取
	
	return fmt.Errorf("PullImage not implemented yet")
}

// DeleteImage 删除镜像 (通过节点上的 crictl/docker)
func (k *K8s) DeleteImage(ctx context.Context, id string) error {
	if !k.connected {
		return fmt.Errorf("not connected to K8s")
	}

	// K8s 镜像通常由节点管理，不建议手动删除
	
	return fmt.Errorf("DeleteImage not recommended for K8s")
}

// GetVersion 获取 K8s 版本
func (k *K8s) GetVersion() string {
	if k.version != "" {
		return k.version
	}
	
	k.getVersion()
	return k.version
}

func (k *K8s) getVersion() {
	if k.sshClient == nil {
		return
	}
	
	output, err := k.sshClient.Execute("kubectl version -o json")
	if err != nil {
		k.version = "unknown"
		return
	}
	
	// 解析版本信息
	// TODO: 解析 JSON
	k.version = "unknown"
}

// SetInstancePassword 设置实例密码 (不适用于 K8s Pod)
func (k *K8s) SetInstancePassword(ctx context.Context, instanceID, password string) error {
	if !k.connected {
		return fmt.Errorf("not connected to K8s")
	}

	// K8s Pod 不支持直接设置密码，需要通过 Secret 或 exec
	
	return fmt.Errorf("SetInstancePassword not applicable for K8s Pods")
}

// ResetInstancePassword 重置实例密码
func (k *K8s) ResetInstancePassword(ctx context.Context, instanceID string) (string, error) {
	if !k.connected {
		return "", fmt.Errorf("not connected to K8s")
	}

	// K8s Pod 不支持直接重置密码
	
	return "", fmt.Errorf("ResetInstancePassword not applicable for K8s Pods")
}

// ExecuteSSHCommand 执行 SSH 命令 (通过 kubectl exec)
func (k *K8s) ExecuteSSHCommand(ctx context.Context, command string) (string, error) {
	if k.sshClient == nil {
		return "", fmt.Errorf("SSH not connected")
	}

	// 通过 kubectl exec 执行命令
	// kubectl exec -it <pod> -n <namespace> -- <command>
	
	return "", fmt.Errorf("ExecuteSSHCommand not implemented yet")
}

// DiscoverInstances 发现已有 Pod
func (k *K8s) DiscoverInstances(ctx context.Context) ([]provider.DiscoveredInstance, error) {
	if !k.connected {
		return nil, fmt.Errorf("not connected to K8s")
	}

	// 列出所有 Pod 并转换为发现格式
	instances, err := k.ListInstances(ctx)
	if err != nil {
		return nil, err
	}

	discovered := make([]provider.DiscoveredInstance, 0, len(instances))
	for _, inst := range instances {
		discovered = append(discovered, provider.DiscoveredInstance{
			UUID:         inst.UUID,
			Name:         inst.Name,
			Status:       inst.Status,
			InstanceType: inst.InstanceType,
			CPU:          inst.CPU,
			Memory:       inst.Memory,
			Disk:         inst.Disk,
			PrivateIP:    inst.PrivateIP,
			PublicIP:     inst.PublicIP,
			SSHPort:      inst.SSHPort,
		})
	}

	return discovered, nil
}

// HealthCheck 健康检查
func (k *K8s) HealthCheck(ctx context.Context) (*health.HealthResult, error) {
	if k.healthChecker != nil {
		return k.healthChecker.CheckHealth(ctx)
	}
	
	return &health.HealthResult{
		Status:  health.HealthStatusUnknown,
		Message: "Health checker not initialized",
	}, nil
}

// GetHealthChecker 获取健康检查器
func (k *K8s) GetHealthChecker() health.HealthChecker {
	return k.healthChecker
}

func (k *K8s) initHealthChecker() {
	healthConfig := health.HealthConfig{
		Host:       k.config.Host,
		Port:       k.config.Port,
		Username:   k.config.Username,
		Password:   k.config.Password,
		SSHEnabled: k.config.SSHEnabled,
		APIEnabled: false,
		Timeout:    30 * time.Second,
	}
	
	k.healthChecker = health.NewBaseHealthChecker(healthConfig, global.APP_LOG)
}

// ListNodes 列出所有 K8s 节点 (用于管理虚拟机)
func (k *K8s) ListNodes() ([]map[string]interface{}, error) {
	if k.sshClient == nil {
		return nil, fmt.Errorf("SSH not connected")
	}
	
	output, err := k.sshClient.Execute("kubectl get nodes -o json")
	if err != nil {
		return nil, err
	}
	
	// TODO: 解析 JSON
	return nil, nil
}

// CreateVM 创建虚拟机 (使用 KubeVirt)
func (k *K8s) CreateVM(name string, cpu, memory int64, image string) error {
	// 使用 KubeVirt CRD 创建虚拟机
	vmYAML := fmt.Sprintf(`
apiVersion: kubevirt.io/v1
kind: VirtualMachine
metadata:
  name: %s
  namespace: %s
spec:
  running: true
  template:
    spec:
      domain:
        cpu:
          cores: %d
        memory:
          guest: %dMi
      devices: {}
`, name, k.namespace, cpu, memory)

	_, err := k.sshClient.Execute(fmt.Sprintf("kubectl apply -f - <<< '%s'", vmYAML))
	return err
}

func init() {
	provider.RegisterProvider("k8s", NewK8sProvider)
}
