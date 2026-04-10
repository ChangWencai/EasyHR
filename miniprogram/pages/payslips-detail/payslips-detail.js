// payslips-detail.js
const { get, post } = require('../../utils/request')

Page({
  data: { detail: null, verified: false, showModal: false, code: '' },
  onLoad(opts) {
    this._id = opts.id
    get('/wxmp/payslips/' + this._id).then(res => {
      const verified = wx.getStorageSync('verify_' + this._id) === '1'
      this.setData({ detail: res.data, verified })
    })
  },
  showVerify() { this.setData({ showModal: true }) },
  cancelVerify() { this.setData({ showModal: false }) },
  onCodeInput(e) { this.setData({ code: e.detail.value }) },
  verifyPayslip() {
    post('/wxmp/payslips/' + this._id + '/verify', { code: this.data.code }).then(() => {
      wx.setStorageSync('verify_' + this._id, '1')
      this.setData({ verified: true, showModal: false })
      wx.showToast({ title: '验证成功', icon: 'success' })
    }).catch(() => wx.showToast({ title: '验证码错误', icon: 'none' }))
  },
  signPayslip() {
    post('/wxmp/payslips/' + this._id + '/sign').then(() => {
      wx.showToast({ title: '签收成功', icon: 'success' })
      this.onLoad({ id: this._id })
    })
  },
})
