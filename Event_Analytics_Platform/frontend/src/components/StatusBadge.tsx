import React from 'react';

interface StatusBadgeProps {
  status: 'success' | 'error' | 'warning' | 'info';
  label: string;
}

const StatusBadge: React.FC<StatusBadgeProps> = ({ status, label }) => (
  <span className={`status-badge status-badge--${status}`}>{label}</span>
);

export default StatusBadge;
