import request from '@/utils/request'

// 获取集群列表
export const getClusterList = (params) => {
  return request({
    url: '/v1/admin/clusters',
    method: 'get',
    params
  })
}

// 获取集群详情
export const getClusterDetail = (params) => {
  return request({
    url: `/v1/admin/clusters/${params.id}`,
    method: 'get'
  })
}

// 创建集群
export const createCluster = (data) => {
  return request({
    url: '/v1/admin/clusters',
    method: 'post',
    data
  })
}

// 更新集群
export const updateCluster = (data) => {
  return request({
    url: `/v1/admin/clusters/${data.id}`,
    method: 'put',
    data
  })
}

// 删除集群
export const deleteCluster = (data) => {
  return request({
    url: `/v1/admin/clusters/${data.id}`,
    method: 'delete'
  })
}

// 测试连接
export const testConnection = (data) => {
  return request({
    url: '/v1/admin/clusters/test-connection',
    method: 'post',
    data
  })
}

// 获取集群实例列表
export const getClusterInstances = (params) => {
  return request({
    url: `/v1/admin/clusters/${params.id}/instances`,
    method: 'get',
    params
  })
}
