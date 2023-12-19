import { Navigate, useLocation } from "react-router-dom";
import { useUser } from "./UserContext";

function RequireAuth({ children }: { children: JSX.Element }) {
  const auth = useUser();
  const location = useLocation();

  if (auth.loading) {
    return null;
  }

  if (!auth.user) {
    return <Navigate to="/login" state={{ from: location }} />;
  }

  return children;
}

export default RequireAuth;