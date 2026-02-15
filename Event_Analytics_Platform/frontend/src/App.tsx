import React from 'react';
import { BrowserRouter, Routes, Route } from 'react-router-dom';
import Layout from './components/Layout';
import Dashboard from './pages/Dashboard';
import Events from './pages/Events';
import EventDetail from './pages/EventDetail';
import Analytics from './pages/Analytics';
import DLQ from './pages/DLQ';

const App: React.FC = () => (
  <BrowserRouter>
    <Routes>
      <Route element={<Layout />}>
        <Route path="/" element={<Dashboard />} />
        <Route path="/events" element={<Events />} />
        <Route path="/events/:id" element={<EventDetail />} />
        <Route path="/analytics" element={<Analytics />} />
        <Route path="/dlq" element={<DLQ />} />
      </Route>
    </Routes>
  </BrowserRouter>
);

export default App;