"use client";

import React, { createContext, useContext, useState, useEffect, useCallback } from "react";
import { apiClient } from "./api/client";

interface User {
  id: string;
  email: string;
  name: string;
  role: string;
}

interface AuthState {
  user: User | null;
  isLoading: boolean;
  login: (email: string, password: string) => Promise<void>;
  register: (email: string, password: string, name: string) => Promise<void>;
  logout: () => Promise<void>;
}

const AuthContext = createContext<AuthState | null>(null);

export function AuthProvider({ children }: { children: React.ReactNode }) {
  const [user, setUser] = useState<User | null>(null);
  const [isLoading, setIsLoading] = useState(true);

  // Check for existing session on mount
  useEffect(() => {
    if (apiClient.isLoggedIn()) {
      apiClient
        .getProfile()
        .then((profile) => {
          setUser({ id: profile.id, email: profile.email, name: profile.name, role: profile.role });
        })
        .catch(() => {
          // Token expired or invalid
        })
        .finally(() => setIsLoading(false));
    } else {
      setIsLoading(false);
    }
  }, []);

  const login = useCallback(async (email: string, password: string) => {
    const res = await apiClient.login(email, password);
    setUser({ id: res.user.id, email: res.user.email, name: res.user.name, role: res.user.role });
  }, []);

  const register = useCallback(async (email: string, password: string, name: string) => {
    const res = await apiClient.register(email, password, name);
    setUser({ id: res.user.id, email: res.user.email, name: res.user.name, role: res.user.role });
  }, []);

  const logout = useCallback(async () => {
    await apiClient.logout();
    setUser(null);
  }, []);

  return (
    <AuthContext.Provider value={{ user, isLoading, login, register, logout }}>
      {children}
    </AuthContext.Provider>
  );
}

export function useAuth() {
  const ctx = useContext(AuthContext);
  if (!ctx) throw new Error("useAuth must be used within AuthProvider");
  return ctx;
}
