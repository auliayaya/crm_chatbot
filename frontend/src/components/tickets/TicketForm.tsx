import { useState } from 'react'
import { useMutation, useQuery } from '@tanstack/react-query'
import { useForm } from '@tanstack/react-form'
import { z } from 'zod'
import { ticketService } from '../../services/ticketService'
import { customerService } from '../../services/customerService'
import Button from '../common/Button'

const ticketSchema = z.object({
  subject: z.string().min(1, 'Subject is required'),
  description: z.string().min(1, 'Description is required'),
  customerId: z.string().min(1, 'Customer is required'),
  priority: z.enum(['low', 'medium', 'high']),
  status: z.enum(['new', 'open', 'in_progress', 'resolved', 'closed']),
})

type TicketFormValues = z.infer<typeof ticketSchema>

interface TicketFormProps {
  ticketId?: string
  initialValues?: Partial<TicketFormValues>
  onSuccess: () => void
  onCancel: () => void
}

export default function TicketForm({
  ticketId,
  initialValues,
  onSuccess,
  onCancel,
}: TicketFormProps) {
  const isEditing = Boolean(ticketId)
  const [formError, setFormError] = useState<string | null>(null)
  
  const { data: customers } = useQuery({
    queryKey: ['customers', 'all'],
    queryFn: () => customerService.getCustomers({ pageSize: 100 }),
  })
  
  const mutation = useMutation({
    mutationFn: (data: TicketFormValues) => {
      if (isEditing && ticketId) {
        return ticketService.updateTicket(ticketId, data)
      } else {
        return ticketService.createTicket(data)
      }
    },
    onSuccess: () => {
      onSuccess()
    },
    onError: (error: any) => {
      setFormError(error.response?.data?.message || 'An error occurred while saving')
    }
  })
  
  const form = useForm({
    defaultValues: {
      subject: initialValues?.subject || '',
      description: initialValues?.description || '',
      customerId: initialValues?.customerId || '',
      priority: initialValues?.priority || 'medium',
      status: initialValues?.status || 'new',
    } as TicketFormValues,
    onSubmit: async ({ value }) => {
      mutation.mutate(value)
    },
    validators: {
      onChange: ticketSchema
    }
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
      
      <div className="space-y-4">
        <div>
          <label htmlFor="subject" className="block text-sm font-medium text-gray-700">
            Subject
          </label>
          <div className="mt-1">
            <input
              type="text"
              id="subject"
              className="shadow-sm focus:ring-primary-500 focus:border-primary-500 block w-full sm:text-sm border-gray-300 rounded-md"
              value={form.getFieldValue('subject')}
              onChange={(e) => form.setFieldValue('subject', e.target.value)}
              onBlur={() => form.validateField('subject', 'blur')}
            />
            {form.getFieldMeta('subject')?.errors && (
              <p className="mt-1 text-sm text-red-600">
                {form.getFieldMeta('subject')?.errors}
              </p>
            )}
          </div>
        </div>

        <div>
          <label htmlFor="customerId" className="block text-sm font-medium text-gray-700">
            Customer
          </label>
          <div className="mt-1">
            <select
              id="customerId"
              className="shadow-sm focus:ring-primary-500 focus:border-primary-500 block w-full sm:text-sm border-gray-300 rounded-md"
              value={form.getFieldValue('customerId')}
              onChange={(e) => form.setFieldValue('customerId', e.target.value)}
              onBlur={() => form.validateField('customerId', 'blur' )}
            >
              <option value="">Select a customer</option>
              {customers?.customers.map(customer => (
                <option key={customer.id} value={customer.id}>
                  {customer.firstName} {customer.lastName} - {customer.email}
                </option>
              ))}
            </select>
            {form.getFieldMeta('customerId')?.errors && (
              <p className="mt-1 text-sm text-red-600">
                {form.getFieldMeta('customerId')?.errors}
              </p>
            )}
          </div>
        </div>

        <div>
          <label htmlFor="description" className="block text-sm font-medium text-gray-700">
            Description
          </label>
          <div className="mt-1">
            <textarea
              id="description"
              rows={4}
              className="shadow-sm focus:ring-primary-500 focus:border-primary-500 block w-full sm:text-sm border-gray-300 rounded-md"
              value={form.getFieldValue('description')}
              onChange={(e) => form.setFieldValue('description', e.target.value)}
              onBlur={() => form.validateField('description',  'blur' )}
            ></textarea>
            {form.getFieldMeta('description')?.errors && (
              <p className="mt-1 text-sm text-red-600">
                {form.getFieldMeta('description')?.errors}
              </p>
            )}
          </div>
        </div>

        <div className="grid grid-cols-1 gap-y-4 sm:grid-cols-2 sm:gap-x-4">
          <div>
            <label htmlFor="priority" className="block text-sm font-medium text-gray-700">
              Priority
            </label>
            <div className="mt-1">
              <select
                id="priority"
                className="shadow-sm focus:ring-primary-500 focus:border-primary-500 block w-full sm:text-sm border-gray-300 rounded-md"
                value={form.getFieldValue('priority')}
                onChange={(e) => form.setFieldValue('priority', e.target.value as 'low' | 'medium' | 'high')}
              >
                <option value="low">Low</option>
                <option value="medium">Medium</option>
                <option value="high">High</option>
              </select>
            </div>
          </div>

          <div>
            <label htmlFor="status" className="block text-sm font-medium text-gray-700">
              Status
            </label>
            <div className="mt-1">
              <select
                id="status"
                className="shadow-sm focus:ring-primary-500 focus:border-primary-500 block w-full sm:text-sm border-gray-300 rounded-md"
                value={form.getFieldValue('status')}
                onChange={(e) => form.setFieldValue('status', e.target.value as 'new' | 'open' | 'in_progress' | 'resolved' | 'closed')}
              >
                <option value="new">New</option>
                <option value="open">Open</option>
                <option value="in_progress">In Progress</option>
                <option value="resolved">Resolved</option>
                <option value="closed">Closed</option>
              </select>
            </div>
          </div>
        </div>
      </div>

      <div className="mt-5 sm:mt-6 sm:grid sm:grid-cols-2 sm:gap-3 sm:grid-flow-row-dense">
        <Button
          type="submit"
          disabled={mutation.isPending}
          className="sm:col-start-2"
        >
          {mutation.isPending ? 'Saving...' : isEditing ? 'Update Ticket' : 'Create Ticket'}
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