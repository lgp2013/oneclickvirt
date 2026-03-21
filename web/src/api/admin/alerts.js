import request from '@/utils/request'

// 获取告警列表
export const getAlerts = (params) => {
  return request({
    url: '/v1/admin/alerts',
    method: 'get',
    params
  })
}

// 获取未读告警数量
export const getUnreadCount = () => {
  return request({
    url: '/v1/admin/alerts/unread-count',
    method: 'get'
  })
}

// 标记告警已读
export const markAsRead = (data) => {
  return request({
    url: `/v1/admin/alerts/${data.id}/read`,
    method: 'put'
  })
}

// 标记全部已读
export const markAllAsRead = () => {
  return request({
    url: '/v1/admin/alerts/read-all',
    method: 'put'
  })
}

// 解决告警
export const resolveAlert = (data) => {
  return request({
    url: `/v1/admin/alerts/${data.id}/resolve`,
    method: 'put'
  })
}

// 删除告警
export const deleteAlert = (data) => {
  return request({
    url: `/v1/admin/alerts/${data.id}`,
    method: 'delete'
  })
}
