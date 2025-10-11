import { commitSession, getSession } from "~/session.server";
import { camelToSnake, keysToCamel } from "./case";
import { redirect } from "react-router";
import type { Chat, Member, Message, Profile } from "~/model";

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

export async function login(username: string, password: string): Promise<{accessToken: string, userId: number}> {
    const res = await fetch(process.env.AUTH_URL + "/login", {
        method: "POST",
        headers: {
            "Content-Type": "application/json",
            "Authorization": "Basic " + btoa(`${username}:${password}`),
        },
    });

    if (!res.ok) {
        const err = await res.json();
        throw new Error(err.error);
    }
    const {AccessToken, UserId} = await res.json();

    return { accessToken: AccessToken, userId: UserId }; 
}


export async function register(username: string, password: string, name: string) {
    const resp = await fetch(process.env.API_URL + "/register", {
        method: "POST",
        headers: {
            "Content-Type": "application/json",
        },
        body: JSON.stringify({username, password, name}),
    });

    if (!resp.ok) {
        const err = await resp.json();
        throw new Error(err.error)
    }
    
    return {};
}

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

export async function getChatMessages(request: Request, chatId: string): Promise<Message[]> {
    const session = await getSession(request.headers.get("Cookie"))
    const token = session.get("accessToken") as string | undefined;
    
    if (!token) throw new Error("No access token");

    const res = await fetch(process.env.API_URL + `/chats/${chatId}/messages`, {
        headers: {
            "Authorization": "Bearer " + token,
        }
    });

    if (!res.ok) throw new Error("Failed to fetch messages " + res.statusText);

    const jsonData = await res.json();
    const data = await keysToCamel(jsonData);

    return data as Message[];
}

export async function getGroupMembers(request: Request, chatId: string): Promise<Member[]> {
    const session = await getSession(request.headers.get("Cookie"))
    const token = session.get("accessToken") as string | undefined;
    
    if (!token) throw new Error("No access token");

    const res = await fetch(process.env.API_URL + `/chats/${chatId}/participants`, {
        headers: {
            "Authorization": "Bearer " + token,
        }
    });

    if (!res.ok) throw new Error("Failed to fetch chat members " + res.statusText);

    const jsonData = await res.json();
    const members = await keysToCamel(jsonData);

    return members as Member[];

}

export async function sendMessage(request: Request, sendParams: {chatId: number, content: string}) {
    const accessToken = await requireAccessToken(request);

    const res = await fetch(process.env.API_URL + `/messages`, {
        method: "POST",
        headers: {
            "Content-Type": "application/json",
            "Authorization": "Bearer " + accessToken,
        },
        body: camelToSnake(JSON.stringify(sendParams)),
    });

    if (!res.ok) {
        throw new Error("Failed to send message " + res.statusText);
    }

    return res.json();
}

type CreateChatParams = {
    name?: string | null;
    participantIds?: Array<string | number>;
};

export async function createChat(request: Request, params: CreateChatParams): Promise<Chat> {
    const accessToken = await requireAccessToken(request);

    const participantIds = (params.participantIds ?? [])
        .map((value) => (typeof value === "string" && value.trim().length > 0 ? value.trim() : value))
        .filter((value) => value !== "" && value !== null && value !== undefined)
        .map((value) => {
            if (typeof value === "string" && /^\d+$/.test(value)) {
                return Number(value);
            }
            return value;
        });

    const payload: Record<string, unknown> = {};

    if (params.name && params.name.trim().length > 0) {
        payload.name = params.name.trim();
    }

    if (participantIds.length > 0) {
        payload.participant_ids = participantIds;
    }

    const res = await fetch(process.env.API_URL + "/chats", {
        method: "POST",
        headers: {
            "Content-Type": "application/json",
            Authorization: `Bearer ${accessToken}`,
        },
        body: JSON.stringify(payload),
    });

    if (!res.ok) {
        let message = "Failed to create chat";
        try {
            const data = await res.json();
            if (typeof data?.error === "string" && data.error.trim().length > 0) {
                message = data.error;
            }
        } catch {
            // ignore parse errors
        }
        throw new Error(message);
    }

    const data = await res.json();
    const chat: Chat = {
        id: data.id,
        name: data.name ?? "Untitled chat",
        lastMessage: data.last_message ?? "",
        lastMessageAt: data.last_message_at ?? Date.now(),
        unreadCount: data.unread_count ?? 0,
    };

    return chat;
}



function toCamelCase(str: string): string {
    return str.replace(/_([a-z])/g, (g) => g[1].toUpperCase());
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

export class ErrUnauthorized extends Error {}

export async function fetchChatList(request: Request): Promise<Chat[]> {
    const token = await getAccessToken(request);

    if (!token) {
        throw new ErrUnauthorized("No access token");
    }
    
    const res = await fetch(process.env.API_URL + "/chats", {
        headers: {
            Authorization: `Bearer ${token}`
        }
    });

    if (!res.ok) {
        if (res.status === 401) {
            throw new ErrUnauthorized("Unauthorized access");
        }
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

export async function fetchContacts(request: Request): Promise<Profile[]> {
    const token = await getAccessToken(request);

    if (!token) {
        throw new ErrUnauthorized("No access token");
    }
    
    const res = await fetch(process.env.API_URL + "/contacts", {
        headers: {
            Authorization: `Bearer ${token}`
        }
    });

    if (!res.ok) {
        if (res.status === 401) {
            throw new ErrUnauthorized("Unauthorized access");
        }
        throw new Error("Failed to fetch contacts");
    }
    const jsonData = await res.json();

    const contacts: Profile[] = jsonData.map((contact: any) => ({
        userId: contact.user_id || contact.userId,
        name: contact.name,
        status: contact.status || "offline",
        avatarUrl: contact.avatar_url || contact.avatarUrl,
    }));

    return contacts;
}

export async function sendContactRequest(request: Request, identifier: string): Promise<void> {
    const token = await getAccessToken(request);

    if (!token) {
        throw new ErrUnauthorized("No access token");
    }

    const res = await fetch(process.env.API_URL + "/contacts", {
        method: "POST",
        headers: {
            "Content-Type": "application/json",
            Authorization: `Bearer ${token}`,
        },
        body: JSON.stringify({ "target_id": Number(identifier) || identifier }),
    });

    if (!res.ok) {
        if (res.status === 401) {
            throw new ErrUnauthorized("Unauthorized access");
        }

        let message = "Failed to send contact request";

        try {
            const data = await res.json();
            if (data?.error) {
                message = data.error;
            }
        } catch {
            // ignore json parse errors
        }

        throw new Error(message);
    }
}
