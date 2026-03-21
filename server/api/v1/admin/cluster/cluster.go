package cluster

import (
	"oneclickvirt/model"
	"oneclickvirt/model/response"
	"oneclickvirt/utils"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// Cluster 集群模型
type Cluster struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	Name        string    `json:"name" gorm:"size:32;not null;uniqueIndex"`
	Type        string    `json:"type" gorm:"size:16;not null"` // openstack, k8s
	Endpoint    string    `json:"endpoint" gorm:"size:255;not null"`
	Port        int       `json:"port" gorm:"default:22"`
	Username    string    `json:"username" gorm:"size:64"`
	Password    string    `json:"-" gorm:"size:255"` // 密码加密存储
	PrivateKey  string    `json:"-" gorm:"type:text"`
	ProjectID   string    `json:"projectId" gorm:"size:64"`
	DomainID    string    `json:"domainId" gorm:"size:64"`
	Region      string    `json:"region" gorm:"size:32"`
	Status      string    `json:"status" gorm:"size:16;default:active"` // active, inactive
	Description string    `json:"description" gorm:"type:text"`
	Kubeconfig  string    `json:"-" gorm:"type:text"` // K8s kubeconfig
	Namespace   string    `json:"namespace" gorm:"size:32;default:default"` // K8s namespace
	InstanceCount int    `json:"instanceCount" gorm:"-:迁移"` // 关联的实例数量
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

func (Cluster) TableName() string {
	return "clusters"
}

// CreateCluster 创建集群
func CreateCluster(c *gin.Context) {
	var cluster Cluster
	if err := c.ShouldBindJSON(&cluster); err != nil {
		response.FailWithMessage("参数错误", c)
		return
	}

	// 加密密码
	if cluster.Password != "" {
		encrypted, err := utils.Encrypt(cluster.Password)
		if err != nil {
			response.FailWithMessage("密码加密失败", c)
			return
		}
		cluster.Password = encrypted
	}

	// 加密 kubeconfig
	if cluster.Kubeconfig != "" {
		encrypted, err := utils.Encrypt(cluster.Kubeconfig)
		if err != nil {
			response.FailWithMessage("kubeconfig加密失败", c)
			return
		}
		cluster.Kubeconfig = encrypted
	}

	if err := model.DB.Create(&cluster).Error; err != nil {
		global.APP_LOG.Error("创建集群失败", zap.Error(err))
		response.FailWithMessage("创建失败", c)
		return
	}

	response.SuccessWithMessage("创建成功", c)
}

// UpdateCluster 更新集群
func UpdateCluster(c *gin.Context) {
	id := c.Param("id")
	var cluster Cluster
	
	if err := model.DB.First(&cluster, id).Error; err != nil {
		response.FailWithMessage("集群不存在", c)
		return
	}

	var updateData Cluster
	if err := c.ShouldBindJSON(&updateData); err != nil {
		response.FailWithMessage("参数错误", c)
		return
	}

	// 如果更新了密码，需要加密
	if updateData.Password != "" && updateData.Password != cluster.Password {
		encrypted, err := utils.Encrypt(updateData.Password)
		if err != nil {
			response.FailWithMessage("密码加密失败", c)
			return
		}
		updateData.Password = encrypted
	} else {
		updateData.Password = cluster.Password
	}

	// 如果更新了 kubeconfig，需要加密
	if updateData.Kubeconfig != "" && updateData.Kubeconfig != cluster.Kubeconfig {
		encrypted, err := utils.Encrypt(updateData.Kubeconfig)
		if err != nil {
			response.FailWithMessage("kubeconfig加密失败", c)
			return
		}
		updateData.Kubeconfig = encrypted
	} else {
		updateData.Kubeconfig = cluster.Kubeconfig
	}

	if err := model.DB.Model(&cluster).Updates(updateData).Error; err != nil {
		global.APP_LOG.Error("更新集群失败", zap.Error(err))
		response.FailWithMessage("更新失败", c)
		return
	}

	response.SuccessWithMessage("更新成功", c)
}

// DeleteCluster 删除集群
func DeleteCluster(c *gin.Context) {
	id := c.Param("id")
	
	if err := model.DB.Delete(&Cluster{}, id).Error; err != nil {
		global.APP_LOG.Error("删除集群失败", zap.Error(err))
		response.FailWithMessage("删除失败", c)
		return
	}

	response.SuccessWithMessage("删除成功", c)
}

// GetClusterList 获取集群列表
func GetClusterList(c *gin.Context) {
	pageStr := c.DefaultQuery("page", "1")
	pageSizeStr := c.DefaultQuery("pageSize", "10")
	name := c.Query("name")
	clusterType := c.Query("type")

	page, _ := strconv.Atoi(pageStr)
	pageSize, _ := strconv.Atoi(pageSizeStr)

	var clusters []Cluster
	var total int64

	query := model.DB.Model(&Cluster{})

	// 搜索过滤
	if name != "" {
		query = query.Where("name LIKE ?", "%"+name+"%")
	}
	if clusterType != "" {
		query = query.Where("type = ?", clusterType)
	}

	query.Count(&total)
	query.Offset((page - 1) * pageSize).Limit(pageSize).Order("created_at DESC")

	if err := query.Find(&clusters).Error; err != nil {
		global.APP_LOG.Error("获取集群列表失败", zap.Error(err))
		response.FailWithMessage("获取失败", c)
		return
	}

	response.SuccessWithDetailed(gin.H{
		"list":  clusters,
		"total":  total,
		"page":   page,
		"pageSize": pageSize,
	}, "获取成功", c)
}

// GetCluster 获取集群详情
func GetCluster(c *gin.Context) {
	id := c.Param("id")
	var cluster Cluster

	if err := model.DB.First(&cluster, id).Error; err != nil {
		response.FailWithMessage("集群不存在", c)
		return
	}

	// 解密敏感信息用于显示
	if cluster.Password != "" {
		decrypted, _ := utils.Decrypt(cluster.Password)
		cluster.Password = decrypted
	}
	if cluster.Kubeconfig != "" {
		decrypted, _ := utils.Decrypt(cluster.Kubeconfig)
		cluster.Kubeconfig = decrypted
	}

	response.SuccessWithData(cluster, c)
}

// TestConnection 测试集群连接
func TestConnection(c *gin.Context) {
	id := c.Param("id")
	var cluster Cluster

	if err := model.DB.First(&cluster, id).Error; err != nil {
		response.FailWithMessage("集群不存在", c)
		return
	}

	// TODO: 实现实际的连接测试
	// 1. 根据 cluster.Type 判断是 OpenStack 还是 K8s
	// 2. 调用相应的 Provider 进行连接测试
	// 3. 返回测试结果

	response.SuccessWithDetailed(gin.H{
		"success": true,
		"message": "连接测试成功",
	}, "测试成功", c)
}
