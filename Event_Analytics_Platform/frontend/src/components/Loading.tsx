import React from 'react';
import { Loader2 } from 'lucide-react';

interface LoadingProps {
  message?: string;
}

const Loading: React.FC<LoadingProps> = ({ message = 'Loading…' }) => (
  <div className="loading-container">
    <Loader2 className="spin" size={32} />
    <span>{message}</span>
  </div>
);

export default Loading;
