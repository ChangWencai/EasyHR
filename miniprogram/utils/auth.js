const TOKEN_KEY = 'token'
const MEMBER_KEY = 'member_info'

function getToken() { return wx.getStorageSync(TOKEN_KEY) }
function setToken(t) { wx.setStorageSync(TOKEN_KEY, t) }
function clearToken() {
  wx.removeStorageSync(TOKEN_KEY)
  wx.removeStorageSync(MEMBER_KEY)
}
function getMemberInfo() { return wx.getStorageSync(MEMBER_KEY) || null }
function setMemberInfo(info) { wx.setStorageSync(MEMBER_KEY, info) }
function requireAuth() {
  if (!getToken()) {
    wx.redirectTo({ url: '/pages/login/login' })
    return false
  }
  return true
}

module.exports = { getToken, setToken, clearToken, getMemberInfo, setMemberInfo, requireAuth }
