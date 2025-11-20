import type { ActionFunctionArgs, LoaderFunctionArgs } from "react-router";
import { logout } from "~/session.server";

export async function action({ request }: ActionFunctionArgs) {
    return logout(request);
}

export async function loader({ request }: LoaderFunctionArgs) {
    return logout(request);
}
