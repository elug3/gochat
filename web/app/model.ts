
type Chat = {
    id: number;
    name: string;
    lastMessage: string;
    lastMessageAt: number;
    unreadCount: number;
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