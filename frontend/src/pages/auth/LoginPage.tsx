// src/pages/auth/LoginPage.tsx
import { useState } from 'react'
import { useMutation } from '@tanstack/react-query'
import { z } from 'zod'
import { Link, useNavigate } from 'react-router-dom'
import { useForm, type SubmitHandler } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { useAuthStore } from '../../store/authStore'
import AuthLayout from '../../layouts/AuthLayout'
import Button from '../../components/common/Button'
import { authService } from '../../services/authService'

// Modified schema to allow username or email
const schema = z.object({
  emailOrUsername: z.string().min(1, 'Email or username is required'),
  password: z.string().min(1, 'Password is required'),
})

type FormValues = z.infer<typeof schema>

export default function LoginPage() {
  const navigate = useNavigate()
  const { setAuth } = useAuthStore()
  const [formError, setFormError] = useState<string | null>(null)
  const [rememberMe, setRememberMe] = useState(false)

  const {
    register,
    handleSubmit,
    formState: { errors, isSubmitting },
  } = useForm<FormValues>({
    resolver: zodResolver(schema),
    defaultValues: {
      emailOrUsername: '',
      password: '',
    },
  })

  const loginMutation = useMutation({
    mutationFn: (credentials: FormValues) =>
      authService.login(credentials.emailOrUsername, credentials.password),
    onSuccess: (data) => {
      setAuth(data.user, data.token)
      navigate('/dashboard')
    },
    onError: (error: { response?: { data?: { message?: string } } }) => {
      setFormError(error.response?.data?.message || 'Invalid email or password')
    },
  })

  const onSubmit: SubmitHandler<FormValues> = (data) => {
    loginMutation.mutate(data)
  }

  return (
    <AuthLayout>
      <h2 className="mt-6 text-center text-3xl font-extrabold text-gray-900">
        Sign in to your account
      </h2>

      <div className="mt-8">
        <form onSubmit={handleSubmit(onSubmit)} className="space-y-6">
          {formError && (
            <div className="bg-red-50 text-red-700 p-3 rounded-md text-sm">
              {formError}
            </div>
          )}

          {/* Email or Username field */}
          <div>
            <label
              htmlFor="emailOrUsername"
              className="block text-sm font-medium text-gray-700"
            >
              Email or Username
            </label>
            <div className="mt-1">
              <input
                id="emailOrUsername"
                type="text"
                className="appearance-none block w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm placeholder-gray-400 focus:outline-none focus:ring-primary-500 focus:border-primary-500 sm:text-sm"
                placeholder="Email address or username"
                {...register('emailOrUsername')}
              />
              {errors.emailOrUsername && (
                <p className="mt-1 text-sm text-red-600">
                  {errors.emailOrUsername.message}
                </p>
              )}
            </div>
          </div>

          {/* Password field */}
          <div>
            <label
              htmlFor="password"
              className="block text-sm font-medium text-gray-700"
            >
              Password
            </label>
            <div className="mt-1">
              <input
                id="password"
                type="password"
                className="appearance-none block w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm placeholder-gray-400 focus:outline-none focus:ring-primary-500 focus:border-primary-500 sm:text-sm"
                {...register('password')}
              />
              {errors.password && (
                <p className="mt-1 text-sm text-red-600">
                  {errors.password.message}
                </p>
              )}
            </div>
          </div>

          {/* Remember me & Forgot password */}
          <div className="flex items-center justify-between">
            <div className="flex items-center">
              <input
                id="remember-me"
                name="remember-me"
                type="checkbox"
                className="h-4 w-4 text-primary-600 focus:ring-primary-500 border-gray-300 rounded"
                checked={rememberMe}
                onChange={(e) => setRememberMe(e.target.checked)}
              />
              <label
                htmlFor="remember-me"
                className="ml-2 block text-sm text-gray-900"
              >
                Remember me
              </label>
            </div>

            <div className="text-sm">
              <Link
                to="/forgot-password"
                className="font-medium text-primary-600 hover:text-primary-500"
              >
                Forgot your password?
              </Link>
            </div>
          </div>

          <div>
            <Button
              type="submit"
              className="w-full"
              disabled={isSubmitting || loginMutation.isPending}
            >
              {loginMutation.isPending ? 'Signing in...' : 'Sign in'}
            </Button>
          </div>

          <div className="text-center">
            <p className="text-sm text-gray-600">
              Don't have an account?{' '}
              <Link
                to="/register"
                className="font-medium text-primary-600 hover:text-primary-500"
              >
                Create an account
              </Link>
            </p>
          </div>
        </form>
      </div>
    </AuthLayout>
  )
}
