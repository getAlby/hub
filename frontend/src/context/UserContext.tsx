/* eslint-disable react-refresh/only-export-components */
import {
  createContext,
  useContext,
  useEffect,
  useState,
} from "react";
import { InfoResponse, UserInfo } from "../types";

interface UserContextType {
  info: UserInfo | null;
  loading: boolean;
  logout: () => void;
}

const UserContext = createContext({} as UserContextType);

export function UserProvider({ children }: { children: React.ReactNode }) {
  const [info, setInfo] = useState<UserContextType["info"]>(null);
  const [loading, setLoading] = useState(true);

  const logout = async () => {
    try {
      await fetch('/logout');
    } catch (error) {
      // TODO: Handle failure
      console.error('Error during logout:', error);
    }
    setInfo(null);
  };

  const getInfo = async () => {
    try {
      const response = await fetch('/api/info');
      const data: InfoResponse = await response.json();
      setInfo(data);
    } catch (error) {
      console.error('Error getting user info:', error);
    } finally {
      setLoading(false);
    }
  }

  // Invoked only on on mount.
  useEffect(() => {
    getInfo()
  }, []);

  const value = {
    info,
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
