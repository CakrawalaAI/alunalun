import { MapCanvas, ControlPanel, useMapOrchestrator } from "@/features/map";

/**
 * Map block - pure composition with zero business logic.
 * All state and logic is managed by the orchestrator hook.
 * Components are pure UI that receive props.
 */
export function Map() {
  // Single hook provides everything
  const map = useMapOrchestrator();

  // Pure composition - inject logic into UI
  return (
    <div className="relative h-full w-full overflow-hidden">
      <MapCanvas {...map.canvas} />
      <ControlPanel {...map.controls} />
    </div>
  );
}
