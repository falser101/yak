const app = getApp()

Page({
  data: {
    activity: null,
    isSignedUp: false,
    showDisclaimer: false,
    disclaimerAgreed: false
  },

  onLoad(options) {
    if (options.id) {
      this.loadActivity(options.id)
    }
  },

  loadActivity(id) {
    wx.showLoading({ title: '加载中...' })

    const token = wx.getStorageSync('token')
    wx.request({
      url: `${app.globalData.apiBaseUrl}/activities/${id}`,
      method: 'GET',
      header: {
        'Authorization': token ? 'Bearer ' + token : ''
      },
      success: (res) => {
        wx.hideLoading()
        if (res.statusCode === 200 && res.data.data) {
          const activity = res.data.data
          // 格式化报名截止时间
          if (activity.signupEndTime) {
            const d = new Date(activity.signupEndTime)
            activity.signupEndTime = `${d.getMonth() + 1}-${d.getDate()} ${d.getHours().toString().padStart(2, '0')}:${d.getMinutes().toString().padStart(2, '0')}`
          }
          // 格式化所有报名时间
          if (activity.allSignups && activity.allSignups.length > 0) {
            activity.allSignups.forEach(item => {
              const d = new Date(item.createdAt)
              item.createdAt = `${d.getMonth() + 1}-${d.getDate()} ${d.getHours().toString().padStart(2, '0')}:${d.getMinutes().toString().padStart(2, '0')}`
            })
          }
          this.setData({
            activity,
            isSignedUp: activity.isSignedUp || false,
            currentUser: activity.currentUser
          })
        } else {
          wx.showToast({
            title: '活动不存在',
            icon: 'none'
          })
        }
      },
      fail: (err) => {
        wx.hideLoading()
        wx.showToast({
          title: '加载失败',
          icon: 'none'
        })
        console.error(err)
      }
    })
  },

  // 检查登录
  checkLogin() {
    const userInfo = wx.getStorageSync('userInfo')
    if (!userInfo) {
      wx.showModal({
        title: '提示',
        content: '请先登录',
        success: (res) => {
          if (res.confirm) {
            wx.switchTab({
              url: '/pages/profile/profile'
            })
          }
        }
      })
      return false
    }
    return true
  },

  handleSignup() {
    const { activity, isSignedUp } = this.data

    if (!activity) return

    // 检查是否已满
    if (activity.participants >= activity.maxParticipants) {
      wx.showToast({
        title: '报名已满',
        icon: 'none'
      })
      return
    }

    // 检查是否已报名
    if (isSignedUp) {
      wx.showToast({
        title: '您已报名',
        icon: 'none'
      })
      return
    }

    // 检查登录
    if (!this.checkLogin()) {
      return
    }

    // 检查实名认证状态
    const userInfo = wx.getStorageSync('userInfo')
    if (!userInfo.rzStatus || userInfo.rzStatus < 2) {
      wx.showModal({
        title: '提示',
        content: '请先完成实名认证才能报名',
        confirmText: '去认证',
        cancelText: '取消',
        success: (res) => {
          if (res.confirm) {
            wx.navigateTo({
              url: '/pages/auth/auth'
            })
          }
        }
      })
      return
    }

    // 显示免责声明
    this.setData({ showDisclaimer: true, disclaimerAgreed: false })
  },

  // 关闭免责声明
  closeDisclaimer() {
    this.setData({ showDisclaimer: false, disclaimerAgreed: false })
  },

  // 免责声明勾选变化
  onDisclaimerAgreeChange(e) {
    this.setData({
      disclaimerAgreed: e.detail.value.includes('agreed')
    })
  },

  // 确认免责声明并直接报名
  confirmDisclaimer() {
    if (!this.data.disclaimerAgreed) {
      wx.showToast({
        title: '请先阅读并同意声明',
        icon: 'none'
      })
      return
    }
    this.setData({ showDisclaimer: false })
    // 直接提交报名
    this.submitSignup()
  },

  // 提交报名
  submitSignup() {
    const { activity } = this.data
    const activityId = activity.id

    // 获取当前用户信息（已实名认证）
    const userInfo = wx.getStorageSync('userInfo')
    const token = wx.getStorageSync('token')

    // 直接使用用户实名信息（包含身份证号）
    const formData = {
      name: userInfo.rzRealName || userInfo.nickname || '微信用户',
      phone: userInfo.phone || '',
      idNumber: userInfo.rzIdCard || '',
      emergencyContact: userInfo.rzEmergencyName || '',
      emergencyPhone: userInfo.rzEmergencyPhone || '',
      remark: ''
    }

    wx.showLoading({ title: '报名中...' })

    wx.request({
      url: `${app.globalData.apiBaseUrl}/activities/${activityId}/signup`,
      method: 'POST',
      data: formData,
      header: {
        'Authorization': token ? 'Bearer ' + token : ''
      },
      success: (res) => {
        wx.hideLoading()
        if (res.statusCode === 200) {
          wx.showToast({
            title: '报名成功',
            icon: 'success'
          })
          this.loadActivity(activityId)
        } else {
          wx.showToast({
            title: res.data.error || '报名失败',
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
  }
})
