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
