const app = getApp()

Page({
  data: {
    currentTab: 'myBike',
    tabs: [
      { key: 'myBike', label: '我的自行车' },
      { key: 'brands', label: '品牌' }
    ],
    bikes: [],
    brands: []
  },

  onLoad() {
    this.loadBikes()
    this.loadBrands()
  },

  onShow() {
    if (this.data.currentTab === 'myBike') {
      this.loadBikes()
    }
  },

  onTabChange(e) {
    const key = e.currentTarget.dataset.key
    this.setData({ currentTab: key })
  },

  onViewAllBrands() {
    this.setData({ currentTab: 'brands' })
  },

  loadBikes() {
    wx.showLoading({ title: '加载中...' })
    const token = wx.getStorageSync('token')

    wx.request({
      url: `${app.globalData.apiBaseUrl}/bikes`,
      method: 'GET',
      header: {
        'Authorization': token ? 'Bearer ' + token : ''
      },
      success: (res) => {
        wx.hideLoading()
        if (res.statusCode === 200 && res.data.data) {
          this.setData({ bikes: res.data.data })
        } else if (res.statusCode === 401) {
          this.setData({ bikes: [] })
        }
      },
      fail: () => {
        wx.hideLoading()
        wx.showToast({ title: '加载失败', icon: 'none' })
      }
    })
  },

  loadBrands() {
    wx.request({
      url: `${app.globalData.apiBaseUrl}/brands`,
      method: 'GET',
      success: (res) => {
        if (res.statusCode === 200 && res.data.data) {
          this.setData({ brands: res.data.data })
        }
      }
    })
  },

  goBikeDetail(e) {
    const id = e.currentTarget.dataset.id
    wx.showToast({ title: '查看自行车详情', icon: 'none' })
  },

  goBrandDetail(e) {
    const { id, name } = e.currentTarget.dataset
    wx.navigateTo({
      url: `/pages/brand/detail?id=${id}&name=${name}`
    })
  }
})
