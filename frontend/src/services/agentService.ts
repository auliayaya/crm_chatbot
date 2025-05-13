import { api } from './api'; // Assuming you have a central api instance like in customerService
import type { Agent, AgentListResponse, CreateAgentData, UpdateAgentData } from "../types/agent"; // We'll define these types next

// Define params for fetching agents if you need pagination/filtering later
interface GetAgentsParams {
  search?: string;
  page?: number;
  pageSize?: number;
  status?: 'active' | 'inactive';
  role?: 'admin' | 'agent' | 'viewer';
}

export const agentService = {
  async getAgents(params?: GetAgentsParams): Promise<AgentListResponse> {
    // Adjust the endpoint if your API returns a paginated structure
    // For now, assuming it returns an array of agents directly or an object with an 'agents' key
    const response = await api.get('/api/crm/agents', { params });
    // If API returns { agents: [], total: ... }, adapt like customerService
    // If API returns Agent[] directly:
    // return { agents: response.data, total: response.data.length, page: 1, pageSize: response.data.length };
    // For now, let's assume it's like customerService and returns an object:
    if (Array.isArray(response.data)) { // Handle if API directly returns an array
        return { agents: response.data, total: response.data.length, page:1, pageSize: response.data.length, totalPages: 1 };
    }
    return response.data; // Assuming response.data is AgentListResponse
  },

  async getAgentById(id: string): Promise<Agent> {
    const response = await api.get(`/api/crm/agents/${id}`);
    return response.data;
  },

  async createAgent(data: CreateAgentData): Promise<Agent> {
    const response = await api.post('/api/crm/agents', data);
    return response.data;
  },

  async updateAgent(id: string, data: UpdateAgentData): Promise<Agent> {
    const response = await api.put(`/api/crm/agents/${id}`, data);
    return response.data;
  },

  async deleteAgent(id: string): Promise<void> {
    await api.delete(`/api/crm/agents/${id}`);
  },
};