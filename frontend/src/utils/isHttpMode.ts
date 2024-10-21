export const isHttpMode = () => {
  return (
    window.location.protocol.startsWith("http") &&
    !window.location.hostname.startsWith("wails")
  );
};
