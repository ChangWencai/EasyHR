<template>
  <div class="employee-detail">
    <el-card v-loading="loading">
      <template #header>
        <div class="header">
          <span>员工详情</span>
          <el-button @click="$router.back()">返回</el-button>
        </div>
      </template>

      <el-descriptions :column="2" border v-if="employee">
        <el-descriptions-item label="姓名">{{ employee.name }}</el-descriptions-item>
        <el-descriptions-item label="手机号">{{ employee.phone }}</el-descriptions-item>
        <el-descriptions-item label="身份证号">{{ employee.id_number }}</el-descriptions-item>
        <el-descriptions-item label="岗位">{{ employee.position }}</el-descriptions-item>
        <el-descriptions-item label="入职日期">{{ employee.entry_date }}</el-descriptions-item>
        <el-descriptions-item label="状态">
          <el-tag :type="statusTagType[employee.status]" size="small">{{ statusMap[employee.status] }}</el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="薪资" v-if="employee.salary">¥{{ employee.salary }}</el-descriptions-item>
        <el-descriptions-item label="试用期薪资" v-if="employee.probation_salary">¥{{ employee.probation_salary }}</el-descriptions-item>
        <el-descriptions-item label="工资卡" :span="2" v-if="employee.bank_card">{{ employee.bank_card }}</el-descriptions-item>
        <el-descriptions-item label="紧急联系人" v-if="employee.emergency_contact">{{ employee.emergency_contact }}</el-descriptions-item>
        <el-descriptions-item label="紧急联系电话" v-if="employee.emergency_phone">{{ employee.emergency_phone }}</el-descriptions-item>
      </el-descriptions>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { employeeApi } from '@/api/employee'
import { ElMessage } from 'element-plus'
import { statusMap, statusTagType } from './statusMap'

const route = useRoute()
const loading = ref(false)
const employee = ref<any>(null)

async function load() {
  loading.value = true
  try {
    employee.value = await employeeApi.get(Number(route.params.id))
  } catch {
    ElMessage.error('加载失败')
  } finally {
    loading.value = false
  }
}

onMounted(() => load())
</script>

<style scoped lang="scss">
.employee-detail {
  padding-bottom: 70px;
}
.header {
  display: flex;
  align-items: center;
  justify-content: space-between;
}
</style>
