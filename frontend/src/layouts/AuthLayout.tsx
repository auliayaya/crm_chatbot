
import type { ReactNode } from 'react'
import { Link } from 'react-router-dom'

interface AuthLayoutProps {
  children: ReactNode
}

export default function AuthLayout({ children }: AuthLayoutProps) {
  return (
    <div className="min-h-screen bg-gray-50 flex flex-col justify-center py-12 sm:px-6 lg:px-8">
      <div className="sm:mx-auto sm:w-full sm:max-w-md">
        <div className="flex justify-center">
          <Link to="/">
            <img
              className="h-12 w-auto"
              src="/logo.svg"
              alt="CRM Chatbot"
            />
          </Link>
        </div>
        <h1 className="mt-6 text-center text-3xl font-extrabold text-gray-900">
          CRM Chatbot
        </h1>
      </div>

      <div className="mt-8 sm:mx-auto sm:w-full sm:max-w-md">
        <div className="bg-white py-8 px-4 shadow sm:rounded-lg sm:px-10">
          {children}
        </div>
      </div>
      
      <div className="mt-8 text-center text-sm text-gray-600">
        <p>
          &copy; {new Date().getFullYear()} CRM Chatbot. All rights reserved.
        </p>
      </div>
    </div>
  )
}