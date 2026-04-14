const app = getApp()

Page({
  data: {
    title: '',
    cover: '',
    coverId: '',
    actionDate: '',
    location: '',
    maxParticipants: '',
    price: '',
    description: '',
    dateTimeArray: [
      ['2026 年', '2027 年'],
      ['01 月', '02 月', '03 月', '04 月', '05 月', '06 月', '07 月', '08 月', '09 月', '10 月', '11 月', '12 月'],
      ['01 日', '02 日', '03 日', '04 日', '05 日', '06 日', '07 日', '08 日', '09 日', '10 日', '11 日', '12 日', '13 日', '14 日', '15 日', '16 日', '17 日', '18 日', '19 日', '20 日', '21 日', '22 日', '23 日', '24 日', '25 日', '26 日', '27 日', '28 日', '29 日', '30 日', '31 日'],
      ['00:00', '01:00', '02:00', '03:00', '04:00', '05:00', '06:00', '07:00', '08:00', '09:00', '10:00', '11:00', '12:00', '13:00', '14:00', '15:00', '16:00', '17:00', '18:00', '19:00', '20:00', '21:00', '22:00', '23:00']
    ],
    dateTimeValue: [0, 3, 0, 8]
  },

  onTitleInput(e) {
    this.setData({ title: e.detail.value })
  },

  onLocationInput(e) {
    this.setData({ location: e.detail.value })
  },

  onMaxInput(e) {
    this.setData({ maxParticipants: parseInt(e.detail.value) || 0 })
  },

  onPriceInput(e) {
    this.setData({ price: parseFloat(e.detail.value) || 0 })
  },

  onDescInput(e) {
    this.setData({ description: e.detail.value })
  },

  onDateChange(e) {
    const values = e.detail.value
    const year = this.data.dateTimeArray[0][values[0]]
    const month = this.data.dateTimeArray[1][values[1]]
    const day = this.data.dateTimeArray[2][values[2]]
    const time = this.data.dateTimeArray[3][values[3]]
    this.setData({
      actionDate: year.replace('年', '-') + month.replace('月', '-') + day.replace('日', '') + ' ' + time,
      dateTimeValue: values
    })
  },

  chooseCover() {
    wx.chooseImage({
      count: 1,
      success: (res) => {
        const tempFilePath = res.tempFilePaths[0]
        this.setData({ cover: tempFilePath })
        // 本地调试先不上传图片，直接用临时路径
        // 后续可添加上传到后端的逻辑
      }
    })
  },

  submitForm() {
    const { title, cover, actionDate, location, maxParticipants, price, description } = this.data

    if (!title || !actionDate || !location || !maxParticipants || !price || !description) {
      wx.showToast({
        title: '请填写完整信息',
        icon: 'none'
      })
      return
    }

    wx.showLoading({ title: '发布中...' })

    wx.request({
      url: `${app.globalData.apiBaseUrl}/activities`,
      method: 'POST',
      data: {
        title,
        cover: cover || '',
        date: actionDate,
        location,
        maxParticipants: parseInt(maxParticipants),
        price: parseFloat(price),
        description,
        createdBy: 'unknown'
      },
      success: (res) => {
        wx.hideLoading()
        if (res.statusCode === 201) {
          wx.showToast({
            title: '发布成功',
            icon: 'success'
          })
          setTimeout(() => {
            wx.navigateBack()
          }, 1500)
        } else {
          wx.showToast({
            title: res.data.error || '发布失败',
            icon: 'none'
          })
        }
      },
      fail: (err) => {
        wx.hideLoading()
        wx.showToast({
          title: '网络错误',
          icon: 'none'
        })
        console.error(err)
      }
    })
  }
})
