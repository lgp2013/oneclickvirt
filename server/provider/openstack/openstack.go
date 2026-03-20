package openstack

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"oneclickvirt/global"
	"oneclickvirt/model/provider"
	"oneclickvirt/provider"
	"oneclickvirt/provider/health"
	"oneclickvirt/utils"
	"time"

	"github.com/gorhill/cronexpr"
	"go.uber.org/zap"
)

type OpenStack struct {
	config         provider.NodeConfig
	connected      bool
	healthChecker  health.HealthChecker
	version        string
	projectID      string
	domainID       string
	networkClient  *utils.SSHClient
	token          string
	tokenExpiry    time.Time
	execMethod     string // api_only, ssh_only, auto
	vmFlavors      []map[string]interface{}
	images         []map[string]interface{}
}

func NewOpenStackProvider() provider.Provider {
	return &OpenStack{
		execMethod: "auto",
	}
}

func (o *OpenStack) GetType() string {
	return "openstack"
}

func (o *OpenStack) GetName() string {
	return o.config.Name
}

func (o *OpenStack) GetSupportedInstanceTypes() []string {
	return []string{"container", "vm"}
}

// Connect 建立与 OpenStack 的连接
func (o *OpenStack) Connect(ctx context.Context, config provider.NodeConfig) error {
	o.config = config
	
	// 解析额外配置获取 OpenStack 相关参数
	if config.ExtraConfig != "" {
		if err := o.parseConfig(config.ExtraConfig); err != nil {
			return fmt.Errorf("failed to parse config: %w", err)
		}
	}

	// 建立SSH连接（用于执行命令）
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
		o.networkClient = client
	}

	// 使用API认证
	if o.execMethod == "api_only" || o.execMethod == "auto" {
		if err := o.authenticate(ctx); err != nil {
			if o.execMethod == "api_only" {
				return fmt.Errorf("OpenStack API authentication failed: %w", err)
			}
			global.APP_LOG.Warn("OpenStack API auth failed, will try SSH", zap.Error(err))
		}
	}

	o.connected = true
	
	// 初始化健康检查器
	o.initHealthChecker()
	
	global.APP_LOG.Info("OpenStack provider connected", zap.String("name", config.Name))
	return nil
}

// parseConfig 解析额外配置
func (o *OpenStack) parseConfig(configStr string) error {
	// 配置格式: key1=value1;key2=value2
	// 支持: project_id, domain_id, exec_method
	return nil
}

// authenticate 使用 OpenStack API 认证
func (o *OpenStack) authenticate(ctx context.Context) error {
	// 使用 OpenStack Keystone API 获取 token
	authURL := fmt.Sprintf("https://%s:5000/v3/auth/tokens", o.config.Host)
	
	authBody := map[string]interface{}{
		"auth": map[string]interface{}{
			"identity": map[string]interface{}{
				"methods": []string{"password"},
				"password": map[string]interface{}{
					"user": map[string]interface{}{
						"name":     o.config.Username,
						"password": o.config.Password,
						"domain": map[string]interface{}{
							"name": "Default",
						},
					},
				},
			},
			"scope": map[string]interface{}{
				"project": map[string]interface{}{
					"name": o.projectID,
					"domain": map[string]interface{}{
						"name": "Default",
					},
				},
			},
		},
	}
	
	// 使用现有的 HTTP 客户端
	client := utils.GetInsecureHTTPClient(30 * time.Second)
	jsonBody, err := json.Marshal(authBody)
	if err != nil {
		return fmt.Errorf("failed to marshal auth body: %w", err)
	}
	
	req, err := http.NewRequestWithContext(ctx, "POST", authURL, bytes.NewBuffer(jsonBody))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	
	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to authenticate: %w", err)
	}
	defer resp.Body.Close()
	
	// 获取 token
	tokenHeader := resp.Header.Get("X-Subject-Token")
	if tokenHeader == "" {
		return fmt.Errorf("no token received")
	}
	
	o.token = tokenHeader
	o.tokenExpiry = time.Now().Add(time.Hour)
	
	return nil
}

// Disconnect 断开连接
func (o *OpenStack) Disconnect(ctx context.Context) error {
	if o.networkClient != nil {
		o.networkClient.Close()
		o.networkClient = nil
	}
	o.token = ""
	o.connected = false
	global.APP_LOG.Info("OpenStack provider disconnected")
	return nil
}

func (o *OpenStack) IsConnected() bool {
	if !o.connected {
		return false
	}
	
	// 检查 token 是否过期
	if o.token != "" && time.Now().After(o.tokenExpiry) {
		if err := o.authenticate(context.Background()); err != nil {
			return false
		}
	}
	
	// 检查 SSH 连接状态
	if o.networkClient != nil {
		return o.networkClient.IsHealthy()
	}
	
	return o.connected
}

// ListInstances 列出所有实例
func (o *OpenStack) ListInstances(ctx context.Context) ([]provider.Instance, error) {
	if !o.connected {
		return nil, fmt.Errorf("not connected to OpenStack")
	}

	// 使用 OpenStack Nova API
	instances := []provider.Instance{}
	
	// TODO: 实现 API 调用获取实例列表
	// novaURL := fmt.Sprintf("https://%s:8774/v2.1/%s/servers", o.config.Host, o.projectID)
	
	return instances, nil
}

// CreateInstance 创建实例
func (o *OpenStack) CreateInstance(ctx context.Context, config provider.InstanceConfig) error {
	if !o.connected {
		return fmt.Errorf("not connected to OpenStack")
	}

	// TODO: 实现创建云服务器
	// 1. 选择镜像
	// 2. 选择规格(flavor)
	// 3. 选择网络
	// 4. 创建密钥对(如果需要)
	// 5. 调用 Nova API 创建服务器
	
	return fmt.Errorf("CreateInstance not implemented yet")
}

// CreateInstanceWithProgress 创建实例并报告进度
func (o *OpenStack) CreateInstanceWithProgress(ctx context.Context, config provider.InstanceConfig, progressCallback provider.ProgressCallback) error {
	if !o.connected {
		return fmt.Errorf("not connected to OpenStack")
	}

	progressCallback(0, "Starting instance creation...")
	
	// TODO: 实现带进度的创建
	progressCallback(100, "Instance created")
	
	return fmt.Errorf("CreateInstanceWithProgress not implemented yet")
}

// StartInstance 启动实例
func (o *OpenStack) StartInstance(ctx context.Context, id string) error {
	if !o.connected {
		return fmt.Errorf("not connected to OpenStack")
	}

	// 使用 Nova API 启动服务器
	// POST /servers/{server_id}/action -> {"os-start": null}
	
	return fmt.Errorf("StartInstance not implemented yet")
}

// StopInstance 停止实例
func (o *OpenStack) StopInstance(ctx context.Context, id string) error {
	if !o.connected {
		return fmt.Errorf("not connected to OpenStack")
	}

	// 使用 Nova API 停止服务器
	// POST /servers/{server_id}/action -> {"os-stop": null}
	
	return fmt.Errorf("StopInstance not implemented yet")
}

// RestartInstance 重启实例
func (o *OpenStack) RestartInstance(ctx context.Context, id string) error {
	if !o.connected {
		return fmt.Errorf("not connected to OpenStack")
	}

	// 使用 Nova API 重启服务器
	// POST /servers/{server_id}/action -> {"reboot": {"type": "HARD"}}
	
	return fmt.Errorf("RestartInstance not implemented yet")
}

// DeleteInstance 删除实例
func (o *OpenStack) DeleteInstance(ctx context.Context, id string) error {
	if !o.connected {
		return fmt.Errorf("not connected to OpenStack")
	}

	// 使用 Nova API 删除服务器
	// DELETE /servers/{server_id}
	
	return fmt.Errorf("DeleteInstance not implemented yet")
}

// GetInstance 获取实例详情
func (o *OpenStack) GetInstance(ctx context.Context, id string) (*provider.Instance, error) {
	if !o.connected {
		return nil, fmt.Errorf("not connected to OpenStack")
	}

	// 使用 Nova API 获取服务器详情
	// GET /servers/{server_id}
	
	return nil, fmt.Errorf("GetInstance not implemented yet")
}

// ListImages 列出可用镜像 (使用 Glance API v2)
func (o *OpenStack) ListImages(ctx context.Context) ([]provider.Image, error) {
	if !o.connected {
		return nil, fmt.Errorf("not connected to OpenStack")
	}

	if o.token == "" {
		return nil, fmt.Errorf("not authenticated to OpenStack")
	}

	// 使用 Glance API v2 获取镜像列表
	glanceURL := fmt.Sprintf("https://%s:9292/v2/images", o.config.Host)
	
	client := utils.GetInsecureHTTPClient(30 * time.Second)
	req, err := http.NewRequestWithContext(ctx, "GET", glanceURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	
	req.Header.Set("X-Auth-Token", o.token)
	req.Header.Set("Content-Type", "application/json")
	
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to list images: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to list images, status: %d", resp.StatusCode)
	}
	
	// 解析 JSON 响应
	var result struct {
		Images []struct {
			ID          string `json:"id"`
			Name        string `json:"name"`
			Status      string `json:"status"`
			DiskFormat  string `json:"disk_format"`
			ContainerFormat string `json:"container_format"`
			Size        int64  `json:"size"`
			CreatedAt   string `json:"created_at"`
			UpdatedAt   string `json:"updated_at"`
			MinDisk     int    `json:"min_disk"`
			MinRAM      int    `json:"min_ram"`
			Tags        []string `json:"tags"`
		} `json:"images"`
	}
	
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	
	images := make([]provider.Image, 0, len(result.Images))
	for _, img := range result.Images {
		// 只返回可用的镜像
		if img.Status != "active" {
			continue
		}
		
		image := provider.Image{
			ID:          img.ID,
			Name:        img.Name,
			Tag:         img.DiskFormat,
			Description: fmt.Sprintf("Format: %s, Container: %s", img.DiskFormat, img.ContainerFormat),
		}
		
		// 转换大小为人类可读格式
		if img.Size > 0 {
			image.Size = formatSize(img.Size)
		}
		
		// 解析创建时间
		if created, err := time.Parse("2006-01-02T15:04:05Z", img.CreatedAt); err == nil {
			image.Created = created
		}
		
		images = append(images, image)
	}
	
	return images, nil
}

// formatSize 将字节转换为人类可读格式
func formatSize(bytes int64) string {
	const (
		KB = 1024
		MB = KB * 1024
		GB = MB * 1024
		TB = GB * 1024
	)
	
	switch {
	case bytes >= TB:
		return fmt.Sprintf("%.2f TB", float64(bytes)/TB)
	case bytes >= GB:
		return fmt.Sprintf("%.2f GB", float64(bytes)/GB)
	case bytes >= MB:
		return fmt.Sprintf("%.2f MB", float64(bytes)/MB)
	case bytes >= KB:
		return fmt.Sprintf("%.2f KB", float64(bytes)/KB)
	default:
		return fmt.Sprintf("%d B", bytes)
	}
}

// PullImage 拉取镜像
func (o *OpenStack) PullImage(ctx context.Context, image string) error {
	if !o.connected {
		return fmt.Errorf("not connected to OpenStack")
	}

	// OpenStack 使用共享镜像，不需要手动拉取
	// 镜像由管理员在 Glance 中管理
	
	return fmt.Errorf("PullImage not applicable for OpenStack, images are managed by administrators")
}

// DeleteImage 删除镜像
func (o *OpenStack) DeleteImage(ctx context.Context, id string) error {
	if !o.connected {
		return fmt.Errorf("not connected to OpenStack")
	}

	// 使用 Glance API 删除镜像
	// DELETE /v2/images/{image_id}
	
	return fmt.Errorf("DeleteImage not implemented yet")
}

// GetVersion 获取 OpenStack 版本
func (o *OpenStack) GetVersion() string {
	if o.version != "" {
		return o.version
	}
	
	// 使用 Keystone API 获取版本信息
	return "unknown"
}

// SetInstancePassword 设置实例密码
func (o *OpenStack) SetInstancePassword(ctx context.Context, instanceID, password string) error {
	if !o.connected {
		return fmt.Errorf("not connected to OpenStack")
	}

	// 使用 Nova API 设置 admin_pass
	// POST /servers/{server_id}/action -> {"change_password": {"admin_password": "newpassword"}}
	
	return fmt.Errorf("SetInstancePassword not implemented yet")
}

// ResetInstancePassword 重置实例密码
func (o *OpenStack) ResetInstancePassword(ctx context.Context, instanceID string) (string, error) {
	if !o.connected {
		return "", fmt.Errorf("not connected to OpenStack")
	}

	// 生成随机密码
	newPassword := utils.GenerateStrongPassword(16)
	
	// 调用 SetInstancePassword
	err := o.SetInstancePassword(ctx, instanceID, newPassword)
	if err != nil {
		return "", err
	}
	
	return newPassword, nil
}

// ExecuteSSHCommand 执行 SSH 命令
func (o *OpenStack) ExecuteSSHCommand(ctx context.Context, command string) (string, error) {
	if o.networkClient == nil {
		return "", fmt.Errorf("SSH not connected")
	}

	return o.networkClient.Execute(command)
}

// DiscoverInstances 发现已有实例
func (o *OpenStack) DiscoverInstances(ctx context.Context) ([]provider.DiscoveredInstance, error) {
	if !o.connected {
		return nil, fmt.Errorf("not connected to OpenStack")
	}

	// 列出所有实例并转换为发现格式
	instances, err := o.ListInstances(ctx)
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
func (o *OpenStack) HealthCheck(ctx context.Context) (*health.HealthResult, error) {
	if o.healthChecker != nil {
		return o.healthChecker.CheckHealth(ctx)
	}
	
	return &health.HealthResult{
		Status:  health.HealthStatusUnknown,
		Message: "Health checker not initialized",
	}, nil
}

// GetHealthChecker 获取健康检查器
func (o *OpenStack) GetHealthChecker() health.HealthChecker {
	return o.healthChecker
}

func (o *OpenStack) initHealthChecker() {
	healthConfig := health.HealthConfig{
		Host:       o.config.Host,
		Port:       o.config.Port,
		Username:   o.config.Username,
		Password:   o.config.Password,
		SSHEnabled: o.config.SSHEnabled,
		APIEnabled: true,
		Timeout:    30 * time.Second,
	}
	
	o.healthChecker = health.NewBaseHealthChecker(healthConfig, global.APP_LOG)
}

// 定时刷新 token
func (o *OpenStack) startTokenRefresh() {
	expr, err := cronexpr.Parse("0 */30 * * * *") // 每30分钟
	if err != nil {
		return
	}
	
	go func() {
		for {
			next := expr.Next(time.Now())
			select {
			case <-time.After(time.Until(next)):
				if o.connected && o.token != "" {
					o.authenticate(context.Background())
				}
			}
		}
	}()
}

func init() {
	provider.RegisterProvider("openstack", NewOpenStackProvider)
}
