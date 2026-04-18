import request from '@/api/request'

export interface CarouselItem {
  id: number
  image_url: string
  link_url: string
  sort_order: number
  active: boolean
  start_at: string
  end_at: string
}

export function fetchCarousels(): Promise<{ data: CarouselItem[] }> {
  return request.get('/carousels').then((res) => res.data)
}

// Admin management functions

export interface CarouselFormData {
  image_url: string
  link_url?: string
  sort_order?: number
  active?: boolean
  start_at?: string
  end_at?: string
}

export function listAllCarousels(): Promise<{ data: CarouselItem[] }> {
  return request.get('/carousels/admin').then((res) => res.data)
}

export function createCarousel(data: CarouselFormData): Promise<{ data: CarouselItem }> {
  return request.post('/carousels', data).then((res) => res.data)
}

export function updateCarousel(id: number, data: Partial<CarouselFormData>): Promise<void> {
  return request.put(`/carousels/${id}`, data).then((res) => res.data)
}

export function deleteCarousel(id: number): Promise<void> {
  return request.delete(`/carousels/${id}`).then((res) => res.data)
}

export function uploadImage(file: File): Promise<string> {
  const formData = new FormData()
  formData.append('file', file)
  return request.post('/upload/image', formData, {
    headers: { 'Content-Type': 'multipart/form-data' },
  }).then((res) => res.data.url as string)
}
