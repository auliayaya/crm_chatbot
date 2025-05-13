import { Routes, Route, Navigate } from 'react-router-dom'
import { useAuthStore } from './store/authStore'
import LoginPage from './pages/auth/LoginPage'
// import ForgotPasswordPage from './pages/auth/ForgotPasswordPage'
// import ResetPasswordPage from './pages/auth/ResetPasswordPage'
import DashboardPage from './pages/dashboard/DashboardPage'

import TicketsPage from './pages/tickets/TicketsPage'
// import TicketDetailPage from './pages/tickets/TicketDetailPage'
// import ChatPage from './pages/chat/ChatPage'
// import SettingsPage from './pages/settings/SettingsPage'
// import NotFoundPage from './pages/NotFoundPage'
import type { JSX } from 'react'
import RegisterPage from './pages/auth/RegisterPage'
import CustomerDetailPage from './pages/customers/CustomerDetailPage'
import CustomerPage from './pages/customers/CustomerPage'
import ChatPage from './pages/chat/ChatPage'
import AgentsPage from './pages/agents/AgentsPage'

interface ProtectedRouteProps {
  children: JSX.Element
}

function ProtectedRoute({ children }: ProtectedRouteProps) {
  const { isAuthenticated } = useAuthStore()

  if (!isAuthenticated) {
    return <Navigate to="/login" replace />
  }

  return children
}

export default function App() {
  return (
    <Routes>
      {/* Public routes */}
      <Route path="/login" element={<LoginPage />} />
      <Route path="/register" element={<RegisterPage />} />
      {/* <Route path="/forgot-password" element={<ForgotPasswordPage />} />
      <Route path="/reset-password" element={<ResetPasswordPage />} /> */}

      {/* Protected routes */}
      <Route
        path="/"
        element={
          <ProtectedRoute>
            <Navigate to="/dashboard" replace />
          </ProtectedRoute>
        }
      />
      <Route
        path="/dashboard"
        element={
          <ProtectedRoute>
            <DashboardPage />
          </ProtectedRoute>
        }
      />
      <Route
        path="/customers"
        element={
          <ProtectedRoute>
            <CustomerPage />
          </ProtectedRoute>
        }
      />
      <Route
        path="/customers/:id"
        element={
          <ProtectedRoute>
            <CustomerDetailPage />
          </ProtectedRoute>
        }
      />
      <Route
        path="/tickets"
        element={
          <ProtectedRoute>
            <TicketsPage />
          </ProtectedRoute>
        }
      />
       {/* <Route 
        path="/tickets/:id"
        element={
          <ProtectedRoute>
            <TicketDetailPage />
          </ProtectedRoute>
        }
      /> */}
      <Route
        path="/agents"
        element={
          <ProtectedRoute>
            <AgentsPage />
          </ProtectedRoute>
        }
      /> 
      <Route
        path="/chat/:customerId"
        element={
          <ProtectedRoute>
            <ChatPage />
          </ProtectedRoute>
        }
      />
      {/* <Route
        path="/settings"
        element={
          <ProtectedRoute>
            <SettingsPage />
          </ProtectedRoute>
        }
      /> */}

      {/* 404 page */}
      {/* <Route path="*" element={<NotFoundPage />} /> *} */}
    </Routes>
  )
}
