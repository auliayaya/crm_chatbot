// src/store/authStore.ts
import { create } from 'zustand'
import { persist } from 'zustand/middleware'
import type { AuthState, User } from '../types/auth'


interface AuthActions {
  setAuth: (user: User, token: string) => void
  logout: () => void
  updateUser: (userData: Partial<User>) => void
}

export const useAuthStore = create<AuthState & AuthActions>()(
  persist(
    (set) => ({
      user: null,
      token: null,
      isAuthenticated: false,
      setAuth: (user, token) => set({ user, token, isAuthenticated: true }),
      logout: () => set({ user: null, token: null, isAuthenticated: false }),
      updateUser: (userData) =>
        set((state) => ({
          user: state.user ? { ...state.user, ...userData } : null,
        })),
    }),
    {
      name: 'crm-auth-storage',
    }
  )
)

