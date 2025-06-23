import { useState, useCallback } from 'react';

export interface AppError {
  message: string;
  code?: string;
  details?: any;
  timestamp: Date;
}

export const useError = () => {
  const [error, setError] = useState<AppError | null>(null);

  const showError = useCallback((message: string, code?: string, details?: any) => {
    const newError: AppError = {
      message,
      code,
      details,
      timestamp: new Date(),
    };
    setError(newError);
    
    // Log error
    console.error('Application Error:', newError);
  }, []);

  const clearError = useCallback(() => {
    setError(null);
  }, []);

  const handleError = useCallback((err: unknown) => {
    if (err instanceof Error) {
      showError(err.message, 'UNKNOWN_ERROR', err.stack);
    } else if (typeof err === 'string') {
      showError(err);
    } else {
      showError('An unexpected error occurred', 'UNKNOWN_ERROR', err);
    }
  }, [showError]);

  const handleApiError = useCallback(async (response: Response) => {
    try {
      const errorData = await response.json();
      showError(
        errorData.message || `HTTP ${response.status}: ${response.statusText}`,
        errorData.code || `HTTP_${response.status}`,
        errorData
      );
    } catch {
      showError(
        `HTTP ${response.status}: ${response.statusText}`,
        `HTTP_${response.status}`
      );
    }
  }, [showError]);

  return {
    error,
    showError,
    clearError,
    handleError,
    handleApiError,
  };
};