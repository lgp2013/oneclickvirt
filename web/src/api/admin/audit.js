import request from '@/utils/request'

// 获取审计日志列表
export const getAuditLogs = (params) => {
  return request({
    url: '/v1/admin/monitoring/audit-logs',
    method: 'get',
    params
  })
}
