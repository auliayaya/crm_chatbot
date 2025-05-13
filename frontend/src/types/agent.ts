export interface Agent {
  id: string;
  first_name: string;
  last_name: string;
  email: string;
  department?: string; // As per API spec
  role: 'admin' | 'agent' | 'viewer'; // Keep if your system uses this, API spec doesn't explicitly list for GET
  status: 'active' | 'inactive';
  lastLogin?: string; // Optional, from your mock
  createdAt?: string; // Standard timestamp
  updatedAt?: string; // Standard timestamp
}

export interface AgentListResponse {
  agents: Agent[];
  total: number;
  page: number;
  pageSize: number;
  totalPages: number;
  // Add other pagination fields if your API returns them
}

import { z } from 'zod';

export const agentStatusSchema = z.enum(['active', 'inactive']);
export const agentRoleSchema = z.enum(['admin', 'agent', 'viewer']);

export const agentFormSchema = z.object({
  first_name: z.string().min(1, 'First name is required'),
  last_name: z.string().min(1, 'Last name is required'),
  email: z.string().email('Invalid email address').min(1, 'Email is required'),
  department: z.string().optional(),
  status: agentStatusSchema,
  role: agentRoleSchema, // Assuming role is managed and settable
  // Password only for creation, and typically not part of the 'Agent' type fetched from GET
  password: z.string().min(6, 'Password must be at least 6 characters').optional(),
  confirmPassword: z.string().optional(),
}).refine(data => {
    // If password is provided, confirmPassword must match
    if (data.password && data.password !== data.confirmPassword) {
        return false;
    }
    return true;
}, {
    message: "Passwords don't match",
    path: ['confirmPassword'], // path of error
});

export type AgentFormData = z.infer<typeof agentFormSchema>;

// Adjust CreateAgentData and UpdateAgentData if needed based on the form
export type CreateAgentData = Omit<AgentFormData, 'confirmPassword'>; // Password is included
export type UpdateAgentData = Omit<AgentFormData, 'password' | 'confirmPassword'>; // Typically password is not updated here or has a separate flow