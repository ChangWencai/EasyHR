// mine.js
const { getMemberInfo, clearToken } = require('../../utils/auth')

Page({
  data: { info: {} },
  onLoad() { this.setData({ info: getMemberInfo() || {} }) },
  onShow() { this.setData({ info: getMemberInfo() || {} }) },
  logout() {
    wx.showModal({
      title: '确认退出', content: '确定退出登录？',
      success: res => {
        if (!res.confirm) return
        clearToken()
        wx.redirectTo({ url: '/pages/login/login' })
      },
    })
  },
})
