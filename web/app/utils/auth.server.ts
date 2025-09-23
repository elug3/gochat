import { getSession } from "~/session.server";

export async function getAccessToken(request: Request) {
    const session = await getSession(request.headers.get("Cookie"));
    return session.get("accessToken") as string | undefined;
}

export async function requireAccessToken(request: Request) {
    const accessToken = await getAccessToken(request);
    if (!accessToken) {
        throw new Error("Access token is required");
    }
    return accessToken;
}

export async function getUser(request: Request) {
    const session = await getSession(request.headers.get("Cookie"));
    const token = session.get("accessToken") as string | undefined;

    if (!token) return null;

    const res = await fetch(process.env.AUTH_URL + "/me", {
        headers: {
            "Authorization": "Bearer " + token,
        }
    });

    if (!res.ok) {
        return null;
    }

    return res.json();
}

// export async function requireUser(request: Request) {
//     const user = await getUser(request);
//     if (!user) {
//         return redirect("/login");
//     }
//     return user;
// }

export async function getChats(request: Request) {
    const session = await getSession(request.headers.get("Cookie"));
    const token = session.get("accessToken") as string | undefined;

    if (!token) return [];

    const res = await fetch(process.env.API_URL + "/chats", {
        headers: {
            "Authorization": "Bearer " + token,
        }
    });

    if (!res.ok) return [];

    return res.json();
}

export async function getChatMessages(request: Request, chatId: string) {
    const session = await getSession(request.headers.get("Cookie"))
    const token = session.get("accessToken") as string | undefined;
    
    if (!token) throw new Error("No access token");

    const res = await fetch(process.env.API_URL + `/chats/${chatId}/messages`, {
        headers: {
            "Authorization": "Bearer " + token,
        }
    });

    if (!res.ok) throw new Error("Failed to fetch messages " + res.statusText);

    const data = await res.json();
    return data;
}


export async function sendMessage(request: Request, sendParams: {chatId: number, content: string}) {
    const accessToken = await requireAccessToken(request);

    const res = await fetch(process.env.API_URL + `/messages`, {
        method: "POST",
        headers: {
            "Content-Type": "application/json",
            "Authorization": "Bearer " + accessToken,
        },
        body: toSnakeCase(JSON.stringify(sendParams)),
    });

    if (!res.ok) {
        throw new Error("Failed to send message " + res.statusText);
    }

    return res.json();
}

function toSnakeCase(str: string): string {
    return str
    .replace(/([a-z0-9])([A-Z])/g, "$1_$2") 
    .replace(/([A-Z])([A-Z][a-z])/g, "$1_$2") 
    .toLowerCase();
}

export async function getWsToken(request: Request): Promise<string> {
    const accessToken = await getAccessToken(request);
    
    if (!accessToken) {
        throw new Error("No access token");
    }

    const res = await fetch(process.env.AUTH_URL + `/auth/ws`, {
        method: "POST",
        headers: {
            "Authorization": "Bearer " + accessToken,
        }
    })
    if (!res.ok) {
        throw new Error("Failed to get WS token: " + res.statusText);
    }

    const data = await res.json();
    const wsToken = data.ws_token;

    return wsToken;
}

export async function fetchChatList(request: Request): Promise<Chat[]> {
    const token = await getAccessToken(request);

    if (!token) {
        throw new Error("No access token");
    }
    
    const res = await fetch(process.env.API_URL + "/chats", {
        headers: {
            Authorization: `Bearer ${token}`
        }
    });

    if (!res.ok) {
        throw new Error("Failed to fetch chats");
    }
    const jsonData = await res.json();

    const chats: Chat[] = jsonData.map((chat: any) => ({
        id: chat.id,
        name: chat.name,
        lastMessage: chat.last_message,
        lastMessageAt: new Date(chat.last_message_at * 1000).toString(),
        unreadCount: chat.unread_count || 0,
    }));

    return chats;
}