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

// CinderService OpenStack Cinder 卷管理服务
type CinderService struct {
	*OpenStack
}

// NewCinderService 创建 Cinder 服务
func NewCinderService(os *OpenStack) *CinderService {
	return &CinderService{OpenStack: os}
}

// Volume 卷实体
type Volume struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Status    string    `json:"status"`
	Size      int       `json:"size"` // GB
	VolumeType string   `json:"volume_type"`
	Zone      string    `json:"availability_zone"`
	CreatedAt time.Time `json:"created_at"`
}

// VolumeAttachment 卷挂载信息
type VolumeAttachment struct {
	VolumeID   string `json:"volumeId"`
	InstanceID string `json:"serverId"`
	Device     string `json:"device"`
}

// ListVolumes 列出所有卷
func (c *CinderService) ListVolumes(ctx context.Context) ([]Volume, error) {
	if !c.IsConnected() {
		return nil, fmt.Errorf("not connected to OpenStack")
	}

	cinderURL := fmt.Sprintf("https://%s:8776/v3/%s/volumes", c.config.Host, c.projectID)
	
	client := utils.GetInsecureHTTPClient(30 * time.Second)
	req, err := http.NewRequestWithContext(ctx, "GET", cinderURL, nil)
	if err != nil {
		return nil, err
	}
	
	req.Header.Set("X-Auth-Token", c.token)
	
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	
	var result struct {
		Volumes []Volume `json:"volumes"`
	}
	
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	
	return result.Volumes, nil
}

// GetVolume 获取卷详情
func (c *CinderService) GetVolume(ctx context.Context, volumeID string) (*Volume, error) {
	if !c.IsConnected() {
		return nil, fmt.Errorf("not connected to OpenStack")
	}

	cinderURL := fmt.Sprintf("https://%s:8776/v3/%s/volumes/%s", c.config.Host, c.projectID, volumeID)
	
	client := utils.GetInsecureHTTPClient(30 * time.Second)
	req, err := http.NewRequestWithContext(ctx, "GET", cinderURL, nil)
	if err != nil {
		return nil, err
	}
	
	req.Header.Set("X-Auth-Token", c.token)
	
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	
	var result struct {
		Volume Volume `json:"volume"`
	}
	
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	
	return &result.Volume, nil
}

// CreateVolume 创建卷
func (c *CinderService) CreateVolume(ctx context.Context, name, volumeType string, size int, zone string) (*Volume, error) {
	if !c.IsConnected() {
		return nil, fmt.Errorf("not connected to OpenStack")
	}

	cinderURL := fmt.Sprintf("https://%s:8776/v3/%s/volumes", c.config.Host, c.projectID)
	
	createReq := map[string]interface{}{
		"volume": map[string]interface{}{
			"name":          name,
			"volume_type":  volumeType,
			"size":          size,
			"availability_zone": zone,
		}
	}
	
	jsonBody, _ := json.Marshal(createReq)
	
	client := utils.GetInsecureHTTPClient(60 * time.Second)
	req, err := http.NewRequestWithContext(ctx, "POST", cinderURL, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, err
	}
	
	req.Header.Set("X-Auth-Token", c.token)
	req.Header.Set("Content-Type", "application/json")
	
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	
	var result struct {
		Volume Volume `json:"volume"`
	}
	
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	
	global.APP_LOG.Info("Volume created", zap.String("name", name), zap.String("volume_id", result.Volume.ID))
	
	return &result.Volume, nil
}

// DeleteVolume 删除卷
func (c *CinderService) DeleteVolume(ctx context.Context, volumeID string) error {
	if !c.IsConnected() {
		return fmt.Errorf("not connected to OpenStack")
	}

	cinderURL := fmt.Sprintf("https://%s:8776/v3/%s/volumes/%s", c.config.Host, c.projectID, volumeID)
	
	client := utils.GetInsecureHTTPClient(30 * time.Second)
	req, err := http.NewRequestWithContext(ctx, "DELETE", cinderURL, nil)
	if err != nil {
		return err
	}
	
	req.Header.Set("X-Auth-Token", c.token)
	
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != 202 {
		return fmt.Errorf("failed to delete volume, status: %d", resp.StatusCode)
	}
	
	global.APP_LOG.Info("Volume deleted", zap.String("volume_id", volumeID))
	
	return nil
}

// AttachVolume 挂载卷到虚拟机
func (c *CinderService) AttachVolume(ctx context.Context, volumeID, instanceID string) error {
	if !c.IsConnected() {
		return fmt.Errorf("not connected to OpenStack")
	}

	cinderURL := fmt.Sprintf("https://%s:8776/v3/%s/volumes/%s/action", c.config.Host, c.projectID, volumeID)
	
	attachReq := map[string]interface{}{
		"attach": map[string]interface{}{
			"instance_uuid": instanceID,
		}
	}
	
	jsonBody, _ := json.Marshal(attachReq)
	
	client := utils.GetInsecureHTTPClient(30 * time.Second)
	req, err := http.NewRequestWithContext(ctx, "POST", cinderURL, bytes.NewBuffer(jsonBody))
	if err != nil {
		return err
	}
	
	req.Header.Set("X-Auth-Token", c.token)
	req.Header.Set("Content-Type", "application/json")
	
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != 202 {
		return fmt.Errorf("failed to attach volume, status: %d", resp.StatusCode)
	}
	
	global.APP_LOG.Info("Volume attached", zap.String("volume_id", volumeID), zap.String("instance_id", instanceID))
	
	return nil
}

// DetachVolume 卸载卷
func (c *CinderService) DetachVolume(ctx context.Context, volumeID string) error {
	if !c.IsConnected() {
		return fmt.Errorf("not connected to OpenStack")
	}

	cinderURL := fmt.Sprintf("https://%s:8776/v3/%s/volumes/%s/action", c.config.Host, c.projectID, volumeID)
	
	detachReq := map[string]interface{}{
		"detach": map[string]interface{}{}
	}
	
	jsonBody, _ := json.Marshal(detachReq)
	
	client := utils.GetInsecureHTTPClient(30 * time.Second)
	req, err := http.NewRequestWithContext(ctx, "POST", cinderURL, bytes.NewBuffer(jsonBody))
	if err != nil {
		return err
	}
	
	req.Header.Set("X-Auth-Token", c.token)
	req.Header.Set("Content-Type", "application/json")
	
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != 202 {
		return fmt.Errorf("failed to detach volume, status: %d", resp.StatusCode)
	}
	
	global.APP_LOG.Info("Volume detached", zap.String("volume_id", volumeID))
	
	return nil
}

// ExtendVolume 扩容卷
func (c *CinderService) ExtendVolume(ctx context.Context, volumeID string, newSize int) error {
	if !c.IsConnected() {
		return fmt.Errorf("not connected to OpenStack")
	}

	cinderURL := fmt.Sprintf("https://%s:8776/v3/%s/volumes/%s/action", c.config.Host, c.projectID, volumeID)
	
	extendReq := map[string]interface{}{
		"os-extend": map[string]interface{}{
			"new_size": newSize,
		}
	}
	
	jsonBody, _ := json.Marshal(extendReq)
	
	client := utils.GetInsecureHTTPClient(60 * time.Second)
	req, err := http.NewRequestWithContext(ctx, "POST", cinderURL, bytes.NewBuffer(jsonBody))
	if err != nil {
		return err
	}
	
	req.Header.Set("X-Auth-Token", c.token)
	req.Header.Set("Content-Type", "application/json")
	
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != 202 {
		return fmt.Errorf("failed to extend volume, status: %d", resp.StatusCode)
	}
	
	global.APP_LOG.Info("Volume extended", zap.String("volume_id", volumeID), zap.Int("new_size", newSize))
	
	return nil
}

// ListVolumeTypes 列出卷类型
func (c *CinderService) ListVolumeTypes(ctx context.Context) ([]map[string]interface{}, error) {
	if !c.IsConnected() {
		return nil, fmt.Errorf("not connected to OpenStack")
	}

	cinderURL := fmt.Sprintf("https://%s:8776/v3/%s/types", c.config.Host, c.projectID)
	
	client := utils.GetInsecureHTTPClient(30 * time.Second)
	req, err := http.NewRequestWithContext(ctx, "GET", cinderURL, nil)
	if err != nil {
		return nil, err
	}
	
	req.Header.Set("X-Auth-Token", c.token)
	
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	
	var result struct {
		VolumeTypes []map[string]interface{} `json:"volume_types"`
	}
	
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	
	return result.VolumeTypes, nil
}

// ListAvailabilityZones 列出可用区域
func (c *CinderService) ListAvailabilityZones(ctx context.Context) ([]map[string]interface{}, error) {
	if !c.IsConnected() {
		return nil, fmt.Errorf("not connected to OpenStack")
	}

	cinderURL := fmt.Sprintf("https://%s:8776/v3/%s/availability_zones", c.config.Host, c.projectID)
	
	client := utils.GetInsecureHTTPClient(30 * time.Second)
	req, err := http.NewRequestWithContext(ctx, "GET", cinderURL, nil)
	if err != nil {
		return nil, err
	}
	
	req.Header.Set("X-Auth-Token", c.token)
	
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	
	var result struct {
		AvailabilityZones []map[string]interface{} `json:"availability_zones"`
	}
	
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	
	return result.AvailabilityZones, nil
}
