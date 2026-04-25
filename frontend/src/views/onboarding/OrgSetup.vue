<template>
  <div class="org-setup">
    <el-card>
      <template #header>
        <span>企业信息录入</span>
      </template>

      <el-form ref="formRef" :model="form" :rules="rules" label-width="140px">
        <el-form-item label="企业名称" prop="name">
          <el-input v-model="form.name" placeholder="请输入企业全称" maxlength="100" />
        </el-form-item>
        <el-form-item label="统一社会信用代码" prop="credit_code">
          <el-input v-model="form.credit_code" placeholder="18位统一社会信用代码" maxlength="18" />
        </el-form-item>
        <el-form-item label="所在城市" prop="city_id">
          <div class="city-select">
            <el-select
              v-model="form.city_id"
              placeholder="请选择城市"
              filterable
              style="width: 200px"
              @change="onCityChange"
            >
              <el-option
                v-for="city in cityList"
                :key="city.code"
                :label="city.name"
                :value="city.code"
              />
            </el-select>
            <el-button size="small" @click="detectCity" :loading="detecting">
              {{ cityAutoDetected ? '已定位' : '自动定位' }}
            </el-button>
            <el-tag v-if="cityAutoDetected" type="success" size="small">已定位: {{ detectedCityName }}</el-tag>
          </div>
        </el-form-item>
        <el-form-item label="联系人" prop="contact_name">
          <el-input v-model="form.contact_name" placeholder="请输入联系人姓名" maxlength="50" />
        </el-form-item>
        <el-form-item label="联系电话" prop="contact_phone">
          <el-input v-model="form.contact_phone" placeholder="请输入手机号" maxlength="11" />
        </el-form-item>

        <el-form-item>
          <el-button type="primary" @click="handleSubmit" :loading="saving">保存并继续</el-button>
          <el-button @click="handleSkip">跳过</el-button>
        </el-form-item>
      </el-form>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import request from '@/api/request'

const router = useRouter()
const formRef = ref()
const saving = ref(false)
const detecting = ref(false)
const cityAutoDetected = ref(false)
const detectedCityName = ref('')
const cityList = ref<{ code: number; name: string }[]>([])

const form = reactive({
  name: '',
  credit_code: '',
  city_id: undefined as number | undefined,
  contact_name: '',
  contact_phone: '',
})

const rules = {
  name: [{ required: true, message: '请输入企业名称', trigger: 'blur' }],
  credit_code: [
    { required: true, message: '请输入统一社会信用代码', trigger: 'blur' },
    { pattern: /^[1-9][0-9A-HJ-NPQRTUWXY]{17}$/, message: '统一社会信用代码格式不正确', trigger: 'blur' },
  ],
  city_id: [{ required: true, message: '请选择城市', trigger: 'change' }],
  contact_name: [{ required: true, message: '请输入联系人', trigger: 'blur' }],
  contact_phone: [
    { required: true, message: '请输入联系电话', trigger: 'blur' },
    { pattern: /^1[3-9]\d{9}$/, message: '手机号格式不正确', trigger: 'blur' },
  ],
}

async function loadCities() {
  try {
    const res = await request.get('/cities')
    cityList.value = res.data
  } catch {
    // ignore
  }
}

async function detectCity() {
  detecting.value = true
  try {
    const res = await fetch('https://ipapi.co/json/', { signal: AbortSignal.timeout(5000) })
    const data = await res.json()
    if (data.city && data.country_code === 'CN') {
      const matched = cityList.value.find(
        (c) => c.name.includes(data.city) || data.city.includes(c.name),
      )
      if (matched) {
        form.city_id = matched.code
        cityAutoDetected.value = true
        detectedCityName.value = matched.name
      } else {
        ElMessage.warning(`未匹配到城市"${data.city}"，请手动选择`)
      }
    } else {
      ElMessage.warning('无法自动定位，请手动选择城市')
    }
  } catch {
    ElMessage.warning('定位失败，请手动选择城市')
  } finally {
    detecting.value = false
  }
}

function onCityChange() {
  cityAutoDetected.value = false
  detectedCityName.value = ''
}

async function handleSubmit() {
  if (saving.value) return
  if (!formRef.value) return
  try {
    await formRef.value.validate()
  } catch {
    return
  }
  saving.value = true
  try {
    const res = await request.post('/org/onboarding', {
      name: form.name,
      credit_code: form.credit_code,
      city: cityList.value.find(c => c.code === form.city_id)?.name || '',
      contact_name: form.contact_name,
      contact_phone: form.contact_phone,
    })
    // 保存新 token（onboarding 完成后 org_id 已更新到 token 中）
    localStorage.setItem('token', res.data.access_token)
    localStorage.setItem('refresh_token', res.data.refresh_token)
    ElMessage.success('保存成功')
    router.push('/home')
  } catch (err: any) {
    ElMessage.error(err.response?.data?.message || '保存失败')
  } finally {
    saving.value = false
  }
}

function handleSkip() {
  router.push('/home')
}

onMounted(() => {
  loadCities()
})
</script>

<style scoped lang="scss">
.org-setup {
  padding: 16px;
}
.city-select {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-wrap: wrap;
}
</style>
