import { NavLink } from "@remix-run/react";
import { ChatSummary } from "~/routes/chat";

function ChatItem({chat}: {chat: ChatSummary}) {
    return (
        <li>
            <NavLink to={`${chat.Id}`}>
                {chat.Name}
            </NavLink>
        </li>
    );
}

export function ChatMenu({recentChats}: {recentChats: ChatSummary[]}) {
    return (
        <ul className="menu">
            {recentChats.map((chat) => (
                <ChatItem key={chat.Name} chat={chat}/>
            ))}
        </ul>
    );
}