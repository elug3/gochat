import { Form, redirect, useActionData, type ActionFunctionArgs } from "react-router";
import { commitSession, getSession } from "~/session.server";

export async function action({ request }: ActionFunctionArgs) {
  const formData = await request.formData();
  const username = formData.get("username");
  const password = formData.get("password");

  const res = await fetch(process.env.AUTH_URL + "/login", {
    method: "POST",
    headers: {
        "Authorization": "Basic " + btoa(username + ":" + password),
    }
  });

  if (!res.ok) {
    return { error: "Invalid username or password" };
  }
  const payload = await res.json();
  const accessToken = payload.AccessToken;

  const userId = payload.UserId;

  const session = await getSession();
  session.set("accessToken", accessToken);
  session.set("userId", userId);

  return redirect("/", {
    headers: {
      "Set-Cookie": await commitSession(session),
    }
  });
}

export default function LoginPage() {
  const actionData = useActionData<typeof action>();

  return (
    <div className="flex items-center justify-center h-screen bg-gray-100">
      <div className="bg-white p-8 rounded shadow w-96">
        <h1 className="text-2xl font-bold mb-6 text-center">Login</h1>

        {actionData?.error && (
          <p className="mb-4 text-red-600 text-sm">{actionData.error}</p>
        )}

        <Form method="post" className="space-y-4">
          <div>
            <label className="block text-sm font-medium mb-1">username</label>
            <input
              type="username"
              name="username"
              required
              className="w-full border rounded p-2"
            />
          </div>

          <div>
            <label className="block text-sm font-medium mb-1">Password</label>
            <input
              type="password"
              name="password"
              required
              className="w-full border rounded p-2"
            />
          </div>

          <button
            type="submit"
            className="w-full bg-blue-600 text-white p-2 rounded font-medium"
          >
            Sign in
          </button>
        </Form>
      </div>
    </div>
  );
}
