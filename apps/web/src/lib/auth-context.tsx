'use client';

import { createContext, useContext, useState, useCallback, type ReactNode } from 'react';

interface AuthContextValue {
  apiKey: string | null;
  isAuthenticated: boolean;
  login: (apiKey: string) => void;
  logout: () => void;
  getStoredWalletIds: () => string[];
  addStoredWalletId: (id: string) => void;
}

const AuthContext = createContext<AuthContextValue | null>(null);

const WALLET_IDS_KEY = 'fluxa_wallet_ids';

export function AuthProvider({ children }: { children: ReactNode }) {
  const [apiKey, setApiKey] = useState<string | null>(() => {
    if (typeof window === 'undefined') return null;
    return localStorage.getItem('fluxa_api_key');
  });

  const login = useCallback((key: string) => {
    localStorage.setItem('fluxa_api_key', key);
    setApiKey(key);
  }, []);

  const logout = useCallback(() => {
    localStorage.removeItem('fluxa_api_key');
    localStorage.removeItem(WALLET_IDS_KEY);
    setApiKey(null);
  }, []);

  const getStoredWalletIds = useCallback((): string[] => {
    try {
      const raw = localStorage.getItem(WALLET_IDS_KEY);
      return raw ? JSON.parse(raw) : [];
    } catch {
      return [];
    }
  }, []);

  const addStoredWalletId = useCallback(
    (id: string) => {
      const existing = getStoredWalletIds();
      if (!existing.includes(id)) {
        localStorage.setItem(WALLET_IDS_KEY, JSON.stringify([...existing, id]));
      }
    },
    [getStoredWalletIds],
  );

  return (
    <AuthContext.Provider
      value={{
        apiKey,
        isAuthenticated: !!apiKey,
        login,
        logout,
        getStoredWalletIds,
        addStoredWalletId,
      }}
    >
      {children}
    </AuthContext.Provider>
  );
}

export function useAuth() {
  const ctx = useContext(AuthContext);
  if (!ctx) throw new Error('useAuth must be used within AuthProvider');
  return ctx;
}
