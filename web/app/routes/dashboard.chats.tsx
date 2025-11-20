import { useEffect, useState } from "react";
import { Link, NavLink, Outlet, useLoaderData, useNavigate, type LoaderFunctionArgs, type ShouldRevalidateFunctionArgs } from "react-router";
import { useChatStore } from "~/store/chatStore";
import { fetchChatList } from "~/utils/auth.server";
import { PopupDrawer } from "~/components/PopupDrawer";
import type { Chat } from "~/model";

export type ChatData = {
  chats: Chat[];
}

export async function loader({request}: LoaderFunctionArgs): Promise<{ chats?: Chat[], error?: string }> {
  try {
    const chats = await fetchChatList(request);
    return { chats };
  } catch (err: any) {
    return { error: err.message };
  }
}

export function shouldRevalidate({ formAction }: ShouldRevalidateFunctionArgs) {
  return false;
}

export function Sidebar() {
  const chats = useChatStore((s) => s.chats);
  const wsState = useChatStore((s) => s.wsState);
  const navigate = useNavigate();
  const [isNewChatDrawerOpen, setIsNewChatDrawerOpen] = useState(false);

  const chatList = Object.values(chats || {}).flat();

  const wsBadgeStyles = wsState === "open"
    ? "bg-emerald-100 text-emerald-700 dark:bg-emerald-900/50 dark:text-emerald-300"
    : wsState === "connecting"
      ? "bg-amber-100 text-amber-700 dark:bg-amber-900/50 dark:text-amber-300"
      : "bg-rose-100 text-rose-700 dark:bg-rose-900/50 dark:text-rose-300";


  const handleCreateGroup = () => {
    setIsNewChatDrawerOpen(false);
    navigate('/dashboard/chats/new?type=group');
  };

  const handleCreateDirectChat = () => {
    setIsNewChatDrawerOpen(false);
    navigate('/dashboard/chats/new?type=direct');
  };

  return (
    <aside className="flex h-full flex-col gap-4 bg-slate-50/80 p-4 text-slate-900 dark:bg-slate-950/60 dark:text-slate-100">
      <div className="space-y-3">
        <div className="flex items-center justify-between">
          <div>
            <p className="text-[11px] font-semibold uppercase tracking-[0.2em] text-slate-500/70 dark:text-slate-400/70">Messages</p>
            <h2 className="mt-1 text-xl font-semibold">Chats</h2>
          </div>
          <span className="rounded-full border border-slate-200 px-2 py-0.5 text-xs font-medium text-slate-500 dark:border-slate-800 dark:text-slate-400">
            {chatList.length}
          </span>
        </div>

        <button
          onClick={() => setIsNewChatDrawerOpen(true)}
          className="inline-flex w-full items-center justify-center gap-2 rounded-xl bg-blue-600 px-3 py-2 text-sm font-medium text-white shadow-sm transition hover:bg-blue-700 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-blue-400"
        >
          <svg className="h-4 w-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 4v16m8-8H4" />
          </svg>
          New Chat
        </button>
      </div>

      <PopupDrawer
        isOpen={isNewChatDrawerOpen}
        onClose={() => setIsNewChatDrawerOpen(false)}
        title="Create New Chat"
      >
        <div className="space-y-3">
          <button
            onClick={handleCreateDirectChat}
            className="w-full flex items-center gap-3 p-4 text-left hover:bg-gray-50 dark:hover:bg-gray-700 rounded-lg transition-colors duration-200 group"
          >
            <div className="w-10 h-10 bg-blue-100 dark:bg-blue-900 rounded-lg flex items-center justify-center group-hover:bg-blue-200 dark:group-hover:bg-blue-800 transition-colors">
              <svg className="w-5 h-5 text-blue-600 dark:text-blue-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M8 12h.01M12 12h.01M16 12h.01M21 12c0 4.418-4.03 8-9 8a9.863 9.863 0 01-4.255-.949L3 20l1.395-3.72C3.512 15.042 3 13.574 3 12c0-4.418 4.03-8 9-8s9 3.582 9 8z" />
              </svg>
            </div>
            <div className="flex-1">
              <h3 className="font-medium text-gray-900 dark:text-white">Direct Chat</h3>
              <p className="text-sm text-gray-500 dark:text-gray-400">Start a conversation with someone</p>
            </div>
            <svg className="w-5 h-5 text-gray-400 group-hover:text-gray-600 dark:group-hover:text-gray-300" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 5l7 7-7 7" />
            </svg>
          </button>

          <button
            onClick={handleCreateGroup}
            className="w-full flex items-center gap-3 p-4 text-left hover:bg-gray-50 dark:hover:bg-gray-700 rounded-lg transition-colors duration-200 group"
          >
            <div className="w-10 h-10 bg-green-100 dark:bg-green-900 rounded-lg flex items-center justify-center group-hover:bg-green-200 dark:group-hover:bg-green-800 transition-colors">
              <svg className="w-5 h-5 text-green-600 dark:text-green-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v-2c0-.656-.126-1.283-.356-1.857M7 20H2v-2a3 3 0 015.356-1.857M7 20v-2c0-.656.126-1.283.356-1.857m0 0a5.002 5.002 0 019.288 0M15 7a3 3 0 11-6 0 3 3 0 016 0zm6 3a2 2 0 11-4 0 2 2 0 014 0zM7 10a2 2 0 11-4 0 2 2 0 014 0z" />
              </svg>
            </div>
            <div className="flex-1">
              <h3 className="font-medium text-gray-900 dark:text-white">New Group</h3>
              <p className="text-sm text-gray-500 dark:text-gray-400">Create a group chat with multiple people</p>
            </div>
            <svg className="w-5 h-5 text-gray-400 group-hover:text-gray-600 dark:group-hover:text-gray-300" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 5l7 7-7 7" />
            </svg>
          </button>
        </div>
      </PopupDrawer>

      <nav className="flex-1 overflow-hidden">
        {chatList.length === 0 ? (
          <div className="flex h-full flex-col items-center justify-center gap-3 rounded-xl border border-dashed border-slate-300/70 bg-white/60 p-6 text-center text-sm text-slate-500 shadow-sm dark:border-slate-800 dark:bg-slate-900/60 dark:text-slate-400">
            <svg className="h-8 w-8 text-slate-300 dark:text-slate-700" fill="none" stroke="currentColor" strokeWidth="1.5" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" d="M7 8h10M7 12h6m-3 8a9 9 0 110-18 9 9 0 010 18z" />
            </svg>
            <div>
              <p className="font-medium text-slate-600 dark:text-slate-200">No chats yet</p>
              <p className="mt-1 text-xs text-slate-500 dark:text-slate-400">Create a new chat or add contacts to start messaging.</p>
            </div>
            <Link
              to="/dashboard/contacts"
              className="inline-flex items-center gap-2 rounded-full border border-slate-200 px-3 py-1.5 text-xs font-medium text-slate-600 transition hover:border-blue-500 hover:text-blue-600 dark:border-slate-700 dark:text-slate-300 dark:hover:border-blue-400 dark:hover:text-blue-200"
            >
              View contacts
            </Link>
          </div>
        ) : (
          <div className="h-full overflow-y-auto pr-1">
            <ul className="space-y-1">
              {chatList.map((chat) => (
                <li key={chat.id}>
                  <NavLink
                    to={`${chat.id}`}
                    className={({ isActive }) => {
                      const base = "group flex items-center gap-3 rounded-xl px-3 py-3 text-sm transition";
                      const active = "bg-blue-600 text-white shadow-sm ring-1 ring-blue-500/80";
                      const inactive = "bg-white/60 text-slate-700 hover:bg-slate-100/80 dark:bg-slate-900/60 dark:text-slate-200 dark:hover:bg-slate-900";
                      return `${base} ${isActive ? active : inactive}`;
                    }}
                  >
                    <div className="flex min-w-0 flex-1 flex-col">
                      <div className="flex items-center justify-between gap-2">
                        <span className="truncate font-medium">{chat.name}</span>
                        {chat.unreadCount > 0 && (
                          <span className="inline-flex h-5 min-w-[1.25rem] items-center justify-center rounded-full bg-white/90 px-1 text-xs font-semibold text-blue-600 shadow-sm group-hover:bg-white">
                            {chat.unreadCount}
                          </span>
                        )}
                      </div>
                      <span className="mt-1 line-clamp-1 text-xs text-slate-500 group-hover:text-slate-600 dark:text-slate-400 dark:group-hover:text-slate-300">
                        {chat.lastMessage || "No messages yet."}
                      </span>
                    </div>
                    <svg className="h-4 w-4 flex-none text-slate-300 opacity-0 transition group-hover:opacity-100 dark:text-slate-600" fill="none" stroke="currentColor" strokeWidth="1.5" viewBox="0 0 24 24">
                      <path strokeLinecap="round" strokeLinejoin="round" d="M9 5l7 7-7 7" />
                    </svg>
                  </NavLink>
                </li>
              ))}
            </ul>
          </div>
        )}
      </nav>

      <div className="mt-auto rounded-xl border border-slate-200/70 bg-white/70 px-3 py-3 text-xs shadow-sm dark:border-slate-800 dark:bg-slate-900/60">
        <p className="flex items-center justify-between text-slate-500 dark:text-slate-400">
          <span className="font-medium text-slate-600 dark:text-slate-200">Connection</span>
          <span className={`inline-flex items-center gap-1 rounded-full px-2 py-0.5 text-[11px] font-medium ${wsBadgeStyles}`}>
            <span className="block h-2 w-2 rounded-full bg-current" />
            {wsState}
          </span>
        </p>
        <p className="mt-2 text-[11px] text-slate-400 dark:text-slate-500">WebSocket keeps chats live in real time.</p>
      </div>
    </aside>
  );
}

export default function Chats() {
  const loaderData = useLoaderData<typeof loader>();
  if (loaderData.error) {
    return <div>Error: {loaderData.error}</div>;
  }
  if (!loaderData.chats) {
    return <div>Failed to load chats</div>;
  }
  const setChats = useChatStore((s) => s.setChats);

  useEffect(() => {
    const chats = loaderData.chats?.reduce((acc, chat) => {
      acc[chat.id] = chat;
      return acc;
    }, {} as Record<string, Chat>);

    setChats(chats!);
  }, [loaderData, setChats]);

  return (
    <div className="page">
        <Sidebar />

        {/* content */}
        <main className="">
          <Outlet /> {/* <p>loading</p> will be centered */}
        </main>
    </div>
  );
}
