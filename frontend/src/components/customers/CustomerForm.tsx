// src/components/customers/CustomerForm.tsx
import { useMutation } from '@tanstack/react-query'
import { useForm } from '@tanstack/react-form'
import { z } from 'zod'

import Button from '../common/Button'
import { useState } from 'react'
import { customerService } from '../../services/customerService'

const customerSchema = z.object({
  firstName: z.string().min(1, 'First name is required'),
  lastName: z.string().min(1, 'Last name is required'),
  email: z.string().email('Invalid email address'),
  phoneNumber: z.string().min(1, 'Phone number is required'),
  companyName: z.string().optional(),
  status: z.enum(['active', 'inactive']),
})

type CustomerFormValues = z.infer<typeof customerSchema>

interface CustomerFormProps {
  customerId?: string
  initialValues?: Partial<CustomerFormValues>
  onSuccess: () => void
  onCancel: () => void
}

export default function CustomerForm({
  customerId,
  initialValues,
  onSuccess,
  onCancel,
}: CustomerFormProps) {
  const isEditing = Boolean(customerId)
  const [formError, setFormError] = useState<string | null>(null)

  const mutation = useMutation({
    mutationFn: (data: CustomerFormValues) => {
      if (isEditing && customerId) {
        return customerService.updateCustomer(customerId, data)
      } else {
        return customerService.createCustomer(data)
      }
    },
    onSuccess: () => {
      onSuccess()
    },
    onError: (error: any) => {
      setFormError(
        error.response?.data?.message || 'An error occurred while saving'
      )
    },
  })

  const form = useForm({
    defaultValues: {
      firstName: initialValues?.firstName || '',
      lastName: initialValues?.lastName || '',
      email: initialValues?.email || '',
      phoneNumber: initialValues?.phoneNumber || '',
      companyName: initialValues?.companyName || '',
      status: initialValues?.status || 'active',
    } as CustomerFormValues,
    onSubmit: async ({ value }) => {
      mutation.mutate(value)
    },
    validators: {
      onChange: customerSchema,
    },
  })

  return (
    <form
      onSubmit={(e) => {
        e.preventDefault()
        form.handleSubmit()
      }}
      className="space-y-4 mt-4"
    >
      {formError && (
        <div className="bg-red-50 text-red-700 p-3 rounded-md text-sm">
          {formError}
        </div>
      )}

      <div className="grid grid-cols-1 gap-y-6 gap-x-4 sm:grid-cols-6">
        <div className="sm:col-span-3">
          <label
            htmlFor="firstName"
            className="block text-sm font-medium text-gray-700"
          >
            First name
          </label>
          <div className="mt-1">
            <input
              type="text"
              id="firstName"
              className="shadow-sm focus:ring-primary-500 focus:border-primary-500 block w-full sm:text-sm border-gray-300 rounded-md"
              value={form.getFieldValue('firstName')}
              onChange={(e) => form.setFieldValue('firstName', e.target.value)}
              onBlur={() => form.validateField('firstName','blur')}
            />
            {form.getFieldMeta('firstName')?.errors && (
              <p className="mt-1 text-sm text-red-600">
                {form.getFieldMeta('firstName')?.errors}
              </p>
            )}
          </div>
        </div>

        <div className="sm:col-span-3">
          <label
            htmlFor="lastName"
            className="block text-sm font-medium text-gray-700"
          >
            Last name
          </label>
          <div className="mt-1">
            <input
              type="text"
              id="lastName"
              className="shadow-sm focus:ring-primary-500 focus:border-primary-500 block w-full sm:text-sm border-gray-300 rounded-md"
              value={form.getFieldValue('lastName')}
              onChange={(e) => form.setFieldValue('lastName', e.target.value)}
              onBlur={() => form.validateField('lastName', 'blur' )}
            />
            {form.getFieldMeta('lastName')?.errors && (
              <p className="mt-1 text-sm text-red-600">
                {form.getFieldMeta('lastName')?.errors}
              </p>
            )}
          </div>
        </div>

        <div className="sm:col-span-6">
          <label
            htmlFor="email"
            className="block text-sm font-medium text-gray-700"
          >
            Email address
          </label>
          <div className="mt-1">
            <input
              type="email"
              id="email"
              className="shadow-sm focus:ring-primary-500 focus:border-primary-500 block w-full sm:text-sm border-gray-300 rounded-md"
              value={form.getFieldValue('email')}
              onChange={(e) => form.setFieldValue('email', e.target.value)}
              onBlur={() => form.validateField('email', 'blur' )}
            />
            {form.getFieldMeta('email')?.errors && (
              <p className="mt-1 text-sm text-red-600">
                {form.getFieldMeta('email')?.errors}
              </p>
            )}
          </div>
        </div>

        <div className="sm:col-span-3">
          <label
            htmlFor="phoneNumber"
            className="block text-sm font-medium text-gray-700"
          >
            Phone number
          </label>
          <div className="mt-1">
            <input
              type="text"
              id="phoneNumber"
              className="shadow-sm focus:ring-primary-500 focus:border-primary-500 block w-full sm:text-sm border-gray-300 rounded-md"
              value={form.getFieldValue('phoneNumber')}
              onChange={(e) =>
                form.setFieldValue('phoneNumber', e.target.value)
              }
              onBlur={() => form.validateField('phoneNumber', 'blur' )}
            />
            {form.getFieldMeta('phoneNumber')?.errors && (
              <p className="mt-1 text-sm text-red-600">
                {form.getFieldMeta('phoneNumber')?.errors}
              </p>
            )}
          </div>
        </div>

        <div className="sm:col-span-3">
          <label
            htmlFor="companyName"
            className="block text-sm font-medium text-gray-700"
          >
            Company name
          </label>
          <div className="mt-1">
            <input
              type="text"
              id="companyName"
              className="shadow-sm focus:ring-primary-500 focus:border-primary-500 block w-full sm:text-sm border-gray-300 rounded-md"
              value={form.getFieldValue('companyName')}
              onChange={(e) =>
                form.setFieldValue('companyName', e.target.value)
              }
            />
          </div>
        </div>

        <div className="sm:col-span-3">
          <label
            htmlFor="status"
            className="block text-sm font-medium text-gray-700"
          >
            Status
          </label>
          <div className="mt-1">
            <select
              id="status"
              className="shadow-sm focus:ring-primary-500 focus:border-primary-500 block w-full sm:text-sm border-gray-300 rounded-md"
              value={form.getFieldValue('status')}
              onChange={(e) =>
                form.setFieldValue(
                  'status',
                  e.target.value as 'active' | 'inactive'
                )
              }
            >
              <option value="active">Active</option>
              <option value="inactive">Inactive</option>
            </select>
          </div>
        </div>
      </div>

      <div className="mt-5 sm:mt-6 sm:grid sm:grid-cols-2 sm:gap-3 sm:grid-flow-row-dense">
        <Button
          type="submit"
          disabled={mutation.isPending}
          className="sm:col-start-2"
        >
          {mutation.isPending
            ? 'Saving...'
            : isEditing
            ? 'Update Customer'
            : 'Add Customer'}
        </Button>
        <Button
          type="button"
          variant="outline"
          onClick={onCancel}
          className="mt-3 sm:mt-0"
        >
          Cancel
        </Button>
      </div>
    </form>
  )
}
