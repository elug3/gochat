import { createCookieSessionStorage, redirect } from "react-router";


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
