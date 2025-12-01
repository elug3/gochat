import { useEffect, useState } from "react";
import { Form, redirect, useActionData, useFetcher, type ActionFunctionArgs } from "react-router";
import { commitSession, getSession } from "~/session.server";
import { login } from "~/utils/auth.server";

const WEBAUTHN_LOGIN_START_URL = `http://localhost:8080/webauthn/login/begin`;
const WEBAUTHN_LOGIN_FINISH_URL = `http://localhost:8080/webauthn/login/finish`;

export async function action({ request }: ActionFunctionArgs) {

  const formData = await request.formData();
  const intent = formData.get("_action");

  switch (intent) {
    case "passkey-login-start": {
      try {
        const startRes = await fetch(`${WEBAUTHN_LOGIN_START_URL}`, {
          method: "POST",
          headers: {
            "Content-Type": "application/json",
          },
        });

        if (!startRes.ok) {
          console.error("Passkey login start failed:", await startRes.text());
          return Response.json({ ok: false, error: "Failed to initiate passkey login." }, { status: startRes.status });
        }
        const { publicKey } = await startRes.json();
        return Response.json({ ok: true, publicKey });
      } catch (error) {
        console.error("Passkey login start failed:", error);
        return Response.json({ ok: false, error: "Failed to initiate passkey login." }, { status: 500 });
      }
    }

    case "passkey-login-finish": {
      const credential = formData.get("credential");
      if (typeof credential !== "string") {
        return Response.json({ ok: false, error: "Invalid credential data." }, { status: 400 });
      }

      let parsedCredential: any;
      try {
        parsedCredential = JSON.parse(credential);
      } catch (error) {
        console.error("Failed to parse credential", error);
        return Response.json({ ok: false, error: "Invalid credential data." }, { status: 400 });
      }

      console.log("Received credential:", parsedCredential?.id);
      const finishRes = await fetch(WEBAUTHN_LOGIN_FINISH_URL, {
        method: "POST",
        body: JSON.stringify(parsedCredential),
        headers: {"Content-Type": "application/json"},
      });

      if (!finishRes.ok) {
        console.error("Passkey login finish failed:", await finishRes.text());
        return Response.json({ ok: false, error: "Failed to complete passkey login." }, { status: finishRes.status });
      }
      const token = await finishRes.json();

      const session = await getSession(request.headers.get("Cookie"));
      const accessToken = token.AccessToken ?? token.accessToken;
      const userId = token.UserId ?? token.userId;

      if (!accessToken || !userId) {
        return Response.json({ ok: false, error: "Invalid login response from server." }, { status: 500 });
      }

      session.set("accessToken", accessToken);
      session.set("userId", userId);

      return redirect("/dashboard", {
        headers: {
          "Set-Cookie": await commitSession(session),
        },
      });
    }
    case "password-login":
    case null:
    case undefined: {
      const username = formData.get("username");
      const password = formData.get("password");

      if (typeof username !== "string" || username.trim() === "") {
        return {
          error: "Please provide your username.",
          usernameError: "Username is required.",
        };
      }

      if (typeof password !== "string" || password.length < 8) {
        return {
          error: "Please provide your password.",
          passwordError: "Password must be at least 8 characters.",
        };
      }

      try {
        const session = await getSession(request.headers.get("Cookie"));
        const { accessToken, userId } = await login(username, password);

        session.set("accessToken", accessToken);
        session.set("userId", userId);

        return redirect("/dashboard", {
          headers: { "Set-Cookie": await commitSession(session) },
        });
      } catch (err: any) {
        console.error("Password login failed:", err);
        return {
          error: err?.message || "Login failed. Please try again.",
        };
      }
    }
    default:
      return Response.json({ ok: false, error: "Unknown action" }, { status: 400 });
  // if (intent === "passkey-login-start") {
  //   const startRes = await fetch(`${WEBAUTHN_LOGIN_START_URL}`, {
  //     method: "POST",
  //     headers: {
  //       "Content-Type": "application/json",
  //     },
  //   });
  //   console.log("Passkey login start response:", startRes);
  // }

  // const username = formData.get("username");
  // const password = formData.get("password");

  // if (typeof username !== "string" || typeof password !== "string") {
  //   return {error: "Invalid form data"};
  // }

  // try {
  //   const session = await getSession();
  //   const { accessToken, userId } = await login(username, password);
  //   session.set("accessToken", accessToken);
  //   session.set("userId", userId);
    
  //   return redirect("/", {
  //     headers: {"Set-Cookie": await commitSession(session),},
  //   });
  // } catch (err: any) {
  //   return {
  //     error: err.message,
  //   };
  // }
  }
}

export default function LoginPage() {
  type PasskeyActionResponse = {
    ok?: boolean;
    error?: string;
    publicKey?: PublicKeyCredentialRequestOptions;
  };

  const fetcher = useFetcher<PasskeyActionResponse>();
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

  useEffect(() => {
    if (fetcher.state !== "idle") return;
    if (!fetcher.data) return;

    if (fetcher.data.error) {
      setPasskeyError(fetcher.data.error);
      setIsPasskeyClientBusy(false);
      return;
    }

    if (isPasskeyClientBusy && fetcher.data.ok && fetcher.data.publicKey) {
      const request = normalizePublicKeyRequest({ ...fetcher.data.publicKey });
      navigator.credentials
        .get({ publicKey: request })
        .then((credential: Credential | null) => {
          if (!credential) {
            throw new Error("No credential returned");
          }
          const serialized = serializeAssertionResponse(credential as PublicKeyCredential);
          fetcher.submit(
            {
              _action: "passkey-login-finish",
              credential: JSON.stringify(serialized),
            },
            { method: "post", action: "/login" }
          );
        })
        .catch((err) => {
          console.error(err);
          setPasskeyError("Failed to get passkey credential.");
        })
        .finally(() => {
          setIsPasskeyClientBusy(false);
        });
    }
  }, [fetcher.data, fetcher.state, isPasskeyClientBusy, fetcher]);

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
    fetcher.submit({ _action: "passkey-login-start" }, { method: "post", action: "/login" });
  }

  return (
    <div className="flex min-h-screen items-center justify-center bg-slate-50 dark:bg-slate-950 px-4">
      <div className="card w-full max-w-md">
        <h1 className="text-2xl font-bold mb-2 text-slate-900 dark:text-slate-100">Sign in</h1>
        <p className="lead mb-6">Enter your credentials to access your account.</p>
        <Form method="post" id="loginForm" noValidate className="space-y-4">
          <input type="hidden" name="_action" value="password-login" />
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
          <div aria-live="polite" className="text-sm text-red-600 dark:text-red-400 bg-red-50 dark:bg-red-900 border border-red-200 dark:border-red-800 rounded-md p-3">
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
              className="text-sm text-amber-700 dark:text-amber-300 bg-amber-50 dark:bg-amber-900 border border-amber-200 dark:border-amber-800 rounded-md p-3"
            >
              {combinedPasskeyError}
            </div>
          )}
        </Form>
      </div>
    </div>
  );
}

function serializeAssertionResponse(credential: PublicKeyCredential) {
  const response = credential.response as AuthenticatorAssertionResponse;
  return {
    id: credential.id,
    rawId: bufferToBase64URL(credential.rawId),
    type: credential.type,
    response: {
      clientDataJSON: bufferToBase64URL(response.clientDataJSON),
      authenticatorData: bufferToBase64URL(response.authenticatorData),
      signature: bufferToBase64URL(response.signature),
      userHandle: bufferSourceToBase64URL(response.userHandle),
    },
    clientExtensionResults:
      typeof credential.getClientExtensionResults === "function"
        ? credential.getClientExtensionResults()
        : {},
  };
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

function bufferToBase64URL(buffer: ArrayBuffer | ArrayBufferLike) {
  const bytes = new Uint8Array(buffer);
  let binary = '';
  for (let i = 0; i < bytes.byteLength; i++) {
    binary += String.fromCharCode(bytes[i]);
  }
  const base64 = btoa(binary);
  return base64.replace(/\+/g, '-').replace(/\//g, '_').replace(/=+$/, '');
}

function bufferSourceToBase64URL(buffer?: ArrayBuffer | ArrayBufferView | null) {
  if (!buffer) return undefined;
  if (buffer instanceof ArrayBuffer) {
    return bufferToBase64URL(buffer);
  }
  return bufferToBase64URL(buffer.buffer);
}
