import { createContext, useContext, useEffect, useRef, useState, type Dispatch, type ReactNode, type SetStateAction } from "react";
import type { ChatData } from "~/routes/dashboard.chats";

type ChatContextType = {
    chats: Record<string, Chat>;
    chatsMessages: Record<string, Message[]>;
    wsState: WsState;
}

type WsState = "connecting" | "open" | "closed" | "error" | "idle";

const ChatContext = createContext<ChatContextType | null>(null);

export function ChatProvider({children, initialChatData}: {children: ReactNode, initialChatData: ChatData})  {
    const initialChats = initialChatData.chats.reduce((acc, chat) => {
        acc[chat.id] = chat;
        return acc;
    }, {} as Record<string, Chat>);

    const wsRef = useRef<WebSocket | null>(null);
    const [wsState, setWsState] = useState<WsState>("idle");
    const [chats, setChats] = useState<Record<string, Chat>>(initialChats);
    const [chatsMessages, setChatsMessages] = useState<Record<string, Message[]>>({});

    useEffect(() => {
        if (wsRef.current) return; // already connected

        const ws = new WebSocket(`ws://172.18.0.10:12345/ws?token=${initialChatData.wsToken}`);
        wsRef.current = ws;

        ws.onopen = () => {
            setWsState("open");
            console.log("WebSocket connected");
        };

        ws.onclose = () => {
            setWsState("closed");
            console.log("WebSocket disconnected");
        };

        ws.onmessage = (event) => {
            console.log("WebSocket message received:", event.data);
        };

        return () => {
            ws.close();
        }
    }, []);

    return (
        <ChatContext.Provider value={{
            chats,
            chatsMessages,
            wsState,
        }}>
            {children}
        </ChatContext.Provider>
    )
}

export function useChat() {
    const context = useContext(ChatContext);
    if (!context) {
        throw new Error("useChat must be used within a ChatProvider");
    }
    return context;
}