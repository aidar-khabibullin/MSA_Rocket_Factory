const priceFormat = new Intl.NumberFormat('ru-RU', { minimumFractionDigits: 2, maximumFractionDigits: 2 })
const timeFormat = new Intl.DateTimeFormat('ru-RU', { hour: '2-digit', minute: '2-digit', second: '2-digit' })

export const formatPrice = (value: number): string => priceFormat.format(value)
export const formatTime = (date: Date): string => timeFormat.format(date)
