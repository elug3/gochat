import { Form, useNavigation } from "react-router";
import { useMemo, useState } from "react";

type SettingsTab = "profile" | "security";

const TAB_LABELS: Record<SettingsTab, string> = {
  profile: "Profile",
  security: "Security",
};

export default function Settings() {
  const navigation = useNavigation();
  const isSubmitting = navigation.state === "submitting";
  const [activeTab, setActiveTab] = useState<SettingsTab>("profile");

  const tabList = useMemo(() => (Object.keys(TAB_LABELS) as SettingsTab[]), []);

  return (
    <div className="flex h-full flex-col gap-6 bg-slate-50/70 px-6 py-6 text-slate-900 dark:bg-slate-950 dark:text-slate-100">
      <header className="flex flex-col gap-1">
        <p className="text-[11px] font-semibold uppercase tracking-[0.2em] text-slate-500/70 dark:text-slate-400/70">Account settings</p>
        <h1 className="text-2xl font-semibold">Manage your account</h1>
        <p className="text-sm text-slate-500 dark:text-slate-400">Update your personal details and keep your account secure.</p>
      </header>

      <nav aria-label="Settings sections" className="flex flex-wrap gap-3">
        {tabList.map((tab) => {
          const isActive = activeTab === tab;
          const baseClasses = "inline-flex items-center gap-2 rounded-full border px-4 py-2 text-sm font-medium transition";
          const activeClasses = "border-blue-500 bg-blue-50 text-blue-600 shadow-sm dark:border-blue-400 dark:bg-blue-500/10 dark:text-blue-200";
          const inactiveClasses = "border-slate-200 text-slate-500 hover:border-blue-400 hover:text-blue-600 dark:border-slate-800 dark:text-slate-300 dark:hover:border-blue-400 dark:hover:text-blue-200";

          return (
            <button
              key={tab}
              type="button"
              onClick={() => setActiveTab(tab)}
              className={`${baseClasses} ${isActive ? activeClasses : inactiveClasses}`}
              aria-pressed={isActive}
            >
              {TAB_LABELS[tab]}
            </button>
          );
        })}
      </nav>

      <section className="flex-1 overflow-hidden">
        {{
          profile: <ProfileForm isSubmitting={isSubmitting} />,
          security: <SecurityForm isSubmitting={isSubmitting} />
        }[activeTab]}
      </section>
    </div>
  );
}

interface SettingsSectionProps {
  isSubmitting: boolean;
  onOpenPasskeys?: () => void;
}

interface SavedPasskey {
  id: string;
  label: string;
  addedOn: string;
  lastUsed: string;
  location: string;
}

const SAVED_PASSKEYS: SavedPasskey[] = [
];

function ProfileForm({ isSubmitting }: SettingsSectionProps) {
  return (
    <div className="rounded-xl border border-slate-200/70 bg-white/70 p-6 shadow-sm dark:border-slate-800 dark:bg-slate-900/60">
      <header className="mb-6 space-y-1">
        <h2 className="text-lg font-semibold text-slate-800 dark:text-slate-100">Profile information</h2>
        <p className="text-sm text-slate-500 dark:text-slate-400">Keep your personal details up to date so teammates know who they are chatting with.</p>
      </header>

      <Form method="post" className="space-y-5">
        <input type="hidden" name="_action" value="update-profile" />

        <div className="grid gap-4 md:grid-cols-2">
          <label className="flex flex-col gap-2 text-sm font-medium text-slate-600 dark:text-slate-300">
            Display name
            <input
              type="text"
              name="displayName"
              placeholder="Alex Johnson"
              className="w-full rounded-lg border border-slate-200 bg-white px-3 py-2 text-sm font-normal text-slate-700 shadow-sm outline-none transition focus:border-blue-500 focus:ring-2 focus:ring-blue-200 dark:border-slate-700 dark:bg-slate-950 dark:text-slate-100"
              disabled={isSubmitting}
              required
            />
          </label>

          <label className="flex flex-col gap-2 text-sm font-medium text-slate-600 dark:text-slate-300">
            Email address
            <input
              type="email"
              name="email"
              placeholder="alex@email.com"
              className="w-full rounded-lg border border-slate-200 bg-white px-3 py-2 text-sm font-normal text-slate-700 shadow-sm outline-none transition focus:border-blue-500 focus:ring-2 focus:ring-blue-200 dark:border-slate-700 dark:bg-slate-950 dark:text-slate-100"
              disabled={isSubmitting}
              required
            />
          </label>
        </div>

        <label className="flex flex-col gap-2 text-sm font-medium text-slate-600 dark:text-slate-300">
          Bio
          <textarea
            name="bio"
            rows={4}
            placeholder="Share a short description about yourself or your role."
            className="w-full rounded-lg border border-slate-200 bg-white px-3 py-2 text-sm font-normal text-slate-700 shadow-sm outline-none transition focus:border-blue-500 focus:ring-2 focus:ring-blue-200 dark:border-slate-700 dark:bg-slate-950 dark:text-slate-100"
            disabled={isSubmitting}
          />
        </label>

        <div className="flex items-center justify-end gap-2">
          <button
            type="submit"
            className="inline-flex items-center gap-2 rounded-lg bg-blue-600 px-4 py-2 text-sm font-medium text-white shadow-sm transition hover:bg-blue-700 disabled:cursor-not-allowed disabled:opacity-60"
            disabled={isSubmitting}
          >
            Save profile
          </button>
        </div>
      </Form>
    </div>
  );
}

function SecurityForm({ isSubmitting, onOpenPasskeys }: SettingsSectionProps) {
  return (
    <div className="rounded-xl border border-slate-200/70 bg-white/70 p-6 shadow-sm dark:border-slate-800 dark:bg-slate-900/60">
      <header className="mb-6 space-y-1">
        <h2 className="text-lg font-semibold text-slate-800 dark:text-slate-100">Security</h2>
        <p className="text-sm text-slate-500 dark:text-slate-400">Change your password regularly to keep your account secure.</p>
      </header>

      <Form method="post" className="space-y-5">
        <input type="hidden" name="_action" value="update-password" />

        <label className="flex flex-col gap-2 text-sm font-medium text-slate-600 dark:text-slate-300">
          Current password
          <input
            type="password"
            name="currentPassword"
            placeholder="••••••••"
            className="w-full rounded-lg border border-slate-200 bg-white px-3 py-2 text-sm font-normal text-slate-700 shadow-sm outline-none transition focus:border-blue-500 focus:ring-2 focus:ring-blue-200 dark:border-slate-700 dark:bg-slate-950 dark:text-slate-100"
            disabled={isSubmitting}
            required
          />
        </label>

        <div className="grid gap-4 md:grid-cols-2">
          <label className="flex flex-col gap-2 text-sm font-medium text-slate-600 dark:text-slate-300">
            New password
            <input
              type="password"
              name="newPassword"
              placeholder="At least 8 characters"
              className="w-full rounded-lg border border-slate-200 bg-white px-3 py-2 text-sm font-normal text-slate-700 shadow-sm outline-none transition focus:border-blue-500 focus:ring-2 focus:ring-blue-200 dark:border-slate-700 dark:bg-slate-950 dark:text-slate-100"
              disabled={isSubmitting}
              required
              minLength={8}
            />
          </label>

          <label className="flex flex-col gap-2 text-sm font-medium text-slate-600 dark:text-slate-300">
            Confirm password
            <input
              type="password"
              name="confirmPassword"
              placeholder="Repeat new password"
              className="w-full rounded-lg border border-slate-200 bg-white px-3 py-2 text-sm font-normal text-slate-700 shadow-sm outline-none transition focus:border-blue-500 focus:ring-2 focus:ring-blue-200 dark:border-slate-700 dark:bg-slate-950 dark:text-slate-100"
              disabled={isSubmitting}
              required
              minLength={8}
            />
          </label>
        </div>

        <div className="flex flex-col gap-4 rounded-lg border border-amber-200/70 bg-amber-50/70 p-4 text-sm text-amber-700 dark:border-amber-900/40 dark:bg-amber-900/20 dark:text-amber-200">
          <p className="font-semibold">Password tips</p>
          <ul className="space-y-1 text-xs">
            <li>• Use a mix of uppercase, lowercase, numbers, and symbols.</li>
            <li>• Avoid using the same password across multiple services.</li>
            <li>• Consider a password manager to generate secure passwords.</li>
          </ul>
        </div>

        <div className="flex items-center justify-end gap-2">
          <button
            type="submit"
            className="inline-flex items-center gap-2 rounded-lg bg-slate-900 px-4 py-2 text-sm font-medium text-white shadow-sm transition hover:bg-slate-700 disabled:cursor-not-allowed disabled:opacity-60 dark:bg-slate-200 dark:text-slate-900 dark:hover:bg-slate-100"
            disabled={isSubmitting}
          >
            Update password
          </button>
        </div>
      </Form>

      <PasskeyMenu isSubmitting={isSubmitting} onOpenPasskeys={onOpenPasskeys} />
    </div>
  );
}

function PasskeyMenu({ isSubmitting, onOpenPasskeys }: SettingsSectionProps) {
  const openPasskeysButtonClasses = onOpenPasskeys
    ? "text-blue-600 hover:text-blue-700 dark:text-blue-300 dark:hover:text-blue-200"
    : "cursor-not-allowed text-slate-400 dark:text-slate-600";

  const [successNotice, setSuccessNotice] = useState<string | null>(null);

  const registerPasskey = async () => {
    setSuccessNotice(null);

    // TODO: use backend url
    const startRes = await fetch("http://localhost:8001/webauthn/register/start", {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify({ userId: "user-123", name: "larry" }),
    });
    const { publicKey } = await startRes.json();
    console.log("PublicKey Credential Creation Options:", publicKey);
    if (!publicKey || !publicKey.challenge || !publicKey.user.id) {
      throw new Error("invalid publicKey from server");
    }

    publicKey.challenge = decodeBase64URL(publicKey.challenge);
    publicKey.user.id = decodeBase64URL(publicKey.user.id);

    const cred = await navigator.credentials.create({
      publicKey,
    });

    console.log("Created Credentials:", cred);

    const finishRes = await fetch("http://localhost:8001/register/finish", {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify(cred),
    });

    if (!finishRes.ok) {
      throw new Error("Failed to finish registration");
    }

    setSuccessNotice("Passkey registered successfully! This device can now be used for passwordless sign in.");
  };

  return (
    <section className="mt-8 rounded-xl border border-slate-200/70 bg-gradient-to-b from-white/90 to-slate-50/70 p-5 shadow-sm dark:border-slate-800 dark:from-slate-900/70 dark:to-slate-900/30">
      <header className="space-y-1">
        <p className="text-[11px] font-semibold uppercase tracking-[0.2em] text-slate-500/80 dark:text-slate-400/70">Passkey menu</p>
        <h3 className="text-base font-semibold text-slate-800 dark:text-slate-100">Fast passkey actions</h3>
        <p className="text-sm text-slate-500 dark:text-slate-400">Add or revoke devices without leaving the security tab.</p>
      </header>

      {successNotice && (
        <p
          role="status"
          aria-live="polite"
          className="mt-4 text-sm font-semibold text-emerald-700 dark:text-emerald-300"
        >
          {successNotice}
          <button
            type="button"
            onClick={() => setSuccessNotice(null)}
            className="ml-3 text-xs font-semibold uppercase tracking-wide text-emerald-700 underline underline-offset-2 transition hover:text-emerald-900 dark:text-emerald-200 dark:hover:text-emerald-50"
          >
            Dismiss
          </button>
        </p>
      )}

      <div className="mt-4 flex flex-col gap-4 rounded-lg border border-dashed border-slate-300/80 bg-white/70 p-4 dark:border-slate-700 dark:bg-slate-950/40 md:flex-row md:items-center md:justify-between">
        <div>
          <p className="text-sm font-semibold text-slate-800 dark:text-slate-100">Add a trusted device</p>
          <p className="text-xs text-slate-500 dark:text-slate-400">Your browser will prompt for Touch ID, Face ID, or a security key.</p>
        </div>
        <div>
          <button
            type="submit"
            className="inline-flex items-center justify-center rounded-lg bg-blue-600 px-4 py-2 text-sm font-medium text-white shadow-sm transition hover:bg-blue-700 disabled:cursor-not-allowed disabled:opacity-60"
            disabled={isSubmitting}
            onClick={async () => {
              try {
                await registerPasskey();
              } catch (error) {
                const reason = error instanceof Error ? error.message : String(error);
                alert("Failed to create passkey: " + reason);
              }
            }}
          >
            Create passkey
          </button>
        </div>
      </div>

      <ul role="menu" className="mt-4 divide-y divide-slate-200 rounded-lg border border-slate-200/80 bg-white/90 dark:divide-slate-800 dark:border-slate-800 dark:bg-slate-950/40">
        {SAVED_PASSKEYS.length === 0 ? (
          <li className="px-4 py-4 text-sm text-slate-500 dark:text-slate-400">
            <span className="font-semibold text-slate-700 dark:text-slate-200">No passkeys yet.</span>{" "}
          </li>
        ) : (
          SAVED_PASSKEYS.map((passkey) => (
            <li key={passkey.id} role="presentation" className="flex flex-col gap-3 px-4 py-3 text-sm text-slate-700 dark:text-slate-200 md:flex-row md:items-center md:justify-between">
              <div>
                <p className="font-semibold text-slate-800 dark:text-slate-100">{passkey.label}</p>
                <p className="text-xs text-slate-500 dark:text-slate-400">{`Added ${passkey.addedOn} • Last used ${passkey.lastUsed}`}</p>
                <p className="text-xs text-slate-400 dark:text-slate-500">{passkey.location}</p>
              </div>

              <Form method="post" className="flex gap-2 md:items-center">
                <input type="hidden" name="_action" value="delete-passkey" />
                <input type="hidden" name="passkeyId" value={passkey.id} />
                <button
                  type="submit"
                  className="inline-flex items-center justify-center rounded-lg border border-slate-300 px-3 py-2 text-xs font-semibold text-slate-700 transition hover:border-slate-400 hover:text-slate-900 disabled:cursor-not-allowed disabled:opacity-60 dark:border-slate-600 dark:text-slate-200 dark:hover:border-slate-500"
                  disabled={isSubmitting}
                >
                  Remove
                </button>
              </Form>
            </li>
          ))
        )}
      </ul>

      <div className="mt-4 flex flex-wrap items-center gap-3 text-xs text-slate-500 dark:text-slate-400">
        <button
          type="button"
          onClick={onOpenPasskeys}
          className={`inline-flex items-center rounded-lg border border-transparent px-3 py-2 font-semibold transition focus-visible:outline focus-visible:outline-2 focus-visible:outline-blue-400 ${openPasskeysButtonClasses}`}
          disabled={!onOpenPasskeys}
        >
          Open full passkey settings
        </button>
      </div>
    </section>
  );
}

function decodeBase64URL(value: string) {
  const pad = value.length % 4 ? 4 - (value.length % 4) : 0;
  const base64 = (value + "=".repeat(pad)).replace(/-/g, "+").replace(/_/g, "/");
  return Uint8Array.from(atob(base64), c => c.charCodeAt(0));
}
