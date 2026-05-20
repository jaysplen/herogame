import { useEffect } from "react";
import { connect, disconnect } from "./net/ws";
import { useGameStore, useServerNow } from "./state/store";

export default function App() {
  const connection = useGameStore((s) => s.connection);
  const bootstrap = useGameStore((s) => s.bootstrap);
  const lastError = useGameStore((s) => s.lastError);
  const serverNow = useServerNow();

  useEffect(() => {
    connect(1);
    return () => disconnect();
  }, []);

  return (
    <main className="app">
      <h1>herogame</h1>
      <p>
        Connection: <strong>{connection.status}</strong>
        {connection.error ? ` — ${connection.error}` : null}
      </p>
      <p>Server now (skew-corrected): {new Date(serverNow).toISOString()}</p>
      {lastError ? (
        <pre className="panel error">
          {JSON.stringify(lastError, null, 2)}
        </pre>
      ) : null}
      <section>
        <h2>Bootstrap snapshot</h2>
        {bootstrap ? (
          <pre className="panel">{JSON.stringify(bootstrap, null, 2)}</pre>
        ) : (
          <p className="muted">Waiting for hello.ack…</p>
        )}
      </section>
    </main>
  );
}
