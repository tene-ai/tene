import { create } from "zustand";
import { persist } from "zustand/middleware";
import { api } from "./api";

interface UserProfile {
  id: string;
  plan: "free" | "pro";
  email: string;
  name: string;
  avatar_url: string;
}

interface AuthState {
  accessToken: string | null;
  refreshToken: string | null;
  user: UserProfile | null;
  isAuthenticated: boolean;
  login: (accessToken: string, refreshToken: string) => void;
  logout: () => void;
  setUser: (user: UserProfile) => void;
}

export const useAuthStore = create<AuthState>()(
  persist(
    (set) => ({
      accessToken: null,
      refreshToken: null,
      user: null,
      isAuthenticated: false,

      login: (accessToken, refreshToken) => {
        api.setToken(accessToken);
        api.setRefreshToken(refreshToken);
        set({ accessToken, refreshToken, isAuthenticated: true });

        // Wire up token refresh callback
        api.onTokenRefresh = (newAccess, newRefresh) => {
          set({ accessToken: newAccess, refreshToken: newRefresh });
        };

        // Wire up logout handler
        api.setLogoutHandler(() => {
          useAuthStore.getState().logout();
        });

        // Fetch full user profile
        api.getMe().then((me) => {
          set({
            user: {
              id: me.user_id,
              plan: me.plan,
              email: me.email || "",
              name: me.name || "",
              avatar_url: me.avatar_url || "",
            },
          });
        }).catch(() => {});
      },

      logout: () => {
        api.clearToken();
        api.onTokenRefresh = null;
        set({ accessToken: null, refreshToken: null, user: null, isAuthenticated: false });
      },

      setUser: (user) => set({ user }),
    }),
    {
      name: "tene-auth",
      partialize: (state) => ({
        accessToken: state.accessToken,
        refreshToken: state.refreshToken,
        user: state.user,
        isAuthenticated: state.isAuthenticated,
      }),
      onRehydrateStorage: () => (state) => {
        if (state?.accessToken) {
          api.setToken(state.accessToken);
        }
        if (state?.refreshToken) {
          api.setRefreshToken(state.refreshToken);
        }
        if (state?.isAuthenticated) {
          api.onTokenRefresh = (newAccess, newRefresh) => {
            useAuthStore.setState({ accessToken: newAccess, refreshToken: newRefresh });
          };
          api.setLogoutHandler(() => {
            useAuthStore.getState().logout();
          });
        }
      },
    }
  )
);
