import { useState, type FormEvent } from 'react';
import { useActionData, useFetcher, type ActionFunctionArgs } from "react-router";
import { register } from '~/utils/auth.server';


export async function action({request}: ActionFunctionArgs): Promise<{ok: boolean, error?: string}> {
    const data = await request.formData();
    const name = data.get("name");
    const username = data.get("username");
    const password = data.get("password");

    if (typeof name !== "string" || typeof username !== "string" || typeof password !== "string") {
      return {ok: false, error: "Invalid form data"};
    }
    if (password.length < 8) {
      return {ok: false, error: "Password must be at least 8 characters long"};
    }
    try {
      await register(username, password, name);
      return {ok: true, error: undefined};
    } catch (err: any) {
      return {ok: false, error: err.message};
    }
}

export default function RegisterPage() {
    const fetcher = useFetcher();
    const [showPassword, setShowPassword] = useState(false);

    console.log("fetcher data", fetcher.data);

    if (fetcher.data?.ok) {
        return (
            <div className="flex min-h-screen items-center justify-center bg-slate-50 dark:bg-slate-950 px-4">
                <div className="card w-full max-w-md text-center">
                    <h1 className="text-2xl font-bold mb-2 text-slate-900 dark:text-slate-100">Registration Successful!</h1>
                    <p className="text-slate-600 dark:text-slate-400 mb-4">Your account has been created.</p>
                    <a 
                        href="/login" 
                        className="inline-block bg-blue-600 hover:bg-blue-700 text-white font-medium py-2 px-4 rounded-md transition-colors dark:bg-blue-500 dark:hover:bg-blue-600"
                    >
                        Go to Login
                    </a>
                </div>
            </div>
        )
    }

    return (
    <div className="flex min-h-screen items-center justify-center bg-slate-50 dark:bg-slate-950 px-4">
        <div className="card w-full max-w-md">
            <h1 id="register-heading" className="text-2xl font-bold mb-2 text-slate-900 dark:text-slate-100">Create a new account</h1>
            <p className="lead mb-6">Fill in your details to register.</p>
            <fetcher.Form method="post" id="registerForm" className="space-y-4" noValidate>
              <div>
                <label htmlFor="name" className="text-slate-700 dark:text-slate-300">Display name</label>
                <input 
                  id="name" 
                  name="name" 
                  type="text" 
                  required 
                  placeholder="name"
                  className="text-slate-900 dark:text-slate-100"
                />
              </div>

              <div>
                <label htmlFor="username" className="text-slate-700 dark:text-slate-300">Username</label>
                <input 
                  id="username" 
                  name="username" 
                  type="text" 
                  required 
                  placeholder="username"
                  className="text-slate-900 dark:text-slate-100"
                />
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
                    placeholder="********"
                    className="flex-1 text-slate-900 dark:text-slate-100"
                  />
                  <button
                    type="button"
                    onClick={() => setShowPassword(prev => !prev)}
                    className="px-3 py-2 text-sm text-slate-600 dark:text-slate-400 hover:text-slate-900 dark:hover:text-slate-200 border border-slate-200 dark:border-slate-700 rounded-md hover:bg-slate-100 dark:hover:bg-slate-800 transition-colors"
                    aria-pressed={showPassword}
                    aria-label="Show password"
                  >
                    {showPassword ? 'Hide' : 'Show'}
                  </button>
                </div>
              </div>
              {fetcher.data?.error && (
                <div role="alert" className="text-sm text-red-600 dark:text-red-400 bg-red-50 dark:bg-red-900 border border-red-200 dark:border-red-800 rounded-md p-3">
                    <p>{fetcher.data.error}</p>
                </div>
              )}
              <button 
                type="submit" 
                className="w-full bg-blue-600 hover:bg-blue-700 text-white font-medium py-2 px-4 rounded-md transition-colors dark:bg-blue-500 dark:hover:bg-blue-600"
              >
                Register
              </button>

              <div className="h-px bg-slate-200 dark:bg-slate-700 my-4" role="separator"></div>
              <p className="text-sm text-center text-slate-600 dark:text-slate-400">
                Already have an account? <a href="/login" className="text-blue-600 dark:text-blue-400 hover:underline">Sign in</a>
              </p>
            </fetcher.Form>
        </div>
    </div>
  );
}
