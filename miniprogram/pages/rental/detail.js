const app = getApp()

Page({
  data: {
    bike: null,
    selectedPackage: 'day',
    quantity: 1,
    rentalDate: '',
    currentPrice: 0,
    priceUnit: '天',
    totalPrice: 0,
    specsList: []
  },

  onLoad(options) {
    if (options.id) {
      this.loadBike(options.id)
    }
    // 默认日期为明天
    const tomorrow = new Date()
    tomorrow.setDate(tomorrow.getDate() + 1)
    this.setData({
      rentalDate: tomorrow.toISOString().split('T')[0]
    })
  },

  loadBike(id) {
    wx.showLoading({ title: '加载中...' })
    const token = wx.getStorageSync('token')
    wx.request({
      url: `${app.globalData.apiBaseUrl}/rental/bikes/${id}`,
      method: 'GET',
      header: {
        'Authorization': token ? 'Bearer ' + token : ''
      },
      success: (res) => {
        wx.hideLoading()
        if (res.statusCode === 200 && res.data.data) {
          const bike = res.data.data
          const specsList = this.formatSpecs(bike.specs)
          this.setData({
            bike,
            specsList,
            currentPrice: bike.priceDay,
            totalPrice: bike.priceDay
          })
        }
      },
      fail: () => {
        wx.hideLoading()
        wx.showToast({ title: '加载失败', icon: 'none' })
      }
    })
  },

  formatSpecs(specs) {
    if (!specs || typeof specs !== 'object') return []
    const labelMap = {
      frame: '车架',
      derailleur: '变速器',
      brake: '刹车',
      wheel: '轮径',
      height: '适合身高'
    }
    return Object.entries(specs)
      .filter(([key, val]) => val && labelMap[key])
      .map(([key, val]) => ({ label: labelMap[key], value: val }))
  },

  onPackageChange(e) {
    const pkg = e.currentTarget.dataset.package
    const { bike } = this.data
    const unitMap = { day: '天', hour: '时', team: '人' }
    const priceMap = { day: bike.priceDay, hour: bike.priceHour, team: bike.priceTeam }

    this.setData({
      selectedPackage: pkg,
      currentPrice: priceMap[pkg],
      priceUnit: unitMap[pkg],
      totalPrice: priceMap[pkg] * this.data.quantity
    })
  },

  onQuantityChange(e) {
    const delta = e.currentTarget.dataset.delta
    const newQty = this.data.quantity + delta
    if (newQty >= 1 && newQty <= 10) {
      const { bike, selectedPackage } = this.data
      const priceMap = { day: bike.priceDay, hour: bike.priceHour, team: bike.priceTeam }
      this.setData({
        quantity: newQty,
        totalPrice: priceMap[selectedPackage] * newQty
      })
    }
  },

  onDateChange(e) {
    this.setData({ rentalDate: e.detail.value })
  },

  goBack() {
    wx.navigateBack()
  },

  handleBooking() {
    const userInfo = wx.getStorageSync('userInfo')
    if (!userInfo) {
      wx.showModal({
        title: '提示',
        content: '请先登录',
        success: (res) => {
          if (res.confirm) {
            wx.switchTab({ url: '/pages/profile/profile' })
          }
        }
      })
      return
    }

    const { bike, selectedPackage, quantity } = this.data
    const token = wx.getStorageSync('token')

    const formData = {
      bikeId: bike.id,
      package: selectedPackage,
      quantity,
      contactName: userInfo.rzRealName || userInfo.nickname || '微信用户',
      contactPhone: userInfo.phone || '',
      remark: ''
    }

    wx.showLoading({ title: '提交中...' })
    wx.request({
      url: `${app.globalData.apiBaseUrl}/rental/orders`,
      method: 'POST',
      data: formData,
      header: {
        'Authorization': token ? 'Bearer ' + token : '',
        'Content-Type': 'application/json'
      },
      success: (res) => {
        wx.hideLoading()
        if ((res.statusCode === 200 || res.statusCode === 201) && res.data.data) {
          const order = res.data.data
          wx.showModal({
            title: '订单已创建',
            content: `订单号: ${order.orderNo}\n租金: ¥${order.amount}\n押金: ¥${order.deposit}\n合计: ¥${order.total}\n\n请到店支付`,
            showCancel: false,
            confirmText: '我知道了',
            success: () => {
              wx.switchTab({ url: '/pages/profile/profile' })
            }
          })
        } else {
          wx.showToast({ title: res.data.error || '创建订单失败', icon: 'none' })
        }
      },
      fail: () => {
        wx.hideLoading()
        wx.showToast({ title: '网络错误', icon: 'none' })
      }
    })
  }
})
