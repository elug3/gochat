import { create } from "zustand";
import type { Chat, Message, Profile } from "~/model";

type WsState = "connecting" | "open" | "closed" | "idle";

type ChatStore = {
    chats: Record<string, Chat>;
    chatsParticipants: Record<string, Record<string, Profile>>;
    chatsMessages: Record<string, Record<number, Message>>;

    wsState: WsState;
    setWsState: (state: WsState) => void;
    setChats: (chats: Record<string, Chat>) => void;
    upsertChat: (chat: Chat) => void;
    setChatsParticipants: (participants: Record<string, Record<string, Profile>>) => void;
    setChatMessages: (chatId: string, messages: Message[]) => void;
    upsertMessage: (message: Message) => void;
}

export const useChatStore = create<ChatStore>((set, get) => ({
    chats: {},
    chatsMessages: {},
    chatsParticipants: {},
    wsState: "closed",

    setWsState: (state: WsState) => set({ wsState: state }),
    setChats: (chats: Record<string, Chat>) => set({ chats }),
    upsertChat: (chat: Chat) => set((state) => ({
        chats: {
            ...state.chats,
            [chat.id]: chat,
        }
    })),
    setChatsParticipants: (participants: Record<string, Record<string, Profile>>) => set(() => ({ chatsParticipants: participants })),

    setChatMessages: (chatId: string, messages: Message[]) => {
        const { chatsMessages } = get();
        const chatKey = String(chatId);
        const normalizedMessages = messages.reduce<Record<number, Message>>((acc, message) => {
            acc[message.id] = message;
            return acc;
        }, {});

        const updatedMessages = {
            ...chatsMessages,
            [chatKey]: normalizedMessages,
        };
        set({ chatsMessages: updatedMessages });
    },

    upsertMessage: (message: Message) => {
        const { chats, chatsMessages } = get();

        const chatKey = String(message.chatId);
        const chatMessages = chatsMessages[chatKey] ?? {};
        const updatedMessages = {
            ...chatsMessages,
            [chatKey]: {
                ...chatMessages,
                [message.id]: message,
            },
        };

        // update chats
        const chat = chats[chatKey];
        const updatedChats = chat ? {...chats, [chatKey]: {
            ...chat,
            lastMessage: message.content,
            lastMessageAt: message.sentAt,
            unreadCount: chat.unreadCount + 1
        }} : chats;

        set({
            chats: updatedChats,
            chatsMessages: updatedMessages
        });
    },

}));
