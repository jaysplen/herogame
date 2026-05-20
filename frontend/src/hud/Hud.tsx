import { ArmyPanel } from "./ArmyPanel";
import { CombatModal } from "./CombatModal";
import { Gold } from "./Gold";
import { HeroPanel } from "./HeroPanel";
import { useGameStore } from "../state/store";

export function Hud() {
  const lastCombat = useGameStore((s) => s.lastCombat);
  const combatModalOpen = useGameStore((s) => s.combatModalOpen);
  const dismissCombatModal = useGameStore((s) => s.dismissCombatModal);

  return (
    <>
      <aside className="hud">
        <Gold />
        <HeroPanel />
        <ArmyPanel />
      </aside>
      <CombatModal
        combat={lastCombat}
        open={combatModalOpen}
        onClose={dismissCombatModal}
      />
    </>
  );
}
