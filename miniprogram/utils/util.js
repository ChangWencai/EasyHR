function formatAmount(amount) {
  const num = parseFloat(amount)
  return isNaN(num) ? '0.00' : num.toLocaleString('zh-CN', { minimumFractionDigits: 2, maximumFractionDigits: 2 })
}

function formatMonth(year, month) {
  return year + '年' + String(month).padStart(2, '0') + '月'
}

function formatDate(dateStr) {
  if (!dateStr) return ''
  const d = new Date(dateStr)
  return d.getFullYear() + '年' + (d.getMonth() + 1) + '月' + d.getDate() + '日'
}

module.exports = { formatAmount, formatMonth, formatDate }
