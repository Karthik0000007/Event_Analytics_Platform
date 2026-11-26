import React, { useCallback, useEffect, useState } from 'react';

export interface ToastMessage {
  kind: 'success' | 'error';
  text: string;
}

/** Small hook: show() a message, it hides itself after a few seconds. */
export function useToast(durationMs = 4000) {
  const [toast, setToast] = useState<ToastMessage | null>(null);

  useEffect(() => {
    if (!toast) return;
    const id = setTimeout(() => setToast(null), durationMs);
    return () => clearTimeout(id);
  }, [toast, durationMs]);

  const show = useCallback((t: ToastMessage) => setToast(t), []);
  return { toast, show, hide: () => setToast(null) };
}

interface ToastProps {
  toast: ToastMessage | null;
  onClose: () => void;
}

const Toast: React.FC<ToastProps> = ({ toast, onClose }) => {
  if (!toast) return null;
  return (
    <div className={`toast toast--${toast.kind}`} role="status" onClick={onClose}>
      {toast.text}
    </div>
  );
};

export default Toast;
