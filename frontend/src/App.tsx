import { useEffect } from "react";
import { E2EControls } from "./e2e/E2EControls";
import { Hud } from "./hud/Hud";
import { MapView } from "./map/Map";
import { connect, disconnect } from "./net/ws";
import { useGameStore } from "./state/store";

const e2eMode = import.meta.env.VITE_E2E === "1";

export default function App() {
  const connection = useGameStore((s) => s.connection);
  const lastError = useGameStore((s) => s.lastError);

  useEffect(() => {
    connect();
    return () => disconnect();
  }, []);

  return (
    <main className="app">
      <header className="app-header">
        <h1>herogame</h1>
        <p data-testid="connection-status">
          Connection: <strong>{connection.status}</strong>
          {connection.error ? ` — ${connection.error}` : null}
        </p>
      </header>

      {lastError ? (
        <pre className="panel error">
          {JSON.stringify(lastError, null, 2)}
        </pre>
      ) : null}

      {e2eMode ? <E2EControls /> : null}

      <div className="layout">
        <Hud />
        <div className="layout-main">
          <MapView />
        </div>
      </div>
    </main>
  );
}
