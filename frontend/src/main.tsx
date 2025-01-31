import React from "react";
import ReactDOM from "react-dom/client";
import App from "src/App.tsx";
import "src/fonts.css";
import "src/index.css";
import { isHttpMode } from "src/utils/isHttpMode";

// redirect hash router links to browser router links
// TODO: remove after 2026-01-01
if (isHttpMode() && window.location.href.indexOf("/#/") > -1) {
  window.location.href = window.location.href.replace("/#/", "/");
} else {
  ReactDOM.createRoot(document.getElementById("root")!).render(
    <React.StrictMode>
      <App />
    </React.StrictMode>
  );
}
