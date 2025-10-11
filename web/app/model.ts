export type Chat = {
    id: number;
    name: string;
    lastMessage: string;
    lastMessageAt: number;
    unreadCount: number;
}

export type Member = {
    userId: string;
    role: "member" | "owner",
}

export type Group = {
    id: number;
    name: string;
    createdAt: number;
}

export type Message = {
    id: number;
    chatId: number;
    sender: number;
    content: string;
    sentAt: number;
}

export type Profile = {
    userId: string;
    name: string;
    status: "online" | "offline";
    avatarUrl?: string;
}