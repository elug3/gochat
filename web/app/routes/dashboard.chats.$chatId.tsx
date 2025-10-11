import type { H } from "node_modules/react-router/dist/development/context-CIdFp11b.mjs";
import { useEffect, useRef, useState, type MouseEvent as ReactMouseEvent } from "react";
import { useFetcher, useLoaderData, useParams, type ActionFunctionArgs, type LoaderFunctionArgs, type ShouldRevalidateFunctionArgs } from "react-router";
import { getChatMessages, sendMessage } from "~/utils/auth.server";
import type { Member, Message, Profile } from "~/model";

export async function loader({ request, params }: LoaderFunctionArgs): Promise<{ messages?: Message[], members?: Member[], error?: string }> {
  const chatId = params.chatId;
  try {
    const messages = await getChatMessages(request, chatId!);
    // const members = await getGroupMembers(request, chatId!);
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
  profile: Profile;
  messages: Message[];
}

type SpeechBubbleProps = {
  messageGroup : MessageGroup;
  currentUserId?: number;
  handleAvatarClick: (event: ReactMouseEvent<HTMLImageElement>, profile: Profile) => void;
}

function SpeechBubble({ messageGroup, handleAvatarClick }: SpeechBubbleProps) {
  return (
    <div className="speech-bubble">
      <div className="message-header">
  <img src="https://picsum.photos/200" alt="Profile" className="message-icon" onClick={(event) => {handleAvatarClick(event, messageGroup.profile)}}/>
        <div className="message-group">
          {messageGroup.messages.map((msg) => (
            <div key={msg.id} className="message-text">{msg.content}</div>
          ))}
        </div>
      </div>
    </div>
  );
}

type ProfileCardProps = {
    profile: Profile;
    position?: {x: number, y: number};
    onClose?: () => void;
}

function ProfileCard({profile, position, onClose}: ProfileCardProps) {
  return (
    <div className="profile-card card" style={{ top: position?.y, left: position?.x }}>
        <button onClick={onClose}>x</button>

        <div className="profile-content">
            image here

            <div>
                <div>{profile.name}</div>
            </div>
            <div>
                <button>Message</button>
                <button>Add friend</button>
            </div>
        </div>
    </div>
  );
}

function groupMessagesBySender(messages: Message[]): MessageGroup[] {
  const groupedMessages: MessageGroup[] = [];
  let currentGroup: MessageGroup | null = null;

  messages.forEach((msg) => {
    // new speech bubble if:
    // // 1. first message
    // // 2. different sender
    if (!currentGroup || currentGroup.senderId !== msg.sender) {
        currentGroup = { senderId: msg.sender, messages: [msg], profile: {} as Profile };
        groupedMessages.push(currentGroup);
    } else {
        // same sender, add to same speech bubble
        currentGroup.messages.push(msg);
    }});
    // push the last group if not already pushed
    if (currentGroup && groupedMessages[groupedMessages.length - 1] !== currentGroup) {
        groupedMessages.push(currentGroup);
    }

    return groupedMessages;
}


export default function ChatDetail() {
    const { chatId } = useParams();
    const loaderData = useLoaderData<typeof loader>();

    const groupedMessages = groupMessagesBySender(loaderData.messages || []);
    const bottomRef = useRef<HTMLDivElement>(null);

  const [cardProfile, setCardProfile] = useState<Profile | null>(null);
  const [profilePosition, setProfilePosition] = useState<{x: number, y: number} | null>(null);
  const handleAvatarClick = (event: ReactMouseEvent<HTMLImageElement>, profile: Profile) => {
        const rect = (event.target as HTMLElement).getBoundingClientRect();
        setCardProfile(profile);
        setProfilePosition({ x: rect.x, y: rect.y });
    }

    useEffect(() => {
      bottomRef.current?.scrollIntoView({ behavior: "auto"});
    }, [chatId]);

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
              <SpeechBubble key={idx} messageGroup={group} handleAvatarClick={handleAvatarClick} />
            )
          ))}

          <div ref={bottomRef}></div>
        </div>

        {/* Input */}
        <div>
          <InputComponent />
        </div>

        {/* Profile card */}
        {cardProfile && (
            <ProfileCard
            profile={cardProfile}
            position={profilePosition || undefined}
            onClose={() => setCardProfile(null)}
            />
        )}
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

