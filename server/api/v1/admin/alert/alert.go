package alert

import (
	"oneclickvirt/global"
	"oneclickvirt/model"
	"oneclickvirt/model/response"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// Alert 告警模型
type Alert struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	Title       string    `json:"title" gorm:"size:128;not null"`
	Content     string    `json:"content" gorm:"type:text"`
	Level       string    `json:"level" gorm:"size:16;default:info"` // info, warning, error, critical
	Type        string    `json:"type" gorm:"size:32"` // system, instance, cluster, user
	Source      string    `json:"source" gorm:"size:64"` // 来源ID
	Status      string    `json:"status" gorm:"size:16;default:unread"` // unread, read, resolved
	IsRead      bool      `json:"isRead" gorm:"default:false"`
	ReadAt      *time.Time `json:"readAt"`
	ResolvedAt  *time.Time `json:"resolvedAt"`
	CreatedAt   time.Time `json:"createdAt"`
}

func (Alert) TableName() string {
	return "alerts"
}

// CreateAlert 创建告警
func CreateAlert(title, content, level, alertType, source string) {
	alert := Alert{
		Title:   title,
		Content: content,
		Level:   level,
		Type:    alertType,
		Source:  source,
		Status:  "unread",
	}
	model.DB.Create(&alert)
}

// GetAlerts 获取告警列表
func GetAlerts(c *gin.Context) {
	level := c.Query("level")
	status := c.Query("status")
	alertType := c.Query("type")
	page := c.DefaultQuery("page", "1")
	pageSize := c.DefaultQuery("pageSize", "10")

	var alerts []Alert
	var total int64

	query := model.DB.Model(&Alert{})

	if level != "" {
		query = query.Where("level = ?", level)
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}
	if alertType != "" {
		query = query.Where("type = ?", alertType)
	}

	query.Count(&total)
	query.Offset((parseInt(page) - 1) * parseInt(pageSize)).Limit(parseInt(pageSize)).Order("created_at DESC")

	if err := query.Find(&alerts).Error; err != nil {
		global.APP_LOG.Error("获取告警列表失败", zap.Error(err))
		response.FailWithMessage("获取失败", c)
		return
	}

	response.SuccessWithDetailed(gin.H{
		"list":  alerts,
		"total": total,
	}, "获取成功", c)
}

// MarkAsRead 标记已读
func MarkAsRead(c *gin.Context) {
	id := c.Param("id")
	
	if err := model.DB.Model(&Alert{}).Where("id = ?", id).Updates(map[string]interface{}{
		"is_read": true,
		"status":  "read",
		"read_at": time.Now(),
	}).Error; err != nil {
		response.FailWithMessage("标记失败", c)
		return
	}

	response.SuccessWithMessage("标记成功", c)
}

// MarkAllAsRead 标记全部已读
func MarkAllAsRead(c *gin.Context) {
	if err := model.DB.Model(&Alert{}).Where("is_read = ?", false).Updates(map[string]interface{}{
		"is_read":  true,
		"status":   "read",
		"read_at":   time.Now(),
	}).Error; err != nil {
		response.FailWithMessage("标记失败", c)
		return
	}

	response.SuccessWithMessage("全部标记成功", c)
}

// ResolveAlert 解决告警
func ResolveAlert(c *gin.Context) {
	id := c.Param("id")
	
	if err := model.DB.Model(&Alert{}).Where("id = ?", id).Updates(map[string]interface{}{
		"status":     "resolved",
		"resolved_at": time.Now(),
	}).Error; err != nil {
		response.FailWithMessage("解决失败", c)
		return
	}

	response.SuccessWithMessage("解决成功", c)
}

// DeleteAlert 删除告警
func DeleteAlert(c *gin.Context) {
	id := c.Param("id")
	
	if err := model.DB.Delete(&Alert{}, id).Error; err != nil {
		response.FailWithMessage("删除失败", c)
		return
	}

	response.SuccessWithMessage("删除成功", c)
}

// GetUnreadCount 获取未读告警数量
func GetUnreadCount(c *gin.Context) {
	var count int64
	model.DB.Model(&Alert{}).Where("is_read = ?", false).Count(&count)
	response.SuccessWithData(gin.H{"count": count}, c)
}

func parseInt(s string) int {
	if s == "" {
		return 1
	}
	var n int
	for _, c := range s {
		if c >= '0' && c <= '9' {
			n = n*10 + int(c-'0')
		}
	}
	return n
}
