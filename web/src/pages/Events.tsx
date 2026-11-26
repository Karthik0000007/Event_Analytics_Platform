import React, { useState, useCallback } from 'react';
import { ChevronLeft, ChevronRight, Send, AlertTriangle } from 'lucide-react';
import EventTable from '../components/EventTable';
import Filters from '../components/Filters';
import Loading from '../components/Loading';
import Toast, { useToast } from '../components/Toast';
import { useApi } from '../hooks/useApi';
import { fetchEvents, fetchTypeCounts, sendBadEvent, sendTestEvent } from '../services/api';
import type { AcceptedResponse, EventFilters } from '../types';

const PAGE_SIZE = 25;
// The consumer usually stores an event well within a second. Wait a little, then reload.
const RELOAD_AFTER_MS = 1500;

const Events: React.FC = () => {
  const [filters, setFilters] = useState<EventFilters>({
    limit: PAGE_SIZE,
    offset: 0,
  });
  const [sending, setSending] = useState(false);
  const { toast, show, hide } = useToast();

  const { data, loading, error, refetch } = useApi(() => fetchEvents(filters), [filters]);

  const { data: typeCounts } = useApi(() => fetchTypeCounts(), []);
  const typeOptions = typeCounts?.map((t) => t.event_type) ?? [];

  const total = data?.total ?? 0;
  const currentPage = Math.floor((filters.offset ?? 0) / PAGE_SIZE) + 1;
  const totalPages = Math.ceil(total / PAGE_SIZE);

  const goPage = useCallback((page: number) => {
    setFilters((f) => ({ ...f, offset: (page - 1) * PAGE_SIZE }));
  }, []);

  const send = async (kind: 'good' | 'bad') => {
    setSending(true);
    try {
      const res: AcceptedResponse = kind === 'good' ? await sendTestEvent() : await sendBadEvent();
      const id = res.event_id.slice(0, 8);
      show({
        kind: 'success',
        text:
          kind === 'good'
            ? `Event ${id} accepted (202). It will appear in the list below.`
            : `Bad event ${id} accepted (202). Postgres will reject it, so watch the Dead Letters page.`,
      });
      setTimeout(refetch, RELOAD_AFTER_MS);
    } catch (err) {
      show({ kind: 'error', text: `Send failed: ${(err as Error).message}` });
    } finally {
      setSending(false);
    }
  };

  return (
    <div className="page">
      <div className="page-header-row">
        <h2 className="page-title">Events</h2>
        <div className="button-row">
          <button className="btn btn--primary" disabled={sending} onClick={() => send('good')}>
            <Send size={14} /> Send test event
          </button>
          <button className="btn btn--danger" disabled={sending} onClick={() => send('bad')}>
            <AlertTriangle size={14} /> Send bad event
          </button>
        </div>
      </div>

      <Filters filters={filters} onChange={setFilters} typeOptions={typeOptions} />

      {error && <div className="error-banner">{error}</div>}

      {loading ? (
        <Loading />
      ) : (
        <>
          <div className="table-info">
            Showing {data?.events.length ?? 0} of {total} events
          </div>
          <EventTable events={data?.events ?? []} />

          {totalPages > 1 && (
            <div className="pagination">
              <button
                className="btn btn--ghost"
                disabled={currentPage <= 1}
                onClick={() => goPage(currentPage - 1)}
              >
                <ChevronLeft size={16} /> Prev
              </button>
              <span className="pagination-info">
                Page {currentPage} of {totalPages}
              </span>
              <button
                className="btn btn--ghost"
                disabled={currentPage >= totalPages}
                onClick={() => goPage(currentPage + 1)}
              >
                Next <ChevronRight size={16} />
              </button>
            </div>
          )}
        </>
      )}

      <Toast toast={toast} onClose={hide} />
    </div>
  );
};

export default Events;
