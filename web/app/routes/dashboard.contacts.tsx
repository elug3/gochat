
import { NavLink, Outlet, useLoaderData, type LoaderFunctionArgs } from "react-router";
import { fetchContacts } from "~/utils/auth.server";
import type { Profile } from "~/model";

export async function loader({ request }: LoaderFunctionArgs): Promise<{ contacts?: Profile[], error?: string }> {
    try {
        const contacts = await fetchContacts(request);
        return { contacts };
    } catch (err: any) {
        return { error: err.message };
    }
}

export type ContactsOutletContext = {
    contacts: Profile[];
};

export type ContactsListProps = {
    contacts: Profile[];
}

export function ContactsList({ contacts }: ContactsListProps) {
    return (
        <div className="sidebar-container">
            {contacts.length === 0 ? (
                <div className="text-center text-gray-500 py-8">
                    <p>No contacts found.</p>
                    <p className="text-sm mt-2">Add friends to start chatting!</p>
                </div>
            ) : (
                contacts.map((contact) => (
                    <div
                        key={contact.userId}
                        className="sidebar-item flex items-center justify-between group"
                    >
                        <div className="flex items-center gap-3 flex-1 min-w-0">
                            {/* Avatar placeholder */}
                            <div className="w-8 h-8 rounded-full bg-gray-300 flex-shrink-0 flex items-center justify-center">
                                {contact.avatarUrl ? (
                                    <img
                                        src={contact.avatarUrl}
                                        alt={contact.name}
                                        className="w-8 h-8 rounded-full object-cover"
                                    />
                                ) : (
                                    <span className="text-sm font-medium text-gray-600">
                                        {contact.name.charAt(0).toUpperCase()}
                                    </span>
                                )}
                            </div>
                            
                            {/* Contact info */}
                            <div className="flex-1 min-w-0">
                                <div className="flex items-center gap-2">
                                    <span className="text-sm font-medium truncate">
                                        {contact.name}
                                    </span>
                                    <span
                                        className={`w-2 h-2 rounded-full flex-shrink-0 ${
                                            contact.status === "online" 
                                                ? "bg-green-500" 
                                                : "bg-gray-400"
                                        }`}
                                    />
                                </div>
                                <div className="text-xs text-gray-500 capitalize">
                                    {contact.status}
                                </div>
                            </div>
                        </div>
                        
                        {/* Options button */}
                        <div className="flex-shrink-0 opacity-0 group-hover:opacity-100 transition-opacity">
                            <button
                                className="p-1 rounded hover:bg-gray-200 text-gray-500 hover:text-gray-700"
                                onClick={(e) => {
                                    e.preventDefault();
                                    // TODO: Show options menu
                                    console.log("Options for", contact.name);
                                }}
                            >
                                <svg
                                    className="w-4 h-4"
                                    fill="currentColor"
                                    viewBox="0 0 24 24"
                                >
                                    <path d="M12 8c1.1 0 2-.9 2-2s-.9-2-2-2-2 .9-2 2 .9 2 2 2zm0 2c-1.1 0-2 .9-2 2s.9 2 2 2 2-.9 2-2-.9-2-2-2zm0 6c-1.1 0-2 .9-2 2s.9 2 2 2 2-.9 2-2-.9-2-2-2z" />
                                </svg>
                            </button>
                        </div>
                    </div>
                ))
            )}
        </div>
    );
}

export default function Contacts() {
    const loaderData = useLoaderData<typeof loader>();
    
    if (loaderData.error) {
        return (
            <div className="page">
                <div className="card">
                    <h1>Error</h1>
                    <p className="text-error">{loaderData.error}</p>
                </div>
            </div>
        );
    }
    
    if (!loaderData.contacts) {
        return (
            <div className="page">
                <div className="card">
                    <h1>Contacts</h1>
                    <p>Failed to load contacts</p>
                </div>
            </div>
        );
    }

    return (
        <div className="page">
            <main className="flex-1 flex items-center justify-center">
                <div className="w-full max-w-2xl">
                    <h2 className="text-center mb-6 text-lg font-semibold">Contacts</h2>

                    <div className="flex border-b border-gray-200">
                        <NavLink
                            to="."
                            end
                            className={({ isActive }) =>
                                `flex-1 py-2 px-4 text-sm font-medium text-center border-b-2 transition-colors ${
                                    isActive
                                        ? "border-blue-500 text-blue-600"
                                        : "border-transparent text-gray-500 hover:text-gray-700"
                                }`
                            }
                        >
                            All Contacts ({loaderData.contacts.length})
                        </NavLink>
                        <NavLink
                            to="new"
                            className={({ isActive }) =>
                                `flex-1 py-2 px-4 text-sm font-medium text-center border-b-2 transition-colors ${
                                    isActive
                                        ? "border-blue-500 text-blue-600"
                                        : "border-transparent text-gray-500 hover:text-gray-700"
                                }`
                            }
                        >
                            Add New
                        </NavLink>
                    </div>

                    <div className="mt-6">
                        <Outlet context={{ contacts: loaderData.contacts }} />
                    </div>
                </div>
            </main>
        </div>
    );
}
