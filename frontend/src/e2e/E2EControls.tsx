import { send } from "../net/ws";
import { MsgCastleBuild, MsgMoveRequest } from "../proto/types";
import { useGameStore, useServerNow } from "../state/store";

/** DOM shortcuts for Playwright — only rendered when VITE_E2E=1. */
export function E2EControls() {
  const bootstrap = useGameStore((s) => s.bootstrap);
  const inFlight = useGameStore((s) => s.inFlight);
  const hero = useGameStore((s) => s.hero);
  const now = useServerNow(250);

  if (!bootstrap) return null;

  const move = (targetNodeId: number) => {
    send(MsgMoveRequest, {
      heroId: bootstrap.heroId,
      targetNodeId,
    });
  };
  const build = (buildingCode: string) => {
    send(MsgCastleBuild, {
      castleId: bootstrap.castleId,
      buildingCode,
    });
  };

  return (
    <div className="e2e-controls" data-testid="e2e-controls">
      <button
        type="button"
        data-testid="e2e-build-defense"
        onClick={() => build("defense")}
      >
        E2E: Build defense
      </button>
      <button
        type="button"
        data-testid="e2e-move-2"
        disabled={!!inFlight || (!!hero?.respawnUntil && now < hero.respawnUntil)}
        onClick={() => move(2)}
      >
        E2E: Move to Crossroads (2)
      </button>
      <button
        type="button"
        data-testid="e2e-move-5"
        disabled={!!inFlight || (!!hero?.respawnUntil && now < hero.respawnUntil)}
        onClick={() => move(5)}
      >
        E2E: Move to Bandit Camp (5)
      </button>
    </div>
  );
}
