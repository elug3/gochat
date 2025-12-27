import { createCookie } from "react-router";
import { CSRF } from "remix-utils/csrf/server";

export const cookie = createCookie("csrf", {
    path: "/",
    httpOnly: true,
    sameSite: "lax",
    secure: process.env.NODE_ENV === "production",
    secrets: ["secret"],
})

export const csrf = new CSRF({
    cookie,
    formDataKey: "csrf",
    secret: "secret",
});