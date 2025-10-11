import { Link } from "react-router";
import { useMemo } from "react";
import { useChatStore } from "~/store/chatStore";

export default function ChatsIndexRoute() {
	const chatsRecord = useChatStore((state) => state.chats);

	const chats = useMemo(() => Object.values(chatsRecord || {}), [chatsRecord]);

	if (chats.length === 0) {
		return (
			<div className="w-full h-full flex items-center justify-center p-6">
				<div className="card max-w-md text-center space-y-4">
					<h2 className="text-lg font-semibold">No chats yet</h2>
					<p className="text-sm text-gray-600">
						Add a contact or start a new conversation to see it here.
					</p>
					<Link
						to="new"
						className="inline-flex items-center justify-center px-4 py-2 rounded-md bg-blue-500 text-white text-sm font-medium hover:bg-blue-600 transition-colors"
					>
						Start a chat
					</Link>
				</div>
			</div>
		);
	}

	return (
		<div className="w-full h-full flex items-center justify-center p-6">
			<div className="card w-full max-w-xl space-y-4">
				<header className="text-center space-y-1">
					<h2 className="text-lg font-semibold">Select a chat to view messages</h2>
					<p className="text-sm text-gray-600">Quick recap of your recent conversations</p>
				</header>

				<ul className="divide-y divide-gray-200 border border-gray-200 rounded-md">
					{chats.map((chat) => (
						<li key={chat.id}>
							<Link
								to={`${chat.id}`}
								className="flex gap-3 items-center px-4 py-3 hover:bg-gray-50 transition-colors"
							>
								<div className="flex-1 min-w-0">
									<div className="flex items-center justify-between gap-2">
										<span className="text-sm font-medium truncate">{chat.name}</span>
										{chat.unreadCount > 0 && (
											<span className="text-xs font-semibold text-blue-600">{chat.unreadCount} unread</span>
										)}
									</div>
									<p className="text-xs text-gray-500 truncate mt-1">
										{chat.lastMessage || "No messages yet"}
									</p>
								</div>
								<span className="text-xs text-gray-400">View</span>
							</Link>
						</li>
					))}
				</ul>
			</div>
		</div>
	);
}
