const app = getApp()

Page({
  data: {
    userInfo: null,
    isLoggedIn: false,
    signupCount: 0
  },

  onLoad() {
    this.checkLogin()
  },

  onShow() {
    this.checkLogin()
  },

  checkLogin() {
    const userInfo = wx.getStorageSync('userInfo')
    if (userInfo) {
      this.setData({
        userInfo,
        isLoggedIn: true
      })
      this.loadUserStats()
    }
  },

  doLogin() {
    wx.showLoading({ title: '登录中...' })

    wx.login({
      success: (res) => {
        if (res.code) {
          wx.request({
            url: `${app.globalData.apiBaseUrl}/auth/login`,
            method: 'POST',
            data: { code: res.code },
            success: (result) => {
              if (result.statusCode === 200 && result.data.user) {
                this.handleLoginSuccess(result.data.user)
              } else {
                wx.hideLoading()
                wx.showToast({ title: result.data.error || '登录失败', icon: 'none' })
              }
            },
            fail: () => {
              wx.hideLoading()
              wx.showToast({ title: '网络错误', icon: 'none' })
            }
          })
        } else {
          wx.hideLoading()
          wx.showToast({ title: '微信登录失败', icon: 'none' })
        }
      },
      fail: () => {
        wx.hideLoading()
        wx.showToast({ title: '微信登录失败', icon: 'none' })
      }
    })
  },

  handleLoginSuccess(userInfo) {
    wx.setStorageSync('userInfo', userInfo)
    wx.setStorageSync('token', userInfo.id)
    this.setData({ userInfo, isLoggedIn: true })
    wx.hideLoading()
    wx.showToast({ title: '登录成功', icon: 'success' })
    this.loadUserStats()
  },

  loadUserStats() {
    const token = wx.getStorageSync('token')
    wx.request({
      url: `${app.globalData.apiBaseUrl}/activity-signups/my`,
      method: 'GET',
      header: { 'Authorization': token ? 'Bearer ' + token : '' },
      success: (res) => {
        if (res.statusCode === 200 && res.data.data) {
          this.setData({ signupCount: res.data.data.length })
        }
      }
    })
  },

  logout() {
    wx.removeStorageSync('userInfo')
    wx.removeStorageSync('token')
    this.setData({ userInfo: null, isLoggedIn: false, signupCount: 0 })
    wx.showToast({ title: '已退出', icon: 'none' })
  },

  goPage(e) {
    const page = e.currentTarget.dataset.page
    if (!this.data.isLoggedIn && ['my-activity', 'my-orders', 'profile-edit'].includes(page)) {
      wx.showToast({ title: '请先登录', icon: 'none' })
      return
    }
    switch (page) {
      case 'my-activity':
        wx.navigateTo({ url: '/pages/my-activity/my-activity' })
        break
      case 'my-orders':
        wx.navigateTo({ url: '/pages/profile/my-orders' })
        break
      default:
        wx.showToast({ title: '功能开发中...', icon: 'none' })
    }
  }
})
