import { useEffect, useRef } from "react";
import { useFetcher, useLoaderData, useParams, type ActionFunctionArgs, type LoaderFunctionArgs, type ShouldRevalidateFunctionArgs } from "react-router";
import { useChat } from "~/context/chat";
import { getChatMessages, sendMessage } from "~/utils/auth.server";

export async function loader({ request, params }: LoaderFunctionArgs): Promise<{ messages: Message[] }> {
  const chatId = params.chatId;
  const messages = await getChatMessages(request, chatId!);
  return { messages };
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

export function shouldRevalidate({ formAction }: ShouldRevalidateFunctionArgs) {
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
    <fetcher.Form method="post" className="flex border-t p-2">
      <input
        type="text"
        name="content"
        ref={inputRef}
        placeholder="Type a message..."
        className="flex-1 border rounded p-2"
      />
      <button
        type="submit"
        className="ml-2 bg-blue-600 text-white px-4 rounded"
      >
        Send
      </button>
    </fetcher.Form>
  );
}

export default function ChatDetail() {
  const { chatId } = useParams();
  if (!chatId) {
    return <div>No chat selected.</div>;
  }

  const { chatsMessages, setChatsMessages} = useChat();
  if (!setChatsMessages) {
    return <div>Loading...</div>;
  }
  
  const data = useLoaderData<typeof loader>();
  if (chatsMessages[chatId] === undefined) {
    setChatsMessages(prev => ({...prev, [chatId]: data.messages}));
  }

  const messages = chatsMessages[chatId];


  const bottomRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    bottomRef.current?.scrollIntoView({ behavior: "instant"});
  }, [messages])

  return (
    <div className="flex flex-col h-full">
      {/* Chat header */}
      <header className="border-b pb-2 mb-2 font-bold">
        Chat with 
      </header>

      {/* Messages */}
      <div className="flex-1 overflow-y-auto space-y-2 p-2">
        {messages === undefined || messages.length === 0 ? (
          <div>No messages yet.</div>
        ) : (
          messages.map((msg) => (
            <div key={msg.id} className="p-2 bg-gray-100 rounded">
              <div className="text-sm text-gray-600">{msg.chatId}</div>
              <div>{msg.content}</div>
              <div className="text-xs text-gray-500">{msg.sentAt}</div>
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
