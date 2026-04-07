const API_BASE = process.env.NEXT_PUBLIC_API_URL || "https://api.tene.sh";

interface ApiResponse<T> {
  ok: boolean;
  data: T;
  meta?: { timestamp: string; request_id: string };
  error?: string;
  message?: string;
}

class ApiClient {
  private accessToken: string | null = null;

  setToken(token: string) {
    this.accessToken = token;
  }

  clearToken() {
    this.accessToken = null;
  }

  private async request<T>(path: string, options: RequestInit = {}): Promise<T> {
    const headers: Record<string, string> = {
      "Content-Type": "application/json",
      ...(options.headers as Record<string, string>),
    };

    if (this.accessToken) {
      headers["Authorization"] = `Bearer ${this.accessToken}`;
    }

    const res = await fetch(`${API_BASE}${path}`, { ...options, headers });
    const json: ApiResponse<T> = await res.json();

    if (!json.ok) {
      throw new Error(json.message || json.error || "API error");
    }

    return json.data;
  }

  // Auth
  getMe() {
    return this.request<{ user_id: string; plan: string }>("/api/v1/auth/me");
  }

  // Vaults
  listVaults() {
    return this.request<any[]>("/api/v1/vaults");
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

  // Waitlist
  joinWaitlist(email: string) {
    return this.request<{ message: string }>("/api/v1/waitlist", {
      method: "POST",
      body: JSON.stringify({ email, source: "dashboard" }),
    });
  }
}

export const api = new ApiClient();
