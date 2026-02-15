import React from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { ArrowLeft, Clock, Tag, Hash } from 'lucide-react';
import Loading from '../components/Loading';
import { useApi } from '../hooks/useApi';
import { fetchEvent } from '../services/api';
import { formatDate, timeAgo, prettyJson } from '../utils/formatters';

const EventDetail: React.FC = () => {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const { data: event, loading, error } = useApi(
    () => fetchEvent(id!),
    [id],
  );

  if (loading) return <Loading message="Loading event…" />;
  if (error) return <div className="error-banner">{error}</div>;
  if (!event) return <div className="empty-state">Event not found.</div>;

  return (
    <div className="page">
      <button className="btn btn--ghost mb-4" onClick={() => navigate(-1)}>
        <ArrowLeft size={16} /> Back
      </button>

      <h2 className="page-title">Event Detail</h2>

      <div className="detail-grid">
        {/* Meta info */}
        <div className="card">
          <h3 className="card-title">Metadata</h3>
          <dl className="detail-list">
            <div className="detail-row">
              <dt><Hash size={14} /> Event ID</dt>
              <dd className="mono">{event.event_id}</dd>
            </div>
            <div className="detail-row">
              <dt><Tag size={14} /> Type</dt>
              <dd><span className="type-badge">{event.event_type}</span></dd>
            </div>
            <div className="detail-row">
              <dt><Clock size={14} /> Received</dt>
              <dd>
                {formatDate(event.received_at)}
                <span className="dim ml-2">({timeAgo(event.received_at)})</span>
              </dd>
            </div>
          </dl>
        </div>

        {/* Payload */}
        <div className="card">
          <h3 className="card-title">Payload</h3>
          <pre className="json-viewer">{prettyJson(event.payload)}</pre>
        </div>
      </div>
    </div>
  );
};

export default EventDetail;
