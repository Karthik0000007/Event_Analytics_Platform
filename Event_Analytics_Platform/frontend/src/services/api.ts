import type {
  EventListResponse,
  EventRecord,
  EventFilters,
  Summary,
  TypeCount,
  TimelinePoint,
} from '../types';

const BASE = '/v1';

async function request<T>(url: string, init?: RequestInit): Promise<T> {
  const res = await fetch(url, init);
  if (!res.ok) {
    const text = await res.text().catch(() => res.statusText);
    throw new Error(text || `HTTP ${res.status}`);
  }
  return res.json();
}

/* ── Events ─────────────────────────────────────────── */

export async function fetchEvents(filters: EventFilters = {}): Promise<EventListResponse> {
  const params = new URLSearchParams();
  if (filters.type) params.set('type', filters.type);
  if (filters.from) params.set('from', filters.from);
  if (filters.to) params.set('to', filters.to);
  if (filters.limit) params.set('limit', String(filters.limit));
  if (filters.offset !== undefined) params.set('offset', String(filters.offset));

  const qs = params.toString();
  return request<EventListResponse>(`${BASE}/events${qs ? `?${qs}` : ''}`);
}

export async function fetchEvent(id: string): Promise<EventRecord> {
  return request<EventRecord>(`${BASE}/events/${id}`);
}

export async function createEvent(data: {
  event_id: string;
  event_type: string;
  payload: Record<string, unknown>;
}): Promise<{ status: string; event_id: string }> {
  return request(`${BASE}/events`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(data),
  });
}

/* ── Analytics ──────────────────────────────────────── */

export async function fetchSummary(): Promise<Summary> {
  return request<Summary>(`${BASE}/analytics/summary`);
}

export async function fetchTypeCounts(): Promise<TypeCount[]> {
  return request<TypeCount[]>(`${BASE}/analytics/types`);
}

export async function fetchTimeline(hours = 24): Promise<TimelinePoint[]> {
  return request<TimelinePoint[]>(`${BASE}/analytics/timeline?hours=${hours}`);
}

/* ── Health ──────────────────────────────────────────── */

export async function checkHealth(): Promise<boolean> {
  try {
    const res = await fetch('/healthz');
    return res.ok;
  } catch {
    return false;
  }
}