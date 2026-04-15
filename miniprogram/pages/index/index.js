const app = getApp()

Page({
  data: {
    currentTab: 'activity',
    tabs: [
      { key: 'activity', label: '活动' },
      { key: 'challenge', label: '挑战' },
      { key: 'race', label: '赛事' },
      { key: 'club', label: '俱乐部' }
    ],
    activities: [],
    challenges: [
      {
        id: 1,
        icon: '🚴',
        title: '百公里挑战',
        description: '累计骑行 100 公里',
        progress: 65,
        current: 65,
        target: 100,
        unit: '公里',
        badge: '🏅',
        completed: false
      },
      {
        id: 2,
        icon: '⛰️',
        title: '爬坡王者',
        description: '累计爬升 1000 米',
        progress: 30,
        current: 300,
        target: 1000,
        unit: '米',
        badge: '🏔️',
        completed: false
      },
      {
        id: 3,
        icon: '👥',
        title: '社交达人',
        description: '参加 10 场活动',
        progress: 100,
        current: 10,
        target: 10,
        unit: '场',
        badge: '🌟',
        completed: true
      }
    ]
  },

  onLoad() {
    this.loadActivities()
  },

  onShow() {
    if (this.data.currentTab === 'activity') {
      this.loadActivities()
    }
  },

  onPullDownRefresh() {
    if (this.data.currentTab === 'activity') {
      this.loadActivities()
    } else {
      wx.stopPullDownRefresh()
    }
  },

  onTabChange(e) {
    const key = e.currentTarget.dataset.key
    this.setData({ currentTab: key })
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
