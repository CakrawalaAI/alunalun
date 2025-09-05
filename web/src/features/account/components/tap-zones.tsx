import { useCallback } from "react";
import { useOverlayStore } from "../store/overlayStore";

/**
 * Tap Zones Component
 * 
 * Implements the center/edge tap zones for overlay toggle functionality.
 * Based on UI specification:
 * - Center 30% area: tap to hide overlay
 * - Edge areas: tap to show overlay
 * 
 * This component creates invisible click zones overlaid on the map.
 * It's designed to not interfere with other interactive elements.
 */
export function TapZones() {
  const { isVisible, setVisible } = useOverlayStore();

  const handleTap = useCallback((event: React.MouseEvent, zone: 'center' | 'edge') => {
    // Don't handle if clicking on interactive elements
    const target = event.target as HTMLElement;
    if (target.closest('[data-interactive]') || 
        target.closest('button') || 
        target.closest('[role="button"]') ||
        target.closest('[data-ui-overlay] > *:not([data-tap-zones])')) {
      return;
    }

    // Prevent event from bubbling to map interactions
    event.preventDefault();
    event.stopPropagation();

    if (zone === 'center') {
      // Center tap always hides the overlay
      setVisible(false);
    } else {
      // Edge tap shows the overlay if it's hidden
      if (!isVisible) {
        setVisible(true);
      }
    }
  }, [isVisible, setVisible]);

  const handleCenterTap = useCallback((event: React.MouseEvent) => {
    handleTap(event, 'center');
  }, [handleTap]);

  const handleEdgeTap = useCallback((event: React.MouseEvent) => {
    handleTap(event, 'edge');
  }, [handleTap]);

  return (
    <div 
      data-tap-zones="true" 
      className="fixed inset-0 pointer-events-none"
      style={{
        zIndex: 5, // Below overlay controls but above map
      }}
    >
      {/* Edge zones - show overlay when tapped */}
      <div
        className="absolute inset-0"
        onClick={handleEdgeTap}
        style={{
          // FIXED: Enable pointer events and proper positioning
          pointerEvents: 'auto',
          zIndex: 10,
          width: '100%',
          height: '100%',
          top: 0,
          left: 0,
          // Invisible background for production
          // backgroundColor: 'rgba(255, 0, 0, 0.1)', // Uncomment for debugging
        }}
        aria-label="Map area - tap to show controls"
        data-testid="tap-zone-edge"
      />

      {/* Center zone - hide overlay when tapped */}
      <div
        className="absolute"
        onClick={handleCenterTap}
        style={{
          // FIXED: Enable pointer events and proper positioning
          pointerEvents: 'auto',
          // Center 30% of viewport
          top: '35%',
          left: '35%',
          width: '30%',
          height: '30%',
          zIndex: 11, // Above edge zone
          // Invisible background for production
          // backgroundColor: 'rgba(0, 255, 0, 0.1)', // Uncomment for debugging
        }}
        aria-label="Center area - tap to hide controls"
        data-testid="tap-zone-center"
      />
    </div>
  );
}

/**
 * Hook to toggle tap zones on/off (useful for debugging or disabling the feature)
 */
export function useTapZones() {
  // Could be extended with additional configuration
  return {
    isEnabled: true,
  };
}