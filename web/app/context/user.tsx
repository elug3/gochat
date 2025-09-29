import { createContext, useContext, useState, type PropsWithChildren, type ReactNode } from "react";

interface User {
  id: number;
  name?: string;
  status?: "online" | "offline";
}

interface UserContextType {
  user?: User;
  profile? : Profile;
  setUser?: (user: User | undefined) => void;
  setProfile?: (profile: Profile | undefined) => void;
}

const UserContext = createContext<UserContextType>({});

export function useUser() {
  return useContext(UserContext);
}


interface Props {
    children: ReactNode;
    profile?: Profile;
    user?: User;
}

export function UserProvider({ children, profile, user }: Props) {

  return (
    <UserContext.Provider value={{ profile, user }}>
      {children}
    </UserContext.Provider>
  );
}
