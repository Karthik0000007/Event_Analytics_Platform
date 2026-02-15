import React from 'react';
import { Search, X } from 'lucide-react';
import type { EventFilters } from '../types';

interface FiltersProps {
  filters: EventFilters;
  onChange: (filters: EventFilters) => void;
  typeOptions?: string[];
}

const Filters: React.FC<FiltersProps> = ({ filters, onChange, typeOptions }) => {
  const update = (patch: Partial<EventFilters>) =>
    onChange({ ...filters, ...patch, offset: 0 });

  const clear = () => onChange({ limit: filters.limit ?? 50, offset: 0 });

  const hasFilters = !!(filters.type || filters.from || filters.to);

  return (
    <div className="filters-bar">
      {/* Type selector */}
      <div className="filter-group">
        <label>Type</label>
        <select
          value={filters.type ?? ''}
          onChange={(e) => update({ type: e.target.value || undefined })}
        >
          <option value="">All types</option>
          {typeOptions?.map((t) => (
            <option key={t} value={t}>
              {t}
            </option>
          ))}
        </select>
      </div>

      {/* Date range */}
      <div className="filter-group">
        <label>From</label>
        <input
          type="datetime-local"
          value={filters.from?.slice(0, 16) ?? ''}
          onChange={(e) =>
            update({ from: e.target.value ? new Date(e.target.value).toISOString() : undefined })
          }
        />
      </div>

      <div className="filter-group">
        <label>To</label>
        <input
          type="datetime-local"
          value={filters.to?.slice(0, 16) ?? ''}
          onChange={(e) =>
            update({ to: e.target.value ? new Date(e.target.value).toISOString() : undefined })
          }
        />
      </div>

      {/* Actions */}
      <div className="filter-actions">
        {hasFilters && (
          <button className="btn btn--ghost" onClick={clear}>
            <X size={14} /> Clear
          </button>
        )}
        <button className="btn btn--primary" onClick={() => onChange(filters)}>
          <Search size={14} /> Search
        </button>
      </div>
    </div>
  );
};

export default Filters;
