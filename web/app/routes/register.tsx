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
            <div>
                <h1>Registration Successful!</h1>
                <p>Your account has been created.</p>
                <a href="/login" >Go to Login</a>
            </div>
        )
    }

    return (
    <div>
        <h1 id="register-heading">Create a new account</h1>
        <p className="lead">Fill in your details to register.</p>
        <fetcher.Form method="post" id="registerForm" className="space-y-4" noValidate>
          <div>
            <label htmlFor="name">display name</label>
            <input 
              id="name" 
              name="name" 
              type="text" 
              required 
              placeholder="name"
              className="input"
            />
          </div>

          <div>
            <label htmlFor="username">Username </label>
            <input 
              id="username" 
              name="username" 
              type="text" 
              required 
              placeholder="username"
              className="input"
            />
          </div>

          <div>
            <label htmlFor="password">Password</label>
            <div className="row">
              <input
                id="password"
                name="password"
                type={showPassword ? 'text' : 'password'}
                required
                minLength={8}
                placeholder="********"
                className="input"
              />
              <button
                type="button"
                onClick={() => setShowPassword(prev => !prev)}
                className="toggle-pass"
                aria-pressed={showPassword}
                aria-label="Show password"
              >{showPassword ? 'Hide' : 'Show'}</button>
            </div>
          </div>
          {fetcher.data?.error && (
            <div role="alert">
                <p className='error'>
                    {fetcher.data.error}
                </p>
            </div>
          )}
          <button type="submit" className="submit-btn">Register</button>

          <div className="divider" role="separator"></div>
          <p className="small text-center">Already have an account? <a href="/login">Sign in</a></p>
        </fetcher.Form>
    </div>
  );
}
