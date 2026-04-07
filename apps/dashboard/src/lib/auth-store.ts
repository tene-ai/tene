import { create } from "zustand";
import { persist } from "zustand/middleware";
import { api } from "./api";

interface AuthState {
  accessToken: string | null;
  user: { id: string; plan: string; email?: string } | null;
  isAuthenticated: boolean;
  login: (accessToken: string, refreshToken: string) => void;
  logout: () => void;
  setUser: (user: AuthState["user"]) => void;
}

export const useAuthStore = create<AuthState>()(
  persist(
    (set) => ({
      accessToken: null,
      user: null,
      isAuthenticated: false,

      login: (accessToken, _refreshToken) => {
        api.setToken(accessToken);
        set({ accessToken, isAuthenticated: true });
      },

      logout: () => {
        api.clearToken();
        set({ accessToken: null, user: null, isAuthenticated: false });
      },

      setUser: (user) => set({ user }),
    }),
    {
      name: "tene-auth",
      partialize: (state) => ({
        accessToken: state.accessToken,
        user: state.user,
        isAuthenticated: state.isAuthenticated,
      }),
      onRehydrateStorage: () => (state) => {
        if (state?.accessToken) {
          api.setToken(state.accessToken);
        }
      },
    }
  )
);
