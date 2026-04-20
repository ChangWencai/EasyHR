import request from './request'
import type { AxiosRequestConfig } from 'axios'

// === Types ===

export type ContractStatus =
  | 'draft'        // 草稿
  | 'pending_sign' // 待签署
  | 'signed'       // 已签
  | 'active'       // 生效中
  | 'terminated'   // 已终止
  | 'expired'      // 已过期

export type ContractType = 'fixed_term' | 'indefinite' | 'intern'

export interface Contract {
  id: number
  employee_id: number
  employee_name: string
  contract_type: ContractType
  type_label: string         // e.g. "劳动合同（固定期限）"
  start_date: string         // YYYY-MM-DD
  end_date: string | null   // null for indefinite
  status: ContractStatus
  status_label: string
  expiry_days: number | null // null for indefinite/no expiry
  created_at: string
  signed_at: string | null
  pdf_url: string | null
  signed_pdf_url: string | null
  // salary 字段用于传给后端，但不显示在 PDF 正文中（D-11-02）
  salary?: number
}

export interface ContractListResponse {
  list: Contract[]
  total: number
}

export interface CreateContractData {
  employee_id: number
  contract_type: ContractType
  start_date: string
  end_date: string | null
  // D-11-02: salary stored in DB but NOT in PDF body
  salary?: number
}

export interface GeneratePdfResponse {
  pdf_url: string
}

export interface SendSignLinkResponse {
  message: string
  expires_at: string
}

export interface SendSignCodeData {
  contract_id: number
  phone: string
}

export interface SignVerifyCodeData {
  contract_id: number
  phone: string
  code: string
}

export interface SignVerifyCodeResponse {
  sign_token: string   // short-lived token for Step 3 confirm
  expires_in: number
  employee_name: string
  contract_type: string
  start_date: string
  end_date?: string
  org_name: string
}

export interface ConfirmSignData {
  contract_id: number
  sign_token: string
}

export interface ConfirmSignResponse {
  signed_pdf_url: string
  message: string
}

export interface GetSignedPdfResponse {
  url: string
}

// === API Methods ===

export const contractApi = {
  // 合同列表
  list: (employeeId: number) =>
    (request.get<ContractListResponse>(`/employees/${employeeId}/contracts`) as Promise<{ data: ContractListResponse }>)
      .then(r => r.data),

  // 创建合同（draft）
  // BLOCKER-5 fix: D-11-02 means salary not in PDF body, not absent from data model.
  // Send actual salary to satisfy backend binding:"required,gt=0".
  create: (data: CreateContractData, employeeSalary?: number) =>
    (request.post<Contract>(`/employees/${data.employee_id}/contracts`, {
      contract_type: data.contract_type,
      start_date: data.start_date,
      end_date: data.end_date,
      salary: employeeSalary ?? data.salary ?? 0, // send actual salary to satisfy backend validation
    }) as Promise<{ data: Contract }>)
      .then(r => r.data),

  // 发起签署（生成PDF + 上传OSS + 发短信）
  sendSignLink: (contractId: number) =>
    (request.post(`/contracts/${contractId}/send-sign-link`) as Promise<{ data: SendSignLinkResponse }>)
      .then(r => r.data),

  // 员工H5签署 - 发送验证码
  sendSignCode: (data: SendSignCodeData) =>
    request.post('/contracts/sign/send-code', data),

  // 员工H5签署 - 校验验证码
  verifySignCode: (data: SignVerifyCodeData) =>
    (request.post<SignVerifyCodeResponse>('/contracts/sign/verify-code', data) as Promise<{ data: SignVerifyCodeResponse }>)
      .then(r => r.data),

  // 员工H5签署 - 确认签署
  confirmSign: (data: ConfirmSignData) =>
    (request.post<ConfirmSignResponse>('/contracts/sign/confirm', data) as Promise<{ data: ConfirmSignResponse }>)
      .then(r => r.data),

  // 获取已签PDF URL
  getSignedPdf: (contractId: number) =>
    (request.get<GetSignedPdfResponse>(`/contracts/${contractId}/signed-pdf`) as Promise<{ data: GetSignedPdfResponse }>)
      .then(r => r.data),

  // 终止合同
  terminate: (contractId: number, reason: string, terminateDate: string) =>
    request.put(`/contracts/${contractId}/terminate`, {
      terminate_reason: reason,
      terminate_date: terminateDate,
    }),

  // 生成PDF预览（返回PDF文件）
  generatePdfBlob: (contractId: number): Promise<Blob> =>
    (request.get(`/contracts/${contractId}/generate-pdf`, {
      responseType: 'blob',
    } as AxiosRequestConfig) as Promise<Blob>),
}
