import { useEffect } from "react";
import { E2EControls } from "./e2e/E2EControls";
import { Hud } from "./hud/Hud";
import { MapView } from "./map/Map";
import { connect, disconnect } from "./net/ws";
import { useGameStore } from "./state/store";

const e2eMode = import.meta.env.VITE_E2E === "1";

function statusModifier(status: string): string {
  switch (status) {
    case "connected":
      return "is-open";
    case "connecting":
      return "is-connecting";
    case "disconnected":
      return "is-closed";
    case "error":
      return "is-error";
    default:
      return "";
  }
}

export default function App() {
  const connection = useGameStore((s) => s.connection);
  const lastError = useGameStore((s) => s.lastError);

  useEffect(() => {
    connect();
    return () => disconnect();
  }, []);

  return (
    <main className="app">
      <header className="app-banner">
        <div className="app-banner-title">
          <h1>Herogame</h1>
          <span className="subtitle">— Realm of Ardent Flame —</span>
        </div>
        <p
          className={`connection-status ${statusModifier(connection.status)}`}
          data-testid="connection-status"
        >
          <span className="connection-dot" aria-hidden="true" />
          <span>
            Aether link: <strong>{connection.status}</strong>
            {connection.error ? ` — ${connection.error}` : null}
          </span>
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
