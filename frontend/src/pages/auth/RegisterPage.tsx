import { useState } from 'react'
import { useMutation } from '@tanstack/react-query'
import { z } from 'zod'
import { Link, useNavigate } from 'react-router-dom'
import { useForm, type SubmitHandler } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { useAuthStore } from '../../store/authStore'
import AuthLayout from '../../layouts/AuthLayout.tsx'
import Button from '../../components/common/Button'
import { authService } from '../../services/authService'

const schema = z
  .object({
    firstName: z.string().min(1, 'First name is required'),
    lastName: z.string().min(1, 'Last name is required'),
    email: z.string().email('Invalid email address'),
    username: z.string().min(3, 'Username must be at least 3 characters'),
    password: z.string().min(6, 'Password must be at least 6 characters'),
    confirmPassword: z.string().min(6, 'Confirm password is required'),
  })
  .refine((data) => data.password === data.confirmPassword, {
    message: "Passwords don't match",
    path: ['confirmPassword'],
  })

type FormValues = z.infer<typeof schema>

export default function RegisterPage() {
  const navigate = useNavigate()
  const { setAuth } = useAuthStore()
  const [formError, setFormError] = useState<string | null>(null)

  // React Hook Form setup with Zod resolver
  const {
    register,
    handleSubmit,
    formState: { errors, isSubmitting },
  } = useForm<FormValues>({
    resolver: zodResolver(schema),
    defaultValues: {
      firstName: '',
      lastName: '',
      email: '',
      username: '',
      password: '',
      confirmPassword: '',
    },
  })

  const registerMutation = useMutation({
    mutationFn: (userData: Omit<FormValues, 'confirmPassword'>) =>
      authService.register(userData),
    onSuccess: (data) => {
      setAuth(data.user, data.token)
      navigate('/dashboard')
    },
    onError: (error: { response?: { data?: { message?: string } } }) => {
      setFormError(error.response?.data?.message || 'Registration failed')
    },
  })

  // Form submission handler
  const onSubmit: SubmitHandler<FormValues> = (data) => {
    const {  ...userData } = data
    registerMutation.mutate(userData)
  }

  return (
    <AuthLayout>
      <h2 className="mt-6 text-center text-3xl font-extrabold text-gray-900">
        Create an account
      </h2>

      <div className="mt-8">
        <form onSubmit={handleSubmit(onSubmit)} className="space-y-6">
          {formError && (
            <div className="bg-red-50 text-red-700 p-3 rounded-md text-sm">
              {formError}
            </div>
          )}

          {/* Map through fields */}
          {(
            [
              'firstName',
              'lastName',
              'email',
              'username',
              'password',
              'confirmPassword',
            ] as const
          ).map((field) => (
            <div key={field}>
              <label
                htmlFor={field}
                className="block text-sm font-medium text-gray-700"
              >
                {field === 'confirmPassword'
                  ? 'Confirm Password'
                  : field.charAt(0).toUpperCase() + field.slice(1)}
              </label>
              <div className="mt-1">
                <input
                  id={field}
                  type={
                    field.toLowerCase().includes('password')
                      ? 'password'
                      : field === 'email'
                      ? 'email'
                      : 'text'
                  }
                  className="appearance-none block w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm placeholder-gray-400 focus:outline-none focus:ring-primary-500 focus:border-primary-500 sm:text-sm"
                  {...register(field)}
                />
                {errors[field] && (
                  <p className="mt-1 text-sm text-red-600">
                    {errors[field]?.message}
                  </p>
                )}
              </div>
            </div>
          ))}

          {/* Submit button */}
          <Button
            type="submit"
            className="w-full"
            disabled={isSubmitting || registerMutation.isPending}
          >
            {registerMutation.isPending
              ? 'Creating account...'
              : 'Create account'}
          </Button>

          <div className="text-center">
            <p className="text-sm text-gray-600">
              Already have an account?{' '}
              <Link
                to="/login"
                className="font-medium text-primary-600 hover:text-primary-500"
              >
                Sign in
              </Link>
            </p>
          </div>
        </form>
      </div>
    </AuthLayout>
  )
}
