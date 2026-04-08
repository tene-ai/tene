const API_BASE = process.env.NEXT_PUBLIC_API_URL || "https://api.tene.sh";

// --- Domain Types ---

export interface Vault {
  id: string;
  user_id: string;
  team_id?: string;
  project_name: string;
  s3_key: string;
  vault_version: number;
  vault_hash: string;
  secret_count: number;
  size: number;
  created_at: string;
  updated_at: string;
}

export interface Team {
  id: string;
  name: string;
  slug: string;
  owner_id: string;
  created_at: string;
  updated_at: string;
}

export interface TeamMember {
  team_id: string;
  user_id: string;
  role: string;
  env_permissions: string[];
  joined_at: string;
  name?: string;
  avatar_url?: string;
}

export interface Device {
  id: string;
  device_name: string;
  last_seen_at: string;
  created_at: string;
}

export interface AuditLog {
  id: string;
  user_id: string;
  vault_id?: string;
  action: string;
  detail?: string;
  ip_address?: string;
  created_at: string;
}

export interface VaultKeyEntry {
  name: string;
  version: number;
  updated_at: string;
}

export interface VaultKeysResponse {
  vault_id: string;
  environment: string;
  keys: VaultKeyEntry[];
  environments: string[];
}

export interface OnboardingStatus {
  cli_installed: boolean;
  first_push: boolean;
  second_device: boolean;
  completed: boolean;
  dismissed: boolean;
}

export interface AuditFilter {
  action?: string;
  limit?: number;
  offset?: number;
}

// --- API Response ---

interface ApiResponse<T> {
  ok: boolean;
  data: T;
  meta?: { timestamp: string; request_id: string };
  error?: string;
  message?: string;
}

class ApiClient {
  private accessToken: string | null = null;
  private refreshTokenValue: string | null = null;
  private isRefreshing = false;
  private refreshPromise: Promise<boolean> | null = null;

  setToken(token: string) {
    this.accessToken = token;
  }

  clearToken() {
    this.accessToken = null;
    this.refreshTokenValue = null;
  }

  setRefreshToken(token: string) {
    this.refreshTokenValue = token;
  }

  private async request<T>(path: string, options: RequestInit = {}, isRetry = false): Promise<T> {
    const headers: Record<string, string> = {
      "Content-Type": "application/json",
      ...(options.headers as Record<string, string>),
    };

    if (this.accessToken) {
      headers["Authorization"] = `Bearer ${this.accessToken}`;
    }

    const res = await fetch(`${API_BASE}${path}`, { ...options, headers });

    // Handle 401: attempt token refresh (skip for refresh endpoint itself and retries)
    if (res.status === 401 && !isRetry && path !== "/api/v1/auth/refresh") {
      const refreshed = await this.tryRefresh();
      if (refreshed) {
        return this.request<T>(path, options, true);
      }
      // Refresh failed — trigger logout via auth store
      this.triggerLogout();
      throw new Error("Session expired. Please sign in again.");
    }

    const json: ApiResponse<T> = await res.json();

    if (!json.ok) {
      throw new Error(json.message || json.error || "API error");
    }

    return json.data;
  }

  private async tryRefresh(): Promise<boolean> {
    if (!this.refreshTokenValue) return false;

    // Deduplicate concurrent refresh attempts
    if (this.isRefreshing && this.refreshPromise) {
      return this.refreshPromise;
    }

    this.isRefreshing = true;
    this.refreshPromise = this.doRefresh();

    try {
      return await this.refreshPromise;
    } finally {
      this.isRefreshing = false;
      this.refreshPromise = null;
    }
  }

  private async doRefresh(): Promise<boolean> {
    try {
      const res = await fetch(`${API_BASE}/api/v1/auth/refresh`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ refresh_token: this.refreshTokenValue }),
      });

      if (!res.ok) return false;

      const json: ApiResponse<{ access_token: string; refresh_token: string }> = await res.json();
      if (!json.ok || !json.data) return false;

      // Update tokens in api client
      this.accessToken = json.data.access_token;
      this.refreshTokenValue = json.data.refresh_token;

      // Update auth store (persisted state)
      this.onTokenRefresh?.(json.data.access_token, json.data.refresh_token);

      // Update cookie for Next.js middleware
      document.cookie = `tene_access_token=${json.data.access_token}; path=/; max-age=${15 * 60}; SameSite=Lax`;

      return true;
    } catch {
      return false;
    }
  }

  // Callbacks set by auth-store to sync state
  onTokenRefresh: ((accessToken: string, refreshToken: string) => void) | null = null;
  private triggerLogout: () => void = () => {};

  setLogoutHandler(handler: () => void) {
    this.triggerLogout = handler;
  }

  // Auth
  async exchangeAuthCode(code: string): Promise<{ access_token: string; refresh_token: string; expires_in: number }> {
    const res = await fetch(`${API_BASE}/api/v1/auth/exchange`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ code }),
    });
    const json: ApiResponse<{ access_token: string; refresh_token: string; expires_in: number }> = await res.json();
    if (!json.ok) throw new Error(json.error || "exchange failed");
    return json.data;
  }

  getMe() {
    return this.request<{ user_id: string; plan: "free" | "pro"; email: string; name: string; avatar_url: string }>("/api/v1/auth/me");
  }

  signout(refreshToken?: string) {
    return this.request<{ message: string }>("/api/v1/auth/signout", {
      method: "POST",
      body: JSON.stringify(refreshToken ? { refresh_token: refreshToken } : {}),
    });
  }

  // Vaults
  listVaults() {
    return this.request<Vault[]>("/api/v1/vaults");
  }

  getVault(id: string) {
    return this.request<Vault>(`/api/v1/vaults/${id}`);
  }

  // Billing
  getSubscription() {
    return this.request<{ plan: string; status: string }>("/api/v1/billing/subscription");
  }

  createCheckout(email: string) {
    return this.request<{ checkout_url: string }>("/api/v1/billing/checkout", {
      method: "POST",
      body: JSON.stringify({ email }),
    });
  }

  getPortal() {
    return this.request<{ portal_url: string }>("/api/v1/billing/portal", {
      method: "POST",
    });
  }

  // Teams
  listTeams() {
    return this.request<Team[]>("/api/v1/teams");
  }

  createTeam(name: string, slug: string) {
    return this.request<Team>("/api/v1/teams", {
      method: "POST",
      body: JSON.stringify({ name, slug }),
    });
  }

  inviteTeamMember(teamId: string, userId: string, role: string, envPermissions?: string[]) {
    return this.request<TeamMember>(`/api/v1/teams/${teamId}/invite`, {
      method: "POST",
      body: JSON.stringify({ user_id: userId, role, env_permissions: envPermissions }),
    });
  }

  removeTeamMember(teamId: string, userId: string) {
    return this.request<void>(`/api/v1/teams/${teamId}/members/${userId}`, {
      method: "DELETE",
    });
  }

  updateMemberRole(teamId: string, userId: string, role: string) {
    return this.request<TeamMember>(`/api/v1/teams/${teamId}/members/${userId}/role`, {
      method: "PATCH",
      body: JSON.stringify({ role }),
    });
  }

  listTeamMembers(teamId: string) {
    return this.request<TeamMember[]>(`/api/v1/teams/${teamId}/members`);
  }

  // Devices
  listDevices() {
    return this.request<Device[]>("/api/v1/devices");
  }

  deleteDevice(id: string) {
    return this.request<void>(`/api/v1/devices/${id}`, {
      method: "DELETE",
    });
  }

  // Audit
  listAuditLogs(params?: { action?: string; limit?: number; offset?: number }) {
    const query = new URLSearchParams();
    if (params?.action) query.set("action", params.action);
    if (params?.limit) query.set("limit", String(params.limit));
    if (params?.offset) query.set("offset", String(params.offset));
    const qs = query.toString();
    return this.request<AuditLog[]>(`/api/v1/audit${qs ? `?${qs}` : ""}`);
  }

  // Vault Keys (metadata only, no values)
  getVaultKeys(vaultId: string, env?: string) {
    const query = env ? `?env=${encodeURIComponent(env)}` : "";
    return this.request<VaultKeysResponse>(`/api/v1/vaults/${vaultId}/keys${query}`);
  }

  // Onboarding
  getOnboardingStatus() {
    return this.request<OnboardingStatus>("/api/v1/onboarding/status");
  }

  dismissOnboarding() {
    return this.request<{ message: string }>("/api/v1/onboarding/dismiss", {
      method: "POST",
    });
  }

}

export const api = new ApiClient();
