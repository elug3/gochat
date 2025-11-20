import { Form, Link, redirect, useActionData, useLoaderData, useNavigation, type ActionFunctionArgs, type LoaderFunctionArgs } from "react-router";
import { useMemo } from "react";
import { addChatParticipant, deleteChat, fetchChatList, getGroupMembers, removeChatParticipant, updateChatName } from "~/utils/auth.server";
import type { Chat, Member } from "~/model";

interface LoaderData {
  chat?: Chat;
  members: Member[];
  error?: string;
}

interface ActionData {
  ok: boolean;
  message?: string;
  error?: string;
}

export async function loader({ request, params }: LoaderFunctionArgs): Promise<LoaderData> {
  const chatId = params.chatId;

  if (!chatId) {
    return { members: [], error: "Chat not found" };
  }

  try {
    const [chats, members] = await Promise.all([
      fetchChatList(request),
      getGroupMembers(request, chatId)
    ]);

    const chat = chats.find((item) => item.id === Number(chatId));

    if (!chat) {
      return { members, error: "Chat not found" };
    }

    return { chat, members };
  } catch (error: any) {
    return { members: [], error: error?.message ?? "Failed to load chat settings" };
  }
}

export async function action({ request, params }: ActionFunctionArgs): Promise<ActionData> {
  const chatId = params.chatId;

  if (!chatId) {
    return { ok: false, error: "Chat not found" };
  }

  const formData = await request.formData();
  const intent = formData.get("_action");

  try {
    switch (intent) {
      case "rename": {
        const name = (formData.get("name") || "").toString().trim();
        if (!name) {
          return { ok: false, error: "Chat name cannot be empty" };
        }
        await updateChatName(request, chatId, name);
        return { ok: true, message: "Chat name updated" };
      }
      case "add-participant": {
        const participantId = (formData.get("participantId") || "").toString().trim();
        if (!participantId) {
          return { ok: false, error: "Participant ID is required" };
        }
        await addChatParticipant(request, chatId, participantId);
        return { ok: true, message: "Participant added" };
      }
      case "remove-participant": {
        const participantId = (formData.get("participantId") || "").toString().trim();
        if (!participantId) {
          return { ok: false, error: "Participant ID is required" };
        }
        await removeChatParticipant(request, chatId, participantId);
        return { ok: true, message: "Participant removed" };
      }
      case "delete-chat": {
        await deleteChat(request, chatId);
        return redirect("/dashboard/chats");
      }
      default:
        return { ok: false, error: "Unsupported action" };
    }
  } catch (error: any) {
    return { ok: false, error: error?.message ?? "Something went wrong" };
  }
}

export default function ChatSettingsPage() {
  const { chat, members, error } = useLoaderData<typeof loader>();
  const actionData = useActionData<ActionData>();
  const navigation = useNavigation();
  const isSubmitting = navigation.state === "submitting";

  const sortedMembers = useMemo(() => {
    return [...members].sort((a, b) => a.userId.localeCompare(b.userId));
  }, [members]);

  if (error) {
    return (
      <div className="flex h-full flex-col items-center justify-center gap-4 bg-slate-50 p-6 text-center dark:bg-slate-950">
        <div className="rounded-full bg-red-100 p-3 text-red-600 dark:bg-red-900/30 dark:text-red-300">
          <svg className="h-6 w-6" fill="none" stroke="currentColor" strokeWidth="1.5" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" d="M12 9v3.75m-.008 3.6h.016v.2h-.016zM21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
          </svg>
        </div>
        <div>
          <h2 className="text-lg font-semibold text-slate-900 dark:text-slate-100">Unable to load settings</h2>
          <p className="mt-1 text-sm text-slate-600 dark:text-slate-400">{error}</p>
        </div>
        <Link
          to="/dashboard/chats"
          className="rounded-full bg-blue-600 px-4 py-2 text-sm font-medium text-white shadow-sm transition hover:bg-blue-700"
        >
          Back to chats
        </Link>
      </div>
    );
  }

  if (!chat) {
    return null;
  }

  return (
    <div className="flex h-full flex-col gap-6 bg-slate-50/70 px-6 py-6 text-slate-900 dark:bg-slate-950 dark:text-slate-100">
      <header className="flex items-center justify-between">
        <div>
          <p className="text-[11px] font-semibold uppercase tracking-[0.2em] text-slate-500/70 dark:text-slate-400/70">Chat settings</p>
          <h1 className="mt-1 text-2xl font-semibold">{chat.name}</h1>
        </div>
        <Link
          to={`/dashboard/chats/${chat.id}`}
          className="inline-flex items-center gap-2 rounded-full border border-slate-200 px-3 py-1.5 text-sm font-medium text-slate-600 transition hover:border-blue-500 hover:text-blue-600 dark:border-slate-800 dark:text-slate-300 dark:hover:border-blue-400 dark:hover:text-blue-200"
        >
          <svg className="h-4 w-4" fill="none" stroke="currentColor" strokeWidth="1.5" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" d="M15 19l-7-7 7-7" />
          </svg>
          Back to chat
        </Link>
      </header>

      {actionData?.error && (
        <div className="rounded-xl border border-rose-200 bg-rose-50 px-4 py-3 text-sm text-rose-700 dark:border-rose-900/40 dark:bg-rose-900/20 dark:text-rose-200">
          {actionData.error}
        </div>
      )}

      {actionData?.ok && actionData.message && actionData.message !== "Chat deleted" && (
        <div className="rounded-xl border border-emerald-200 bg-emerald-50 px-4 py-3 text-sm text-emerald-700 dark:border-emerald-900/40 dark:bg-emerald-900/20 dark:text-emerald-200">
          {actionData.message}
        </div>
      )}

      <section className="rounded-xl border border-slate-200/70 bg-white/70 p-5 shadow-sm dark:border-slate-800 dark:bg-slate-900/60">
        <h2 className="text-sm font-semibold text-slate-700 dark:text-slate-200">Chat name</h2>
        <p className="mt-1 text-xs text-slate-500 dark:text-slate-400">Update the chat title to help members recognize this conversation.</p>
        <Form method="post" className="mt-4 flex flex-col gap-3">
          <input type="hidden" name="_action" value="rename" />
          <div>
            <input
              type="text"
              name="name"
              defaultValue={chat.name}
              placeholder="Team hangout"
              className="w-full rounded-lg border border-slate-200 bg-white px-3 py-2 text-sm shadow-sm outline-none transition focus:border-blue-500 focus:ring-2 focus:ring-blue-200 dark:border-slate-700 dark:bg-slate-950 dark:text-slate-100"
              disabled={isSubmitting}
            />
          </div>
          <div className="flex items-center justify-end gap-2">
            <button
              type="submit"
              className="inline-flex items-center gap-2 rounded-lg bg-blue-600 px-4 py-2 text-sm font-medium text-white shadow-sm transition hover:bg-blue-700 disabled:cursor-not-allowed disabled:opacity-60"
              disabled={isSubmitting}
            >
              Save changes
            </button>
          </div>
        </Form>
      </section>

      <section className="rounded-xl border border-slate-200/70 bg-white/70 p-5 shadow-sm dark:border-slate-800 dark:bg-slate-900/60">
        <div className="flex items-center justify-between">
          <div>
            <h2 className="text-sm font-semibold text-slate-700 dark:text-slate-200">Participants</h2>
            <p className="mt-1 text-xs text-slate-500 dark:text-slate-400">Manage who is part of this chat.</p>
          </div>
          <span className="rounded-full border border-slate-200 px-2 py-0.5 text-[11px] font-medium text-slate-500 dark:border-slate-700 dark:text-slate-400">
            {members.length}
          </span>
        </div>

        <div className="mt-4 space-y-3">
          {sortedMembers.length === 0 ? (
            <p className="rounded-lg border border-dashed border-slate-300/70 bg-white/60 px-3 py-4 text-center text-sm text-slate-500 dark:border-slate-700 dark:bg-slate-900/50 dark:text-slate-400">
              No participants found.
            </p>
          ) : (
            <ul className="space-y-2">
              {sortedMembers.map((member) => (
                <li key={member.userId} className="flex items-center justify-between rounded-lg border border-slate-200 bg-white px-3 py-2 text-sm shadow-sm dark:border-slate-800 dark:bg-slate-950/80">
                  <div>
                    <p className="font-medium text-slate-700 dark:text-slate-200">User {member.userId}</p>
                    <p className="text-xs text-slate-500 dark:text-slate-400 capitalize">{member.role}</p>
                  </div>
                  <Form method="post" className="flex items-center" replace>
                    <input type="hidden" name="_action" value="remove-participant" />
                    <input type="hidden" name="participantId" value={member.userId} />
                    <button
                      type="submit"
                      className="inline-flex items-center gap-1 rounded-full border border-slate-200 px-3 py-1 text-xs font-medium text-slate-500 transition hover:border-rose-500 hover:text-rose-600 dark:border-slate-700 dark:text-slate-300 dark:hover:border-rose-400 dark:hover:text-rose-300"
                      disabled={isSubmitting || member.role === "owner"}
                    >
                      Remove
                    </button>
                  </Form>
                </li>
              ))}
            </ul>
          )}
        </div>

        <Form method="post" className="mt-5 flex flex-col gap-3">
          <input type="hidden" name="_action" value="add-participant" />
          <div className="flex items-center gap-3">
            <input
              type="text"
              name="participantId"
              placeholder="Enter user ID"
              className="flex-1 rounded-lg border border-slate-200 bg-white px-3 py-2 text-sm shadow-sm outline-none transition focus:border-blue-500 focus:ring-2 focus:ring-blue-200 dark:border-slate-700 dark:bg-slate-950 dark:text-slate-100"
              disabled={isSubmitting}
            />
            <button
              type="submit"
              className="inline-flex items-center gap-2 rounded-lg bg-slate-900 px-4 py-2 text-sm font-medium text-white shadow-sm transition hover:bg-slate-700 disabled:cursor-not-allowed disabled:opacity-60 dark:bg-slate-200 dark:text-slate-900 dark:hover:bg-slate-100"
              disabled={isSubmitting}
            >
              Add
            </button>
          </div>
        </Form>
      </section>

      <section className="rounded-xl border border-rose-200/60 bg-rose-50/80 p-5 shadow-sm dark:border-rose-900/40 dark:bg-rose-900/20">
        <h2 className="text-sm font-semibold text-rose-700 dark:text-rose-200">Danger zone</h2>
        <p className="mt-1 text-xs text-rose-600 dark:text-rose-300">Deleting this chat will remove all messages for every participant. This cannot be undone.</p>
        <Form
          method="post"
          className="mt-4"
          onSubmit={(event) => {
            if (!confirm("Are you sure you want to delete this chat?")) {
              event.preventDefault();
            }
          }}
        >
          <input type="hidden" name="_action" value="delete-chat" />
          <button
            type="submit"
            className="inline-flex items-center gap-2 rounded-lg bg-rose-600 px-4 py-2 text-sm font-medium text-white shadow-sm transition hover:bg-rose-700 disabled:cursor-not-allowed disabled:opacity-60"
            disabled={isSubmitting}
          >
            Delete chat
          </button>
        </Form>
      </section>
    </div>
  );
}
