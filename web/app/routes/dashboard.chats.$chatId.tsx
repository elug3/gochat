import { useEffect, useMemo, useRef, useState } from "react";
import { Link, useFetcher, useLoaderData, useNavigate, useOutlet, useParams, type ActionFunctionArgs, type LoaderFunctionArgs, type ShouldRevalidateFunctionArgs } from "react-router";
import { useUser } from "~/context/user";
import { getChatMessages, getGroupMembers, sendMessage } from "~/utils/auth.server";
import type { Member, Message } from "~/model";
import { useChatStore } from "~/store/chatStore";

type LoaderData = {
  messages?: Message[];
  members?: Member[];
  error?: string;
};

export async function loader({ request, params }: LoaderFunctionArgs): Promise<LoaderData> {
  const chatId = params.chatId;
  try {
    const [messages, members] = await Promise.all([
      getChatMessages(request, chatId!),
      getGroupMembers(request, chatId!)
    ]);
    return { messages, members };
  } catch (err: any) {
    return { error: err.message };
  }
}

export async function action({request, params}: ActionFunctionArgs) {
  const formdata = await request.formData();
  const content = formdata.get("content") as string;
  const chatId = params.chatId;

  const result = await sendMessage(request, {
    chatId: Number(chatId),
    content: content,
  });

  return {ok: true};
}

const loaded: Set<string> = new Set();

function isLoaded(chatId: string) {
  return loaded.has(chatId);
}

function markAsLoaded(chatId: string) {
  loaded.add(chatId);
}

export function shouldRevalidate({nextUrl}: ShouldRevalidateFunctionArgs) {
  if (nextUrl.pathname.includes("/dashboard/chats/")) {
    const chatId = nextUrl.pathname.split("/").pop();
    if (chatId && !isLoaded(chatId)) {
      return true;
    }
  }
  return false;
}

function InputComponent() {
  const fetcher = useFetcher();
  const inputRef = useRef<HTMLInputElement>(null);
  const isSubmitting = fetcher.state !== "idle";

  useEffect(() => {
    if (fetcher.state === "idle") {
      if (inputRef.current) {
        inputRef.current.value = "";
      }
    }
  }, [fetcher.state]);

  return (
  <fetcher.Form method="post" className="flex items-center gap-3">
      <div className="flex-1">
        <input
          type="text"
          name="content"
          ref={inputRef}
          placeholder="Message"
          autoComplete="off"
          className="w-full rounded-full border border-slate-200 bg-white px-4 py-2 text-sm shadow-sm outline-none transition focus:border-blue-500 focus:ring-2 focus:ring-blue-200 dark:border-slate-700 dark:bg-slate-900 dark:text-slate-100"
          disabled={isSubmitting}
        />
      </div>
      <button
        type="submit"
        className="inline-flex items-center gap-2 rounded-full bg-blue-600 px-4 py-2 text-sm font-medium text-white shadow-sm transition hover:bg-blue-700 disabled:cursor-not-allowed disabled:opacity-60"
        disabled={isSubmitting}
      >
        <svg
          className="h-4 w-4"
          fill="none"
          stroke="currentColor"
          strokeWidth="2"
          viewBox="0 0 24 24"
        >
          <path strokeLinecap="round" strokeLinejoin="round" d="M5 12l14-7-4 7 4 7-14-7z" />
        </svg>
        <span>{isSubmitting ? "Sending" : "Send"}</span>
      </button>
    </fetcher.Form>
  );
}

type MessageGroup = {
  senderId: number;
  messages: Message[];
}

function normalizeMessageSender(message: Message): Message {
  if (typeof (message as any).sender === "number") {
    return message;
  }

  const legacySender = (message as any).senderId;
  if (legacySender !== undefined) {
    return { ...message, sender: Number(legacySender) };
  }

  return message;
}

function sortMessages(messages: Message[]): Message[] {
  return [...messages].sort((a, b) => {
    if (a.sentAt === b.sentAt) {
      return a.id - b.id;
    }
    return a.sentAt - b.sentAt;
  });
}

function groupMessagesBySender(messages: Message[]): MessageGroup[] {
  const grouped: MessageGroup[] = [];
  let current: MessageGroup | null = null;

  for (const msg of messages) {
    if (!current || current.senderId !== msg.sender) {
      current = { senderId: msg.sender, messages: [msg] };
      grouped.push(current);
    } else {
      current.messages.push(msg);
    }
  }

  return grouped;
}

export default function ChatDetail() {
    const { chatId } = useParams();
    const loaderData = useLoaderData<typeof loader>();
    const navigate = useNavigate();
  const outlet = useOutlet();
    const { user } = useUser();
    const currentUserId = user?.id ?? null;
    const setChatMessages = useChatStore((s) => s.setChatMessages);

    const loaderMessages = useMemo(() => loaderData.messages ?? [], [loaderData.messages]);
    const members = loaderData.members ?? [];
    const [messages, setMessages] = useState<Message[]>(() => sortMessages(loaderMessages.map(normalizeMessageSender)));
    const [openMenuMessageId, setOpenMenuMessageId] = useState<number | null>(null);

    useEffect(() => {
      if (!chatId) return;
      const existing = useChatStore.getState().chatsMessages[chatId];
      const merged = new Map<number, Message>();
      const normalizedLoader = loaderMessages.map(normalizeMessageSender);
      const normalizedExisting = existing ? Object.values(existing).map(normalizeMessageSender) : [];

      normalizedLoader.forEach((msg) => merged.set(msg.id, msg));
      if (existing) {
        normalizedExisting.forEach((msg) => merged.set(msg.id, msg));
      }
      setChatMessages(chatId, Array.from(merged.values()));
    }, [chatId, loaderMessages, setChatMessages]);

    useEffect(() => {
      if (!chatId) return;

      const syncMessages = (messagesMap?: Record<number, Message>) => {
        const liveList = messagesMap ? Object.values(messagesMap) : loaderMessages;
        const normalized = liveList.map(normalizeMessageSender);
        setMessages(sortMessages(normalized));
      };

      syncMessages(useChatStore.getState().chatsMessages[chatId]);

      const unsubscribe = useChatStore.subscribe((state, prev) => {
        const next = state.chatsMessages[chatId];
        if (next === prev.chatsMessages[chatId]) return;
        syncMessages(next);
      });

      return unsubscribe;
    }, [chatId, loaderMessages]);

    const memberLookup = useMemo(() => {
      const lookup = new Map<number, Member>();
      for (const member of members) {
        const id = Number(member.userId);
        if (!Number.isNaN(id)) {
          lookup.set(id, member);
        }
      }
      return lookup;
    }, [members]);

    const groupedMessages = useMemo(() => groupMessagesBySender(messages), [messages]);
    const bottomRef = useRef<HTMLDivElement>(null);
    const messageListRef = useRef<HTMLDivElement>(null);

    useEffect(() => {
      if (chatId) {
        markAsLoaded(chatId);
      }
    }, [chatId]);

    useEffect(() => {
      bottomRef.current?.scrollIntoView({ behavior: "smooth" });
    }, [messages]);

    useEffect(() => {
      const handleClickOutside = (event: MouseEvent) => {
        if (!messageListRef.current?.contains(event.target as Node)) {
          setOpenMenuMessageId(null);
        }
      };
      document.addEventListener("mousedown", handleClickOutside);
      return () => document.removeEventListener("mousedown", handleClickOutside);
    }, []);

    if (outlet) {
      return outlet;
    }

    if (loaderData.error) {
      return (
        <div className="flex h-full flex-col items-center justify-center gap-4 bg-slate-50 p-6 text-center dark:bg-slate-950">
          <div className="rounded-full bg-red-100 p-3 text-red-600 dark:bg-red-900 dark:text-red-300">
            <svg className="h-6 w-6" fill="none" stroke="currentColor" strokeWidth="1.5" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" d="M12 9v3.75m-.008 3.6h.016v.2h-.016zM21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
            </svg>
          </div>
          <div>
            <h2 className="text-lg font-semibold text-slate-900 dark:text-slate-100">Unable to load chat</h2>
            <p className="mt-1 text-sm text-slate-600 dark:text-slate-400">{loaderData.error}</p>
          </div>
          <button
            onClick={() => navigate("/dashboard/chats")}
            className="rounded-full bg-blue-600 px-4 py-2 text-sm font-medium text-white shadow-sm transition hover:bg-blue-700"
          >
            Back to chats
          </button>
        </div>
      );
    }

    const participantLabel = (senderId: number) => {
      if (currentUserId != null && senderId === currentUserId) {
        return "You";
      }

      const member = memberLookup.get(senderId);
      if (member) {
        return member.userId.toString();
      }

      return `User ${senderId}`;
    };

    const formatTime = (timestamp: number) => {
      if (!timestamp) return "";
      return new Intl.DateTimeFormat(undefined, {
        hour: "numeric",
        minute: "2-digit"
      }).format(new Date(timestamp));
    };

    return (
      <div className="flex h-full w-full flex-col bg-slate-50 text-slate-900 dark:bg-slate-950 dark:text-slate-50">
        <header className="flex items-center justify-between border-b border-slate-200 bg-white px-4 py-3 backdrop-blur-lg dark:border-slate-800 dark:bg-slate-900">
          <div className="flex items-center gap-3">
            <button
              onClick={() => navigate("/dashboard/chats")}
              className="inline-flex h-9 w-9 items-center justify-center rounded-full border border-slate-200 bg-white text-slate-600 shadow-sm transition hover:border-blue-500 hover:text-blue-600 dark:border-slate-800 dark:bg-slate-950 dark:text-slate-300 dark:hover:border-blue-400 dark:hover:text-blue-300"
              aria-label="Back to chats"
            >
              <svg className="h-4 w-4" fill="none" stroke="currentColor" strokeWidth="2" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" d="M15 19l-7-7 7-7" />
              </svg>
            </button>
            <div>
              <div className="flex items-center gap-2">
                <h2 className="text-base font-semibold leading-tight">Chat {chatId}</h2>
                <span className="rounded-full border border-slate-200 px-2 py-0.5 text-xs font-medium text-slate-500 dark:border-slate-700 dark:text-slate-400">
                  {members.length} participants
                </span>
              </div>
              <p className="text-xs text-slate-500 dark:text-slate-400">{messages.length} messages</p>
            </div>
          </div>
          <div className="flex items-center gap-2">
            {chatId && (
              <Link
                to={`/dashboard/chats/${chatId}/settings`}
                className="inline-flex items-center gap-2 rounded-full border border-slate-200 px-3 py-1.5 text-xs font-medium text-slate-600 transition hover:border-blue-500 hover:text-blue-600 dark:border-slate-800 dark:text-slate-300 dark:hover:border-blue-400 dark:hover:text-blue-200"
              >
                <svg className="h-4 w-4" fill="none" stroke="currentColor" strokeWidth="1.5" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" d="M12 3c2.755 0 5 2.254 5 5.033 0 2.171-1.375 4.012-3.3 4.722l-.7 5.545-1.022.7-1.978-1.4-.7-4.845C8.375 12.045 7 10.204 7 8.032 7 5.255 9.245 3 12 3z" />
                </svg>
                Settings
              </Link>
            )}
          </div>
        </header>

        <main className="flex-1 overflow-y-auto px-4 py-6">
          {groupedMessages.length === 0 ? (
            <div className="flex h-full flex-col items-center justify-center gap-3 text-center text-sm text-slate-500 dark:text-slate-400">
              <div className="rounded-full bg-slate-200 p-3 text-slate-500 dark:bg-slate-800 dark:text-slate-300">
                <svg className="h-6 w-6" fill="none" stroke="currentColor" strokeWidth="1.5" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" d="M12 19l-7 2 2-7 9-9a2.828 2.828 0 114 4l-9 9z" />
                </svg>
              </div>
              <div>
                <p className="font-medium text-slate-600 dark:text-slate-200">No messages yet</p>
                <p className="text-xs text-slate-500 dark:text-slate-400">Send the first message to kick things off.</p>
              </div>
            </div>
          ) : (
            <div className="flex flex-col gap-3" ref={messageListRef}>
              {groupedMessages.map((group) => {
                const isOwn = currentUserId != null && group.senderId === currentUserId;
                const bubbleAlignment = isOwn ? "items-end" : "items-start";
                const bubbleBase = "w-full max-w-full rounded-lg px-3 py-2 text-[13px] leading-snug shadow-sm ring-1 ring-slate-200 dark:ring-slate-800 select-text cursor-text flex items-center gap-3 justify-between";
                const bubbleTone = isOwn
                  ? "bg-blue-600 text-white"
                  : "bg-slate-100 text-slate-900 dark:bg-slate-900 dark:text-slate-100";
                const bubbleMeta = isOwn ? "text-blue-100" : "text-slate-400 dark:text-slate-400";

                return (
                  <div key={`${group.senderId}-${group.messages[0]?.id ?? "group"}`} className={`flex w-full flex-col gap-1.5 ${bubbleAlignment}`}>
                    <div className="flex items-center gap-2 text-[11px] uppercase tracking-[0.08em] text-slate-400 dark:text-slate-500">
                      <span>{participantLabel(group.senderId)}</span>
                    </div>
                    <div className="flex flex-col gap-1.5">
                      {group.messages.map((msg) => {
                        const isMenuOpen = openMenuMessageId === msg.id;
                        return (
                          <div key={msg.id} className="group relative flex w-full items-center gap-2">
                            <div className={`${bubbleBase} ${bubbleTone}`}>
                              <p className="whitespace-pre-wrap flex-1 text-left">{msg.content}</p>
                              {msg.sentAt && (
                                <span className={`text-[11px] ${bubbleMeta} whitespace-nowrap`}>
                                  {formatTime(msg.sentAt)}
                                </span>
                              )}
                            </div>
                            <button
                              type="button"
                              aria-label="Message options"
                              onClick={(e) => {
                                e.stopPropagation();
                                setOpenMenuMessageId(isMenuOpen ? null : msg.id);
                              }}
                              className={`h-7 w-7 rounded-full text-xs font-semibold transition opacity-0 group-hover:opacity-100 ${isOwn ? "hover:bg-blue-700/30" : "hover:bg-slate-200 dark:hover:bg-slate-800"} select-none`}
                            >
                              ⋯
                            </button>
                            {isMenuOpen && (
                              <div className="absolute right-0 top-full z-10 mt-1 w-32 rounded-md border border-slate-200 bg-white text-sm shadow-lg dark:border-slate-800 dark:bg-slate-900">
                                {["Copy", "Reply", "Delete"].map((label) => (
                                  <button
                                    key={label}
                                    type="button"
                                    onClick={() => setOpenMenuMessageId(null)}
                                    className="flex w-full items-center px-3 py-2 text-left text-slate-700 hover:bg-slate-100 dark:text-slate-200 dark:hover:bg-slate-800"
                                  >
                                    {label}
                                  </button>
                                ))}
                              </div>
                            )}
                          </div>
                        );
                      })}
                    </div>
                  </div>
                );
              })}
            </div>
          )}

          <div ref={bottomRef} />
        </main>

        <footer className="border-t border-slate-200 bg-white px-4 py-3 backdrop-blur-lg dark:border-slate-800 dark:bg-slate-900">
          <InputComponent />
        </footer>
      </div>
  );
}


// export default function ChatDetail() {
//   const [error, setError] = useState<string | null>(null);
//   const { chatId } = useParams();
//   if (!chatId) {
//     return <div>No chat selected.</div>;
//   }
//   markAsLoaded(chatId!);

//   const data = useLoaderData<typeof loader>();
//   if (data.error) {
//     setError(data.error);
//   }

//   const chatsMessages  = useChatStore((s) => s.chatsMessages);
//   const setChatMessages = useChatStore((s) => s.setChatMessages);

//   const bottomRef = useRef<HTMLDivElement>(null);

//   useEffect(() => {
//     bottomRef.current?.scrollIntoView({ behavior: "auto"});
    
//     if (chatsMessages[chatId] === undefined) {
//       setChatMessages(chatId, data.messages || []);
//     }
//   }, [chatId]);

//   // Handle undefined chatMessages state. It must be after useEffect above.
//   if (chatsMessages[chatId] === undefined) {
//     setError("Chat messages are not loaded yet.");
//   }

//   const messages = chatsMessages[chatId] || [];


//   const [cardProfile, setCardProfile] = useState<Profile | null>(null);
//   function openProfileCard(profile: Profile) {
//     setCardProfile(profile);
//   }
//   function closeProfileCard() {
//     setCardProfile(null);
//   }

//   return (
//     <div className="w-full h-full flex flex-col">
//       {/* Chat header */}
//       <header className="chat-header">
//         <div className="chat-header-left">
//           <button>back</button>
//           <h2 className="chat-name">Chat with {chatId}</h2>
//         </div>
//         <div className="chat-header-right">
//           hello
//         </div>
//       </header>

//       {/* Messages */}
//         <div className="messages-container">
//           {groupedMessages.length === 0 ? (
//             <div>No messages yet. Start the conversation!</div>
//           ) : (
//             groupedMessages.map((group, idx) => (
//               <SpeechBubble key={idx} messageGroup={group} openProfileCard={openProfileCard} />
//             )
//           ))}

//           <div ref={bottomRef}></div>
//         </div>

//         {/* Input */}
//         <div>
//           <InputComponent />
//         </div>

//         {/* Profile card */}
//         <div>
//           {cardProfile && <ProfileCard/>}
//         </div>
//       </div>
//   );
// }
