/* eslint-disable react-refresh/only-export-components */
import {
  createContext,
  useContext,
  useEffect,
  useState,
} from "react";

interface UserContextType {
  user: Record<string, string> | null;
  loading: boolean;
  logout: (callback: VoidFunction) => void;
}

const UserContext = createContext({} as UserContextType);

export function UserProvider({ children }: { children: React.ReactNode }) {
  const [user, setUser] = useState<UserContextType["user"]>(null);
  const [loading, setLoading] = useState(true);

  const logout = () => { //callback: VoidFunction param?
    // do an api request and logout
    return;
    // return msg.request("lock").then(() => {
    //   setUserId("");
    //   callback();
    // });
  };

  // Invoked only on on mount.
  useEffect(() => {
    // do an api call to /user with the cookie and set the user
    setUser({
      id: "1234",
    })
    setLoading(false)
  }, []);

  const value = {
    user,
    loading,
    logout
  };

  return (
    <UserContext.Provider value={value}>{children}</UserContext.Provider>
  );
}

export function useUser() {
  return useContext(UserContext);
}
