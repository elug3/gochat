import { create } from "zustand";
import type { Chat, Message, Profile } from "~/model";

type WsState = "connecting" | "open" | "closed" | "idle";

type ChatStore = {
    chats: Record<string, Chat>;
    chatsParticipants: Record<string, Record<string, Profile>>;
    chatsMessages: Record<string, Message[]>;

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

        const updatedMessages = {
            ...chatsMessages,
            [chatId]: messages,
        };
        set({ chatsMessages: updatedMessages });
    },

    upsertMessage: (message: Message) => {
        const { chats, chatsMessages } = get();

        const updatedMessages = {
            ...chatsMessages,
            [message.chatId]: [
                ...(chatsMessages[message.chatId] || []),
                message,
            ],
        };

        // update chats
        const chat = chats[message.chatId];
        const updatedChats = chat ? {...chats, [chat.id]: {
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