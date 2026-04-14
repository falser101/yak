App({
  globalData: {
    userInfo: null,
    activities: [],
    challenges: [],
    apiBaseUrl: 'http://localhost:8080/api'
  },

  onLaunch() {
    console.log('App launched')
  }
})
