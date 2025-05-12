import { useState } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import { useQuery, useMutation } from '@tanstack/react-query'
import { customerService } from '../../services/customerService'
import { formatDate, formatPhoneNumber } from '../../utils/formatters'
import DashboardLayout from '../../layouts/DashboardLayout'
import Card from '../../components/common/Card'
import Button from '../../components/common/Button'
import CustomerForm from '../../components/customers/CustomerForm'

export default function CustomerDetailPage() {
  const { id } = useParams<{ id: string }>()
  const navigate = useNavigate()
  const [isEditing, setIsEditing] = useState(false)

  const {
    data: customer,
    isLoading,
    refetch,
  } = useQuery({
    queryKey: ['customer', id],
    queryFn: () => customerService.getCustomerById(id!),
    enabled: !!id,
  })

  const deleteMutation = useMutation({
    mutationFn: () => customerService.deleteCustomer(id!),
    onSuccess: () => {
      navigate('/customers')
    },
  })

  const handleDelete = () => {
    if (window.confirm('Are you sure you want to delete this customer?')) {
      deleteMutation.mutate()
    }
  }

  const handleEditSuccess = () => {
    setIsEditing(false)
    refetch()
  }

  if (isLoading) {
    return (
      <DashboardLayout>
        <div className="flex items-center justify-center h-64">
          <p className="text-gray-500">Loading customer details...</p>
        </div>
      </DashboardLayout>
    )
  }

  if (!customer) {
    return (
      <DashboardLayout>
        <div className="flex flex-col items-center justify-center h-64">
          <p className="text-gray-500 mb-4">Customer not found</p>
          <Button onClick={() => navigate('/customers')}>
            Back to Customers
          </Button>
        </div>
      </DashboardLayout>
    )
  }

  return (
    <DashboardLayout>
      <div className="flex justify-between items-center mb-6">
        <h1 className="text-2xl font-semibold text-gray-900">
          Customer Details
        </h1>
        <div className="space-x-2">
          <Button variant="outline" onClick={() => navigate('/customers')}>
            Back to Customers
          </Button>
          {!isEditing && (
            <Button onClick={() => setIsEditing(true)}>Edit Customer</Button>
          )}
        </div>
      </div>

      {isEditing ? (
        <Card>
          <CustomerForm
            customerId={customer.id}
            initialValues={customer}
            onSuccess={handleEditSuccess}
            onCancel={() => setIsEditing(false)}
          />
        </Card>
      ) : (
        <>
          <div className="grid grid-cols-1 gap-6 lg:grid-cols-2">
            <Card>
              <div className="px-4 py-5 sm:p-6">
                <h3 className="text-lg font-medium leading-6 text-gray-900 mb-4">
                  Contact Information
                </h3>
                <div className="grid grid-cols-1 gap-x-4 gap-y-6 sm:grid-cols-2">
                  <div className="sm:col-span-1">
                    <dt className="text-sm font-medium text-gray-500">Name</dt>
                    <dd className="mt-1 text-sm text-gray-900">
                      {customer.firstName} {customer.lastName}
                    </dd>
                  </div>
                  <div className="sm:col-span-1">
                    <dt className="text-sm font-medium text-gray-500">
                      Company
                    </dt>
                    <dd className="mt-1 text-sm text-gray-900">
                      {customer.companyName || '—'}
                    </dd>
                  </div>
                  <div className="sm:col-span-1">
                    <dt className="text-sm font-medium text-gray-500">Email</dt>
                    <dd className="mt-1 text-sm text-gray-900">
                      <a
                        href={`mailto:${customer.email}`}
                        className="text-primary-600 hover:text-primary-500"
                      >
                        {customer.email}
                      </a>
                    </dd>
                  </div>
                  <div className="sm:col-span-1">
                    <dt className="text-sm font-medium text-gray-500">Phone</dt>
                    <dd className="mt-1 text-sm text-gray-900">
                      {formatPhoneNumber(customer.phoneNumber)}
                    </dd>
                  </div>
                  <div className="sm:col-span-1">
                    <dt className="text-sm font-medium text-gray-500">
                      Status
                    </dt>
                    <dd className="mt-1 text-sm text-gray-900">
                      <span
                        className={`px-2 py-1 text-xs rounded-full ${
                          customer.status === 'active'
                            ? 'bg-green-100 text-green-800'
                            : 'bg-red-100 text-red-800'
                        }`}
                      >
                        {customer.status === 'active' ? 'Active' : 'Inactive'}
                      </span>
                    </dd>
                  </div>
                  <div className="sm:col-span-1">
                    <dt className="text-sm font-medium text-gray-500">
                      Created
                    </dt>
                    <dd className="mt-1 text-sm text-gray-900">
                      {formatDate(customer.createdAt)}
                    </dd>
                  </div>
                </div>
              </div>
            </Card>

            <Card title="Recent Activity">
              <div className="px-4 py-5 sm:p-6">
                <p className="text-sm text-gray-500">
                  No recent activity found for this customer.
                </p>
              </div>
            </Card>
          </div>

          {/* Related tickets and other customer information can be added here */}
          <div className="mt-6">
            <Card title="Recent Tickets">
              <div className="px-4 py-5 sm:p-6">
                <p className="text-sm text-gray-500">
                  No recent tickets found for this customer.
                </p>
              </div>
            </Card>
          </div>

          <div className="mt-6">
            <Card>
              <div className="px-4 py-5 sm:p-6">
                <h3 className="text-lg font-medium leading-6 text-gray-900 mb-4">
                  Danger Zone
                </h3>
                <div className="mt-2 max-w-xl text-sm text-gray-500">
                  <p>
                    Once you delete this customer, there is no going back.
                    Please be certain.
                  </p>
                </div>
                <div className="mt-5">
                  <Button
                    variant="danger"
                    onClick={handleDelete}
                    disabled={deleteMutation.isPending}
                  >
                    {deleteMutation.isPending
                      ? 'Deleting...'
                      : 'Delete Customer'}
                  </Button>
                </div>
              </div>
            </Card>
          </div>
        </>
      )}
    </DashboardLayout>
  )
}
