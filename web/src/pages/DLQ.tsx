import React from 'react';
import { AlertTriangle, RefreshCw } from 'lucide-react';

/**
 * DLQ (Dead Letter Queue) page.
 *
 * The backend currently writes failed messages to the `events.dlq` Kafka topic.
 * A dedicated DLQ query endpoint would be needed to surface these here.
 * For now this shows an informational placeholder that can be wired up
 * once a DLQ read API is added.
 */
const DLQ: React.FC = () => (
  <div className="page">
    <h2 className="page-title">Dead Letter Queue</h2>

    <div className="card dlq-placeholder">
      <div className="dlq-icon">
        <AlertTriangle size={48} />
      </div>
      <h3>DLQ Monitoring</h3>
      <p>
        Messages that fail processing after all retries are routed to the{' '}
        <code>events.dlq</code> Kafka topic with a full forensic envelope
        containing the original message, error details, and retry history.
      </p>

      <div className="dlq-info-grid">
        <div className="dlq-info-card">
          <h4>Envelope Fields</h4>
          <ul>
            <li><code>original_topic</code> — Source topic</li>
            <li><code>original_partition</code> — Source partition</li>
            <li><code>original_offset</code> — Message offset</li>
            <li><code>original_key</code> — Event ID</li>
            <li><code>original_value</code> — Raw message bytes</li>
            <li><code>error_message</code> — Failure reason</li>
            <li><code>error_type</code> — transient / permanent</li>
            <li><code>retry_count</code> — Attempts made</li>
            <li><code>failed_at</code> — Timestamp</li>
          </ul>
        </div>
        <div className="dlq-info-card">
          <h4>Next Steps</h4>
          <p>To enable DLQ browsing in this dashboard:</p>
          <ol>
            <li>Add a DLQ consumer that reads from <code>events.dlq</code> and persists to a <code>dlq_events</code> table</li>
            <li>Expose <code>GET /v1/dlq</code> and <code>GET /v1/dlq/:id</code> endpoints</li>
            <li>Wire this page to those endpoints</li>
          </ol>
          <button className="btn btn--primary" disabled>
            <RefreshCw size={14} /> Connect DLQ (coming soon)
          </button>
        </div>
      </div>
    </div>
  </div>
);

export default DLQ;
