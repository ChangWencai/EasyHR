// social.js
const { get } = require('../../utils/request')
const { requireAuth } = require('../../utils/auth')

Page({
  data: { record: null, loading: true },
  onLoad() {
    if (!requireAuth()) return
    this.loadSocial()
  },
  loadSocial() {
    this.setData({ loading: true })
    get('/wxmp/social-insurance').then(res => {
      this.setData({ record: res.data || null, loading: false })
    }).catch(() => this.setData({ loading: false }))
  },
})
