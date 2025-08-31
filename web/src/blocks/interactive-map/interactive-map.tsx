import { useEffect } from "react";
import { MapRenderer, useMapOverlays } from "@/features/map";

/**
 * Example of how to compose overlays from external features
 * This demonstrates the pattern but doesn't create actual features
 */
function MapOverlayManager() {
  const { addOverlay, removeOverlay } = useMapOverlays();

  useEffect(() => {
    // Example: External features could add their overlays here
    // const overlayId = addOverlay(
    //   <SomeExternalComponent />,
    //   { position: { lat: -6.2195, lng: 106.8033 }, zIndex: 100 }
    // );

    // Example: Adding a simple notification overlay
    const notificationId = addOverlay(
      <div
        style={{
          position: "absolute",
          top: "100px",
          left: "50%",
          transform: "translateX(-50%)",
          background: "white",
          padding: "12px 24px",
          borderRadius: "8px",
          boxShadow: "0 2px 8px rgba(0,0,0,0.1)",
          pointerEvents: "auto",
          display: "none", // Hidden by default - external features would show when needed
        }}
      >
        Map Ready for Overlays
      </div>,
      { zIndex: 1000 },
    );

    // Cleanup on unmount
    return () => {
      removeOverlay(notificationId);
    };
  }, [addOverlay, removeOverlay]);

  return null;
}

/**
 * Interactive map block that composes the map with overlay support
 * External features can add their own overlays through the overlay system
 */
export function InteractiveMap() {
  return (
    <MapRenderer showControls>
      {/* This component has access to overlay context */}
      <MapOverlayManager />

      {/* External features could add their components here */}
      {/* Example: <PlacePins /> from a places feature */}
      {/* Example: <UserAvatars /> from a social feature */}
      {/* Example: <ChatBubbles /> from a chat feature */}
    </MapRenderer>
  );
}
