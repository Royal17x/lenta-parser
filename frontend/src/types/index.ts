export interface Product {
    id: number
    name: string
    slug: string
    storeId: number
    count: number
    prices: { price: number; priceRegular: number; isPromoactionPrice: boolean }
    display: { name: string; package: string }
    rating: { rate: number; votes: number }
}

export interface SSEEvent {
    type: 'progress' | 'error' | 'done'
    category?: string
    message?: string
    count?: number
}

export interface LogEntry {
    id: number
    type: SSEEvent['type']
    text: string
    time: string
}