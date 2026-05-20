import { useEffect } from "react";
import { MapView } from "./map/Map";
import { connect, disconnect } from "./net/ws";
import { useGameStore } from "./state/store";

export default function App() {
  const connection = useGameStore((s) => s.connection);
  const bootstrap = useGameStore((s) => s.bootstrap);
  const lastError = useGameStore((s) => s.lastError);
  const hero = useGameStore((s) => s.hero);

  useEffect(() => {
    connect(1);
    return () => disconnect();
  }, []);

  return (
    <main className="app">
      <header className="app-header">
        <h1>herogame</h1>
        <p>
          Connection: <strong>{connection.status}</strong>
          {connection.error ? ` — ${connection.error}` : null}
          {hero ? (
            <>
              {" "}
              · Node <strong>{hero.currentNodeId}</strong> · Army{" "}
              <strong>{hero.armySize}</strong>
            </>
          ) : null}
        </p>
      </header>

      {lastError ? (
        <pre className="panel error">
          {JSON.stringify(lastError, null, 2)}
        </pre>
      ) : null}

      <MapView />

      <details className="bootstrap-details">
        <summary>Bootstrap snapshot</summary>
        {bootstrap ? (
          <pre className="panel">{JSON.stringify(bootstrap, null, 2)}</pre>
        ) : (
          <p className="muted">Waiting for hello.ack…</p>
        )}
      </details>
    </main>
  );
}
