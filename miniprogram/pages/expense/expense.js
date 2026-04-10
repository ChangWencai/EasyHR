// expense.js
const { post } = require('../../utils/request')
const { requireAuth } = require('../../utils/auth')

Page({
  data: { types: ['差旅费', '交通费', '招待费', '办公费', '其他'], typeIndex: -1, amount: '', description: '', photos: [], submitting: false },
  onLoad() { if (!requireAuth()) return },
  onTypeChange(e) { this.setData({ typeIndex: e.detail.value }) },
  onAmountInput(e) { this.setData({ amount: e.detail.value }) },
  onDescInput(e) { this.setData({ description: e.detail.value }) },
  chooseImage() {
    wx.chooseMedia({ count: 9 - this.data.photos.length, mediaType: ['image'], success: res => {
      const tempFiles = res.tempFiles || []
      const photos = [...this.data.photos]
      tempFiles.forEach(f => photos.push(f.tempFilePath))
      this.setData({ photos })
    }})
  },
  delPhoto(e) {
    const idx = e.currentTarget.dataset.index
    const photos = this.data.photos.filter((_, i) => i !== idx)
    this.setData({ photos })
  },
  submitExpense() {
    const { types, typeIndex, amount, description, photos, submitting } = this.data
    if (typeIndex < 0 || !amount) { wx.showToast({ title: '请填写完整信息', icon: 'none' }); return }
    if (submitting) return
    wx.showModal({ title: '确认提交', content: '确定提交这笔 ¥' + amount + ' 的报销申请？', success: res => {
      if (!res.confirm) return
      this.setData({ submitting: true })
      this._uploadAndSubmit(photos, { expense_type: types[typeIndex], amount, description })
        .then(() => {
          wx.showToast({ title: '提交成功', icon: 'success' })
          setTimeout(() => wx.switchTab({ url: '/pages/mine/mine' }), 800)
        })
        .catch(() => wx.showToast({ title: '提交失败', icon: 'none' }))
        .finally(() => this.setData({ submitting: false }))
    }})
  },
  async _uploadAndSubmit(photoPaths, formData) {
    const fileKeys = []
    for (const path of photoPaths) {
      const info = wx.getFileSystemManager().getFileInfo({ filePath: path })
      const fileName = path.split('/').pop()
      const uploadRes = await post('/wxmp/oss/upload-url', { fileName, fileSize: info.size })
      await new Promise((resolve, reject) => {
        wx.uploadFile({ url: uploadRes.url, filePath: path, name: 'file', success: () => resolve(), fail: reject })
      })
      fileKeys.push(uploadRes.key)
    }
    return post('/wxmp/expenses', { ...formData, attachments: fileKeys })
  },
})
