import React, { useEffect, useState } from 'react';
import { Wifi, WifiOff, RefreshCw } from 'lucide-react';
import { checkHealth } from '../services/api';

interface HeaderProps {
  title: string;
}

const Header: React.FC<HeaderProps> = ({ title }) => {
  const [healthy, setHealthy] = useState<boolean | null>(null);

  const probe = () => {
    checkHealth().then(setHealthy);
  };

  useEffect(() => {
    probe();
    const id = setInterval(probe, 30_000);
    return () => clearInterval(id);
  }, []);

  return (
    <header className="topbar">
      <h1 className="topbar-title">{title}</h1>
      <div className="topbar-actions">
        <button className="icon-btn" onClick={probe} title="Refresh health">
          <RefreshCw size={16} />
        </button>
        <span
          className={`health-badge ${healthy === null ? 'health--loading' : healthy ? 'health--ok' : 'health--err'}`}
          title={healthy ? 'API healthy' : 'API unreachable'}
        >
          {healthy ? <Wifi size={14} /> : <WifiOff size={14} />}
          {healthy === null ? 'Checking…' : healthy ? 'Healthy' : 'Unhealthy'}
        </span>
      </div>
    </header>
  );
};

export default Header;