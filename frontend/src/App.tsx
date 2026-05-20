import { useEffect } from "react";
import { Hud } from "./hud/Hud";
import { MapView } from "./map/Map";
import { connect, disconnect } from "./net/ws";
import { useGameStore } from "./state/store";

export default function App() {
  const connection = useGameStore((s) => s.connection);
  const lastError = useGameStore((s) => s.lastError);

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
        </p>
      </header>

      {lastError ? (
        <pre className="panel error">
          {JSON.stringify(lastError, null, 2)}
        </pre>
      ) : null}

      <div className="layout">
        <Hud />
        <div className="layout-main">
          <MapView />
        </div>
      </div>
    </main>
  );
}
