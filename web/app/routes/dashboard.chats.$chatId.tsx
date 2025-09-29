import { useEffect, useRef } from "react";
import { useFetcher, useLoaderData, useParams, type ActionFunctionArgs, type LoaderFunctionArgs, type ShouldRevalidateFunctionArgs } from "react-router";
import { useChat } from "~/context/chat";
import { getChatMessages, sendMessage } from "~/utils/auth.server";

export async function loader({ request, params }: LoaderFunctionArgs): Promise<{ messages?: Message[], error?: string }> {
  const chatId = params.chatId;
  try {
    
    const messages = await getChatMessages(request, chatId!);
    return { messages };
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

  useEffect(() => {
    if (fetcher.state === "idle") {
      if (inputRef.current) {
        inputRef.current.value = "";
      }
    }
  }, [fetcher.state]);

  return (
    <fetcher.Form method="post" className="flex">
      <input
        type="text"
        name="content"
        ref={inputRef}
        placeholder="Type a message..."
        className=""
      />
      <button
        type="submit"
        className="send-button"
      >
        Send
      </button>
    </fetcher.Form>
  );
}

type MessageGroup = {
  senderId: number;
  messages: Message[];
}

export default function ChatDetail() {
  const { chatId } = useParams();
  if (!chatId) {
    return <div>No chatId selected.</div>;
  }
  markAsLoaded(chatId!);

  const data = useLoaderData<typeof loader>();
  if (data.error) {
    return <div>error: {data.error}</div>;
  }

  const { chatsMessages, setChatsMessages } = useChat();
  if (setChatsMessages === undefined) {
    return <div>setChatsMessages is not defined </div>;
  }

  const bottomRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    bottomRef.current?.scrollIntoView({ behavior: "auto"});
    
    if (chatsMessages[chatId] === undefined) {
      setChatsMessages((prev) => {
        return { ...prev, [chatId]: data.messages || [] };
      });
    }
  }, [chatId]);

  // Handle undefined chatMessages state. It must be after useEffect above.
  if (chatsMessages[chatId] === undefined) {
    return <div>chat messages are not loaded yet.</div>;
  }

  const messages = chatsMessages[chatId] || [];

  const groupedMessages: MessageGroup[] = [];
  let currentGroup: MessageGroup | null = null

  console.log(groupedMessages);

  messages.forEach((msg) => {
    // new speech bubble if:
    // 1. first message
    // 2. different sender
    if (!currentGroup || currentGroup.senderId !== msg.sender) {
      currentGroup = { senderId: msg.sender, messages: [msg] };
      groupedMessages.push(currentGroup);
      
    } else {
      // same sender, add to same speech bubble
      currentGroup.messages.push(msg);
    }
  });
  // push the last group if not already pushed
  if (currentGroup && groupedMessages[groupedMessages.length - 1] !== currentGroup) {
    groupedMessages.push(currentGroup);
  }



  return (
    <div className="w-full h-full flex flex-col">
      {/* Chat header */}
      <header className="chat-header">
        <div className="chat-header-left">
          <button>back</button>
          <h2 className="chat-name">Chat with {chatId}</h2>
        </div>
        <div className="chat-header-right">
          hello
        </div>
      </header>

      {/* Messages */}
      <div className="messages-container">
        {groupedMessages.length === 0 ? (
          <div>No messages yet. Start the conversation!</div>
        ) : (
          groupedMessages.map((group, idx) => (
            <div key={idx} className="messages-speech-bubble">
              {group.messages.map((msg) => (
                <div key={msg.id} className="">
                  {msg.content}
                </div>
              ))}
            </div>
          ))
        )}

        <div ref={bottomRef}></div>
      </div>

      {/* Input */}
      <div>
        <InputComponent />
      </div>

    </div>
  );
}

