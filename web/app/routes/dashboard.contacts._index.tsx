import { useOutletContext } from "react-router";
import { ContactsList, type ContactsOutletContext } from "./dashboard.contacts";

export default function ContactsIndexRoute() {
  const { contacts } = useOutletContext<ContactsOutletContext>();

  return <ContactsList contacts={contacts} />;
}
