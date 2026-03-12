import { ApiAuthResponse, ApiEvent, ApiListEventsResponse, ApiBooking, ApiListBookingsResponse, ApiUser, ApiSeat, ApiGetSeatsParams, ApiUpdateSeatStatusParams } from "./types";

class ApiClient {
  private baseUrl: string;

  constructor() {
    // Use relative URL so Next.js rewrites proxy to backend
    this.baseUrl = "/api";
  }

  private getToken(): string | null {
    if (typeof window === "undefined") return null;
    return localStorage.getItem("access_token");
  }

  private getRefreshToken(): string | null {
    if (typeof window === "undefined") return null;
    return localStorage.getItem("refresh_token");
  }

  private setTokens(access: string, refresh: string) {
    localStorage.setItem("access_token", access);
    localStorage.setItem("refresh_token", refresh);
  }

  private clearTokens() {
    localStorage.removeItem("access_token");
    localStorage.removeItem("refresh_token");
  }

  private async request<T>(path: string, options: RequestInit = {}): Promise<T> {
    const token = this.getToken();
    const headers: Record<string, string> = {
      "Content-Type": "application/json",
      ...((options.headers as Record<string, string>) || {}),
    };
    if (token) {
      headers["Authorization"] = `Bearer ${token}`;
    }

    const res = await fetch(`${this.baseUrl}${path}`, {
      ...options,
      headers,
    });

    if (res.status === 401 && token) {
      // Try refresh
      const refreshed = await this.tryRefresh();
      if (refreshed) {
        headers["Authorization"] = `Bearer ${this.getToken()}`;
        const retry = await fetch(`${this.baseUrl}${path}`, { ...options, headers });
        if (!retry.ok) {
          const err = await retry.json().catch(() => ({ error: "Request failed" }));
          throw new Error(err.error || "Request failed");
        }
        return retry.json();
      }
      this.clearTokens();
      throw new Error("Session expired");
    }

    if (!res.ok) {
      const err = await res.json().catch(() => ({ error: "Request failed" }));
      throw new Error(err.error || `Request failed with status ${res.status}`);
    }

    return res.json();
  }

  private async tryRefresh(): Promise<boolean> {
    const refreshToken = this.getRefreshToken();
    if (!refreshToken) return false;
    try {
      const res = await fetch(`${this.baseUrl}/auth/refresh`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ refresh_token: refreshToken }),
      });
      if (!res.ok) return false;
      const data: ApiAuthResponse = await res.json();
      this.setTokens(data.access_token, data.refresh_token);
      return true;
    } catch {
      return false;
    }
  }

  // ── Auth ──────────────────────────────────────────────
  async register(email: string, password: string, name: string): Promise<ApiAuthResponse> {
    const data = await this.request<ApiAuthResponse>("/auth/register", {
      method: "POST",
      body: JSON.stringify({ email, password, name }),
    });
    this.setTokens(data.access_token, data.refresh_token);
    return data;
  }

  async login(email: string, password: string): Promise<ApiAuthResponse> {
    const data = await this.request<ApiAuthResponse>("/auth/login", {
      method: "POST",
      body: JSON.stringify({ email, password }),
    });
    this.setTokens(data.access_token, data.refresh_token);
    return data;
  }

  async logout(): Promise<void> {
    try {
      await this.request("/auth/logout", { method: "POST" });
    } finally {
      this.clearTokens();
    }
  }

  async getProfile(): Promise<ApiUser> {
    return this.request<ApiUser>("/users/me");
  }

  // ── Events ────────────────────────────────────────────
  async listEvents(params?: { category?: string; search?: string; page?: number; page_size?: number }): Promise<ApiListEventsResponse> {
    const query = new URLSearchParams();
    if (params?.category) query.set("category", params.category);
    if (params?.search) query.set("search", params.search);
    if (params?.page) query.set("page", String(params.page));
    if (params?.page_size) query.set("page_size", String(params.page_size));
    const qs = query.toString();
    return this.request<ApiListEventsResponse>(`/events${qs ? `?${qs}` : ""}`);
  }

  async getEvent(id: string): Promise<ApiEvent> {
    return this.request<ApiEvent>(`/events/${id}`);
  }

  // ── Bookings ──────────────────────────────────────────
  async createBooking(eventId: string, items: { ticket_tier_id: string; quantity: number; seat_ids: string[] }[]): Promise<ApiBooking> {
    return this.request<ApiBooking>("/bookings", {
      method: "POST",
      body: JSON.stringify({ event_id: eventId, items }),
    });
  }

  async listBookings(page?: number, pageSize?: number): Promise<ApiListBookingsResponse> {
    const query = new URLSearchParams();
    if (page) query.set("page", String(page));
    if (pageSize) query.set("page_size", String(pageSize));
    const qs = query.toString();
    return this.request<ApiListBookingsResponse>(`/bookings${qs ? `?${qs}` : ""}`);
  }

  async getBooking(id: string): Promise<ApiBooking> {
    return this.request<ApiBooking>(`/bookings/${id}`);
  }

  async cancelBooking(id: string): Promise<ApiBooking> {
    return this.request<ApiBooking>(`/bookings/${id}/cancel`, { method: "POST" });
  }

  // ── Seats ───────────────────────────────────────────────
  async getSeats(params: ApiGetSeatsParams): Promise<{ seats: ApiSeat[] }> {
    const query = new URLSearchParams();
    query.set("event_id", params.event_id);
    if (params.ticket_tier_id) {
      query.set("ticket_tier_id", params.ticket_tier_id);
    }
    const qs = query.toString();
    return this.request<{ seats: ApiSeat[] }>(`/seats${qs ? `?${qs}` : ""}`);
  }

  async updateSeatStatus(params: ApiUpdateSeatStatusParams): Promise<ApiSeat> {
    const body: Record<string, string> = {
      seat_id: params.seat_id,
      status: params.status,
    };
    if (params.booking_id) {
      body.booking_id = params.booking_id;
    }
    return this.request<ApiSeat>("/seats/status", {
      method: "PATCH",
      body: JSON.stringify(body),
    });
  }

  // ── Token state check ─────────────────────────────────
  isLoggedIn(): boolean {
    return !!this.getToken();
  }
}

export const apiClient = new ApiClient();
