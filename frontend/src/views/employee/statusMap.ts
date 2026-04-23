export const statusMap: Record<string, string> = {
  pending: '待入职',
  active: '在职',
  probation: '试用期',
  resigned: '离职',
  archived: '归档',
}

export const statusTagType: Record<string, 'primary' | 'success' | 'warning' | 'info' | 'danger'> = {
  pending: 'primary',
  active: 'success',
  probation: 'warning',
  resigned: 'info',
  archived: 'info',
}

export const invitationStatusMap: Record<string, string> = {
  pending: '待激活',
  used: '已使用',
  expired: '已过期',
  cancelled: '已取消',
}

export const invitationStatusTagType: Record<string, 'primary' | 'success' | 'warning' | 'info' | 'danger'> = {
  pending: 'warning',
  used: 'success',
  expired: 'info',
  cancelled: 'info',
}

export const offboardingStatusMap: Record<string, string> = {
  pending: '待审核',
  approved: '已批准',
  rejected: '已驳回',
  completed: '已完成',
}

export const offboardingStatusTagType: Record<string, 'primary' | 'success' | 'warning' | 'info' | 'danger'> = {
  pending: 'warning',
  approved: 'primary',
  rejected: 'danger',
  completed: 'success',
}
