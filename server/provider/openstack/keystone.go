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
)

// KeystoneService OpenStack Keystone 认证服务
type KeystoneService struct {
	*OpenStack
}

// Token 认证Token
type Token struct {
	ID        string    `json:"id"`
	ExpiresAt time.Time `json:"expires_at"`
	Project   Project   `json:"project"`
	User      User      `json:"user"`
}

// Project 项目实体
type Project struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	DomainID    string `json:"domain_id"`
	Description string `json:"description"`
	Enabled     bool   `json:"enabled"`
}

// User 用户实体
type User struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	DomainID    string `json:"domain_id"`
	Email       string `json:"email"`
	Enabled     bool   `json:"enabled"`
	DefaultProjectID string `json:"default_project_id"`
}

// Role 角色实体
type Role struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// Domain 域实体
type Domain struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Enabled     bool   `json:"enabled"`
}

// NewKeystoneService 创建 Keystone 服务
func NewKeystoneService(os *OpenStack) *KeystoneService {
	return &KeystoneService{OpenStack: os}
}

// AuthenticateWithPassword 使用密码认证
func (k *KeystoneService) AuthenticateWithPassword(ctx context.Context, username, password, domainName string) (*Token, error) {
	authURL := fmt.Sprintf("https://%s:5000/v3/auth/tokens", k.config.Host)
	
	// 如果没有指定 domain，使用默认
	if domainName == "" {
		domainName = "Default"
	}
	
	authBody := map[string]interface{}{
		"auth": map[string]interface{}{
			"identity": map[string]interface{}{
				"methods": []string{"password"},
				"password": map[string]interface{}{
					"user": map[string]interface{}{
						"name":     username,
						"password": password,
						"domain": map[string]interface{}{
							"name": domainName,
						},
					},
				},
			},
			"scope": map[string]interface{}{
				"project": map[string]interface{}{
					"name": k.projectID,
					"domain": map[string]interface{}{
						"name": "Default",
					},
				},
			},
		},
	}
	
	jsonBody, _ := json.Marshal(authBody)
	
	client := utils.GetInsecureHTTPClient(30 * time.Second)
	req, err := http.NewRequestWithContext(ctx, "POST", authURL, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, err
	}
	
	req.Header.Set("Content-Type", "application/json")
	
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != 201 {
		return nil, fmt.Errorf("authentication failed, status: %d", resp.StatusCode)
	}
	
	// 获取 token
	tokenID := resp.Header.Get("X-Subject-Token")
	if tokenID == "" {
		return nil, fmt.Errorf("no token received")
	}
	
	var result struct {
		Token Token `json:"token"`
	}
	
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	
	result.Token.ID = tokenID
	
	global.APP_LOG.Info("Keystone authentication successful", zap.String("username", username), zap.String("project", k.projectID))
	
	return &result.Token, nil
}

// ValidateToken 验证Token
func (k *KeystoneService) ValidateToken(ctx context.Context, token string) (*Token, error) {
	authURL := fmt.Sprintf("https://%s:5000/v3/auth/tokens", k.config.Host)
	
	client := utils.GetInsecureHTTPClient(30 * time.Second)
	req, err := http.NewRequestWithContext(ctx, "GET", authURL, nil)
	if err != nil {
		return nil, err
	}
	
	req.Header.Set("X-Auth-Token", token)
	req.Header.Set("X-Subject-Token", token)
	
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("token validation failed, status: %d", resp.StatusCode)
	}
	
	var result struct {
		Token Token `json:"token"`
	}
	
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	
	return &result.Token, nil
}

// GetTokenInfo 获取Token信息
func (k *KeystoneService) GetTokenInfo(ctx context.Context) (*Token, error) {
	if k.token == "" {
		return nil, fmt.Errorf("not authenticated")
	}
	return k.ValidateToken(ctx, k.token)
}

// ListProjects 列出项目
func (k *KeystoneService) ListProjects(ctx context.Context, domainID string) ([]Project, error) {
	if !k.IsConnected() {
		return nil, fmt.Errorf("not connected to OpenStack")
	}

	keystoneURL := fmt.Sprintf("https://%s:5000/v3/projects", k.config.Host)
	if domainID != "" {
		keystoneURL += "?domain_id=" + domainID
	}
	
	client := utils.GetInsecureHTTPClient(30 * time.Second)
	req, err := http.NewRequestWithContext(ctx, "GET", keystoneURL, nil)
	if err != nil {
		return nil, err
	}
	
	req.Header.Set("X-Auth-Token", k.token)
	
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	
	var result struct {
		Projects []Project `json:"projects"`
	}
	
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	
	return result.Projects, nil
}

// GetProject 获取项目详情
func (k *KeystoneService) GetProject(ctx context.Context, projectID string) (*Project, error) {
	if !k.IsConnected() {
		return nil, fmt.Errorf("not connected to OpenStack")
	}

	keystoneURL := fmt.Sprintf("https://%s:5000/v3/projects/%s", k.config.Host, projectID)
	
	client := utils.GetInsecureHTTPClient(30 * time.Second)
	req, err := http.NewRequestWithContext(ctx, "GET", keystoneURL, nil)
	if err != nil {
		return nil, err
	}
	
	req.Header.Set("X-Auth-Token", k.token)
	
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	
	var result struct {
		Project Project `json:"project"`
	}
	
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	
	return &result.Project, nil
}

// CreateProject 创建项目
func (k *KeystoneService) CreateProject(ctx context.Context, name, domainID, description string, enabled bool) (*Project, error) {
	if !k.IsConnected() {
		return nil, fmt.Errorf("not connected to OpenStack")
	}

	keystoneURL := fmt.Sprintf("https://%s:5000/v3/projects", k.config.Host)
	
	// 如果没有指定 domain，使用默认
	if domainID == "" {
		domainID = "default"
	}
	
	projectReq := map[string]interface{}{
		"project": map[string]interface{}{
			"name":        name,
			"domain_id":   domainID,
			"description": description,
			"enabled":    enabled,
		}
	}
	
	jsonBody, _ := json.Marshal(projectReq)
	
	client := utils.GetInsecureHTTPClient(30 * time.Second)
	req, err := http.NewRequestWithContext(ctx, "POST", keystoneURL, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, err
	}
	
	req.Header.Set("X-Auth-Token", k.token)
	req.Header.Set("Content-Type", "application/json")
	
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	
	var result struct {
		Project Project `json:"project"`
	}
	
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	
	global.APP_LOG.Info("Project created", zap.String("name", name), zap.String("project_id", result.Project.ID))
	
	return &result.Project, nil
}

// UpdateProject 更新项目
func (k *KeystoneService) UpdateProject(ctx context.Context, projectID, name, description string, enabled *bool) (*Project, error) {
	if !k.IsConnected() {
		return nil, fmt.Errorf("not connected to OpenStack")
	}

	keystoneURL := fmt.Sprintf("https://%s:5000/v3/projects/%s", k.config.Host, projectID)
	
	updateReq := map[string]interface{}{
		"project": map[string]interface{}{
			"name":        name,
			"description": description,
		}
	}
	
	if enabled != nil {
		updateReq.(map[string]interface{})["project"].(map[string]interface{})["enabled"] = *enabled
	}
	
	jsonBody, _ := json.Marshal(updateReq)
	
	client := utils.GetInsecureHTTPClient(30 * time.Second)
	req, err := http.NewRequestWithContext(ctx, "PATCH", keystoneURL, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, err
	}
	
	req.Header.Set("X-Auth-Token", k.token)
	req.Header.Set("Content-Type", "application/json")
	
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	
	var result struct {
		Project Project `json:"project"`
	}
	
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	
	return &result.Project, nil
}

// DeleteProject 删除项目
func (k *KeystoneService) DeleteProject(ctx context.Context, projectID string) error {
	if !k.IsConnected() {
		return fmt.Errorf("not connected to OpenStack")
	}

	keystoneURL := fmt.Sprintf("https://%s:5000/v3/projects/%s", k.config.Host, projectID)
	
	client := utils.GetInsecureHTTPClient(30 * time.Second)
	req, err := http.NewRequestWithContext(ctx, "DELETE", keystoneURL, nil)
	if err != nil {
		return err
	}
	
	req.Header.Set("X-Auth-Token", k.token)
	
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != 204 {
		return fmt.Errorf("failed to delete project, status: %d", resp.StatusCode)
	}
	
	global.APP_LOG.Info("Project deleted", zap.String("project_id", projectID))
	
	return nil
}

// ListUsers 列出用户
func (k *KeystoneService) ListUsers(ctx context.Context, projectID string) ([]User, error) {
	if !k.IsConnected() {
		return nil, fmt.Errorf("not connected to OpenStack")
	}

	keystoneURL := fmt.Sprintf("https://%s:5000/v3/users", k.config.Host)
	if projectID != "" {
		keystoneURL += "?default_project_id=" + projectID
	}
	
	client := utils.GetInsecureHTTPClient(30 * time.Second)
	req, err := http.NewRequestWithContext(ctx, "GET", keystoneURL, nil)
	if err != nil {
		return nil, err
	}
	
	req.Header.Set("X-Auth-Token", k.token)
	
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	
	var result struct {
		Users []User `json:"users"`
	}
	
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	
	return result.Users, nil
}

// GetUser 获取用户详情
func (k *KeystoneService) GetUser(ctx context.Context, userID string) (*User, error) {
	if !k.IsConnected() {
		return nil, fmt.Errorf("not connected to OpenStack")
	}

	keystoneURL := fmt.Sprintf("https://%s:5000/v3/users/%s", k.config.Host, userID)
	
	client := utils.GetInsecureHTTPClient(30 * time.Second)
	req, err := http.NewRequestWithContext(ctx, "GET", keystoneURL, nil)
	if err != nil {
		return nil, err
	}
	
	req.Header.Set("X-Auth-Token", k.token)
	
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	
	var result struct {
		User User `json:"user"`
	}
	
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	
	return &result.User, nil
}

// CreateUser 创建用户
func (k *KeystoneService) CreateUser(ctx context.Context, name, password, email, domainID, projectID string, enabled bool) (*User, error) {
	if !k.IsConnected() {
		return nil, fmt.Errorf("not connected to OpenStack")
	}

	keystoneURL := fmt.Sprintf("https://%s:5000/v3/users", k.config.Host)
	
	// 如果没有指定 domain，使用默认
	if domainID == "" {
		domainID = "default"
	}
	
	userReq := map[string]interface{}{
		"user": map[string]interface{}{
			"name":     name,
			"password": password,
			"email":    email,
			"domain_id": domainID,
			"enabled":  enabled,
		}
	}
	
	if projectID != "" {
		userReq.(map[string]interface{})["user"].(map[string]interface{})["default_project_id"] = projectID
	}
	
	jsonBody, _ := json.Marshal(userReq)
	
	client := utils.GetInsecureHTTPClient(30 * time.Second)
	req, err := http.NewRequestWithContext(ctx, "POST", keystoneURL, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, err
	}
	
	req.Header.Set("X-Auth-Token", k.token)
	req.Header.Set("Content-Type", "application/json")
	
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	
	var result struct {
		User User `json:"user"`
	}
	
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	
	global.APP_LOG.Info("User created", zap.String("name", name), zap.String("user_id", result.User.ID))
	
	return &result.User, nil
}

// UpdateUser 更新用户
func (k *KeystoneService) UpdateUser(ctx context.Context, userID, name, email string, enabled *bool) (*User, error) {
	if !k.IsConnected() {
		return nil, fmt.Errorf("not connected to OpenStack")
	}

	keystoneURL := fmt.Sprintf("https://%s:5000/v3/users/%s", k.config.Host, userID)
	
	updateReq := map[string]interface{}{
		"user": map[string]interface{}{
			"name":   name,
			"email":  email,
		}
	}
	
	if enabled != nil {
		updateReq.(map[string]interface{})["user"].(map[string]interface{})["enabled"] = *enabled
	}
	
	jsonBody, _ := json.Marshal(updateReq)
	
	client := utils.GetInsecureHTTPClient(30 * time.Second)
	req, err := http.NewRequestWithContext(ctx, "PATCH", keystoneURL, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, err
	}
	
	req.Header.Set("X-Auth-Token", k.token)
	req.Header.Set("Content-Type", "application/json")
	
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	
	var result struct {
		User User `json:"user"`
	}
	
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	
	return &result.User, nil
}

// ChangeUserPassword 修改用户密码
func (k *KeystoneService) ChangeUserPassword(ctx context.Context, userID, oldPassword, newPassword string) error {
	if !k.IsConnected() {
		return fmt.Errorf("not connected to OpenStack")
	}

	keystoneURL := fmt.Sprintf("https://%s:5000/v3/users/%s/password", k.config.Host, userID)
	
	passwordReq := map[string]interface{}{
		"user": map[string]interface{}{
			"password":     newPassword,
			"original_password": oldPassword,
		}
	}
	
	jsonBody, _ := json.Marshal(passwordReq)
	
	client := utils.GetInsecureHTTPClient(30 * time.Second)
	req, err := http.NewRequestWithContext(ctx, "POST", keystoneURL, bytes.NewBuffer(jsonBody))
	if err != nil {
		return err
	}
	
	req.Header.Set("X-Auth-Token", k.token)
	req.Header.Set("Content-Type", "application/json")
	
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != 204 {
		return fmt.Errorf("failed to change password, status: %d", resp.StatusCode)
	}
	
	global.APP_LOG.Info("User password changed", zap.String("user_id", userID))
	
	return nil
}

// DeleteUser 删除用户
func (k *KeystoneService) DeleteUser(ctx context.Context, userID string) error {
	if !k.IsConnected() {
		return fmt.Errorf("not connected to OpenStack")
	}

	keystoneURL := fmt.Sprintf("https://%s:5000/v3/users/%s", k.config.Host, userID)
	
	client := utils.GetInsecureHTTPClient(30 * time.Second)
	req, err := http.NewRequestWithContext(ctx, "DELETE", keystoneURL, nil)
	if err != nil {
		return err
	}
	
	req.Header.Set("X-Auth-Token", k.token)
	
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != 204 {
		return fmt.Errorf("failed to delete user, status: %d", resp.StatusCode)
	}
	
	global.APP_LOG.Info("User deleted", zap.String("user_id", userID))
	
	return nil
}

// ListRoles 列出角色
func (k *KeystoneService) ListRoles(ctx context.Context) ([]Role, error) {
	if !k.IsConnected() {
		return nil, fmt.Errorf("not connected to OpenStack")
	}

	keystoneURL := fmt.Sprintf("https://%s:5000/v3/roles", k.config.Host)
	
	client := utils.GetInsecureHTTPClient(30 * time.Second)
	req, err := http.NewRequestWithContext(ctx, "GET", keystoneURL, nil)
	if err != nil {
		return nil, err
	}
	
	req.Header.Set("X-Auth-Token", k.token)
	
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	
	var result struct {
		Roles []Role `json:"roles"`
	}
	
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	
	return result.Roles, nil
}

// GrantProjectRole 授予项目角色
func (k *KeystoneService) GrantProjectRole(ctx context.Context, userID, projectID, roleID string) error {
	if !k.IsConnected() {
		return fmt.Errorf("not connected to OpenStack")
	}

	keystoneURL := fmt.Sprintf("https://%s:5000/v3/projects/%s/users/%s/roles/%s", k.config.Host, projectID, userID, roleID)
	
	client := utils.GetInsecureHTTPClient(30 * time.Second)
	req, err := http.NewRequestWithContext(ctx, "PUT", keystoneURL, nil)
	if err != nil {
		return err
	}
	
	req.Header.Set("X-Auth-Token", k.token)
	
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != 204 && resp.StatusCode != 201 {
		return fmt.Errorf("failed to grant role, status: %d", resp.StatusCode)
	}
	
	global.APP_LOG.Info("Role granted", zap.String("user_id", userID), zap.String("project_id", projectID), zap.String("role_id", roleID))
	
	return nil
}

// RevokeProjectRole 撤销项目角色
func (k *KeystoneService) RevokeProjectRole(ctx context.Context, userID, projectID, roleID string) error {
	if !k.IsConnected() {
		return fmt.Errorf("not connected to OpenStack")
	}

	keystoneURL := fmt.Sprintf("https://%s:5000/v3/projects/%s/users/%s/roles/%s", k.config.Host, projectID, userID, roleID)
	
	client := utils.GetInsecureHTTPClient(30 * time.Second)
	req, err := http.NewRequestWithContext(ctx, "DELETE", keystoneURL, nil)
	if err != nil {
		return err
	}
	
	req.Header.Set("X-Auth-Token", k.token)
	
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != 204 {
		return fmt.Errorf("failed to revoke role, status: %d", resp.StatusCode)
	}
	
	global.APP_LOG.Info("Role revoked", zap.String("user_id", userID), zap.String("project_id", projectID), zap.String("role_id", roleID))
	
	return nil
}

// ListDomains 列出域
func (k *KeystoneService) ListDomains(ctx context.Context) ([]Domain, error) {
	if !k.IsConnected() {
		return nil, fmt.Errorf("not connected to OpenStack")
	}

	keystoneURL := fmt.Sprintf("https://%s:5000/v3/domains", k.config.Host)
	
	client := utils.GetInsecureHTTPClient(30 * time.Second)
	req, err := http.NewRequestWithContext(ctx, "GET", keystoneURL, nil)
	if err != nil {
		return nil, err
	}
	
	req.Header.Set("X-Auth-Token", k.token)
	
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	
	var result struct {
		Domains []Domain `json:"domains"`
	}
	
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	
	return result.Domains, nil
}
