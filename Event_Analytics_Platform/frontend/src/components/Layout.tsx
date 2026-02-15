import React from 'react';
import { Outlet } from 'react-router-dom';
import Sidebar from './Sidebar';
import Header from './Header';

interface LayoutProps {
  title?: string;
}

const Layout: React.FC<LayoutProps> = ({ title = 'Event Analytics Platform' }) => (
  <div className="app-layout">
    <Sidebar />
    <div className="app-main">
      <Header title={title} />
      <main className="app-content">
        <Outlet />
      </main>
    </div>
  </div>
);

export default Layout;
