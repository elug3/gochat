
type Chat = {
    id: number;
    name: string;
    lastMessage: string;
    lastMessageAt: number;
    unreadCount: number;
}

type Member = {
    userId: string;
    role: "member" | "owner",
}

type Message = {
    id: number;
    chatId: number;
    sender: number;
    content: string;
    sentAt: number;
}

type Profile = {
    userId: string;
    name: string;
    status: "online" | "offline";
    avatarUrl?: string;
}