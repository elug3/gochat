import { Form, useActionData, useNavigation, type ActionFunctionArgs } from "react-router";
import { useEffect, useRef } from "react";
import { sendContactRequest } from "~/utils/auth.server";

type ActionResponse = {
	ok: boolean;
	error?: string;
};

export async function action({ request }: ActionFunctionArgs): Promise<ActionResponse> {
	const formData = await request.formData();
	const identifier = formData.get("identifier");

	if (typeof identifier !== "string" || identifier.trim().length === 0) {
		return { ok: false, error: "Please enter a username or email" };
	}

	try {
		await sendContactRequest(request, identifier.trim());
		return { ok: true };
	} catch (err: any) {
		return { ok: false, error: err.message ?? "Failed to send contact request" };
	}
}

export default function NewContactRoute() {
	const actionData = useActionData<typeof action>();
	const navigation = useNavigation();
	const isSubmitting = navigation.state === "submitting";

	const formRef = useRef<HTMLFormElement>(null);
	const inputRef = useRef<HTMLInputElement>(null);

	useEffect(() => {
		if (actionData?.ok) {
			formRef.current?.reset();
			inputRef.current?.focus();
		}
	}, [actionData]);

	return (
		<div className="card">
			<h3 className="text-lg font-semibold mb-2 text-center text-slate-900 dark:text-slate-100">Add a new contact</h3>
			<p className="text-sm text-slate-600 dark:text-slate-400 mb-6 text-center">
				Send a contact request by entering their username or email address.
			</p>

					<Form
						method="post"
						replace
						ref={formRef}
						className="space-y-4"
					>
				<div>
					<label htmlFor="identifier" className="block text-sm font-medium mb-2 text-slate-700 dark:text-slate-300">
						Username or email
					</label>
					<input
						id="identifier"
						name="identifier"
						type="text"
						ref={inputRef}
						placeholder="jane@example.com"
						className="w-full px-3 py-2 border border-slate-300 dark:border-slate-700 rounded-md bg-white dark:bg-slate-950 text-slate-900 dark:text-slate-100 focus:outline-none focus:ring-2 focus:ring-blue-500 dark:focus:ring-blue-400 focus:border-transparent"
						disabled={isSubmitting}
						required
					/>
				</div>

				{actionData?.error && (
					<div className="text-sm text-red-600 dark:text-red-400 bg-red-50 dark:bg-red-900 border border-red-200 dark:border-red-800 rounded-md p-3" role="alert">
						{actionData.error}
					</div>
				)}

				{actionData?.ok && !actionData.error && (
					<div className="text-sm text-green-600 dark:text-green-400 bg-green-50 dark:bg-green-900 border border-green-200 dark:border-green-800 rounded-md p-3" role="status">
						Contact request sent successfully.
					</div>
				)}

				<button
					type="submit"
					disabled={isSubmitting}
					className="w-full bg-blue-500 dark:bg-blue-600 text-white py-2 px-4 rounded-md hover:bg-blue-600 dark:hover:bg-blue-700 disabled:bg-slate-400 dark:disabled:bg-slate-600 disabled:cursor-not-allowed transition-colors"
				>
					{isSubmitting ? "Sending..." : "Send Request"}
				</button>
			</Form>
		</div>
	);
}
