import React, { useCallback, useEffect, useState } from 'react';
import { ChevronLeft, ChevronRight, RefreshCw } from 'lucide-react';
import Loading from '../components/Loading';
import StatusBadge from '../components/StatusBadge';
import { fetchDeadLetters } from '../services/api';
import type { DeadLetter, DeadLetterListResponse } from '../types';
import { formatDate, truncate } from '../utils/formatters';

const PAGE_SIZE = 20;
const REFRESH_MS = 3000;

function kindBadge(kind: string) {
  if (kind === 'permanent') return <StatusBadge status="error" label="permanent" />;
  if (kind === 'transient') return <StatusBadge status="warning" label="transient" />;
  return <StatusBadge status="info" label={kind || 'unknown'} />;
}

/**
 * Dead Letter Queue page.
 * Failed messages go to the Kafka topic events.dlq. The dlq-recorder service
 * stores them in Postgres, and this page reads them from GET /v1/dlq.
 * It refreshes every few seconds so new failures appear without a reload.
 */
const DLQ: React.FC = () => {
  const [offset, setOffset] = useState(0);
  const [data, setData] = useState<DeadLetterListResponse | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [openId, setOpenId] = useState<number | null>(null);

  const load = useCallback(() => {
    fetchDeadLetters(PAGE_SIZE, offset)
      .then((res) => {
        setData(res);
        setError(null);
      })
      .catch((err: Error) => setError(err.message));
  }, [offset]);

  useEffect(() => {
    load();
    const id = setInterval(load, REFRESH_MS);
    return () => clearInterval(id);
  }, [load]);

  const total = data?.total ?? 0;
  const page = Math.floor(offset / PAGE_SIZE) + 1;
  const pages = Math.max(1, Math.ceil(total / PAGE_SIZE));

  return (
    <div className="page">
      <div className="page-header-row">
        <h2 className="page-title">Dead Letter Queue</h2>
        <button className="btn btn--ghost" onClick={load} title="Refresh now">
          <RefreshCw size={14} /> Refresh
        </button>
      </div>

      <p className="dim mb-4">
        Messages the consumer could not store. Permanent errors (bad JSON, values Postgres
        rejects) come here at once. Transient errors come here only after the retry budget is
        used up while the database is healthy. Refreshes every {REFRESH_MS / 1000} seconds.
      </p>

      {error && <div className="error-banner">{error}</div>}

      {!data && !error ? (
        <Loading />
      ) : total === 0 ? (
        <div className="card empty-state">
          No dead letters yet. Use <strong>Send bad event</strong> on the Events page to create one.
        </div>
      ) : (
        <>
          <div className="table-info">
            Showing {data?.dead_letters.length ?? 0} of {total} dead letters
          </div>
          <div className="table-wrapper">
            <table className="data-table">
              <thead>
                <tr>
                  <th>Failed at</th>
                  <th>Event key</th>
                  <th>Error kind</th>
                  <th>Message</th>
                  <th>Retries</th>
                </tr>
              </thead>
              <tbody>
                {data?.dead_letters.map((d) => (
                  <DeadLetterRow
                    key={d.id}
                    d={d}
                    open={openId === d.id}
                    onToggle={() => setOpenId(openId === d.id ? null : d.id)}
                  />
                ))}
              </tbody>
            </table>
          </div>

          {pages > 1 && (
            <div className="pagination">
              <button className="btn btn--ghost" disabled={page <= 1} onClick={() => setOffset(offset - PAGE_SIZE)}>
                <ChevronLeft size={16} /> Prev
              </button>
              <span className="pagination-info">
                Page {page} of {pages}
              </span>
              <button className="btn btn--ghost" disabled={page >= pages} onClick={() => setOffset(offset + PAGE_SIZE)}>
                Next <ChevronRight size={16} />
              </button>
            </div>
          )}
        </>
      )}
    </div>
  );
};

interface RowProps {
  d: DeadLetter;
  open: boolean;
  onToggle: () => void;
}

const DeadLetterRow: React.FC<RowProps> = ({ d, open, onToggle }) => (
  <>
    <tr className="data-table-row" onClick={onToggle} title="Click to see the original message">
      <td>{formatDate(d.failed_at)}</td>
      <td className="mono">{d.original_key ? truncate(d.original_key, 13) : '(none)'}</td>
      <td>{kindBadge(d.error_kind)}</td>
      <td className="mono dim" title={d.error_message}>
        {truncate(d.error_message, 70)}
      </td>
      <td>{d.retry_count}</td>
    </tr>
    {open && (
      <tr>
        <td colSpan={5}>
          <div className="dim mb-2">
            events.dlq partition {d.kafka_partition}, offset {d.kafka_offset}. Original message:
          </div>
          <pre className="json-viewer">{d.original_value || '(empty)'}</pre>
        </td>
      </tr>
    )}
  </>
);

export default DLQ;
