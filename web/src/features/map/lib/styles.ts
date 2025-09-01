// Map container styles for full-screen display
export const mapContainerStyles = {
  position: "fixed" as const,
  top: 0,
  left: 0,
  right: 0,
  bottom: 0,
  width: "100vw",
  height: "100vh",
  overflow: "hidden",
  zIndex: 0, // Base layer - everything else appears above
};

// Control panel styles
export const controlPanelStyles = {
  position: "absolute" as const,
  top: "1rem",
  right: "1rem",
  zIndex: 10,
  display: "flex",
  flexDirection: "column" as const,
  gap: "0.5rem",
};

// Control button styles
export const controlButtonStyles = {
  width: "40px",
  height: "40px",
  backgroundColor: "white",
  border: "2px solid #ddd",
  borderRadius: "4px",
  cursor: "pointer",
  display: "flex",
  alignItems: "center",
  justifyContent: "center",
  fontSize: "20px",
  transition: "all 0.2s",
  "&:hover": {
    backgroundColor: "#f5f5f5",
    borderColor: "#999",
  },
};
