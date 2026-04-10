// contracts.js
const { get } = require('../../utils/request')
const { requireAuth } = require('../../utils/auth')

Page({
  data: { list: [], loading: true },
  onLoad() {
    if (!requireAuth()) return
    this.setData({ loading: true })
    get('/wxmp/contracts').then(res => {
      this.setData({ list: res.data || [], loading: false })
    }).catch(() => this.setData({ loading: false }))
  },
  viewPDF(e) {
    wx.showLoading({ title: '加载中...' })
    wx.openDocument({
      filePath: e.currentTarget.dataset.url,
      success: () => wx.hideLoading(),
      fail: () => { wx.hideLoading(); wx.showToast({ title: '打开失败', icon: 'none' }) },
    })
  },
})
