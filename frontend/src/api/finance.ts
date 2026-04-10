import request from './request'

export interface Voucher {
  id: number
  period: string
  voucher_no: string
  status: string
  total_debit: string
  total_credit: string
  creator_name: string
  created_at: string
}

export interface Invoice {
  id: number
  invoice_no: string
  type: string
  amount: string
  tax_amount: string
  status: string
  created_at: string
}

export interface VoucherEntry {
  account_id: number
  account_name?: string
  account_code?: string
  dc: 'debit' | 'credit'
  amount: string
  summary: string
}

export const financeApi = {
  // Accounts
  accountTree: () => request.get('/accounts'),
  createAccount: (data: { name: string; code: string; category: string; parent_id?: number }) =>
    request.post('/accounts', data),

  // Vouchers
  vouchers: (params: { page: number; period_id?: number; account_id?: number; keyword?: string }) =>
    request.get('/vouchers', { params }),
  createVoucher: (data: {
    period_id: number
    entries: { account_id: number; dc: string; amount: string; summary: string }[]
  }) => request.post('/vouchers', data),
  submitVoucher: (id: number) => request.post('/vouchers/submit', { id }),
  auditVoucher: (id: number) => request.post('/vouchers/audit', { id }),
  reverseVoucher: (id: number) => request.post('/vouchers/reverse', { id }),

  // Invoices
  invoices: (params?: { page: number; type?: string }) =>
    request.get('/invoices', { params }),
  createInvoice: (data: {
    invoice_no: string
    type: string
    amount: string
    tax_rate?: string
    tax_amount?: string
    invoice_date?: string
  }) => request.post('/invoices', data),

  // Expenses
  expenses: (params?: { page: number; status?: string }) =>
    request.get('/expenses', { params }),
  approveExpense: (id: number) => request.post(`/expenses/${id}/approve`, {}),
  rejectExpense: (id: number, reason: string) => request.post(`/expenses/${id}/reject`, { reason }),

  // Books
  trialBalance: (periodId: number) =>
    request.get('/books/trial-balance', { params: { period_id: periodId } }),
  ledger: (periodId: number, accountId?: number) =>
    request.get('/books/ledger', { params: { period_id: periodId, account_id: accountId } }),
  bookExport: (periodId: number) =>
    request.get('/books/export', { params: { period_id: periodId }, responseType: 'blob' }),

  // Reports
  balanceSheet: (periodId: number) =>
    request.get('/reports/balance-sheet', { params: { period_id: periodId } }),
  incomeStatement: (periodId: number) =>
    request.get('/reports/income-statement', { params: { period_id: periodId } }),
  taxDeclaration: (year: number, month: number) =>
    request.get('/reports/tax-declaration', { params: { year, month } }),
  taxExport: (year: number, month: number) =>
    request.get('/reports/tax-declaration/export', { params: { year, month }, responseType: 'blob' }),

  // Periods
  periods: () => request.get('/periods'),
  closePeriod: (id: number) => request.post(`/periods/${id}/close`, {}),
  revertPeriod: (id: number) => request.post(`/periods/${id}/revert`, { confirm: true }),
}
