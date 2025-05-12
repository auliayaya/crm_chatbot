import { format, parseISO } from 'date-fns'

export function formatDate(dateString: string, formatStr = 'MMM d, yyyy'): string {
  if (!dateString) return ''
  try {
    return format(parseISO(dateString), formatStr)
  } catch (error) {
    console.error('Invalid date format:', dateString, error
    )
    return dateString
  }
}

export function formatCurrency(amount: number, currency = 'USD'): string {
  return new Intl.NumberFormat('en-US', { 
    style: 'currency', 
    currency 
  }).format(amount)
}

export function formatPhoneNumber(phoneNumber: string): string {
  if (!phoneNumber) return ''
  
  // Remove all non-numeric characters
  const cleaned = phoneNumber.replace(/\D/g, '')
  
  // Format as (XXX) XXX-XXXX for US phone numbers
  if (cleaned.length === 10) {
    return `(${cleaned.slice(0, 3)}) ${cleaned.slice(3, 6)}-${cleaned.slice(6)}`
  }
  
  // Return original if not a standard 10-digit number
  return phoneNumber
}

export function truncateText(text: string, maxLength: number): string {
  if (!text || text.length <= maxLength) return text
  return `${text.slice(0, maxLength)}...`
}