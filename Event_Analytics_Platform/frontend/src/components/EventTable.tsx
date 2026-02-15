import React from 'react';
import { useNavigate } from 'react-router-dom';
import { ExternalLink } from 'lucide-react';
import type { EventRecord } from '../types';
import { formatDate, truncate } from '../utils/formatters';

interface EventTableProps {
  events: EventRecord[];
  loading?: boolean;
}

const EventTable: React.FC<EventTableProps> = ({ events, loading }) => {
  const navigate = useNavigate();

  if (loading) {
    return (
      <div className="table-skeleton">
        {Array.from({ length: 5 }).map((_, i) => (
          <div key={i} className="skeleton-row" />
        ))}
      </div>
    );
  }

  if (events.length === 0) {
    return <div className="empty-state">No events found.</div>;
  }

  return (
    <div className="table-wrapper">
      <table className="data-table">
        <thead>
          <tr>
            <th>Event ID</th>
            <th>Type</th>
            <th>Payload</th>
            <th>Received</th>
            <th></th>
          </tr>
        </thead>
        <tbody>
          {events.map((evt) => (
            <tr
              key={evt.event_id}
              className="data-table-row"
              onClick={() => navigate(`/events/${evt.event_id}`)}
            >
              <td className="mono">{evt.event_id.slice(0, 8)}…</td>
              <td>
                <span className="type-badge">{evt.event_type}</span>
              </td>
              <td className="mono dim">
                {truncate(JSON.stringify(evt.payload), 60)}
              </td>
              <td>{formatDate(evt.received_at)}</td>
              <td>
                <ExternalLink size={14} />
              </td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
};

export default EventTable;
