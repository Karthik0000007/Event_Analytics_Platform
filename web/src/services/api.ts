import type {
  AcceptedResponse,
  DeadLetterListResponse,
  EventListResponse,
  EventRecord,
  EventFilters,
  Summary,
  TypeCount,
  TimelinePoint,
} from '../types';

const BASE = '/v1';

/** Fetch JSON. On an error status, throw the API's {"error": "..."} message. */
async function request<T>(url: string, init?: RequestInit): Promise<T> {
  const res = await fetch(url, init);
  if (!res.ok) {
    const text = await res.text().catch(() => '');
    let message = text || `HTTP ${res.status}`;
    try {
      const body = JSON.parse(text) as { error?: string };
      if (body.error) message = body.error;
    } catch {
      /* not JSON, keep the raw text */
    }
    throw new Error(`${message} (HTTP ${res.status})`);
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
}): Promise<AcceptedResponse> {
  return request(`${BASE}/events`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(data),
  });
}

/** A random UUID. crypto.randomUUID only exists on https or localhost. */
export function newEventId(): string {
  if (typeof crypto !== 'undefined' && 'randomUUID' in crypto) {
    return crypto.randomUUID();
  }
  return 'xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx'.replace(/[xy]/g, (c) => {
    const r = (Math.random() * 16) | 0;
    return (c === 'x' ? r : (r & 0x3) | 0x8).toString(16);
  });
}

const DEMO_TYPES = ['click', 'page_view', 'add_to_cart', 'purchase', 'signup'];

/** Send a valid random event. It should appear on the Events page. */
export function sendTestEvent(): Promise<AcceptedResponse> {
  const type = DEMO_TYPES[Math.floor(Math.random() * DEMO_TYPES.length)];
  return createEvent({
    event_id: newEventId(),
    event_type: type,
    payload: { source: 'dashboard', page: '/demo', value: Math.round(Math.random() * 100) },
  });
}

/**
 * Send an event the API accepts but PostgreSQL rejects.
 * The payload holds the NUL character (\u0000). It is valid JSON, but a JSONB
 * column cannot store it (SQLSTATE 22P05). The consumer classifies that as a
 * permanent error and routes the message to the DLQ.
 */
export function sendBadEvent(): Promise<AcceptedResponse> {
  return createEvent({
    event_id: newEventId(),
    event_type: 'bad_event',
    payload: { note: 'Postgres JSONB cannot store this character: \u0000' },
  });
}

/* ── Dead letters ───────────────────────────────────── */

export async function fetchDeadLetters(limit: number, offset: number): Promise<DeadLetterListResponse> {
  return request<DeadLetterListResponse>(`${BASE}/dlq?limit=${limit}&offset=${offset}`);
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
