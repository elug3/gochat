import { Link, NavLink, Outlet, useLoaderData, type LoaderFunctionArgs, type ShouldRevalidateFunctionArgs } from "react-router";
import { ChatProvider, useChat } from "~/context/chat";
import { fetchChatList, getWsToken } from "~/utils/auth.server";

const API_BASE = "http://localhost:8080";


export type ChatData = {
  chats: Chat[];
  wsToken?: string;
}

export async function loader({request}: LoaderFunctionArgs): Promise<ChatData> {
  const chats = await fetchChatList(request);
  const wsToken = await getWsToken(request);

  return { chats, wsToken };
}

export function shouldRevalidate({ formAction }: ShouldRevalidateFunctionArgs) {
  return false;
}



export function Sidebar() {
  const {chats, wsState} = useChat();

  const chatList = Object.values(chats || {}).flat();

  return (
    <aside className="w-64 border-r bg-white p-4 flex flex-col h-full">
      <h2 className="text-lg font-bold mb-4">Chats</h2>

      <nav className="space-y-2 overflow-auto">
        {chatList.map((chat) => (
          <NavLink
            key={chat.id}
            to={`${chat.id}`}
            className={({ isActive }) =>
              `flex block p-2 rounded ${
                isActive
                  ? "bg-blue-100 text-blue-700 font-semibold"
                  : "hover:bg-gray-100 text-gray-800"
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
            <span className="text-center w-6 h-6 leading-6 rounded-md bg-red-100 text-white text-xs font-semibold ml-2 flex-shrink-0">
              {chat.unreadCount}
            </span>

          </NavLink>
        ))}
      </nav>

      <div className="mt-auto pt-4 border-t text-sm text-gray-500">
        WebSocket: <span className={`font-mono ${wsState === "open" ? "text-green-600" : wsState === "connecting" ? "text-yellow-600" : "text-red-600"}`}>{wsState}</span>
      </div>
    </aside>
  );
}

export default function Chats() {
  const initialChatData = useLoaderData<typeof loader>();

  return (
    <ChatProvider initialChatData={initialChatData}>
      <div className="flex h-full">
          <Sidebar/>

          {/* content */}
          <main>
              <Outlet/>
          </main>

      </div>
    </ChatProvider>
  )
}