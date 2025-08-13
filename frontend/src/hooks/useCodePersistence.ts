import { useState, useEffect, useCallback } from 'react';
import { SubmissionLanguage } from '../types/api';
import { useAuth } from './useAuth';

const BOILERPLATE_CODE = {
  [SubmissionLanguage.CPP]: `#include <iostream>
#include <vector>
#include <string>
#include <algorithm>
using namespace std;

int main() {
    // Your code here
    return 0;
}`,
  [SubmissionLanguage.JAVA]: `import java.util.*;
import java.io.*;

public class Main {
    public static void main(String[] args) {
        Scanner sc = new Scanner(System.in);
        // Your code here
        sc.close();
    }
}`,
  [SubmissionLanguage.PYTHON]: `# Your code here
def solve():
    pass

if __name__ == "__main__":
    solve()`,
  [SubmissionLanguage.JAVASCRIPT]: `const readline = require('readline');

const rl = readline.createInterface({
    input: process.stdin,
    output: process.stdout
});

// Your code here

rl.close();`
};

const CODE_STORAGE_KEY_PREFIX = 'problem_code_';
const LANGUAGE_STORAGE_KEY_PREFIX = 'problem_language_';

const getStorageKey = (problemId: number, language: SubmissionLanguage, userId?: string) => {
  const userSuffix = userId ? `_user_${userId}` : '_anonymous';
  return `${CODE_STORAGE_KEY_PREFIX}${problemId}_${language}${userSuffix}`;
};

const getLanguageStorageKey = (problemId: number, userId?: string) => {
  const userSuffix = userId ? `_user_${userId}` : '_anonymous';
  return `${LANGUAGE_STORAGE_KEY_PREFIX}${problemId}${userSuffix}`;
};

const saveCodeToStorage = (problemId: number, language: SubmissionLanguage, code: string, userId?: string) => {
  const key = getStorageKey(problemId, language, userId);
  try {
    localStorage.setItem(key, code);
  } catch (error) {
    console.warn('Failed to save code to localStorage:', error);
  }
};

const loadCodeFromStorage = (problemId: number, language: SubmissionLanguage, userId?: string): string | null => {
  const key = getStorageKey(problemId, language, userId);
  try {
    return localStorage.getItem(key);
  } catch (error) {
    console.warn('Failed to load code from localStorage:', error);
    return null;
  }
};

const saveLanguageToStorage = (problemId: number, language: SubmissionLanguage, userId?: string) => {
  const key = getLanguageStorageKey(problemId, userId);
  try {
    localStorage.setItem(key, language);
  } catch (error) {
    console.warn('Failed to save language to localStorage:', error);
  }
};

const loadLanguageFromStorage = (problemId: number, userId?: string): SubmissionLanguage | null => {
  const key = getLanguageStorageKey(problemId, userId);
  try {
    const stored = localStorage.getItem(key);
    if (stored && Object.values(SubmissionLanguage).includes(stored as SubmissionLanguage)) {
      return stored as SubmissionLanguage;
    }
  } catch (error) {
    console.warn('Failed to load language from localStorage:', error);
  }
  return null;
};

const migrateAnonymousCodeToUser = (problemId: number, language: SubmissionLanguage, userId: string) => {
  const anonymousCode = loadCodeFromStorage(problemId, language);
  const userCode = loadCodeFromStorage(problemId, language, userId);
  
  if (anonymousCode && !userCode) {
    saveCodeToStorage(problemId, language, anonymousCode, userId);
  }
  
  try {
    const anonymousKey = getStorageKey(problemId, language);
    localStorage.removeItem(anonymousKey);
  } catch (error) {
    console.warn('Failed to remove anonymous code from localStorage:', error);
  }
};

const clearUserCode = (problemId: number, userId: string) => {
  try {
    Object.values(SubmissionLanguage).forEach(language => {
      const key = getStorageKey(problemId, language, userId);
      localStorage.removeItem(key);
    });
    const languageKey = getLanguageStorageKey(problemId, userId);
    localStorage.removeItem(languageKey);
  } catch (error) {
    console.warn('Failed to clear user code from localStorage:', error);
  }
};


export interface UseCodePersistenceResult {
  selectedLanguage: SubmissionLanguage;
  code: string;
  setSelectedLanguage: (language: SubmissionLanguage) => void;
  updateCode: (code: string) => void;
  resetCode: () => void;
  resetAllCode: () => void;
}

export const useCodePersistence = (problemId: number): UseCodePersistenceResult => {
  const { user } = useAuth();
  const userId = user?.handle;
  const [previousUserId, setPreviousUserId] = useState<string | undefined>(userId);

  const [selectedLanguage, setSelectedLanguageState] = useState<SubmissionLanguage>(() => {
    return loadLanguageFromStorage(problemId, userId) || SubmissionLanguage.CPP;
  });

  const [code, setCodeState] = useState<string>(() => {
    const stored = loadCodeFromStorage(problemId, selectedLanguage, userId);
    return stored || BOILERPLATE_CODE[selectedLanguage];
  });

  useEffect(() => {
    if (userId !== previousUserId) {
      if (userId && !previousUserId) {
        // Anonymous user logged in - migrate their code if no user code exists
        Object.values(SubmissionLanguage).forEach(lang => {
          migrateAnonymousCodeToUser(problemId, lang, userId);
        });
        
        const anonymousLanguage = loadLanguageFromStorage(problemId);
        if (anonymousLanguage) {
          try {
            const anonymousLanguageKey = getLanguageStorageKey(problemId);
            localStorage.removeItem(anonymousLanguageKey);
          } catch (error) {
            console.warn('Failed to remove anonymous language from localStorage:', error);
          }
        }
      } else if (!userId && previousUserId) {
        clearUserCode(problemId, previousUserId);
      }
      
      setPreviousUserId(userId);
      
      const newLanguage = loadLanguageFromStorage(problemId, userId) || SubmissionLanguage.CPP;
      setSelectedLanguageState(newLanguage);
      
      const newCode = loadCodeFromStorage(problemId, newLanguage, userId) || BOILERPLATE_CODE[newLanguage];
      setCodeState(newCode);
    }
  }, [userId, previousUserId, problemId]);

  useEffect(() => {
    const stored = loadCodeFromStorage(problemId, selectedLanguage, userId);
    if (stored) {
      setCodeState(stored);
    } else {
      setCodeState(BOILERPLATE_CODE[selectedLanguage]);
    }
  }, [problemId, selectedLanguage, userId]);

  const setSelectedLanguage = useCallback((language: SubmissionLanguage) => {
    saveLanguageToStorage(problemId, language, userId);
    setSelectedLanguageState(language);
  }, [problemId, userId]);

  const updateCode = useCallback((newCode: string) => {
    setCodeState(newCode);
    saveCodeToStorage(problemId, selectedLanguage, newCode, userId);
  }, [problemId, selectedLanguage, userId]);

  const resetCode = useCallback(() => {
    const boilerplate = BOILERPLATE_CODE[selectedLanguage];
    setCodeState(boilerplate);
    saveCodeToStorage(problemId, selectedLanguage, boilerplate, userId);
  }, [problemId, selectedLanguage, userId]);

  const resetAllCode = useCallback(() => {
    try {
      Object.values(SubmissionLanguage).forEach(language => {
        const key = getStorageKey(problemId, language, userId);
        localStorage.removeItem(key);
      });
      const languageKey = getLanguageStorageKey(problemId, userId);
      localStorage.removeItem(languageKey);
      
      setSelectedLanguageState(SubmissionLanguage.CPP);
      setCodeState(BOILERPLATE_CODE[SubmissionLanguage.CPP]);
    } catch (error) {
      console.warn('Failed to reset code from localStorage:', error);
    }
  }, [problemId, userId]);

  return {
    selectedLanguage,
    code,
    setSelectedLanguage,
    updateCode,
    resetCode,
    resetAllCode
  };
};