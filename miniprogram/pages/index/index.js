const app = getApp()

Page({
  data: {
    activities: []
  },

  onLoad() {
    console.log('活动列表页加载')
    this.loadActivities()
  },

  onShow() {
    // 每次显示页面时刷新活动列表
    this.loadActivities()
  },

  onPullDownRefresh() {
    this.loadActivities()
  },

  loadActivities() {
    wx.showLoading({ title: '加载中...' })

    const token = wx.getStorageSync('token')
    wx.request({
      url: `${app.globalData.apiBaseUrl}/activities`,
      method: 'GET',
      header: {
        'Authorization': token ? 'Bearer ' + token : ''
      },
      success: (res) => {
        if (res.statusCode === 200 && res.data.data) {
          const activities = res.data.data.map(item => {
            // 格式化报名截止时间
            if (item.signupEndTime) {
              const d = new Date(item.signupEndTime)
              item.signupEndTime = `${d.getMonth() + 1}/${d.getDate()} ${d.getHours().toString().padStart(2, '0')}:${d.getMinutes().toString().padStart(2, '0')}`
            }
            return item
          })
          this.setData({ activities })
        }
        wx.hideLoading()
        wx.stopPullDownRefresh()
      },
      fail: (err) => {
        console.error('加载活动失败', err)
        wx.hideLoading()
        wx.stopPullDownRefresh()
        wx.showToast({
          title: '加载失败',
          icon: 'none'
        })
      }
    })
  },

  onSearch(e) {
    const keyword = e.detail.value.trim()
    if (!keyword) {
      this.loadActivities()
      return
    }

    // 前端简单过滤
    const filtered = this.data.activities.filter(item =>
      item.title.toLowerCase().includes(keyword.toLowerCase())
    )
    this.setData({ activities: filtered })
  },

  goDetail(e) {
    const id = e.currentTarget.dataset.id
    wx.navigateTo({
      url: '/pages/activity/detail?id=' + id
    })
  }
})
