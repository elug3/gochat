import type { LoaderFunctionArgs, MetaArgs, ShouldRevalidateFunctionArgs } from "react-router";
import { Form, NavLink, Outlet, useLoaderData } from "react-router";
import { useTheme } from "~/components/theme";

import { Sun, Moon } from "lucide-react";
import { requireProfile, requireUserId } from "~/session.server";
import { useEffect, useRef, useState, type ReactNode } from "react";
import { UserProvider, useUser } from "~/context/user";
import { getWsToken } from "~/utils/auth.server";
import { useChatWebSocket } from "~/hooks/useChatWebsocket";

export function meta({}: MetaArgs) {
  return [
    { title: "New React Router App" },
    { name: "description", content: "Welcome to React Router!" },
  ];
}

export async function loader({ request }: LoaderFunctionArgs) {
  const userId = await requireUserId(request);
  const profile = await requireProfile(request);
  const wsToken = await getWsToken(request);
  return { userId, profile, wsToken };
}

export function shouldRevalidate({ currentUrl, nextUrl, formAction }: ShouldRevalidateFunctionArgs) {
  if (formAction) {
    const actionPath = formAction.startsWith("http")
      ? new URL(formAction).pathname
      : formAction;
    // Keep existing socket token when handling forms inside the dashboard (e.g., chat send)
    if (actionPath.startsWith("/dashboard")) {
      return false;
    }
    return true;
  }
  const stayingInDashboard =
    currentUrl?.pathname.startsWith("/dashboard") &&
    nextUrl?.pathname.startsWith("/dashboard");
  return !stayingInDashboard;
}

type Tab = "chats" | "contacts" | "settings";

function ToggleThemeButton() {
  const { theme, toggleTheme } = useTheme();

  return (
    <button 
      onClick={() => toggleTheme()} 
      className="inline-flex items-center justify-center w-9 h-9 rounded-lg border border-slate-200 dark:border-slate-800 bg-white dark:bg-slate-900 text-slate-600 dark:text-slate-300 hover:bg-slate-100 dark:hover:bg-slate-800 transition-colors"
      aria-label={`Switch to ${theme === "light" ? "dark" : "light"} mode`}
    >
      {theme === "light" ? <Sun className="w-4 h-4" /> : <Moon className="w-4 h-4" />}
    </button>
  );
}

function UserBadge() {
  const { profile, user } = useUser();
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
        <span>{user?.id?.toString() || ""}</span>
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
  const { userId, profile, wsToken } = useLoaderData<typeof loader>();

  return (
    <DashboardSocketBoundary token={wsToken}>
      <UserProvider profile={profile} user={{ id: userId }}>
        {/* tabs */}
        <TopNav />
        {/* content */}
        <Outlet/>
      </UserProvider>
    </DashboardSocketBoundary>
  );
}

function DashboardSocketBoundary({ token, children }: { token: string; children: ReactNode }) {
  useChatWebSocket(token);
  return <>{children}</>;
}
