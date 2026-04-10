App({
  globalData: {
    API_BASE: 'https://api.easyhr.example.com/api/v1',
    userInfo: null,
  },
  onLaunch() {
    const token = wx.getStorageSync('token')
    if (!token) {
      wx.redirectTo({ url: '/pages/login/login' })
    }
  },
})
