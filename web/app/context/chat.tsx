import { createContext, useContext, useEffect, useRef, useState, type Dispatch, type ReactNode, type SetStateAction } from "react";
import type { ChatData } from "~/routes/dashboard.chats";

type ChatContextType = {
    chats: Record<string, Chat>;
    chatsParticipants?: Record<string, Record<string, Profile>>;
    chatsMessages: Record<string, Message[]>;
    setChatsMessages?: Dispatch<SetStateAction<Record<string, Message[]>>>;
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

    const onMessage = (message: Message) => {
        setChatsMessages((prev) => {
            if (!prev[message.chatId]) {
                return {};
            }
            return {...prev, [message.chatId]: [...prev[message.chatId], message]};
        });

        setChats((prev) => {
            const chat = prev[message.chatId];
            if (!chat) return prev;

            return {
                ...prev,
                [chat.id]: {
                    ...chat,
                    lastMessage: message.content,
                    lastMessageAt: message.sentAt,
                    unreadCount: chat.unreadCount + 1,
                }
            }
        });
    }

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
            const data = JSON.parse(toCamelCase(event.data));
            console.log("Parsed data:", data);
            switch (data.type) {
                case "message":
                    onMessage(data.payload);
                    break;
                    
            }
        };

        return () => {
            ws.close();
        }
    }, []);

    return (
        <ChatContext.Provider value={{
            chats,
            chatsMessages,
            setChatsMessages,
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

function toCamelCase(str: string): string {
    return str.replace(/_([a-z])/g, (g) => g[1].toUpperCase());
}
