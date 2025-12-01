import { Form, useActionData, useNavigate, useNavigation, useSearchParams, useLoaderData, type ActionFunctionArgs, type LoaderFunctionArgs } from "react-router";
import { useEffect, useRef, useState } from "react";
import { createGroup, createDirectChat, fetchContacts } from "~/utils/auth.server";
import { useChatStore } from "~/store/chatStore";
import type { Chat, Group, Profile } from "~/model";

type LoaderData = {
	contacts?: Profile[];
	error?: string;
};

type ActionResponse = {
	ok: boolean;
	group?: Group;
	chat?: Chat;
	error?: string;
};

export async function loader({ request }: LoaderFunctionArgs): Promise<LoaderData> {
	try {
		const contacts = await fetchContacts(request);
		return { contacts };
	} catch (err: any) {
		return { error: err.message };
	}
}

export async function action({ request }: ActionFunctionArgs): Promise<ActionResponse> {
	const formData = await request.formData();
	const chatType = formData.get("chatType");
	
	try {
		if (chatType === "direct") {
			// Handle direct chat creation
			const contactId = formData.get("contactId");
			if (!contactId || typeof contactId !== "string") {
				return { ok: false, error: "Please select a contact" };
			}

			const chat = await createDirectChat(request, { contactId });
			return { ok: true, chat };
		} else {
			// Handle group creation
			const name = formData.get("name");
			const trimmedName = typeof name === "string" ? name.trim() : "";

			const group = await createGroup(request, { name: trimmedName });
			return { ok: true, group };
		}
	} catch (err: any) {
		return { ok: false, error: err?.message ?? "Failed to create chat" };
	}
}

export default function NewChatPage() {
	const actionData = useActionData<typeof action>();
	const loaderData = useLoaderData<typeof loader>();
	const navigation = useNavigation();
	const navigate = useNavigate();
	const [searchParams] = useSearchParams();
	const [selectedContact, setSelectedContact] = useState<string>("");
	const upsertChat = useChatStore((state) => state.upsertChat);
	const lastNavigatedChatId = useRef<number | null>(null);

	const chatType = searchParams.get("type") || "group"; // default to group if no type specified
	const isDirectChat = chatType === "direct";
	const isGroupChat = chatType === "group";

	const contacts = loaderData?.contacts || [];

	const isSubmitting = navigation.state === "submitting";

	useEffect(() => {
		if (actionData?.ok) {
			// Handle group creation
			if (actionData.group) {
				if (lastNavigatedChatId.current === actionData.group.id) return;

				// Convert Group to Chat format for the chat store
				const chat: Chat = {
					id: actionData.group.id,
					name: actionData.group.name,
					lastMessage: "",
					lastMessageAt: actionData.group.createdAt,
					unreadCount: 0
				};

				upsertChat(chat);
				lastNavigatedChatId.current = actionData.group.id;
				navigate(`/dashboard/chats/${actionData.group.id}`);
			}
			// Handle direct chat creation
			else if (actionData.chat) {
				if (lastNavigatedChatId.current === actionData.chat.id) return;

				upsertChat(actionData.chat);
				lastNavigatedChatId.current = actionData.chat.id;
				navigate(`/dashboard/chats/${actionData.chat.id}`);
			}
		}
	}, [actionData, navigate, upsertChat]);

	// Handle loader errors
	if (loaderData?.error) {
		return (
			<div className="page">
				<div className="w-full max-w-xl mx-auto">
					<div className="card">
						<div className="text-center">
							<h1 className="text-xl font-semibold mb-2 text-slate-900 dark:text-slate-100">Error</h1>
							<p className="text-red-600 dark:text-red-400 mb-4">{loaderData.error}</p>
							<button
								onClick={() => navigate('/dashboard/chats')}
								className="px-4 py-2 bg-slate-500 dark:bg-slate-600 text-white rounded-md hover:bg-slate-600 dark:hover:bg-slate-700 transition-colors"
							>
								Go Back
							</button>
						</div>
					</div>
				</div>
			</div>
		);
	}

	return (
		<div className="page">
			<div className="w-full max-w-xl mx-auto">
				<div className="card">
					<h1 className="text-xl font-semibold text-center mb-2 text-slate-900 dark:text-slate-100">
						{isDirectChat ? "Start a Direct Chat" : "Create a New Group"}
					</h1>
					<p className="text-sm text-slate-600 dark:text-slate-400 text-center mb-6">
						{isDirectChat 
							? "Select a contact to start a private conversation with."
							: "Give your group a name. You can invite people later from the chat view."
						}
					</p>

					<Form method="post" className="space-y-6" replace>
						{/* Hidden input to pass chat type */}
						<input type="hidden" name="chatType" value={chatType} />

						{isDirectChat ? (
							/* Direct Chat - Contact Selection */
							<div>
								<label htmlFor="contact" className="block text-sm font-medium mb-1 text-slate-700 dark:text-slate-300">
									Select Contact
								</label>
								{contacts.length > 0 ? (
									<select
										id="contact"
										name="contactId"
										value={selectedContact}
										onChange={(e) => setSelectedContact(e.target.value)}
										className="w-full px-3 py-2 border border-slate-300 dark:border-slate-700 rounded-md bg-white dark:bg-slate-950 text-slate-900 dark:text-slate-100 focus:outline-none focus:ring-2 focus:ring-blue-500 dark:focus:ring-blue-400 focus:border-transparent"
										disabled={isSubmitting}
										required
									>
										<option value="">Choose a contact...</option>
										{contacts.map((contact) => (
											<option key={contact.userId} value={contact.userId}>
												{contact.name}
												{contact.status === "online" && " • Online"}
											</option>
										))}
									</select>
								) : (
									<div className="text-center py-8">
										<p className="text-slate-500 dark:text-slate-400 mb-4">No contacts available</p>
										<button
											type="button"
											onClick={() => navigate('/dashboard/contacts')}
											className="text-blue-600 dark:text-blue-400 hover:text-blue-800 dark:hover:text-blue-300 underline"
										>
											Add contacts first
										</button>
									</div>
								)}
							</div>
						) : (
							/* Group Chat - Name Input */
							<div>
								<label htmlFor="name" className="block text-sm font-medium mb-1 text-slate-700 dark:text-slate-300">
									Group name <span className="text-slate-400 dark:text-slate-500">(optional)</span>
								</label>
								<input
									id="name"
									name="name"
									type="text"
									placeholder="Team hangout"
									className="w-full px-3 py-2 border border-slate-300 dark:border-slate-700 rounded-md bg-white dark:bg-slate-950 text-slate-900 dark:text-slate-100 focus:outline-none focus:ring-2 focus:ring-blue-500 dark:focus:ring-blue-400 focus:border-transparent"
									disabled={isSubmitting}
								/>
								<p className="text-xs text-slate-500 dark:text-slate-400 mt-1">
									Leave blank to use a default name.
								</p>
							</div>
						)}

						{actionData?.error && (
							<div className="text-sm text-red-600 dark:text-red-400 bg-red-50 dark:bg-red-900 border border-red-200 dark:border-red-800 rounded-md p-3" role="alert">
								{actionData.error}
							</div>
						)}

						<button
							type="submit"
							className="w-full py-2 px-4 rounded-md bg-blue-500 dark:bg-blue-600 text-white font-medium hover:bg-blue-600 dark:hover:bg-blue-700 disabled:bg-slate-400 dark:disabled:bg-slate-600 disabled:cursor-not-allowed transition-colors"
							disabled={isSubmitting || (isDirectChat && !selectedContact)}
						>
							{isSubmitting 
								? (isDirectChat ? "Starting chat..." : "Creating group...") 
								: (isDirectChat ? "Start Chat" : "Create Group")
							}
						</button>
					</Form>
				</div>
			</div>
		</div>
	);
}
