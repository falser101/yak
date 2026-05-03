const app = getApp()

Page({
  data: {
    rentalBikes: [],
    rentalTypes: [
      { key: '', label: '全部' },
      { key: '公路车', label: '公路车' },
      { key: '山地车', label: '山地车' },
      { key: '平把公路', label: '平把公路' },
      { key: '电动车', label: '电动车' }
    ],
    currentType: ''
  },

  onLoad() {
    this.loadRentalBikes()
  },

  onTypeChange(e) {
    const type = e.currentTarget.dataset.type || ''
    this.setData({ currentType: type })
    this.loadRentalBikes()
  },

  loadRentalBikes() {
    wx.showLoading({ title: '加载中...' })
    const token = wx.getStorageSync('token')
    let url = `${app.globalData.apiBaseUrl}/rental/bikes`
    if (this.data.currentType) {
      url += `?bikeType=${encodeURIComponent(this.data.currentType)}`
    }

    wx.request({
      url,
      method: 'GET',
      header: {
        'Authorization': token ? 'Bearer ' + token : ''
      },
      success: (res) => {
        wx.hideLoading()
        if (res.statusCode === 200 && res.data.data) {
          this.setData({ rentalBikes: res.data.data })
        }
      },
      fail: () => {
        wx.hideLoading()
        wx.showToast({ title: '加载失败', icon: 'none' })
      }
    })
  },

  goBikeDetail(e) {
    const id = e.currentTarget.dataset.id
    wx.navigateTo({
      url: `/pages/rental/detail?id=${id}`
    })
  }
})
