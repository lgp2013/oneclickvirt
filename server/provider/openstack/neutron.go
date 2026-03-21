package openstack

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"oneclickvirt/global"
	"oneclickvirt/utils"
	"time"

	"go.uber.org/zap"
)

// NeutronService OpenStack Neutron 网络管理服务
type NeutronService struct {
	*OpenStack
}

// Network 网络实体
type Network struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Status      string `json:"status"`
	AdminStateUp bool  `json:"admin_state_up"`
	NetworkType string `json:"provider:network_type"`
	PhysicalNetwork string `json:"provider:physical_network"`
	SegmentationID int    `json:"provider:segmentation_id"`
	Shared      bool    `json:"shared"`
	TenantID    string `json:"tenant_id"`
	CreatedAt   string `json:"created_at"`
}

// Subnet 子网实体
type Subnet struct {
	ID              string   `json:"id"`
	Name            string   `json:"name"`
	NetworkID       string   `json:"network_id"`
	CIDR            string   `json:"cidr"`
	IPVersion       int      `json:"ip_version"`
	GatewayIP       string   `json:"gateway_ip"`
	DNSNameservers  []string `json:"dns_nameservers"`
	AllocationPool  []map[string]string `json:"allocation_pools"`
	HostRoutes      []map[string]string `json:"host_routes"`
	EnableDHCP      bool     `json:"enable_dhcp"`
}

// Router 路由器实体
type Router struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	Status        string `json:"status"`
	AdminStateUp  bool   `json:"admin_state_up"`
	TenantID      string `json:"tenant_id"`
	ExternalGatewayInfo map[string]interface{} `json:"external_gateway_info"`
}

// Port 端口实体
type Port struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	NetworkID   string `json:"network_id"`
	Status      string `json:"status"`
	AdminStateUp bool  `json:"admin_state_up"`
	MACAddress   string `json:"mac_address"`
	FixedIPs     []map[string]string `json:"fixed_ips"`
	DeviceID     string `json:"device_id"`
	DeviceOwner  string `json:"device_owner"`
	TenantID     string `json:"tenant_id"`
}

// FloatingIP 浮动IP实体
type FloatingIP struct {
	ID                string `json:"id"`
	FloatingIPAddress string `json:"floating_ip_address"`
	FloatingNetworkID string `json:"floating_network_id"`
	FixedIPAddress    string `json:"fixed_ip_address"`
	PortID           string `json:"port_id"`
	RouterID         string `json:"router_id"`
	TenantID         string `json:"tenant_id"`
	Status           string `json:"status"`
}

// NewNeutronService 创建 Neutron 服务
func NewNeutronService(os *OpenStack) *NeutronService {
	return &NeutronService{OpenStack: os}
}

// ListNetworks 列出所有网络
func (n *NeutronService) ListNetworks(ctx context.Context) ([]Network, error) {
	if !n.IsConnected() {
		return nil, fmt.Errorf("not connected to OpenStack")
	}

	neutronURL := fmt.Sprintf("https://%s:9696/v2.0/networks", n.config.Host)
	
	client := utils.GetInsecureHTTPClient(30 * time.Second)
	req, err := http.NewRequestWithContext(ctx, "GET", neutronURL, nil)
	if err != nil {
		return nil, err
	}
	
	req.Header.Set("X-Auth-Token", n.token)
	
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	
	var result struct {
		Networks []Network `json:"networks"`
	}
	
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	
	return result.Networks, nil
}

// GetNetwork 获取网络详情
func (n *NeutronService) GetNetwork(ctx context.Context, networkID string) (*Network, error) {
	if !n.IsConnected() {
		return nil, fmt.Errorf("not connected to OpenStack")
	}

	neutronURL := fmt.Sprintf("https://%s:9696/v2.0/networks/%s", n.config.Host, networkID)
	
	client := utils.GetInsecureHTTPClient(30 * time.Second)
	req, err := http.NewRequestWithContext(ctx, "GET", neutronURL, nil)
	if err != nil {
		return nil, err
	}
	
	req.Header.Set("X-Auth-Token", n.token)
	
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	
	var result struct {
		Network Network `json:"network"`
	}
	
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	
	return &result.Network, nil
}

// CreateNetwork 创建网络
func (n *NeutronService) CreateNetwork(ctx context.Context, name, networkType string, physicalNetwork string, segmentationID int) (*Network, error) {
	if !n.IsConnected() {
		return nil, fmt.Errorf("not connected to OpenStack")
	}

	neutronURL := fmt.Sprintf("https://%s:9696/v2.0/networks", n.config.Host)
	
	networkReq := map[string]interface{}{
		"network": map[string]interface{}{
			"name": name,
		}
	}
	
	// 添加 provider 网络类型
	if networkType != "" {
		networkReq.(map[string]interface{})["network"].(map[string]interface{})["provider:network_type"] = networkType
	}
	if physicalNetwork != "" {
		networkReq.(map[string]interface{})["network"].(map[string]interface{})["provider:physical_network"] = physicalNetwork
	}
	if segmentationID > 0 {
		networkReq.(map[string]interface{})["network"].(map[string]interface{})["provider:segmentation_id"] = segmentationID
	}
	
	jsonBody, _ := json.Marshal(networkReq)
	
	client := utils.GetInsecureHTTPClient(30 * time.Second)
	req, err := http.NewRequestWithContext(ctx, "POST", neutronURL, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, err
	}
	
	req.Header.Set("X-Auth-Token", n.token)
	req.Header.Set("Content-Type", "application/json")
	
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	
	var result struct {
		Network Network `json:"network"`
	}
	
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	
	global.APP_LOG.Info("Network created", zap.String("name", name), zap.String("network_id", result.Network.ID))
	
	return &result.Network, nil
}

// UpdateNetwork 更新网络
func (n *NeutronService) UpdateNetwork(ctx context.Context, networkID, name string, adminStateUp *bool) (*Network, error) {
	if !n.IsConnected() {
		return nil, fmt.Errorf("not connected to OpenStack")
	}

	neutronURL := fmt.Sprintf("https://%s:9696/v2.0/networks/%s", n.config.Host, networkID)
	
	updateReq := map[string]interface{}{
		"network": map[string]interface{}{
			"name": name,
		}
	}
	
	if adminStateUp != nil {
		updateReq.(map[string]interface{})["network"].(map[string]interface{})["admin_state_up"] = *adminStateUp
	}
	
	jsonBody, _ := json.Marshal(updateReq)
	
	client := utils.GetInsecureHTTPClient(30 * time.Second)
	req, err := http.NewRequestWithContext(ctx, "PUT", neutronURL, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, err
	}
	
	req.Header.Set("X-Auth-Token", n.token)
	req.Header.Set("Content-Type", "application/json")
	
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	
	var result struct {
		Network Network `json:"network"`
	}
	
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	
	return &result.Network, nil
}

// DeleteNetwork 删除网络
func (n *NeutronService) DeleteNetwork(ctx context.Context, networkID string) error {
	if !n.IsConnected() {
		return fmt.Errorf("not connected to OpenStack")
	}

	neutronURL := fmt.Sprintf("https://%s:9696/v2.0/networks/%s", n.config.Host, networkID)
	
	client := utils.GetInsecureHTTPClient(30 * time.Second)
	req, err := http.NewRequestWithContext(ctx, "DELETE", neutronURL, nil)
	if err != nil {
		return err
	}
	
	req.Header.Set("X-Auth-Token", n.token)
	
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != 204 {
		return fmt.Errorf("failed to delete network, status: %d", resp.StatusCode)
	}
	
	global.APP_LOG.Info("Network deleted", zap.String("network_id", networkID))
	
	return nil
}

// ListSubnets 列出所有子网
func (n *NeutronService) ListSubnets(ctx context.Context, networkID string) ([]Subnet, error) {
	if !n.IsConnected() {
		return nil, fmt.Errorf("not connected to OpenStack")
	}

	neutronURL := fmt.Sprintf("https://%s:9696/v2.0/subnets", n.config.Host)
	if networkID != "" {
		neutronURL += "?network_id=" + networkID
	}
	
	client := utils.GetInsecureHTTPClient(30 * time.Second)
	req, err := http.NewRequestWithContext(ctx, "GET", neutronURL, nil)
	if err != nil {
		return nil, err
	}
	
	req.Header.Set("X-Auth-Token", n.token)
	
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	
	var result struct {
		Subnets []Subnet `json:"subnets"`
	}
	
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	
	return result.Subnets, nil
}

// CreateSubnet 创建子网
func (n *NeutronService) CreateSubnet(ctx context.Context, networkID, name, cidr, gatewayIP string, enableDHCP bool, dnsNameservers []string) (*Subnet, error) {
	if !n.IsConnected() {
		return nil, fmt.Errorf("not connected to OpenStack")
	}

	neutronURL := fmt.Sprintf("https://%s:9696/v2.0/subnets", n.config.Host)
	
	subnetReq := map[string]interface{}{
		"subnet": map[string]interface{}{
			"network_id":      networkID,
			"name":            name,
			"cidr":            cidr,
			"gateway_ip":      gatewayIP,
			"enable_dhcp":     enableDHCP,
			"dns_nameservers": dnsNameservers,
		}
	}
	
	jsonBody, _ := json.Marshal(subnetReq)
	
	client := utils.GetInsecureHTTPClient(30 * time.Second)
	req, err := http.NewRequestWithContext(ctx, "POST", neutronURL, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, err
	}
	
	req.Header.Set("X-Auth-Token", n.token)
	req.Header.Set("Content-Type", "application/json")
	
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	
	var result struct {
		Subnet Subnet `json:"subnet"`
	}
	
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	
	global.APP_LOG.Info("Subnet created", zap.String("name", name), zap.String("cidr", cidr))
	
	return &result.Subnet, nil
}

// DeleteSubnet 删除子网
func (n *NeutronService) DeleteSubnet(ctx context.Context, subnetID string) error {
	if !n.IsConnected() {
		return fmt.Errorf("not connected to OpenStack")
	}

	neutronURL := fmt.Sprintf("https://%s:9696/v2.0/subnets/%s", n.config.Host, subnetID)
	
	client := utils.GetInsecureHTTPClient(30 * time.Second)
	req, err := http.NewRequestWithContext(ctx, "DELETE", neutronURL, nil)
	if err != nil {
		return err
	}
	
	req.Header.Set("X-Auth-Token", n.token)
	
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != 204 {
		return fmt.Errorf("failed to delete subnet, status: %d", resp.StatusCode)
	}
	
	global.APP_LOG.Info("Subnet deleted", zap.String("subnet_id", subnetID))
	
	return nil
}

// ListRouters 列出所有路由器
func (n *NeutronService) ListRouters(ctx context.Context) ([]Router, error) {
	if !n.IsConnected() {
		return nil, fmt.Errorf("not connected to OpenStack")
	}

	neutronURL := fmt.Sprintf("https://%s:9696/v2.0/routers", n.config.Host)
	
	client := utils.GetInsecureHTTPClient(30 * time.Second)
	req, err := http.NewRequestWithContext(ctx, "GET", neutronURL, nil)
	if err != nil {
		return nil, err
	}
	
	req.Header.Set("X-Auth-Token", n.token)
	
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	
	var result struct {
		Routers []Router `json:"routers"`
	}
	
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	
	return result.Routers, nil
}

// CreateRouter 创建路由器
func (n *NeutronService) CreateRouter(ctx context.Context, name, externalNetworkID string) (*Router, error) {
	if !n.IsConnected() {
		return nil, fmt.Errorf("not connected to OpenStack")
	}

	neutronURL := fmt.Sprintf("https://%s:9696/v2.0/routers", n.config.Host)
	
	routerReq := map[string]interface{}{
		"router": map[string]interface{}{
			"name": name,
		}
	}
	
	if externalNetworkID != "" {
		routerReq.(map[string]interface{})["router"].(map[string]interface{})["external_gateway_info"] = map[string]interface{}{
			"network_id": externalNetworkID,
		}
	}
	
	jsonBody, _ := json.Marshal(routerReq)
	
	client := utils.GetInsecureHTTPClient(30 * time.Second)
	req, err := http.NewRequestWithContext(ctx, "POST", neutronURL, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, err
	}
	
	req.Header.Set("X-Auth-Token", n.token)
	req.Header.Set("Content-Type", "application/json")
	
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	
	var result struct {
		Router Router `json:"router"`
	}
	
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	
	global.APP_LOG.Info("Router created", zap.String("name", name), zap.String("router_id", result.Router.ID))
	
	return &result.Router, nil
}

// AddRouterInterface 添加路由器接口
func (n *NeutronService) AddRouterInterface(ctx context.Context, routerID, subnetID string) error {
	if !n.IsConnected() {
		return fmt.Errorf("not connected to OpenStack")
	}

	neutronURL := fmt.Sprintf("https://%s:9696/v2.0/routers/%s/add_router_interface", n.config.Host, routerID)
	
	interfaceReq := map[string]interface{}{
		"subnet_id": subnetID,
	}
	
	jsonBody, _ := json.Marshal(interfaceReq)
	
	client := utils.GetInsecureHTTPClient(30 * time.Second)
	req, err := http.NewRequestWithContext(ctx, "PUT", neutronURL, bytes.NewBuffer(jsonBody))
	if err != nil {
		return err
	}
	
	req.Header.Set("X-Auth-Token", n.token)
	req.Header.Set("Content-Type", "application/json")
	
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != 200 {
		return fmt.Errorf("failed to add router interface, status: %d", resp.StatusCode)
	}
	
	global.APP_LOG.Info("Router interface added", zap.String("router_id", routerID), zap.String("subnet_id", subnetID))
	
	return nil
}

// RemoveRouterInterface 移除路由器接口
func (n *NeutronService) RemoveRouterInterface(ctx context.Context, routerID, subnetID string) error {
	if !n.IsConnected() {
		return fmt.Errorf("not connected to OpenStack")
	}

	neutronURL := fmt.Sprintf("https://%s:9696/v2.0/routers/%s/remove_router_interface", n.config.Host, routerID)
	
	interfaceReq := map[string]interface{}{
		"subnet_id": subnetID,
	}
	
	jsonBody, _ := json.Marshal(interfaceReq)
	
	client := utils.GetInsecureHTTPClient(30 * time.Second)
	req, err := http.NewRequestWithContext(ctx, "PUT", neutronURL, bytes.NewBuffer(jsonBody))
	if err != nil {
		return err
	}
	
	req.Header.Set("X-Auth-Token", n.token)
	req.Header.Set("Content-Type", "application/json")
	
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != 200 {
		return fmt.Errorf("failed to remove router interface, status: %d", resp.StatusCode)
	}
	
	global.APP_LOG.Info("Router interface removed", zap.String("router_id", routerID), zap.String("subnet_id", subnetID))
	
	return nil
}

// DeleteRouter 删除路由器
func (n *NeutronService) DeleteRouter(ctx context.Context, routerID string) error {
	if !n.IsConnected() {
		return fmt.Errorf("not connected to OpenStack")
	}

	neutronURL := fmt.Sprintf("https://%s:9696/v2.0/routers/%s", n.config.Host, routerID)
	
	client := utils.GetInsecureHTTPClient(30 * time.Second)
	req, err := http.NewRequestWithContext(ctx, "DELETE", neutronURL, nil)
	if err != nil {
		return err
	}
	
	req.Header.Set("X-Auth-Token", n.token)
	
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != 204 {
		return fmt.Errorf("failed to delete router, status: %d", resp.StatusCode)
	}
	
	global.APP_LOG.Info("Router deleted", zap.String("router_id", routerID))
	
	return nil
}

// ListFloatingIPs 列出浮动IP
func (n *NeutronService) ListFloatingIPs(ctx context.Context) ([]FloatingIP, error) {
	if !n.IsConnected() {
		return nil, fmt.Errorf("not connected to OpenStack")
	}

	neutronURL := fmt.Sprintf("https://%s:9696/v2.0/floatingips", n.config.Host)
	
	client := utils.GetInsecureHTTPClient(30 * time.Second)
	req, err := http.NewRequestWithContext(ctx, "GET", neutronURL, nil)
	if err != nil {
		return nil, err
	}
	
	req.Header.Set("X-Auth-Token", n.token)
	
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	
	var result struct {
		FloatingIPs []FloatingIP `json:"floatingips"`
	}
	
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	
	return result.FloatingIPs, nil
}

// CreateFloatingIP 创建浮动IP
func (n *NeutronService) CreateFloatingIP(ctx context.Context, externalNetworkID, portID string) (*FloatingIP, error) {
	if !n.IsConnected() {
		return nil, fmt.Errorf("not connected to OpenStack")
	}

	neutronURL := fmt.Sprintf("https://%s:9696/v2.0/floatingips", n.config.Host)
	
	fipReq := map[string]interface{}{
		"floatingip": map[string]interface{}{
			"floating_network_id": externalNetworkID,
		}
	}
	
	if portID != "" {
		fipReq.(map[string]interface{})["floatingip"].(map[string]interface{})["port_id"] = portID
	}
	
	jsonBody, _ := json.Marshal(fipReq)
	
	client := utils.GetInsecureHTTPClient(30 * time.Second)
	req, err := http.NewRequestWithContext(ctx, "POST", neutronURL, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, err
	}
	
	req.Header.Set("X-Auth-Token", n.token)
	req.Header.Set("Content-Type", "application/json")
	
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	
	var result struct {
		FloatingIP FloatingIP `json:"floatingip"`
	}
	
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	
	global.APP_LOG.Info("Floating IP created", zap.String("ip", result.FloatingIP.FloatingIPAddress))
	
	return &result.FloatingIP, nil
}

// AssociateFloatingIP 关联浮动IP
func (n *NeutronService) AssociateFloatingIP(ctx context.Context, floatingIPID, portID string) (*FloatingIP, error) {
	if !n.IsConnected() {
		return nil, fmt.Errorf("not connected to OpenStack")
	}

	neutronURL := fmt.Sprintf("https://%s:9696/v2.0/floatingips/%s", n.config.Host, floatingIPID)
	
	updateReq := map[string]interface{}{
		"floatingip": map[string]interface{}{
			"port_id": portID,
		}
	}
	
	jsonBody, _ := json.Marshal(updateReq)
	
	client := utils.GetInsecureHTTPClient(30 * time.Second)
	req, err := http.NewRequestWithContext(ctx, "PUT", neutronURL, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, err
	}
	
	req.Header.Set("X-Auth-Token", n.token)
	req.Header.Set("Content-Type", "application/json")
	
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	
	var result struct {
		FloatingIP FloatingIP `json:"floatingip"`
	}
	
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	
	return &result.FloatingIP, nil
}

// DisassociateFloatingIP 解除浮动IP关联
func (n *NeutronService) DisassociateFloatingIP(ctx context.Context, floatingIPID string) (*FloatingIP, error) {
	if !n.IsConnected() {
		return nil, fmt.Errorf("not connected to OpenStack")
	}

	neutronURL := fmt.Sprintf("https://%s:9696/v2.0/floatingips/%s", n.config.Host, floatingIPID)
	
	updateReq := map[string]interface{}{
		"floatingip": map[string]interface{}{
			"port_id": nil,
		}
	}
	
	jsonBody, _ := json.Marshal(updateReq)
	
	client := utils.GetInsecureHTTPClient(30 * time.Second)
	req, err := http.NewRequestWithContext(ctx, "PUT", neutronURL, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, err
	}
	
	req.Header.Set("X-Auth-Token", n.token)
	req.Header.Set("Content-Type", "application/json")
	
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	
	var result struct {
		FloatingIP FloatingIP `json:"floatingip"`
	}
	
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	
	return &result.FloatingIP, nil
}

// DeleteFloatingIP 删除浮动IP
func (n *NeutronService) DeleteFloatingIP(ctx context.Context, floatingIPID string) error {
	if !n.IsConnected() {
		return fmt.Errorf("not connected to OpenStack")
	}

	neutronURL := fmt.Sprintf("https://%s:9696/v2.0/floatingips/%s", n.config.Host, floatingIPID)
	
	client := utils.GetInsecureHTTPClient(30 * time.Second)
	req, err := http.NewRequestWithContext(ctx, "DELETE", neutronURL, nil)
	if err != nil {
		return err
	}
	
	req.Header.Set("X-Auth-Token", n.token)
	
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != 204 {
		return fmt.Errorf("failed to delete floating IP, status: %d", resp.StatusCode)
	}
	
	global.APP_LOG.Info("Floating IP deleted", zap.String("floating_ip_id", floatingIPID))
	
	return nil
}

// ListPorts 列出所有端口
func (n *NeutronService) ListPorts(ctx context.Context, networkID string) ([]Port, error) {
	if !n.IsConnected() {
		return nil, fmt.Errorf("not connected to OpenStack")
	}

	neutronURL := fmt.Sprintf("https://%s:9696/v2.0/ports", n.config.Host)
	if networkID != "" {
		neutronURL += "?network_id=" + networkID
	}
	
	client := utils.GetInsecureHTTPClient(30 * time.Second)
	req, err := http.NewRequestWithContext(ctx, "GET", neutronURL, nil)
	if err != nil {
		return nil, err
	}
	
	req.Header.Set("X-Auth-Token", n.token)
	
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	
	var result struct {
		Ports []Port `json:"ports"`
	}
	
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	
	return result.Ports, nil
}

// ListExternalNetworks 列出外部网络
func (n *NeutronService) ListExternalNetworks(ctx context.Context) ([]Network, error) {
	if !n.IsConnected() {
		return nil, fmt.Errorf("not connected to OpenStack")
	}

	neutronURL := fmt.Sprintf("https://%s:9696/v2.0/networks?router:external=True")
	
	client := utils.GetInsecureHTTPClient(30 * time.Second)
	req, err := http.NewRequestWithContext(ctx, "GET", neutronURL, nil)
	if err != nil {
		return nil, err
	}
	
	req.Header.Set("X-Auth-Token", n.token)
	
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	
	var result struct {
		Networks []Network `json:"networks"`
	}
	
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	
	return result.Networks, nil
}
