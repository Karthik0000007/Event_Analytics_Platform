/* ── Domain Types ────────────────────────────────────── */

export interface EventRecord {
  event_id: string;
  event_type: string;
  payload: Record<string, unknown>;
  received_at: string; // ISO-8601
}

export interface EventListResponse {
  events: EventRecord[];
  total: number;
  limit: number;
  offset: number;
}

export interface EventFilters {
  type?: string;
  from?: string;
  to?: string;
  limit?: number;
  offset?: number;
}

/* ── Analytics Types ────────────────────────────────── */

export interface Summary {
  total_events: number;
  today_events: number;
  event_types: number;
  top_types: string[];
}

export interface TypeCount {
  event_type: string;
  count: number;
}

export interface TimelinePoint {
  bucket: string; // ISO-8601
  count: number;
}

/* ── Health ──────────────────────────────────────────── */

export type HealthStatus = 'healthy' | 'unhealthy' | 'loading';

/* ── Component Props ────────────────────────────────── */

export interface StatsCardProps {
  title: string;
  value: string | number;
  icon: React.ReactNode;
  trend?: string;
  color?: string;
}

export interface EventTableProps {
  events: EventRecord[];
  loading?: boolean;
  onRowClick?: (event: EventRecord) => void;
}