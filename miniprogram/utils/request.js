// 小程序开发阶段：使用本机局域网 IP（如 192.168.x.x:8089），上线前改为生产域名
// 也可通过 project.private.config.json 的 compileVars 注入
// 微信开发者工具自带请求转发，开发者工具中可请求 http://localhost:8089
// 真机调试需要在微信公众平台配置合法域名（或开发阶段关闭 urlCheck）
const API_BASE = 'http://localhost:8089/api/v1'

function request(options) {
  return new Promise((resolve, reject) => {
    const token = wx.getStorageSync('token')
    const header = Object.assign({ 'Content-Type': 'application/json' }, options.header || {})
    if (token) header['Authorization'] = 'Bearer ' + token

    wx.request({
      url: API_BASE + options.url,
      method: options.method || 'GET',
      data: options.data,
      header,
      timeout: options.timeout || 10000,
      success(res) {
        if (res.statusCode === 401) {
          wx.removeStorageSync('token')
          wx.removeStorageSync('member_info')
          wx.redirectTo({ url: '/pages/login/login' })
          return reject(new Error('登录已过期'))
        }
        if (res.statusCode >= 400) return reject(res.data || { message: '请求失败' })
        resolve(res.data)
      },
      fail(err) {
        wx.showToast({ title: '网络连接失败', icon: 'none' })
        reject(err)
      },
    })
  })
}

module.exports = {
  request,
  get: (url, data) => request({ url, method: 'GET', data }),
  post: (url, data) => request({ url, method: 'POST', data }),
  put: (url, data) => request({ url, method: 'PUT', data }),
  del: (url, data) => request({ url, method: 'DELETE', data }),
}
