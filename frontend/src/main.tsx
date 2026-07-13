import { StrictMode } from "react";
import { createRoot } from "react-dom/client";
import { BrowserRouter } from "react-router-dom";
import App from "./App";
import { ConfirmProvider, ToastProvider } from "./components/common";
import { AuthProvider } from "./services/AuthContext";
import "./styles.css";

createRoot(document.getElementById("root")!).render(
  <StrictMode>
    <BrowserRouter useTransitions={false}>
      <AuthProvider>
        <ConfirmProvider>
          <App />
          <ToastProvider />
        </ConfirmProvider>
      </AuthProvider>
    </BrowserRouter>
  </StrictMode>,
);
