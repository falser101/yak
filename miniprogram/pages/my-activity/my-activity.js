const app = getApp()

Page({
  data: {
    loading: true,
    signups: []
  },

  onLoad() {
    this.loadMySignups()
  },

  onShow() {
    this.loadMySignups()
  },

  loadMySignups() {
    const token = wx.getStorageSync('token')
    if (!token) {
      this.setData({ loading: false, signups: [] })
      wx.showToast({ title: '请先登录', icon: 'none' })
      return
    }

    this.setData({ loading: true })
    wx.request({
      url: `${app.globalData.apiBaseUrl}/my/signups`,
      method: 'GET',
      header: { 'Authorization': 'Bearer ' + token },
      success: (res) => {
        if (res.statusCode === 200 && res.data.data) {
          const signups = res.data.data.map(item => this.formatSignup(item))
          this.setData({ signups })
        } else {
          this.setData({ signups: [] })
        }
      },
      fail: () => {
        wx.showToast({ title: '加载失败', icon: 'none' })
        this.setData({ signups: [] })
      },
      complete: () => {
        this.setData({ loading: false })
      }
    })
  },

  formatSignup(item) {
    // 格式化日期
    const date = new Date(item.date)
    const dateStr = `${date.getMonth() + 1}-${date.getDate()} ${date.getHours().toString().padStart(2, '0')}:${date.getMinutes().toString().padStart(2, '0')}`

    const signupDate = new Date(item.createdAt)
    const signupTimeStr = `${signupDate.getMonth() + 1}-${signupDate.getDate()} ${signupDate.getHours().toString().padStart(2, '0')}:${signupDate.getMinutes().toString().padStart(2, '0')}`

    // 报名状态: 1=已报名, 2=已取消, 3=已完成
    const signupStatusMap = { 1: '已报名', 2: '已取消', 3: '已完成' }
    // 活动状态: 0=报名中, 1=已满, 2=已结束
    const activityStatusMap = { 0: '报名中', 1: '已满', 2: '已结束' }

    return {
      ...item,
      dateStr,
      signupTimeStr,
      signupStatusText: signupStatusMap[item.signupStatus] || '未知',
      activityStatusText: activityStatusMap[item.activityStatus] || '未知'
    }
  },

  goDetail(e) {
    const id = e.currentTarget.dataset.id
    wx.navigateTo({
      url: `/pages/activity/detail?id=${id}`
    })
  }
})
