import { Select } from "@radix-ui/themes";
import { NavLink } from "@remix-run/react";

export default function Nav() {
    return (
        <nav className="nav">
            <NavigationLinks/>
            <ThemeSwitch/>
        </nav>
    )
}

function NavigationLinks() {
    return (
        <div className="nav-links">
            <NavLink
            to={"/chat"}
            className={({isActive}) => (isActive ? "nav-link active" : "nav-link")}
            >
                Chat
            </NavLink>

            <NavLink
            to={"/contacts"}
            className={({isActive}) => (isActive ? "nav-link active" : "nav-link")}
            >
                Contacts
            </NavLink> 
        </div>
    )
}

function ThemeSwitch() {
    return (
        <div>
            <Select.Root defaultValue="System">
                <Select.Trigger/>
                <Select.Content>
                    <Select.Group>
                        <Select.Item value="System">System</Select.Item>
                        <Select.Item value="Dark">Dark</Select.Item>
                        <Select.Item value="Light">Light</Select.Item>
                    </Select.Group>
                </Select.Content>
            </Select.Root>
        </div>
    )
}