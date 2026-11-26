import React from 'react';
import { NavLink } from 'react-router-dom';
import {
  LayoutDashboard,
  List,
  BarChart3,
  AlertTriangle,
  Activity,
} from 'lucide-react';

const links = [
  { to: '/', label: 'Dashboard', icon: <LayoutDashboard size={18} /> },
  { to: '/events', label: 'Events', icon: <List size={18} /> },
  { to: '/analytics', label: 'Analytics', icon: <BarChart3 size={18} /> },
  { to: '/dlq', label: 'Dead Letters', icon: <AlertTriangle size={18} /> },
];

const Sidebar: React.FC = () => (
  <aside className="sidebar">
    <div className="sidebar-brand">
      <Activity size={24} />
      <span>Event Analytics</span>
    </div>
    <nav className="sidebar-nav">
      {links.map((link) => (
        <NavLink
          key={link.to}
          to={link.to}
          end={link.to === '/'}
          className={({ isActive }) =>
            `sidebar-link ${isActive ? 'sidebar-link--active' : ''}`
          }
        >
          {link.icon}
          <span>{link.label}</span>
        </NavLink>
      ))}
    </nav>
    <div className="sidebar-footer">
      <span className="sidebar-version">v1.0.0</span>
    </div>
  </aside>
);

export default Sidebar;
