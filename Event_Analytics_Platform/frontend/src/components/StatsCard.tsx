import React from 'react';
import type { StatsCardProps } from '../types';

const StatsCard: React.FC<StatsCardProps> = ({ title, value, icon, trend, color }) => (
  <div className="stats-card" style={{ borderTopColor: color ?? 'var(--accent)' }}>
    <div className="stats-card-header">
      <span className="stats-card-icon">{icon}</span>
      <span className="stats-card-title">{title}</span>
    </div>
    <div className="stats-card-value">{value}</div>
    {trend && <div className="stats-card-trend">{trend}</div>}
  </div>
);

export default StatsCard;
