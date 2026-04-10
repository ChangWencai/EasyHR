// expense-list.js
const { get } = require('../../utils/request')
const { requireAuth } = require('../../utils/auth')

Page({
  data: {
    tabs: [{ label: '全部', value: '' }, { label: '待审批', value: 'pending' }, { label: '已通过', value: 'approved' }, { label: '已支付', value: 'paid' }, { label: '已驳回', value: 'rejected' }],
    curTab: 0, list: [], filtered: [], loading: true,
  },
  onLoad() {
    if (!requireAuth()) return
    this.loadExpenses()
  },
  onShow() { this.loadExpenses() },
  loadExpenses() {
    this.setData({ loading: true })
    get('/wxmp/expenses').then(res => {
      const list = res.data || []
      list.forEach(i => i.expanded = false)
      this.setData({ list, filtered: list, loading: false })
    }).catch(() => this.setData({ loading: false }))
  },
  filterByStatus(e) {
    const status = e.currentTarget.dataset.status
    const idx = this.data.tabs.findIndex(t => t.value === status)
    const filtered = status ? this.data.list.filter(i => i.status === status) : this.data.list
    this.setData({ curTab: idx, filtered })
  },
  toggleExpand(e) {
    const id = e.currentTarget.dataset.id
    const list = this.data.list.map(i => i.id === id ? { ...i, expanded: !i.expanded } : i)
    const filtered = this.data.curTab ? list.filter(i => i.status === this.data.tabs[this.data.curTab].value) : list
    this.setData({ list, filtered })
  },
})
