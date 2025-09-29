import { createCookieSessionStorage, redirect } from "react-router";
import { requireAccessToken } from "./utils/auth.server";


export const { getSession, commitSession, destroySession } = createCookieSessionStorage({
    cookie: {
        name: "__session",
        httpOnly: true,
        secure: process.env.NODE_ENV === "production",
        sameSite: "lax",
        maxAge: 60 * 60 * 24,
        secrets: [process.env.SESSION_SECRET!],
    },
});

export async function getUserSession(request: Request) {
    return await getSession(request.headers.get("Cookie"));
}

export async function getUserId(request: Request) {
    const session = await getUserSession(request);
    return session.get("userId") as number | undefined;
}

export async function requireUserId(request: Request) {
    const userId = await getUserId(request);

    if (!userId) {
        throw redirect("/login");
    }

    return userId;
}

export async function requireProfile(request: Request): Promise<Profile> {
    const accessToken = await requireAccessToken(request);
    const res = await fetch(`${process.env.API_URL}/profile`, {
        headers: {
            "Authorization": `Bearer ${accessToken}`,
        }
    });

    if (res.status === 401) {
        throw redirect("/login");
    }

    if (!res.ok) {
        throw new Error("Failed to fetch profile");
    }

    const data = await res.json();

    const profile = {
        userId: data.userId,
        name: data.name,
        status: data.status,
        avatarUrl: data.avatarUrl,
    }
    return profile;
}


export async function createUserSession(userId: number, redirectTo: string) {
    const session = await getSession();
    session.set("userId", userId);
    return redirect(redirectTo, {
        headers: {
            "Set-Cookie": await commitSession(session),
        },
    });
}

export async function logout(request: Request) {
    const session = await getUserSession(request);
    
    return redirect("/login", {
        headers: {
            "Set-Cookie": await destroySession(session),
        },
    });
}