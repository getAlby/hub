import React from "react";
import ReactDOM from "react-dom/client";
import App from "src/App.tsx";
import "src/index.css";
import useSetupStore from "src/state/SetupStore";
import { isHttpMode } from "src/utils/isHttpMode";

if (isHttpMode()) {
  captureSetupCode();
}
ReactDOM.createRoot(document.getElementById("root")!).render(
  <React.StrictMode>
    <App />
  </React.StrictMode>
);

// the install script prints a link that includes the setup code
function captureSetupCode() {
  const url = new URL(window.location.href);
  const setupCode = url.searchParams.get("c");
  if (!setupCode) {
    return;
  }
  useSetupStore.getState().setSetupCode(setupCode);
  url.searchParams.delete("c");
  window.history.replaceState(window.history.state, "", url.toString());
}
