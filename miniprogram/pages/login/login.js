// login.js
const { post } = require('../../utils/request')
const { setToken, setMemberInfo } = require('../../utils/auth')
const { sendVerifyCode, countDown } = require('../../utils/sms')

Page({
  data: { phone: '', code: '', countdown: 0, loading: false },
  onPhoneInput(e) { this.setData({ phone: e.detail.value }) },
  onCodeInput(e) { this.setData({ code: e.detail.value }) },
  sendCode() {
    const { phone } = this.data
    if (!phone || phone.length < 11) { wx.showToast({ title: '请输入正确手机号', icon: 'none' }); return }
    sendVerifyCode(phone).then(() => {
      this.setData({ countdown: 60 })
      this._cancelTimer = countDown(60, (s) => this.setData({ countdown: s }))
      wx.showToast({ title: '验证码已发送', icon: 'success' })
    }).catch(() => wx.showToast({ title: '发送失败', icon: 'none' }))
  },
  login() {
    const { phone, code, loading } = this.data
    if (!phone || !code) { wx.showToast({ title: '请填写完整信息', icon: 'none' }); return }
    if (loading) return
    this.setData({ loading: true })
    post('/wxmp/auth/login', { phone, code }).then(res => {
      setToken(res.token)
      setMemberInfo(res.member_info)
      this._cancelTimer && this._cancelTimer()
      wx.showToast({ title: '登录成功', icon: 'success' })
      setTimeout(() => wx.switchTab({ url: '/pages/payslips/payslips' }), 800)
    }).catch(() => wx.showToast({ title: '登录失败', icon: 'none' })).finally(() => this.setData({ loading: false }))
  },
})
