import { use, useEffect } from "react";
import { Link, NavLink, Outlet, redirect, useLoaderData, type LoaderFunctionArgs, type ShouldRevalidateFunctionArgs } from "react-router";
import { useChatWebSocket } from "~/hooks/useChatWebsocket";
import { useChatStore } from "~/store/chatStore";
import { fetchChatList, getWsToken } from "~/utils/auth.server";

export type ChatData = {
  chats: Chat[];
  wsToken?: string;
}

export async function loader({request}: LoaderFunctionArgs): Promise<{ chats?: Chat[], wsToken?: string, error?: string }> {
  try {
    const chats = await fetchChatList(request);
    const wsToken = await getWsToken(request);
    return { chats, wsToken };
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

  const chatList = Object.values(chats || {}).flat();


  return (
    <aside className="sidebar">
      <h2 className="">Chats</h2>

      <nav className="sidebar-container">
        {
          chatList.length === 0 ? (
            <div className="">
              <p>No chats available.</p>
              <Link to="/dashboard/contacts" className="">Add contacts</Link> to start chatting.
            </div>
          ) : (chatList.map((chat) => (
              <NavLink
                key={chat.id}
                to={`${chat.id}`}
                className={({ isActive }) =>
                  `sidebar-item ${
                    isActive
                      ? ""
                      : ""
                  }`
                }
              >
                
                <div className="flex-1 min-w-0">
                  <div className="flex items-center justify-between gap-2">
                    <div className="text-sm font-medium truncate">{chat.name}</div>
                    {/* <div className="text-xs text-gray-400 flex-shrink-0">{chat.lastMessageAt}</div> */}
                  </div>
    
                  <div className="text-xs text-gray-500 truncate mt-0.5">
                    {chat.lastMessage || "No messages yet."}
                  </div>
                </div>
                {chat.unreadCount > 0 && (
                  <span className="text-center w-6 h-6 leading-6 rounded-md bg-red-100 text-white text-xs font-semibold ml-2 flex-shrink-0">
                    {chat.unreadCount}
                  </span>
                )}
    
              </NavLink>
          )))
        }

      </nav>

      <div className="mt-auto pt-4 border-t text-sm text-gray-500">
        WebSocket: <span className={`font-mono ${wsState === "open" ? "text-green-600" : wsState === "connecting" ? "text-yellow-600" : "text-red-600"}`}>{wsState}</span>
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
  if (!loaderData.wsToken) {
    return <div>Failed to get websocket token</div>;
  }

  const setChats = useChatStore((s) => s.setChats);

  useEffect(() => {
    const chats = loaderData.chats?.reduce((acc, chat) => {
      acc[chat.id] = chat;
      return acc;
    }, {} as Record<string, Chat>);

    setChats(chats!);
  }, [loaderData, setChats]);

  // start websocket
  useChatWebSocket(loaderData.wsToken);

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
