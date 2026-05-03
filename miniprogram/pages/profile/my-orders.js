const app = getApp()

Page({
  data: {
    orders: [],
    loading: true
  },

  onLoad() {
    this.loadOrders()
  },

  onShow() {
    this.loadOrders()
  },

  loadOrders() {
    this.setData({ loading: true })
    const token = wx.getStorageSync('token')
    wx.request({
      url: `${app.globalData.apiBaseUrl}/rental/orders`,
      method: 'GET',
      header: { 'Authorization': token ? 'Bearer ' + token : '' },
      success: (res) => {
        if (res.statusCode === 200 && res.data.data) {
          const orders = res.data.data.map(item => {
            const statusMap = {
              0: { text: '待付款', class: 'pending' },
              1: { text: '已付款', class: 'paid' },
              2: { text: '已完成', class: 'done' },
              3: { text: '已取消', class: 'cancel' }
            }
            const pkgMap = { day: '日租', hour: '时租', team: '团队租' }
            const st = statusMap[item.status] || statusMap[0]
            return {
              ...item,
              statusText: st.text,
              statusClass: st.class,
              packageText: pkgMap[item.package] || item.package,
              total: item.amount + item.deposit,
              createdAt: item.createdAt ? item.createdAt.split(' ')[0] : ''
            }
          })
          this.setData({ orders })
        }
      },
      fail: () => {
        wx.showToast({ title: '加载失败', icon: 'none' })
      },
      complete: () => {
        this.setData({ loading: false })
      }
    })
  }
})
