// src/components/common/Card.tsx
import type { ReactNode } from 'react'

interface CardProps {
  title?: string
  children: ReactNode
  className?: string
  footer?: ReactNode
}

export default function Card({
  title,
  children,
  className = '',
  footer,
}: CardProps) {
  return (
    <div className={`bg-white shadow rounded-lg ${className}`}>
      {title && (
        <div className="px-6 py-4 border-b border-gray-200">
          <h3 className="text-lg font-medium text-gray-900">{title}</h3>
        </div>
      )}
      <div className="px-6 py-4">{children}</div>
      {footer && (
        <div className="px-6 py-4 border-t border-gray-200">{footer}</div>
      )}
    </div>
  )
}
