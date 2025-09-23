
type Chat = {
    id: number;
    name: string;
    lastMessage: string;
    lastMessageAt: number;
    unreadCount: number;
}


type Message = {
    Id: number;
    ChatId: number;
    Sender: number;
    Content: string;
    SentAt: string;
}