import type { LoaderFunctionArgs, MetaArgs } from "react-router";
import { Form, NavLink, Outlet, useLoaderData } from "react-router";
import { ThemeProvider, useTheme } from "~/components/theme";

import { Sun, Moon } from "lucide-react";
import { requireProfile, requireUserId } from "~/session.server";
import { useEffect, useRef, useState } from "react";
import { UserProvider, useUser } from "~/context/user";

export function meta({}: MetaArgs) {
  return [
    { title: "New React Router App" },
    { name: "description", content: "Welcome to React Router!" },
  ];
}

export async function loader({ request }: LoaderFunctionArgs) {
  const userId = await requireUserId(request);
  const profile = await requireProfile(request);
  return { userId, profile };
}

type Tab = "chats" | "contacts" | "settings";

function ToggleThemeButton() {
  const { theme, toggleTheme } = useTheme();

  return (
    <button onClick={() => toggleTheme()} className="">
      {theme === "light" ? <Sun /> : <Moon />}
    </button>
  );
}

type UserBadgeProps = {
  userId?: string;
};

function UserBadge({ userId }: UserBadgeProps) {
  const { profile } = useUser();
  const [open, setOpen] = useState(false);

  const containerRef = useRef<HTMLDivElement>(null);


  // Close dropdown if click is outside
  useEffect(() => {
    function handleClickOutside(event: MouseEvent) {
      if (containerRef.current && !containerRef.current.contains(event.target as Node)) {
        setOpen(false);
      }
    }

    document.addEventListener("mousedown", handleClickOutside);
    return () => {
      document.removeEventListener("mousedown", handleClickOutside);
    };
  }, []);


  return (
    <div className="user-badge-container">
      {/* Badge button */}
      <button
        onClick={() => setOpen((prev) => !prev)}
        className="user-badge-button"
      >
        <span>{userId}</span>
        <span className={`arrow ${open ? "open" : ""}`}>{profile?.name}</span>
      </button>

      {/* Dropdown */}
      {open && (
        <div className="user-badge-dropdown" ref={containerRef}>
          <a href="/profile" className="dropdown-item">
            Profile
          </a>
          <Form method="post" action="/logout">
            <button type="submit" className="dropdown-item logout">
              Logout
            </button>
          </Form>
        </div>
      )}
    </div>
  );
}



export function TopNav() {

  const tabs = [
    { name: "Chats", path: "chats" },
    { name: "Contacts", path: "contacts" },
    { name: "Settings", path: "settings" },
  ];

  return(
    <nav className="topnav">
      {/* left side  */}
      <div className="topnav-tab-container">
        {tabs.map((tab) => (
          <NavLink key={tab.path} to={tab.path} className={({ isActive }) => `topnav-tab ${isActive ? "active" : ""}`}>
            {tab.name}
          </NavLink>
        ))}
      </div>
      {/* right side */}
      <div className="flex items-center space-x-4">
        <UserBadge/>
        <ToggleThemeButton/>
      </div>

  </nav>
  )
}


export default function Dashboard() {
  const { userId, profile } = useLoaderData<typeof loader>();

  return (
    <ThemeProvider>
      <UserProvider profile={profile} user={{ id: userId }}>
        {/* tabs */}
        <TopNav />
        {/* content */}
        <Outlet/>
      </UserProvider>
    </ThemeProvider>
  );
}

