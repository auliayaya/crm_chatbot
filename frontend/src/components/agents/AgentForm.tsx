import React, { useEffect } from 'react';
import { useForm, Controller } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { type Agent, type AgentFormData, agentFormSchema, agentRoleSchema, agentStatusSchema } from '../../types/agent'; // Adjust path as needed
import Input from '../common/Input';
import Select from '../common/Select'; // Assuming you have a Select component
import Button from '../common/Button';
import FormError from '../common/FormError'; // Assuming you have a FormError component

interface AgentFormProps {
  agent?: Agent | null; // For pre-filling form in edit mode
  onSubmit: (data: AgentFormData) => void;
  onCancel: () => void;
  isSubmitting?: boolean;
}

const AgentForm: React.FC<AgentFormProps> = ({ agent, onSubmit, onCancel, isSubmitting }) => {
  const isEditMode = !!agent;

  const {
    control,
    handleSubmit,
    reset,
    formState: { errors },
    watch
  } = useForm<AgentFormData>({
    resolver: zodResolver(agentFormSchema),
    defaultValues: {
      first_name: agent?.first_name || '',
      last_name: agent?.last_name || '',
      email: agent?.email || '',
      department: agent?.department || '',
      status: agent?.status || 'active',
      role: agent?.role || 'agent',
      password: '',
      confirmPassword: '',
    },
  });

  useEffect(() => {
    if (agent) {
      reset({
        first_name: agent.first_name,
        last_name: agent.last_name,
        email: agent.email,
        department: agent.department || '',
        status: agent.status,
        role: agent.role,
        password: '', // Password fields are typically not pre-filled in edit mode
        confirmPassword: '',
      });
    } else {
      reset({ // Default for new agent
        first_name: '',
        last_name: '',
        email: '',
        department: '',
        status: 'active',
        role: 'agent',
        password: '',
        confirmPassword: '',
      });
    }
  }, [agent, reset]);

//   const password = watch('password');

  return (
    <form onSubmit={handleSubmit(onSubmit)} className="space-y-4">
      <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
        <Controller
          name="first_name"
          control={control}
          render={({ field }) => (
            <Input label="First Name" {...field} error={errors.first_name?.message} />
          )}
        />
        <Controller
          name="last_name"
          control={control}
          render={({ field }) => (
            <Input label="Last Name" {...field} error={errors.last_name?.message} />
          )}
        />
      </div>
      <Controller
        name="email"
        control={control}
        render={({ field }) => <Input label="Email" type="email" {...field} error={errors.email?.message} />}
      />
      <Controller
        name="department"
        control={control}
        render={({ field }) => <Input label="Department (Optional)" {...field} error={errors.department?.message} />}
      />
      <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
        <Controller
          name="status"
          control={control}
          render={({ field }) => (
            <Select label="Status" {...field} error={errors.status?.message}>
              {agentStatusSchema.options.map(s => <option key={s} value={s}>{s.charAt(0).toUpperCase() + s.slice(1)}</option>)}
            </Select>
          )}
        />
        <Controller
          name="role"
          control={control}
          render={({ field }) => (
            <Select label="Role" {...field} error={errors.role?.message}>
              {agentRoleSchema.options.map(r => <option key={r} value={r}>{r.charAt(0).toUpperCase() + r.slice(1)}</option>)}
            </Select>
          )}
        />
      </div>
      {!isEditMode && (
        <>
          <Controller
            name="password"
            control={control}
            render={({ field }) => (
              <Input label="Password" type="password" {...field} error={errors.password?.message} autoComplete="new-password" />
            )}
          />
          <Controller
            name="confirmPassword"
            control={control}
            render={({ field }) => (
              <Input label="Confirm Password" type="password" {...field} error={errors.confirmPassword?.message} autoComplete="new-password" />
            )}
          />
        </>
      )}
       {errors.confirmPassword && errors.confirmPassword.message === "Passwords don't match" && (
         <FormError message={errors.confirmPassword.message} />
       )}


      <div className="flex justify-end space-x-3 pt-4">
        <Button type="button" variant="outline" onClick={onCancel} disabled={isSubmitting}>
          Cancel
        </Button>
        <Button type="submit" variant="default" disabled={isSubmitting}>
          {isSubmitting ? (isEditMode ? 'Saving...' : 'Creating...') : (isEditMode ? 'Save Changes' : 'Create Agent')}
        </Button>
      </div>
    </form>
  );
};

export default AgentForm;