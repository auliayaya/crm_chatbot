// src/store/authStore.ts
import { create } from 'zustand'
import { persist } from 'zustand/middleware'
import { jwtDecode } from 'jwt-decode' // Import jwt-decode
import type { AuthState, User } from '../types/auth' // Ensure this path is correct

// Define the structure of the login response
interface LoginResponse {
  token: string
}

// Define the expected structure of your JWT payload after decoding
interface DecodedToken {
  sub: string // Subject (usually user ID)
  username: string
  role: string
  exp: number // Expiration time
  // Add other claims you expect, e.g., iat (issued at)
  [key: string]: any // Allow other claims
}

interface AuthActions {
  setAuth: (loginResponse: LoginResponse) => void
  logout: () => void
  updateUser: (userData: Partial<User>) => void
}

export const useAuthStore = create<AuthState & AuthActions>()(
  persist(
    (set) => ({
      user: null,
      token: null,
      isAuthenticated: false,
      setAuth: (loginResponse) => {
        try {
          const decodedToken = jwtDecode<DecodedToken>(loginResponse.token)

          // Construct the user object from the decoded token claims
          const userData: User = {
            id: decodedToken.username,
            username: decodedToken.username,
            role: 'user',
            firstName: '',
            lastName: '',
            email: '',
            // Map other claims to your User interface as needed
            // name: decodedToken.name, (if 'name' claim exists)
            // email: decodedToken.email, (if 'email' claim exists)
          }

          set({
            user: userData,
            token: loginResponse.token,
            isAuthenticated: true,
          })
        } catch (error) {
          console.error('Failed to decode token or set auth state:', error)
          // Optionally clear auth state if token is invalid
          set({ user: null, token: null, isAuthenticated: false })
        }
      },
      logout: () => set({ user: null, token: null, isAuthenticated: false }),
      updateUser: (userData) =>
        set((state) => ({
          user: state.user ? { ...state.user, ...userData } : null,
        })),
    }),
    {
      name: 'crm-auth-storage', // This is the localStorage key
    }
  )
)
