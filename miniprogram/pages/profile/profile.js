const app = getApp()

Page({
  data: {
    userInfo: null,
    isLoggedIn: false,
    stats: {
      activityCount: 0,
      joinCount: 0,
      distance: 0
    }
  },

  onLoad() {
    this.checkLogin()
  },

  onShow() {
    this.checkLogin()
  },

  // 检查登录状态
  checkLogin() {
    const userInfo = wx.getStorageSync('userInfo')
    if (userInfo) {
      this.setData({
        userInfo,
        isLoggedIn: true
      })
      // 加载用户的统计数据
      this.loadUserStats()
    }
  },

  // 微信登录
  doLogin() {
    wx.showLoading({ title: '登录中...' })

    // 微信登录
    wx.login({
      success: (res) => {
        if (res.code) {
          wx.request({
            url: `${app.globalData.apiBaseUrl}/auth/login`,
            method: 'POST',
            data: {
              code: res.code
            },
            success: (result) => {
              if (result.statusCode === 200 && result.data.user) {
                this.handleLoginSuccess(result.data.user)
              } else {
                wx.hideLoading()
                wx.showToast({
                  title: result.data.error || '登录失败',
                  icon: 'none'
                })
              }
            },
            fail: () => {
              wx.hideLoading()
              wx.showToast({
                title: '网络错误',
                icon: 'none'
              })
            }
          })
        } else {
          wx.hideLoading()
          wx.showToast({
            title: '微信登录失败',
            icon: 'none'
          })
        }
      },
      fail: () => {
        wx.hideLoading()
        wx.showToast({
          title: '微信登录失败',
          icon: 'none'
        })
      }
    })
  },

  // 处理登录成功
  handleLoginSuccess(userInfo) {
    wx.setStorageSync('userInfo', userInfo)
    wx.setStorageSync('token', userInfo.id)
    this.setData({
      userInfo,
      isLoggedIn: true
    })
    wx.hideLoading()
    wx.showToast({
      title: '登录成功',
      icon: 'success'
    })
    this.loadUserStats()
  },

  // 加载用户统计数据
  loadUserStats() {
    // TODO: 从 API 加载用户的报名统计
    this.setData({
      stats: {
        activityCount: 0,
        joinCount: 0,
        distance: 0
      }
    })
  },

  // 退出登录
  logout() {
    wx.removeStorageSync('userInfo')
    wx.removeStorageSync('token')
    this.setData({
      userInfo: null,
      isLoggedIn: false
    })
    wx.showToast({
      title: '已退出',
      icon: 'none'
    })
  },

  goPage(e) {
    const page = e.currentTarget.dataset.page
    if (!this.data.isLoggedIn) {
      wx.showToast({
        title: '请先登录',
        icon: 'none'
      })
      return
    }
    if (page === 'my-activity') {
      wx.navigateTo({
        url: '/pages/my-activity/my-activity'
      })
      return
    }
    wx.showToast({
      title: '功能开发中...',
      icon: 'none'
    })
  }
})
