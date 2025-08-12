import type { 
  ListSubmissionsResponse, 
  GetSubmissionByIdResponse, 
  SubmitCodeRequest, 
  SubmitCodeResponse, 
  SubmissionFilterRequest,
  SubmissionDetailsUi
} from '../types/api';
import { apiClient } from './api-client';

export const submissionsService = {
  list: async (filters: SubmissionFilterRequest): Promise<ListSubmissionsResponse> => {
    const params = new URLSearchParams();
    params.append('page', filters.page.toString());
    
    if (filters.onlyMe !== undefined) {
      params.append('onlyMe', filters.onlyMe.toString());
    }
    
    if (filters.problemId !== undefined) {
      params.append('problemId', filters.problemId.toString());
    }
    
    if (filters.statuses?.length) {
      filters.statuses.forEach(status => params.append('status', status));
    }
    
    if (filters.languages?.length) {
      filters.languages.forEach(language => params.append('language', language));
    }
    
    const response = await apiClient.get<ListSubmissionsResponse>(`/api/v1/submissions/list?${params}`);
    return response.data;
  },
  
  getById: async (id: number): Promise<GetSubmissionByIdResponse> => {
    const response = await apiClient.get<GetSubmissionByIdResponse>(`/api/v1/submissions/${id}`);
    return response.data;
  },
  
  submit: async (data: SubmitCodeRequest): Promise<SubmitCodeResponse> => {
    const response = await apiClient.post<{submissionDetails: SubmissionDetailsUi}>('/api/v1/submissions', data);
    return { submission: response.data.submissionDetails };
  }
};