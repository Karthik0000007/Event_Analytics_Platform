import { format, formatDistanceToNow, parseISO } from 'date-fns';

/** Format an ISO timestamp as "Jan 15, 2026 14:32" */
export function formatDate(iso: string): string {
  try {
    return format(parseISO(iso), 'MMM d, yyyy HH:mm');
  } catch {
    return iso;
  }
}

/** "3 minutes ago" */
export function timeAgo(iso: string): string {
  try {
    return formatDistanceToNow(parseISO(iso), { addSuffix: true });
  } catch {
    return iso;
  }
}

/** Format large numbers with commas: 12345 → "12,345" */
export function formatNumber(n: number): string {
  return n.toLocaleString();
}

/** Truncate a string to `len` characters with ellipsis */
export function truncate(s: string, len = 80): string {
  return s.length > len ? s.slice(0, len) + '…' : s;
}

/** Pretty-print JSON with 2-space indentation */
export function prettyJson(obj: unknown): string {
  try {
    return JSON.stringify(obj, null, 2);
  } catch {
    return String(obj);
  }
}
