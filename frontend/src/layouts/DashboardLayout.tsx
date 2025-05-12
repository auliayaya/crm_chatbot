import type { ReactNode } from 'react'

import Sidebar from '../components/common/Sidebar'
import Header from '../components/common/Header'
import Footer from '../components/common/Footer'

interface DashboardLayoutProps {
  children: ReactNode
}

export default function DashboardLayout({ children }: DashboardLayoutProps) {
  return (
    <div className="h-screen flex overflow-hidden bg-gray-100">
      {/* Sidebar */}
      <Sidebar />

      {/* Main content */}
      <div className="flex flex-col w-0 flex-1 overflow-hidden">
        <Header />

        <main className="flex-1 relative overflow-y-auto focus:outline-none">
          <div className="py-6">
            <div className="max-w-7xl mx-auto px-4 sm:px-6 md:px-8">
              {children}
            </div>
          </div>
        </main>

        <Footer />
      </div>
    </div>
  )
}
