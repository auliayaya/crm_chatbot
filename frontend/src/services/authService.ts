// src/services/authService.ts
import { api } from './api'

export const authService = {
  async login(email: string, password: string) {
    const response = await api.post('/auth/login', { username:email, password })
    return response.data
  },

  async register(userData: {
    email: string
    username: string
    password: string
    firstName: string
    lastName: string
  }) {
    const response = await api.post('/auth/register', userData)
    return response.data
  },

  async forgotPassword(email: string) {
    const response = await api.post('/auth/forgot-password', { email })
    return response.data
  },

  async resetPassword(token: string, password: string) {
    const response = await api.post('/auth/reset-password', { token, password })
    return response.data
  },
}
