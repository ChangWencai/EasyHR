import request from './request'

export interface Position {
  id: number
  name: string
  department_id: number | null
  sort_order: number
}

export interface PositionSelectOptions {
  dept_positions: Array<{ id: number; name: string }>
  common_positions: Array<{ id: number; name: string }>
  unassigned_option: { id: null; name: string }
}

export const positionApi = {
  list: (department_id?: number) =>
    request.get<Position[]>('/positions', department_id !== undefined ? { params: { department_id } } : {}),

  getSelectOptions: (department_id?: number) =>
    request.get<PositionSelectOptions>('/positions/select-options', {
      params: department_id !== undefined ? { department_id } : {},
    }),

  create: (data: { name: string; department_id?: number | null; sort_order?: number }) =>
    request.post<Position>('/positions', data),

  update: (id: number, data: Partial<Pick<Position, 'name' | 'department_id' | 'sort_order'>>) =>
    request.put<Position>(`/positions/${id}`, data),

  delete: (id: number) => request.delete(`/positions/${id}`),
}
