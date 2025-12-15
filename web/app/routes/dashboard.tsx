import type { ActionFunctionArgs, LoaderFunctionArgs, MetaArgs, ShouldRevalidateFunctionArgs } from "react-router";
import { Form, NavLink, Outlet, redirect, useActionData, useLoaderData, useNavigation } from "react-router";
import { useTheme } from "~/components/theme";

import { Sun, Moon } from "lucide-react";
import { getProfile, requireUserId } from "~/session.server";
import { useEffect, useRef, useState, type ReactNode } from "react";
import { UserProvider, useUser } from "~/context/user";
import { getWsToken, requireAccessToken } from "~/utils/auth.server";
import { useChatWebSocket } from "~/hooks/useChatWebsocket";

export function meta({}: MetaArgs) {
  return [
    { title: "New React Router App" },
    { name: "description", content: "Welcome to React Router!" },
  ];
}

export async function loader({ request }: LoaderFunctionArgs) {
  const userId = await requireUserId(request);
  const profile = await getProfile(request);
  const wsToken = await getWsToken(request);
  return { userId, profile, wsToken };
}

export async function action({ request }: ActionFunctionArgs) {
  const formData = await request.formData();
  const intent = formData.get("_action");

  if (intent === "create-profile") {
    const name = formData.get("name");
    
    if (typeof name !== "string" || name.trim().length === 0) {
      return Response.json({ ok: false, error: "Name is required" }, { status: 400 });
    }

    try {
      const accessToken = await requireAccessToken(request);
      const res = await fetch(`${process.env.API_URL}/profile`, {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
          "Authorization": `Bearer ${accessToken}`,
        },
        body: JSON.stringify({ name: name.trim() }),
      });

      if (!res.ok) {
        const errorData = await res.json().catch(() => ({ error: "Failed to create profile" }));
        return Response.json({ ok: false, error: errorData.error || "Failed to create profile" }, { status: res.status });
      }

      // Redirect to reload the page and fetch the new profile
      return redirect("/dashboard");
    } catch (error: any) {
      return Response.json({ ok: false, error: error.message || "Failed to create profile" }, { status: 500 });
    }
  }

  return Response.json({ ok: false, error: "Unknown action" }, { status: 400 });
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


function CreateProfilePage() {
  const navigation = useNavigation();
  const actionData = useActionData<{ ok: boolean; error?: string }>();
  const isSubmitting = navigation.state === "submitting";
  const [name, setName] = useState("");

  return (
    <div className="flex h-full items-center justify-center bg-slate-50 px-4 dark:bg-slate-950">
      <div className="card w-full max-w-md">
        <h1 className="text-2xl font-bold mb-2 text-slate-900 dark:text-slate-100">Create Your Profile</h1>
        <p className="lead mb-6 text-slate-600 dark:text-slate-400">
          Welcome! Please create your profile to get started.
        </p>
        
        {actionData?.error && (
          <div className="mb-4 rounded-lg border border-red-200 bg-red-50 p-4 text-sm text-red-700 dark:border-red-800 dark:bg-red-900/20 dark:text-red-300">
            {actionData.error}
          </div>
        )}

        <Form method="post" className="space-y-4" noValidate>
          <input type="hidden" name="_action" value="create-profile" />
          <div>
            <label htmlFor="name" className="text-slate-700 dark:text-slate-300">
              Display Name
            </label>
            <input
              id="name"
              name="name"
              type="text"
              value={name}
              onChange={(e) => setName(e.target.value)}
              required
              placeholder="Enter your display name"
              className="text-slate-900 dark:text-slate-100"
              disabled={isSubmitting}
              autoFocus
            />
            <span className="text-xs text-slate-500 dark:text-slate-400 mt-1 block">
              This is how other people will see you in chats and contacts.
            </span>
          </div>

          <button
            type="submit"
            className="w-full"
            disabled={isSubmitting || name.trim().length === 0}
          >
            {isSubmitting ? "Creating..." : "Create Profile"}
          </button>
        </Form>
      </div>
    </div>
  );
}

export default function Dashboard() {
  const { userId, profile, wsToken } = useLoaderData<typeof loader>();

  // Show create profile page if profile doesn't exist
  if (!profile) {
    return (
      <DashboardSocketBoundary token={wsToken}>
        <UserProvider profile={null} user={{ id: userId }}>
          <div className="flex flex-col h-screen">
            <CreateProfilePage />
          </div>
        </UserProvider>
      </DashboardSocketBoundary>
    );
  }

  return (
    <DashboardSocketBoundary token={wsToken}>
      <UserProvider profile={profile} user={{ id: userId }}>
        <div className="flex flex-col h-screen">
          {/* tabs */}
          <TopNav />
          {/* content */}
          <div className="flex-1 min-h-0">
            <Outlet/>
          </div>
        </div>
      </UserProvider>
    </DashboardSocketBoundary>
  );
}

function DashboardSocketBoundary({ token, children }: { token: string; children: ReactNode }) {
  useChatWebSocket(token);
  return <>{children}</>;
}
