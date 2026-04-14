Page({
  data: {
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
    ],
    badges: [
      { icon: '🏅', name: '百公里', owned: true },
      { icon: '🏔️', name: '爬坡王', owned: false },
      { icon: '🌟', name: '社交达人', owned: true },
      { icon: '🔥', name: '坚持不懈', owned: false },
      { icon: '⚡', name: '速度之星', owned: false },
      { icon: '👑', name: '全能王者', owned: false }
    ]
  }
})
