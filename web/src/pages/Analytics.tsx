import React, { useState } from 'react';
import { TimelineChart, TypeBarChart, TypePieChart } from '../components/EventChart';
import Loading from '../components/Loading';
import StatsCard from '../components/StatsCard';
import { useApi } from '../hooks/useApi';
import { fetchSummary, fetchTimeline, fetchTypeCounts } from '../services/api';
import { formatNumber } from '../utils/formatters';
import { Activity, Layers, Clock } from 'lucide-react';

const HOUR_OPTIONS = [
  { label: '6h', value: 6 },
  { label: '12h', value: 12 },
  { label: '24h', value: 24 },
  { label: '48h', value: 48 },
  { label: '7d', value: 168 },
];

const Analytics: React.FC = () => {
  const [hours, setHours] = useState(24);
  const { data: summary, loading: sLoad } = useApi(() => fetchSummary(), []);
  const { data: timeline, loading: tLoad } = useApi(
    () => fetchTimeline(hours),
    [hours],
  );
  const { data: typeCounts, loading: tcLoad } = useApi(
    () => fetchTypeCounts(),
    [],
  );

  if (sLoad && tLoad && tcLoad) return <Loading message="Loading analytics…" />;

  return (
    <div className="page">
      <h2 className="page-title">Analytics</h2>

      {/* ── Summary Cards ────────────────────────────── */}
      <div className="stats-grid">
        <StatsCard
          title="Total Events"
          value={formatNumber(summary?.total_events ?? 0)}
          icon={<Activity size={20} />}
          color="#6366f1"
        />
        <StatsCard
          title="Distinct Types"
          value={summary?.event_types ?? 0}
          icon={<Layers size={20} />}
          color="#22d3ee"
        />
        <StatsCard
          title="Today"
          value={formatNumber(summary?.today_events ?? 0)}
          icon={<Clock size={20} />}
          color="#f59e0b"
        />
      </div>

      {/* ── Timeline ─────────────────────────────────── */}
      <div className="card">
        <div className="card-header-row">
          <h3 className="card-title">Event Volume</h3>
          <div className="toggle-group">
            {HOUR_OPTIONS.map((opt) => (
              <button
                key={opt.value}
                className={`toggle-btn ${hours === opt.value ? 'toggle-btn--active' : ''}`}
                onClick={() => setHours(opt.value)}
              >
                {opt.label}
              </button>
            ))}
          </div>
        </div>
        {tLoad ? (
          <Loading />
        ) : timeline && timeline.length > 0 ? (
          <TimelineChart data={timeline} height={350} />
        ) : (
          <div className="empty-state">No data for the selected range.</div>
        )}
      </div>

      {/* ── Type distribution ────────────────────────── */}
      <div className="chart-grid">
        <div className="card">
          <h3 className="card-title">Type Distribution</h3>
          {tcLoad ? (
            <Loading />
          ) : typeCounts && typeCounts.length > 0 ? (
            <TypePieChart data={typeCounts} height={320} />
          ) : (
            <div className="empty-state">No type data.</div>
          )}
        </div>
        <div className="card">
          <h3 className="card-title">Events by Type</h3>
          {tcLoad ? (
            <Loading />
          ) : typeCounts && typeCounts.length > 0 ? (
            <TypeBarChart data={typeCounts} height={320} />
          ) : (
            <div className="empty-state">No type data.</div>
          )}
        </div>
      </div>
    </div>
  );
};

export default Analytics;
