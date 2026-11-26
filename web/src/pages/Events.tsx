import React, { useState, useCallback } from 'react';
import { ChevronLeft, ChevronRight } from 'lucide-react';
import EventTable from '../components/EventTable';
import Filters from '../components/Filters';
import Loading from '../components/Loading';
import { useApi } from '../hooks/useApi';
import { fetchEvents, fetchTypeCounts } from '../services/api';
import type { EventFilters } from '../types';

const PAGE_SIZE = 25;

const Events: React.FC = () => {
  const [filters, setFilters] = useState<EventFilters>({
    limit: PAGE_SIZE,
    offset: 0,
  });

  const { data, loading, error } = useApi(
    () => fetchEvents(filters),
    [filters],
  );

  const { data: typeCounts } = useApi(() => fetchTypeCounts(), []);
  const typeOptions = typeCounts?.map((t) => t.event_type) ?? [];

  const total = data?.total ?? 0;
  const currentPage = Math.floor((filters.offset ?? 0) / PAGE_SIZE) + 1;
  const totalPages = Math.ceil(total / PAGE_SIZE);

  const goPage = useCallback(
    (page: number) => {
      setFilters((f) => ({ ...f, offset: (page - 1) * PAGE_SIZE }));
    },
    [],
  );

  return (
    <div className="page">
      <h2 className="page-title">Events</h2>

      <Filters
        filters={filters}
        onChange={setFilters}
        typeOptions={typeOptions}
      />

      {error && <div className="error-banner">{error}</div>}

      {loading ? (
        <Loading />
      ) : (
        <>
          <div className="table-info">
            Showing {data?.events.length ?? 0} of {total} events
          </div>
          <EventTable events={data?.events ?? []} />

          {/* Pagination */}
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
    </div>
  );
};

export default Events;
