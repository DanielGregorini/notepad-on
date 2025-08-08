'use client';

import { useEffect, useRef, useState, type ChangeEvent } from 'react';
import { useParams } from 'next/navigation';

type ServerMsg =
  | { type: 'init'; text: string }
  | { type: 'update'; text: string; clientId?: string };

export default function PageSlug() {
  const params = useParams();
  const raw = params?.slug as string | string[] | undefined;
  const slug = Array.isArray(raw) ? raw[0] : raw;

  const [isConnected, setIsConnected] = useState(false);
  
  const [text, setText] = useState('');

  const wsRef = useRef<WebSocket | null>(null);
  const debounceRef = useRef<number | null>(null);
  const clientIdRef = useRef<string>(
    (typeof crypto !== 'undefined' && 'randomUUID' in crypto
      ? (crypto as any).randomUUID()
      : `${Date.now()}-${Math.random()}`) as string
  );

  useEffect(() => {
    if (!slug) return;

    // sempre feche conexão anterior (se houver)
    wsRef.current?.close();

    const url = `ws://localhost:8888/page/${encodeURIComponent(slug)}`;
    const ws = new WebSocket(url);
    wsRef.current = ws;

    ws.onopen = () => {
      setIsConnected(true);
      // opcional: identifique o cliente
      try {
        ws.send(JSON.stringify({ type: 'hello', clientId: clientIdRef.current }));
      } catch {}
    };

    ws.onmessage = (ev: MessageEvent) => {
      try {
        const msg = JSON.parse(String(ev.data)) as ServerMsg;

        if (msg.type === 'init') {
          setText(msg.text ?? '');
          return;
        }

        if (msg.type === 'update') {
          // ignore eco próprio
          if (msg.clientId && msg.clientId === clientIdRef.current) return;
          setText((curr) => (curr === msg.text ? curr : msg.text ?? ''));
          return;
        }
      } catch {
        // mensagens não-JSON são ignoradas
      }
    };

    ws.onerror = () => {
      setIsConnected(false);
    };

    ws.onclose = () => {
      setIsConnected(false);
      if (wsRef.current === ws) wsRef.current = null;
    };

    return () => {
      // cleanup
      setIsConnected(false);
      if (debounceRef.current) {
        window.clearTimeout(debounceRef.current);
        debounceRef.current = null;
      }
      ws.close();
    };
  }, [slug]);

  if (!slug) return null;

  function sendUpdate(nextText: string) {
    const ws = wsRef.current;
    if (!ws || ws.readyState !== WebSocket.OPEN) return;

    const payload = {
      type: 'update',
      text: nextText,
      clientId: clientIdRef.current,
    } as const;

    try {
      ws.send(JSON.stringify(payload));
    } catch {}
  }

  function scheduleSend(nextText: string) {
    if (debounceRef.current) window.clearTimeout(debounceRef.current);
    debounceRef.current = window.setTimeout(() => {
      sendUpdate(nextText);
      debounceRef.current = null;
    }, 120);
  }

  function handleChange(e: ChangeEvent<HTMLTextAreaElement>) {
    const next = e.target.value;
    setText(next);
    scheduleSend(next);
  }

  return (
    <div className="p-4 space-y-3">
      <div className="text-sm text-gray-600">slug: {slug}</div>

      <div
        className={`text-sm ${
          isConnected ? 'text-green-600' : 'text-red-600'
        }`}
      >
        {isConnected ? 'conectado' : 'desconectado'}
      </div>

      <div className="flex justify-center w-full max-w-xl">
        <textarea
          className="w-full h-64 border border-gray-300 rounded p-2 outline-none"
          value={text}
          onChange={handleChange}
          spellCheck={false}
        />
      </div>
    </div>
  );
}
