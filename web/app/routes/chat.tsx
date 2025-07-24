import { Outlet, useLoaderData } from "@remix-run/react";
import { ChatMenu } from "~/components/chat-menu";

export type ChatSummary = {
    Id: number;
    Name: string;
    LastMessage: string;
    Unread: number;
}

export const loader = () => {
    const exampleData: ChatSummary[] = [
        {
            Id: 0,
            Name: "hello",
            LastMessage: "cya",
            Unread: 0,
        },
    ];

    return exampleData
}

export default function Chat() {
    const recentChats = useLoaderData<typeof loader>()
    return (
        <div className="content">
            <ChatMenu recentChats={recentChats}/>
            <Outlet/>
        </div>
    );
}