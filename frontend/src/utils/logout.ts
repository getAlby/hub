export const logout = async () => {
  try {
    await fetch("/api/logout");
    window.location.href = "/";
  } catch (error) {
    // TODO: Handle failure
    console.error("Error during logout:", error);
  }
};
