import React from "react";
import type { MetaArgs } from "react-router";
import { NavLink, Outlet } from "react-router";

export function meta({}: MetaArgs) {
  return [
    { title: "New React Router App" },
    { name: "description", content: "Welcome to React Router!" },
  ];
}


type Tab = "chats" | "contacts" | "settings";

export default function Dashboard() {
  const tabs = [
    { name: "Chats", path: "chats" },
    { name: "Contacts", path: "contacts" },
    { name: "Settings", path: "settings" },
  ];

  return (
    <div className="flex flex-col h-full">
      {/* tabs */}
      <nav>
        {tabs.map((tab) => (
          <NavLink key={tab.path} to={tab.path} className={({ isActive }) => isActive ? "font-bold" : ""}>
            {tab.name}
          </NavLink>
        ))}
      </nav>
      {/* content */}
      <main className="flex-1 overflow-hidden h-full">
        <Outlet />
      </main>
      
    </div>
  )
}
