import { useEffect, useRef } from "react";
import { useChatStore } from "~/store/chatStore";
import { snakeToCamel } from "~/utils/case";

export function useChatWebSocket(wsToken: string) {
    const wsRef = useRef<WebSocket | null>(null);
    const setWsState = useChatStore((s) => s.setWsState)
    const upsertMessage = useChatStore((s) => s.upsertMessage)

    useEffect(() => {
        if (wsRef.current) return; // already connected

        const ws = new WebSocket(`ws://localhost:12345/ws?token=${wsToken}`); // TODO
        wsRef.current = ws;

        ws.onopen = () => {
            setWsState("open")
            console.log("WebSocket connected");
        };
        ws.onclose = () => {
            setWsState("closed")
            console.log("WebSocket disconnected");
        };

        ws.onmessage = (event) => {
            console.log("WebSocket message received:", event.data);
            const data = JSON.parse(snakeToCamel(event.data));
            switch (data.type) {
                case "message":
                    upsertMessage(data.payload);
                    break;
            }
        };

        return () => {
            ws.close();
        }; 
    }, [wsToken, setWsState, upsertMessage]);
}