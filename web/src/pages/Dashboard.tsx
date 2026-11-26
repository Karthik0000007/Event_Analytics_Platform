import React from 'react';
import {
  Activity,
  Calendar,
  Layers,
  TrendingUp,
} from 'lucide-react';
import StatsCard from '../components/StatsCard';
import EventTable from '../components/EventTable';
import { TimelineChart, TypePieChart } from '../components/EventChart';
import Loading from '../components/Loading';
import { useApi } from '../hooks/useApi';
import { fetchSummary, fetchEvents, fetchTimeline, fetchTypeCounts } from '../services/api';
import { formatNumber } from '../utils/formatters';

const Dashboard: React.FC = () => {
  const { data: summary, loading: sLoad } = useApi(() => fetchSummary(), []);
  const { data: eventsRes, loading: eLoad } = useApi(
    () => fetchEvents({ limit: 10 }),
    [],
  );
  const { data: timeline, loading: tLoad } = useApi(
    () => fetchTimeline(24),
    [],
  );
  const { data: typeCounts, loading: tcLoad } = useApi(
    () => fetchTypeCounts(),
    [],
  );

  if (sLoad || eLoad || tLoad || tcLoad) return <Loading message="Loading dashboard…" />;

  return (
    <div className="page">
      <h2 className="page-title">Dashboard</h2>

      {/* ── Stats Cards ──────────────────────────────── */}
      <div className="stats-grid">
        <StatsCard
          title="Total Events"
          value={formatNumber(summary?.total_events ?? 0)}
          icon={<Activity size={20} />}
          color="#6366f1"
        />
        <StatsCard
          title="Today"
          value={formatNumber(summary?.today_events ?? 0)}
          icon={<Calendar size={20} />}
          color="#22d3ee"
        />
        <StatsCard
          title="Event Types"
          value={summary?.event_types ?? 0}
          icon={<Layers size={20} />}
          color="#f59e0b"
        />
        <StatsCard
          title="Top Type"
          value={summary?.top_types?.[0] ?? '—'}
          icon={<TrendingUp size={20} />}
          color="#10b981"
        />
      </div>

      {/* ── Charts ───────────────────────────────────── */}
      <div className="chart-grid">
        <div className="card">
          <h3 className="card-title">Events — Last 24 Hours</h3>
          {timeline && timeline.length > 0 ? (
            <TimelineChart data={timeline} />
          ) : (
            <div className="empty-state">No timeline data yet.</div>
          )}
        </div>
        <div className="card">
          <h3 className="card-title">Event Type Distribution</h3>
          {typeCounts && typeCounts.length > 0 ? (
            <TypePieChart data={typeCounts} height={280} />
          ) : (
            <div className="empty-state">No type data yet.</div>
          )}
        </div>
      </div>

      {/* ── Recent Events ────────────────────────────── */}
      <div className="card">
        <h3 className="card-title">Recent Events</h3>
        <EventTable events={eventsRes?.events ?? []} />
      </div>
    </div>
  );
};

export default Dashboard;
