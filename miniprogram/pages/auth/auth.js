const app = getApp()

Page({
  data: {
    realName: '',
    idCard: '',
    gender: 0,
    emergencyName: '',
    emergencyPhone: '',
    userPhone: '',
    phoneLoading: false,
    canSubmit: false
  },

  onLoad() {
    // 加载当前认证状态
    this.loadRzStatus()
  },

  onShow() {
    this.checkCanSubmit()
  },

  // 加载当前实名认证状态
  loadRzStatus() {
    const token = wx.getStorageSync('token')
    if (!token) {
      return
    }

    wx.request({
      url: `${app.globalData.apiBaseUrl}/auth/rz_status`,
      method: 'GET',
      header: { 'Authorization': 'Bearer ' + token },
      success: (res) => {
        if (res.statusCode === 200 && res.data.data) {
          const data = res.data.data
          if (data.rzStatus === 2) {
            // 已认证，显示已有信息
            this.setData({
              realName: data.rzRealName || '',
              idCard: '******************', // 脱敏显示
              gender: data.rzGender || 0,
              emergencyName: data.rzEmergencyName || '',
              emergencyPhone: data.rzEmergencyPhone || ''
            })
          }
          // 加载用户手机号
          if (data.phone) {
            this.setData({ userPhone: data.phone })
            const userInfo = wx.getStorageSync('userInfo') || {}
            userInfo.phone = data.phone
            wx.setStorageSync('userInfo', userInfo)
          }
        }
      }
    })
  },

  onInput(e) {
    const field = e.currentTarget.dataset.field
    this.setData({
      [field]: e.detail.value
    })
    this.checkCanSubmit()
  },

  onGenderChange(e) {
    this.setData({
      gender: parseInt(e.detail.value)
    })
    this.checkCanSubmit()
  },

  // 微信获取手机号
  onGetPhoneNumber(e) {
    if (e.detail.errMsg === 'getPhoneNumber:ok') {
      this.setData({ phoneLoading: true })
      const token = wx.getStorageSync('token')
      wx.request({
        url: `${app.globalData.apiBaseUrl}/auth/decrypt_phone`,
        method: 'POST',
        header: {
          'Authorization': 'Bearer ' + token,
          'Content-Type': 'application/json'
        },
        data: {
          encryptData: e.detail.encryptedData,
          iv: e.detail.iv
        },
        success: (res) => {
          if (res.statusCode === 200 && res.data.phone) {
            this.setData({ userPhone: res.data.phone })
            // 更新本地存储
            const userInfo = wx.getStorageSync('userInfo') || {}
            userInfo.phone = res.data.phone
            wx.setStorageSync('userInfo', userInfo)
            wx.showToast({ title: '获取成功', icon: 'success' })
          } else {
            wx.showToast({ title: res.data.error || '获取失败', icon: 'none' })
          }
        },
        fail: () => {
          wx.showToast({ title: '网络错误', icon: 'none' })
        },
        complete: () => {
          this.setData({ phoneLoading: false })
        }
      })
    } else {
      console.log('getPhoneNumber失败:', e.detail.errMsg)
      wx.showToast({ title: '获取失败: ' + e.detail.errMsg, icon: 'none' })
    }
  },

  checkCanSubmit() {
    const { realName, idCard, gender, userPhone, emergencyName, emergencyPhone } = this.data
    const canSubmit = realName.length > 0 &&
                      idCard.length === 18 &&
                      gender > 0 &&
                      userPhone.length > 0 &&
                      emergencyName.length > 0 &&
                      emergencyPhone.length > 0
    this.setData({ canSubmit })
  },

  submitAuth() {
    const { realName, idCard, gender, userPhone, emergencyName, emergencyPhone } = this.data
    const token = wx.getStorageSync('token')

    if (!token) {
      wx.showToast({
        title: '请先登录',
        icon: 'none'
      })
      return
    }

    wx.showLoading({ title: '提交中...' })

    wx.request({
      url: `${app.globalData.apiBaseUrl}/auth/rz`,
      method: 'POST',
      header: {
        'Authorization': 'Bearer ' + token,
        'Content-Type': 'application/json'
      },
      data: {
        realName,
        idCard,
        gender,
        emergencyName,
        emergencyPhone
      },
      success: (res) => {
        wx.hideLoading()
        if (res.statusCode === 200 && res.data.data) {
          // 更新本地用户信息
          const userInfo = wx.getStorageSync('userInfo') || {}
          userInfo.rzStatus = res.data.data.rzStatus
          userInfo.rzRealName = res.data.data.rzRealName
          userInfo.rzGender = res.data.data.rzGender
          userInfo.rzEmergencyName = res.data.data.rzEmergencyName
          userInfo.rzEmergencyPhone = res.data.data.rzEmergencyPhone
          wx.setStorageSync('userInfo', userInfo)

          wx.showToast({
            title: '认证成功',
            icon: 'success',
            success: () => {
              setTimeout(() => {
                wx.navigateBack()
              }, 1500)
            }
          })
        } else {
          wx.showToast({
            title: res.data.error || '提交失败',
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
