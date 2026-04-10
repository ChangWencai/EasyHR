// payslips.js
const { get } = require('../../utils/request')
const { requireAuth } = require('../../utils/auth')

Page({
  data: { list: [], loading: true },
  onLoad() {
    if (!requireAuth()) return
    this.loadPayslips()
  },
  onShow() { this.loadPayslips() },
  loadPayslips() {
    this.setData({ loading: true })
    get('/wxmp/payslips').then(res => {
      this.setData({ list: res.data || [], loading: false })
    }).catch(() => this.setData({ loading: false }))
  },
  onPayslipTap(e) {
    wx.navigateTo({ url: '/pages/payslips-detail/payslips-detail?id=' + e.currentTarget.dataset.id })
  },
})
