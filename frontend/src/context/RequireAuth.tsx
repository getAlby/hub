import { Navigate, useLocation } from "react-router-dom";
import { useUser } from "./UserContext";

function RequireAuth({ children }: { children: JSX.Element }) {
  const { info, loading } = useUser();
  const location = useLocation();

  if (loading) {
    return null;
  }

  if (!info?.user) {
    // TODO: Use the location to redirect back in /alby/auth?c=
    return <Navigate to="/login" state={{ from: location }} />;
  }

  return children;
}

export default RequireAuth;