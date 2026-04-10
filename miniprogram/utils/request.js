const API_BASE = 'https://api.easyhr.example.com/api/v1'

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
