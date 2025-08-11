import type { 
  ListProblemsResponse, 
  GetProblemByIdResponse, 
  ProblemFilterRequest 
} from '../types/api';
import { apiClient } from './api-client';

export const problemsService = {
  list: async (filters: ProblemFilterRequest): Promise<ListProblemsResponse> => {
    const params = new URLSearchParams();
    params.append('page', filters.page.toString());
    
    if (filters.difficulties?.length) {
      filters.difficulties.forEach(difficulty => params.append('difficulty', difficulty));
    }
    
    if (filters.tags?.length) {
      filters.tags.forEach(tag => params.append('tag', tag));
    }
    
    if (filters.solvedStatuses?.length) {
      filters.solvedStatuses.forEach(status => params.append('solvedStatus', status));
    }
    
    const response = await apiClient.get<ListProblemsResponse>(`/api/v1/problems/list?${params}`);
    return response.data;
  },
  
  getById: async (id: number): Promise<GetProblemByIdResponse> => {
    const response = await apiClient.get<{problemDetails: any}>(`/api/v1/problems/${id}`);
    return { problem: response.data.problemDetails };
  }
};