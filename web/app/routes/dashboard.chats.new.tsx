import { Form, useActionData, useNavigate, useNavigation, type ActionFunctionArgs } from "react-router";
import { useEffect, useRef } from "react";
import { createChat } from "~/utils/auth.server";
import { useChatStore } from "~/store/chatStore";
import type { Chat } from "~/model";

type ActionResponse = {
	ok: boolean;
	chat?: Chat;
	error?: string;
};

export async function action({ request }: ActionFunctionArgs): Promise<ActionResponse> {
	const formData = await request.formData();
	const name = formData.get("name");
	const trimmedName = typeof name === "string" ? name.trim() : "";

	try {
		const chat = await createChat(request, {
			name: trimmedName.length > 0 ? trimmedName : undefined,
		});
		return { ok: true, chat };
	} catch (err: any) {
		return { ok: false, error: err?.message ?? "Failed to create chat" };
	}
}

export default function NewChatPage() {
	const actionData = useActionData<typeof action>();
	const navigation = useNavigation();
	const navigate = useNavigate();
	const upsertChat = useChatStore((state) => state.upsertChat);
	const lastNavigatedChatId = useRef<number | null>(null);

	const isSubmitting = navigation.state === "submitting";

	useEffect(() => {
		if (actionData?.ok && actionData.chat) {
			if (lastNavigatedChatId.current === actionData.chat.id) return;

			upsertChat(actionData.chat);
			lastNavigatedChatId.current = actionData.chat.id;
			navigate(`/dashboard/chats/${actionData.chat.id}`);
		}
	}, [actionData, navigate, upsertChat]);

	return (
		<div className="page">
			<div className="w-full max-w-xl mx-auto">
				<div className="card">
					<h1 className="text-xl font-semibold text-center mb-2">Start a new chat</h1>
					<p className="text-sm text-gray-600 text-center mb-6">
						Give your conversation a name. You can invite people later from the chat view.
					</p>

					<Form method="post" className="space-y-6" replace>
						<div>
							<label htmlFor="name" className="block text-sm font-medium mb-1">
								Chat name <span className="text-gray-400">(optional)</span>
							</label>
							<input
								id="name"
								name="name"
								type="text"
								placeholder="Team hangout"
								className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
								disabled={isSubmitting}
							/>
							<p className="text-xs text-gray-500 mt-1">
								Leave blank to use a default name.
							</p>
						</div>

						{actionData?.error && (
							<div className="text-sm text-red-600" role="alert">
								{actionData.error}
							</div>
						)}

						<button
							type="submit"
							className="w-full py-2 px-4 rounded-md bg-blue-500 text-white font-medium hover:bg-blue-600 disabled:bg-gray-400 disabled:cursor-not-allowed transition-colors"
							disabled={isSubmitting}
						>
							{isSubmitting ? "Creating chat..." : "Create chat"}
						</button>
					</Form>
				</div>
			</div>
		</div>
	);
}