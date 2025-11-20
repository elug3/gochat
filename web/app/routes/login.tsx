import { useState } from "react";
import { Form, redirect, useActionData, useFetcher, type ActionFunctionArgs } from "react-router";
import { commitSession, getSession } from "~/session.server";
import { login } from "~/utils/auth.server";

const PASSKEY_SERVER_URL = "http://localhost:8001";

export async function action({ request }: ActionFunctionArgs) {

  const formData = await request.formData();
  const intent = formData.get("_action");

  if (intent === "passkey-login") {
    const accessToken = formData.get("accessToken");
    const userId = formData.get("userId");

    if (typeof accessToken !== "string" || typeof userId !== "string") {
      return { error: "Invalid passkey response" };
    }

    const numericUserId = Number(userId);
    if (!Number.isFinite(numericUserId)) {
      return { error: "Invalid passkey user" };
    }

    const session = await getSession();
    session.set("accessToken", accessToken);
    session.set("userId", numericUserId);

    return redirect("/", {
      headers: {
        "Set-Cookie": await commitSession(session),
      },
    });
  }

  const username = formData.get("username");
  const password = formData.get("password");

  if (typeof username !== "string" || typeof password !== "string") {
    return {error: "Invalid form data"};
  }

  try {
    const session = await getSession();
    const { accessToken, userId } = await login(username, password);
    session.set("accessToken", accessToken);
    session.set("userId", userId);
    
    return redirect("/", {
      headers: {
        "Set-Cookie": await commitSession(session),
      },
    });
  } catch (err: any) {
    return {
      error: err.message,
    };
  }

}

export default function LoginPage() {
  const fetcher = useFetcher<{ error?: string }>();
  const actionData = useActionData<{ error?: string; usernameError?: string; passwordError?: string }>();
  const [showPassword, setShowPassword] = useState(false);
  const [passkeyError, setPasskeyError] = useState<string | null>(null);
  const [isPasskeyClientBusy, setIsPasskeyClientBusy] = useState(false);

  const isPasskeyLoading = isPasskeyClientBusy || fetcher.state !== "idle";
  const serverPasskeyError =
    fetcher.state === "idle" && !isPasskeyClientBusy && typeof fetcher.data?.error === "string"
      ? fetcher.data.error
      : null;
  const combinedPasskeyError = passkeyError ?? serverPasskeyError;

  const handlePasskeyLogin = async () => {
    setPasskeyError(null);

    if (typeof window === "undefined" || typeof window.PublicKeyCredential === "undefined") {
      setPasskeyError("Passkeys are not supported in this browser.");
      return;
    }

    if (!navigator.credentials || typeof navigator.credentials.get !== "function") {
      setPasskeyError("Credential management APIs are unavailable.");
      return;
    }

    setIsPasskeyClientBusy(true);

    try {
      const startRes = await fetch(`${PASSKEY_SERVER_URL}/login/start`, {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        credentials: "include",
      });

      if (!startRes.ok) {
        throw new Error("Failed to start a passkey sign in request.");
      }

      const { publicKey } = await startRes.json();
      if (!publicKey || !publicKey.challenge) {
        throw new Error("Invalid login challenge from server.");
      }

      normalizePublicKeyRequest(publicKey);

      const credential = (await navigator.credentials.get({
        publicKey,
      })) as PublicKeyCredential | null;

      if (!credential) {
        throw new Error("Passkey authentication was cancelled.");
      }

      const finishRes = await fetch(`${PASSKEY_SERVER_URL}/login/finish`, {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        credentials: "include",
        body: JSON.stringify(credential),
      });

      let finishPayload: any = {};
      try {
        finishPayload = await finishRes.json();
      } catch {
        // ignore non JSON responses
      }

      if (!finishRes.ok) {
        throw new Error(finishPayload?.error ?? "Failed to verify passkey.");
      }

      const accessToken = finishPayload?.accessToken ?? finishPayload?.AccessToken;
      const userId = finishPayload?.userId ?? finishPayload?.UserId;

      if (typeof accessToken !== "string" || (typeof userId !== "string" && typeof userId !== "number")) {
        throw new Error("Authentication server returned an invalid response.");
      }

      fetcher.submit(
        {
          _action: "passkey-login",
          accessToken,
          userId: String(userId),
        },
        { method: "post" },
      );
    } catch (err) {
      setPasskeyError(err instanceof Error ? err.message : "Failed to sign in with a passkey.");
    } finally {
      setIsPasskeyClientBusy(false);
    }
  };

  return (
    <div className="flex min-h-screen items-center justify-center bg-slate-50 dark:bg-slate-950 px-4">
      <div className="card w-full max-w-md">
        <h1 className="text-2xl font-bold mb-2 text-slate-900 dark:text-slate-100">Sign in</h1>
        <p className="lead mb-6">Enter your credentials to access your account.</p>
        <Form method="post" id="loginForm" noValidate className="space-y-4">
          <div>
            <label htmlFor="username" className="text-slate-700 dark:text-slate-300">Username</label>
            <input 
              id="username" 
              name="username" 
              type="text" 
              required 
              placeholder="Username"
              className="text-slate-900 dark:text-slate-100"
            />
            {actionData?.usernameError && (
              <div aria-live="polite" className="text-sm text-red-600 dark:text-red-400 mt-1">
                {actionData.usernameError}
              </div>
            )}
          </div>

          <div>
            <label htmlFor="password" className="text-slate-700 dark:text-slate-300">Password</label>
            <div className="flex gap-2">
              <input
                id="password"
                name="password"
                type={showPassword ? 'text' : 'password'}
                required
                minLength={8}
                placeholder="••••••••"
                className="flex-1 text-slate-900 dark:text-slate-100"
              />
              <button
                type="button"
                onClick={() => setShowPassword(prev => !prev)}
                aria-pressed={showPassword}
                aria-label="Show password"
                className="px-3 py-2 text-sm text-slate-600 dark:text-slate-400 hover:text-slate-900 dark:hover:text-slate-200 border border-slate-200 dark:border-slate-700 rounded-md hover:bg-slate-100 dark:hover:bg-slate-800 transition-colors"
              >
                {showPassword ? 'Hide' : 'Show'}
              </button>
            </div>
            {actionData?.passwordError && (
              <div aria-live="polite" className="text-sm text-red-600 dark:text-red-400 mt-1">
                {actionData.passwordError}
              </div>
            )}
          </div>

          <div className="flex flex-col gap-2 text-sm">
            <a href="/forgot" className="text-blue-600 dark:text-blue-400 hover:underline">
              Forgot password?
            </a>
            <a href="/register" className="text-blue-600 dark:text-blue-400 hover:underline">
              Create an account
            </a>
          </div>
          
          {actionData?.error && (
            <div aria-live="polite" className="text-sm text-red-600 dark:text-red-400 bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 rounded-md p-3">
              {actionData.error}
            </div>
          )}
          
          <button 
            type="submit" 
            className="w-full bg-blue-600 hover:bg-blue-700 text-white font-medium py-2 px-4 rounded-md transition-colors dark:bg-blue-500 dark:hover:bg-blue-600"
          >
            Sign in
          </button>

          <div className="flex items-center gap-3 text-xs text-slate-500 dark:text-slate-400" aria-hidden="true">
            <span className="h-px flex-1 bg-slate-200 dark:bg-slate-700" />
            <span>Or</span>
            <span className="h-px flex-1 bg-slate-200 dark:bg-slate-700" />
          </div>

          <button
            type="button"
            onClick={handlePasskeyLogin}
            disabled={isPasskeyLoading}
            className="w-full border border-slate-300 bg-white text-slate-700 hover:bg-slate-50 dark:border-slate-600 dark:bg-slate-900 dark:text-slate-100 dark:hover:bg-slate-800 font-medium py-2 px-4 rounded-md transition-colors disabled:cursor-not-allowed disabled:opacity-60"
          >
            {isPasskeyLoading ? "Waiting for passkey..." : "Sign in with passkey"}
          </button>

          {combinedPasskeyError && (
            <div
              aria-live="polite"
              className="text-sm text-amber-700 dark:text-amber-300 bg-amber-50 dark:bg-amber-900/30 border border-amber-200 dark:border-amber-800 rounded-md p-3"
            >
              {combinedPasskeyError}
            </div>
          )}
        </Form>
      </div>
    </div>
  );
}

function normalizePublicKeyRequest(publicKey: any) {
  if (typeof publicKey.challenge === "string") {
    publicKey.challenge = decodeBase64URL(publicKey.challenge);
  }

  if (Array.isArray(publicKey.allowCredentials)) {
    publicKey.allowCredentials = publicKey.allowCredentials.map((cred: any) => ({
      ...cred,
      id: typeof cred.id === "string" ? decodeBase64URL(cred.id) : cred.id,
    }));
  }

  if (publicKey.user?.id && typeof publicKey.user.id === "string") {
    publicKey.user.id = decodeBase64URL(publicKey.user.id);
  }

  return publicKey;
}

function decodeBase64URL(value: string) {
  const pad = value.length % 4 ? 4 - (value.length % 4) : 0;
  const base64 = (value + "=".repeat(pad)).replace(/-/g, "+").replace(/_/g, "/");
  return Uint8Array.from(atob(base64), (c) => c.charCodeAt(0));
}
