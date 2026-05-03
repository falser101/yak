const app = getApp()

Page({
  data: {
    currentTab: '',
    tabs: [
      { key: '', label: '全部' },
      { key: 'activity', label: '赛事' },
      { key: 'charity', label: '公益' },
      { key: 'club', label: '俱乐部' }
    ],
    activities: [],
    allActivities: []
  },

  onLoad() {
    this.loadActivities()
  },

  onShow() {
    this.loadActivities()
  },

  onPullDownRefresh() {
    this.loadActivities()
  },

  onTabChange(e) {
    const key = e.currentTarget.dataset.key
    this.setData({ currentTab: key })
    this.filterActivities()
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
            // 格式化日期显示
            if (item.date) {
              const d = new Date(item.date)
              item.date = `${d.getMonth() + 1}月${d.getDate()}日 ${d.getHours().toString().padStart(2, '0')}:${d.getMinutes().toString().padStart(2, '0')}`
            }
            if (item.signupEndTime) {
              const d = new Date(item.signupEndTime)
              item.signupEndTime = `${d.getMonth() + 1}/${d.getDate()} ${d.getHours().toString().padStart(2, '0')}:${d.getMinutes().toString().padStart(2, '0')}`
            }
            return item
          })
          this.setData({
            allActivities: activities,
            activities: activities
          })
          this.filterActivities()
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

  filterActivities() {
    const { allActivities, currentTab } = this.data
    if (!currentTab) {
      this.setData({ activities: allActivities })
    } else {
      const filtered = allActivities.filter(item => item.category === currentTab)
      this.setData({ activities: filtered })
    }
  },

  onSearch(e) {
    const keyword = e.detail.value.trim()
    if (!keyword) {
      this.filterActivities()
      return
    }

    const { allActivities, currentTab } = this.data
    let filtered = allActivities.filter(item =>
      item.title.toLowerCase().includes(keyword.toLowerCase())
    )
    if (currentTab) {
      filtered = filtered.filter(item => item.category === currentTab)
    }
    this.setData({ activities: filtered })
  },

  goDetail(e) {
    const id = e.currentTarget.dataset.id
    wx.navigateTo({
      url: '/pages/activity/detail?id=' + id
    })
  }
})
