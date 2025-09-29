import { useState } from "react";
import { Form, redirect, useActionData, useFetcher, type ActionFunctionArgs } from "react-router";
import { commitSession, getSession } from "~/session.server";
import { login } from "~/utils/auth.server";

export async function action({ request }: ActionFunctionArgs) {

  const formData = await request.formData();

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
  const fetcher = useFetcher();
  const actionData = useActionData();
  const [showPassword, setShowPassword] = useState(false);
  return (
    <div>
      <Form method="post" id="loginForm" noValidate>

      <div>
        <label htmlFor="username">Username</label>
        <input 
          id="username" 
          name="username" 
          type="text" 
          required 
          placeholder="Username"
        />
  {actionData?.usernameError && <div aria-live="polite">{actionData.usernameError}</div>}
      </div>

      <div>
        <label htmlFor="password">Password</label>
        <div>
          <input
            id="password"
            name="password"
            type={showPassword ? 'text' : 'password'}
            required
            minLength={8}
            placeholder="••••••••"
          />
          <button
            type="button"
            onClick={() => setShowPassword(prev => !prev)}
            aria-pressed={showPassword}
            aria-label="Show password"
          >{showPassword ? 'Hide' : 'Show'}</button>
        </div>
  {actionData?.passwordError && <div aria-live="polite">{actionData.passwordError}</div>}
      </div>

      <div>
        <a href="/forgot">Forgot password?</a>
      </div>
      <div>
        <a href="/register">Create an account</a>
      </div>
  {actionData?.error && <div aria-live="polite">{actionData.error}</div>}
      <button type="submit">Sign in</button>

      </Form>
    </div>
  );
}