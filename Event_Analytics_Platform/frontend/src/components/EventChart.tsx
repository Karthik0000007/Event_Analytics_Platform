import React from 'react';
import {
  AreaChart,
  Area,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  ResponsiveContainer,
  PieChart,
  Pie,
  Cell,
  Legend,
  BarChart,
  Bar,
} from 'recharts';
import { format, parseISO } from 'date-fns';
import type { TimelinePoint, TypeCount } from '../types';

/* ── Color palette ──────────────────────────────────── */

const COLORS = [
  '#6366f1', '#22d3ee', '#f59e0b', '#ef4444', '#10b981',
  '#8b5cf6', '#ec4899', '#14b8a6', '#f97316', '#3b82f6',
];

/* ── Timeline chart ─────────────────────────────────── */

interface TimelineChartProps {
  data: TimelinePoint[];
  height?: number;
}

export const TimelineChart: React.FC<TimelineChartProps> = ({ data, height = 300 }) => {
  const formatted = data.map((d) => ({
    ...d,
    label: format(parseISO(d.bucket), 'HH:mm'),
  }));

  return (
    <ResponsiveContainer width="100%" height={height}>
      <AreaChart data={formatted}>
        <defs>
          <linearGradient id="colorCount" x1="0" y1="0" x2="0" y2="1">
            <stop offset="5%" stopColor="#6366f1" stopOpacity={0.3} />
            <stop offset="95%" stopColor="#6366f1" stopOpacity={0} />
          </linearGradient>
        </defs>
        <CartesianGrid strokeDasharray="3 3" stroke="var(--border)" />
        <XAxis dataKey="label" stroke="var(--text-dim)" fontSize={12} />
        <YAxis stroke="var(--text-dim)" fontSize={12} />
        <Tooltip
          contentStyle={{
            background: 'var(--surface)',
            border: '1px solid var(--border)',
            borderRadius: 8,
            color: 'var(--text)',
          }}
        />
        <Area
          type="monotone"
          dataKey="count"
          stroke="#6366f1"
          fillOpacity={1}
          fill="url(#colorCount)"
          strokeWidth={2}
        />
      </AreaChart>
    </ResponsiveContainer>
  );
};

/* ── Type distribution (Pie) ────────────────────────── */

interface TypePieChartProps {
  data: TypeCount[];
  height?: number;
}

export const TypePieChart: React.FC<TypePieChartProps> = ({ data, height = 300 }) => (
  <ResponsiveContainer width="100%" height={height}>
    <PieChart>
      <Pie
        data={data}
        dataKey="count"
        nameKey="event_type"
        cx="50%"
        cy="50%"
        outerRadius={100}
        label={({ event_type, percent }) =>
          `${event_type} (${(percent * 100).toFixed(0)}%)`
        }
      >
        {data.map((_, i) => (
          <Cell key={i} fill={COLORS[i % COLORS.length]} />
        ))}
      </Pie>
      <Tooltip
        contentStyle={{
          background: 'var(--surface)',
          border: '1px solid var(--border)',
          borderRadius: 8,
          color: 'var(--text)',
        }}
      />
      <Legend />
    </PieChart>
  </ResponsiveContainer>
);

/* ── Type distribution (Bar) ────────────────────────── */

interface TypeBarChartProps {
  data: TypeCount[];
  height?: number;
}

export const TypeBarChart: React.FC<TypeBarChartProps> = ({ data, height = 300 }) => (
  <ResponsiveContainer width="100%" height={height}>
    <BarChart data={data} layout="vertical" margin={{ left: 80 }}>
      <CartesianGrid strokeDasharray="3 3" stroke="var(--border)" />
      <XAxis type="number" stroke="var(--text-dim)" fontSize={12} />
      <YAxis
        type="category"
        dataKey="event_type"
        stroke="var(--text-dim)"
        fontSize={12}
      />
      <Tooltip
        contentStyle={{
          background: 'var(--surface)',
          border: '1px solid var(--border)',
          borderRadius: 8,
          color: 'var(--text)',
        }}
      />
      <Bar dataKey="count" radius={[0, 6, 6, 0]}>
        {data.map((_, i) => (
          <Cell key={i} fill={COLORS[i % COLORS.length]} />
        ))}
      </Bar>
    </BarChart>
  </ResponsiveContainer>
);
