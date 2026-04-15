const app = getApp()

Page({
  data: {
    brandName: '',
    brand: {
      logo: '',
      desc: '',
      bikes: []
    }
  },

  onLoad(options) {
    const { id, name } = options
    this.setData({ brandName: name })
    this.loadBrandDetail(id)
  },

  loadBrandDetail(id) {
    wx.showLoading({ title: '加载中...' })

    wx.request({
      url: `${app.globalData.apiBaseUrl}/brands/${id}`,
      method: 'GET',
      success: (res) => {
        wx.hideLoading()
        if (res.statusCode === 200 && res.data.data) {
          const { brand, models } = res.data.data
          this.setData({
            brand: {
              logo: brand.logo,
              description: brand.description,
              bikes: models || []
            }
          })
        }
      },
      fail: () => {
        wx.hideLoading()
        wx.showToast({ title: '加载失败', icon: 'none' })
      }
    })
  }
})
