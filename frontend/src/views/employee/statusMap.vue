<script setup lang="ts">
// shared status maps used by employee views
export const statusMap: Record<string, string> = {
  active: '在职',
  probation: '试用期',
  resigned: '离职',
  archived: '归档',
}

export const statusTagType: Record<string, string> = {
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

export const invitationStatusTagType: Record<string, string> = {
  pending: 'warning',
  used: 'success',
  expired: 'info',
  cancelled: 'info',
}

export const offboardingStatusMap: Record<string, string> = {
  pending_review: '待审核',
  approved: '已批准',
  completed: '已完成',
}

export const offboardingStatusTagType: Record<string, string> = {
  pending_review: 'warning',
  approved: 'primary',
  completed: 'success',
}
</script>
