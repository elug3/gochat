import { Form, useFetcher, useLoaderData, useNavigation, useRevalidator, type ActionFunctionArgs } from "react-router";
import { useEffect, useMemo, useState } from "react";
import type { KeyboardEvent } from "react";
import { useUser } from "~/context/user";
import { requireAccessToken } from "~/utils/auth.server";


const AUTH_URL = `http://localhost:8080`;
const WEBAUTHN_REGISTER_BEGIN_URL = `${AUTH_URL}/webauthn/register/begin`;
const WEBAUTHN_REGISTER_FINISH_URL = `${AUTH_URL}/webauthn/register/finish`;


type WebAuthnStartResponse = {
  ok: boolean;
  error?: string;
  publicKey?: PublicKeyCredentialCreationOptions;
};

type WebAuthnFinishResponse = {
  ok: boolean;
  error?: string;
};

type PasskeyRecord = {
  id: number;
  key_name: string;
  user_id?: number;
  created_at: string;
  last_used_at: string;
};

type LoaderData = {
  passkeys: PasskeyRecord[];
};

export async function action({ request }: ActionFunctionArgs) {
  const formData = await request.formData();
  const intent = formData.get("_action");

  const accessToken = await requireAccessToken(request);

  switch (intent) {
    case "update-profile": {
      return Response.json({ ok: true });
    }
    case "register-passkey-begin": {
      try {
        const startRes = await fetch(`${WEBAUTHN_REGISTER_BEGIN_URL}`, {
          method: "POST",
          headers: {
            "Content-Type": "application/json",
            "Authorization": `Bearer ${accessToken}`,
          }
        });
        if (!startRes.ok) {
          const errorText = await startRes.text().catch(() => "");
          console.error("Failed to start passkey registration:", errorText);
          return Response.json({ ok: false, error: "Failed to start passkey registration." }, { status: startRes.status });
        }

        const startData: WebAuthnStartResponse = await startRes.json();
        if (!startData.publicKey) {
          return Response.json({ ok: false, error: startData.error || "Failed to start passkey registration." }, { status: 400 });
        }
        return Response.json({ ok: true, publicKey: startData.publicKey });
      } catch (error) {
        console.error("Failed to start passkey registration:", error);
        return Response.json({ ok: false, error: "Failed to start passkey registration." }, { status: 500 });
      }
    }
    case "register-passkey-finish": {
      const credential = formData.get("credential");
      if (typeof credential !== "string") {
        return Response.json({ ok: false, error: "Invalid credential payload." }, { status: 400 });
      }

      let parsedCredential: unknown;
      try {
        parsedCredential = JSON.parse(credential);
      } catch (error) {
        console.error("Failed to parse credential payload:", error);
        return Response.json({ ok: false, error: "Invalid credential payload." }, { status: 400 });
      }

      try {
        const finishRes = await fetch(`${WEBAUTHN_REGISTER_FINISH_URL}`, {
          method: "POST",
          headers: {
            "Content-Type": "application/json",
            "Authorization": `Bearer ${accessToken}`,
          },
          body: JSON.stringify(parsedCredential),
        });
        if (!finishRes.ok) {
          const errorText = await finishRes.text().catch(() => "");
          console.error("Failed to finish passkey registration:", errorText);
          return Response.json({ ok: false, error: "Failed to finish passkey registration." }, { status: finishRes.status });
        }
        return Response.json({ ok: true });
      } catch (error) {
        console.error("Failed to finish passkey registration:", error);
        return Response.json({ ok: false, error: "Failed to finish passkey registration." }, { status: 500 });
      }

    }
    case "update-password": {
      return Response.json({ ok: true });
    }
    case "delete-passkey": {
      const passkeyIdRaw = formData.get("passkeyId");
      const passkeyId = typeof passkeyIdRaw === "string" ? Number.parseInt(passkeyIdRaw, 10) : NaN;

      if (!Number.isFinite(passkeyId)) {
        return Response.json({ ok: false, error: "Invalid passkey id" }, { status: 400 });
      }

      try {
        const res = await fetch(`${AUTH_URL}/webauthn/passkeys/${passkeyId}`, {
          method: "DELETE",
          headers: {
            "Authorization": `Bearer ${accessToken}`,
          },
        });

        if (!res.ok) {
          const errorText = await res.text().catch(() => "");
          console.error("Failed to delete passkey:", errorText);
          return Response.json({ ok: false, error: "Failed to delete passkey." }, { status: res.status });
        }

        return Response.json({ ok: true, id: passkeyId });
      } catch (error) {
        console.error("Failed to delete passkey:", error);
        return Response.json({ ok: false, error: "Failed to delete passkey." }, { status: 500 });
      }
    }
    case "rename-passkey": {
      const passkeyIdRaw = formData.get("passkeyId");
      const newName = formData.get("newName");

      const passkeyId = typeof passkeyIdRaw === "string" ? Number.parseInt(passkeyIdRaw, 10) : NaN;

      if (!Number.isFinite(passkeyId) || typeof newName !== "string" || newName.trim().length === 0) {
        return Response.json({ ok: false, error: "Invalid rename request" }, { status: 400 });
      }

      try {
        const res = await fetch(`${AUTH_URL}/webauthn/passkeys/${passkeyId}`, {
          method: "PUT",
          headers: {
            "Content-Type": "application/json",
            "Authorization": `Bearer ${accessToken}`,
          },
          body: JSON.stringify({ id: passkeyId, key_name: newName.trim() }),
        });

        if (!res.ok) {
          const errorText = await res.text().catch(() => "");
          console.error("Failed to rename passkey:", errorText);
          return Response.json({ ok: false, error: "Failed to rename passkey." }, { status: res.status });
        }

        const passkey = await res.json();
        return Response.json({ ok: true, passkey });
      } catch (error) {
        console.error("Failed to rename passkey:", error);
        return Response.json({ ok: false, error: "Failed to rename passkey." }, { status: 500 });
      }
    }
    default: {
      return Response.json({ ok: false, error: "Unknown action" });
    }
  }
}

export async function loader({ request }: { request: Request }) {
  const accessToken = await requireAccessToken(request);

  try {
    const res = await fetch(`${AUTH_URL}/webauthn/passkeys`, {
      method: "GET",
      headers: {
        "Content-Type": "application/json",
        "Authorization": `Bearer ${accessToken}`,
      }
    });

    if (!res.ok) {
      console.error("Failed to fetch passkeys");
      return { passkeys: [] as PasskeyRecord[] };
    }

    const passkeys: PasskeyRecord[] = await res.json();
    return { passkeys };
  } catch (error) {
    console.error("Error fetching passkeys", error);
    return { passkeys: [] as PasskeyRecord[] };
  }
}

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
    <div className="flex h-full flex-col gap-6 bg-slate-50 px-6 py-6 text-slate-900 dark:bg-slate-950 dark:text-slate-100">
      <header className="flex flex-col gap-1">
        <p className="text-[11px] font-semibold uppercase tracking-[0.2em] text-slate-500 dark:text-slate-400">Account settings</p>
        <div className="flex items-center gap-3">
          <h1 className="text-2xl font-semibold">Manage your account</h1>
        </div>
        <p className="text-sm text-slate-500 dark:text-slate-400">Update your personal details and keep your account secure.</p>
      </header>

      <div className="flex flex-1 gap-6">
        <nav
          aria-label="Settings sections"
          className="flex w-64 shrink-0 flex-col gap-2 text-sm text-slate-700 dark:text-slate-200"
        >
          <p className="text-[11px] font-semibold uppercase tracking-[0.2em] text-slate-400 dark:text-slate-500">
            Contents
          </p>
          <ul className="flex flex-col gap-1">
            {tabList.map((tab, index) => {
              const isActive = activeTab === tab;
              const baseClasses =
                "group flex w-full items-center gap-2 rounded px-2 py-1 text-left transition hover:underline";
              const activeClasses =
                "font-semibold text-slate-900 underline decoration-slate-400 decoration-2 underline-offset-4 dark:text-slate-50 dark:decoration-slate-500";
              const inactiveClasses =
                "text-slate-600 hover:text-slate-900 dark:text-slate-400 dark:hover:text-slate-100";

              return (
                <li key={tab}>
                  <button
                    type="button"
                    onClick={() => setActiveTab(tab)}
                    className={`${baseClasses} ${isActive ? activeClasses : inactiveClasses}`}
                    aria-pressed={isActive}
                    aria-current={isActive ? "page" : undefined}
                  >
                    <span className="text-xs font-semibold text-slate-400 dark:text-slate-500">
                      {index + 1}.
                    </span>
                    <span>{TAB_LABELS[tab]}</span>
                  </button>
                </li>
              );
            })}
          </ul>
        </nav>

        <section className="flex-1 overflow-hidden">
          {{
            profile: <ProfileForm isSubmitting={isSubmitting} />,
            security: <SecurityForm isSubmitting={isSubmitting} />
          }[activeTab]}
        </section>
      </div>
    </div>
  );
}

interface SettingsSectionProps {
  isSubmitting: boolean;
  onOpenPasskeys?: () => void;
}

type SavedPasskey = PasskeyRecord;

function ProfileForm({ isSubmitting }: SettingsSectionProps) {
  const { user, profile } = useUser();
  const userId = profile?.userId ?? (user?.id ? user.id.toString() : "Unavailable");
  const displayName = profile?.name ?? "Unknown";
  const [name, setName] = useState(displayName);

  useEffect(() => {
    setName(displayName);
  }, [displayName]);

  return (
    <div className="rounded-xl border border-slate-200 bg-white p-6 shadow-sm dark:border-slate-800 dark:bg-slate-900">
      <header className="mb-6 space-y-1">
        <h2 className="text-lg font-semibold text-slate-800 dark:text-slate-100">Profile</h2>
        <p className="text-sm text-slate-500 dark:text-slate-400">Update your personal information.</p>
      </header>

      <Form method="post" className="space-y-6">
        <input type="hidden" name="_action" value="update-profile" />
        <div className="rounded-lg border border-slate-100 bg-slate-50 p-4 text-sm text-slate-700 dark:border-slate-800 dark:bg-slate-900/60 dark:text-slate-200">
          <p className="font-semibold text-slate-800 dark:text-slate-100">Signed-in account</p>
          <p className="mt-1 text-slate-500 dark:text-slate-400">
            These details come from your current session and are read-only here.
          </p>
        </div>

        <div className="grid gap-4 md:grid-cols-2">
          <label className="flex flex-col gap-2 text-sm font-medium text-slate-600 dark:text-slate-300">
            User ID
            <input
              value={userId}
              readOnly
              disabled
              className="w-full rounded-lg border border-slate-200 bg-slate-100 px-3 py-2 text-sm font-semibold text-slate-800 shadow-sm outline-none transition focus:border-blue-500 focus:ring-2 focus:ring-blue-200 dark:border-slate-700 dark:bg-slate-900 dark:text-slate-100"
            />
            <span className="text-xs font-normal text-slate-500 dark:text-slate-400">
              Unique identifier for your account.
            </span>
          </label>

          <label className="flex flex-col gap-2 text-sm font-medium text-slate-600 dark:text-slate-300">
            Name
            <input
              name="name"
              value={name}
              onChange={(event) => setName(event.target.value)}
              className="w-full rounded-lg border border-slate-200 bg-slate-100 px-3 py-2 text-sm font-semibold text-slate-800 shadow-sm outline-none transition focus:border-blue-500 focus:ring-2 focus:ring-blue-200 dark:border-slate-700 dark:bg-slate-900 dark:text-slate-100"
              disabled={isSubmitting}
              required
            />
            <span className="text-xs font-normal text-slate-500 dark:text-slate-400">
              How other people see you in chats and contacts.
            </span>
          </label>
        </div>

        <div className="flex items-center justify-end gap-2">
          <p className="flex-1 text-xs text-slate-500 dark:text-slate-400">
            {isSubmitting ? "Saving your profile..." : "Profile values refresh automatically when your account info updates."}
          </p>
          <button
            type="submit"
            className="inline-flex items-center gap-2 rounded-lg bg-slate-900 px-4 py-2 text-sm font-medium text-white shadow-sm transition hover:bg-slate-700 disabled:cursor-not-allowed disabled:opacity-60 dark:bg-slate-200 dark:text-slate-900 dark:hover:bg-slate-100"
            disabled={isSubmitting}
          >
            {isSubmitting ? "Saving..." : "Save changes"}
          </button>
        </div>
      </Form>
    </div>
  );
}

function SecurityForm({ isSubmitting, onOpenPasskeys }: SettingsSectionProps) {
  return (
    <div className="rounded-xl border border-slate-200 bg-white p-6 shadow-sm dark:border-slate-800 dark:bg-slate-900">
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

        <div className="flex flex-col gap-4 rounded-lg border border-amber-200 bg-amber-50 p-4 text-sm text-amber-700 dark:border-amber-900 dark:bg-amber-900 dark:text-amber-200">
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

  const revalidator = useRevalidator();
  const [successNotice, setSuccessNotice] = useState<string | null>(null);
  const [isPasskeyLoading, setIsPasskeyLoading] = useState(false);
  const beginFetcher = useFetcher();
  const finishFetcher = useFetcher();
  const renameFetcher = useFetcher();
  const [editingPasskeyId, setEditingPasskeyId] = useState<number | null>(null);
  const [renameValue, setRenameValue] = useState("");

  const loaderData = useLoaderData<LoaderData>();
  const loadedPasskeys = loaderData?.passkeys ?? [];
  const [passkeys, setPasskeys] = useState<SavedPasskey[]>(loadedPasskeys);

  useEffect(() => {
    setPasskeys(loadedPasskeys);
  }, [loadedPasskeys]);

  const registerPasskey = async () => {
    setSuccessNotice(null);

    if (
      typeof window === "undefined" ||
      typeof navigator === "undefined" ||
      !navigator.credentials ||
      typeof navigator.credentials.create !== "function"
    ) {
      alert("Passkeys are not supported in this browser.");
      return;
    }

    setIsPasskeyLoading(true);
    beginFetcher.submit({ _action: "register-passkey-begin" }, { method: "post" });
  };

  // Handle begin fetcher response
  useEffect(() => {
    if (beginFetcher.state !== "idle" || !beginFetcher.data) return;

    const data: WebAuthnStartResponse = beginFetcher.data;
    const cleanup = () => beginFetcher.unstable_reset();

    if (data.ok && data.publicKey) {
      const publicKey = normalizePublicKeyCreationOptions(data.publicKey);
      navigator.credentials
        .create({ publicKey })
        .then((credential) => {
          if (!credential) {
            throw new Error("No credential returned");
          }

          const serialized = serializeAttestationResponse(credential as PublicKeyCredential);
          finishFetcher.submit(
            {
              _action: "register-passkey-finish",
              credential: JSON.stringify(serialized),
            },
            { method: "post" }
          );
        })
        .catch((error) => {
          console.error("Error creating credential:", error);
          alert("Could not create passkey.");
          setIsPasskeyLoading(false);
        })
        .finally(cleanup);
      return;
    }

    if (!data.ok) {
      alert(data.error || "Could not start passkey registration.");
      setIsPasskeyLoading(false);
    }
    cleanup();
  }, [beginFetcher.state, beginFetcher.data, finishFetcher]);

  useEffect(() => {
    if (finishFetcher.state !== "idle" || !finishFetcher.data) return;

    const data: WebAuthnFinishResponse = finishFetcher.data;
    if (data.ok) {
      setSuccessNotice("Passkey created successfully.");
      revalidator.revalidate();
    } else {
      alert(data.error || "Could not finish passkey registration.");
    }
    setIsPasskeyLoading(false);
    finishFetcher.unstable_reset();
  }, [finishFetcher.state, finishFetcher.data, revalidator]);


  useEffect(() => {
    if (renameFetcher.state !== "idle" || !renameFetcher.data) return;

    if (renameFetcher.data.ok && renameFetcher.data.passkey) {
      const updatedPasskey = renameFetcher.data.passkey as SavedPasskey;

      setPasskeys((current) =>
        current.map((pk) => (pk.id === updatedPasskey.id ? updatedPasskey : pk))
      );
      setSuccessNotice(`Renamed passkey to "${updatedPasskey.key_name}"`);
      setEditingPasskeyId(null);
      setRenameValue("");
      revalidator.revalidate();
    } else {
      alert(renameFetcher.data.error || "Could not rename passkey.");
    }
    renameFetcher.unstable_reset();
  }, [renameFetcher.state, renameFetcher.data, revalidator]);

  const formatTimestamp = (value: string) => {
    const date = new Date(value);
    return Number.isNaN(date.getTime()) ? value : date.toLocaleString();
  };

  const beginRename = (passkey: SavedPasskey) => {
    setSuccessNotice(null);
    setEditingPasskeyId(passkey.id);
    setRenameValue(passkey.key_name);
  };

  const cancelRename = () => {
    setEditingPasskeyId(null);
    setRenameValue("");
  };

  const submitRename = (passkey: SavedPasskey) => {
    const trimmedName = renameValue.trim();
    if (!trimmedName || trimmedName === passkey.key_name) {
      cancelRename();
      return;
    }

    renameFetcher.submit(
      {
        _action: "rename-passkey",
        passkeyId: passkey.id.toString(),
        newName: trimmedName,
      },
      { method: "post" }
    );
  };

  const onRenameKeyDown = (event: KeyboardEvent<HTMLInputElement>, passkey: SavedPasskey) => {
    if (event.key === "Enter") {
      event.preventDefault();
      submitRename(passkey);
    }
    if (event.key === "Escape") {
      event.preventDefault();
      cancelRename();
    }
  };

  return (
    <section className="mt-8 rounded-xl border border-slate-200 bg-white p-5 shadow-sm dark:border-slate-800 dark:bg-slate-900">
      <header className="space-y-1">
        <p className="text-[11px] font-semibold uppercase tracking-[0.2em] text-slate-500 dark:text-slate-400">Passkey menu</p>
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

      <div className="mt-4 flex flex-col gap-4 rounded-lg border border-dashed border-slate-300 bg-white p-4 dark:border-slate-700 dark:bg-slate-950 md:flex-row md:items-center md:justify-between">
        <div>
          <p className="text-sm font-semibold text-slate-800 dark:text-slate-100">Add a trusted device</p>
          <p className="text-xs text-slate-500 dark:text-slate-400">Your browser will prompt for Touch ID, Face ID, or a security key.</p>
        </div>
        <div>
          <button
            type="button"
            className="inline-flex items-center justify-center rounded-lg bg-blue-600 px-4 py-2 text-sm font-medium text-white shadow-sm transition hover:bg-blue-700 disabled:cursor-not-allowed disabled:opacity-60"
            disabled={isSubmitting || isPasskeyLoading}
            onClick={registerPasskey}
          >
            {isPasskeyLoading ? "Creating..." : "Create passkey"}
          </button>
        </div>
      </div>

      <ul role="menu" className="mt-4 divide-y divide-slate-200 rounded-lg border border-slate-200 bg-white dark:divide-slate-800 dark:border-slate-800 dark:bg-slate-950">
        {passkeys.length === 0 ? (
          <li className="px-4 py-4 text-sm text-slate-500 dark:text-slate-400">
            <span className="font-semibold text-slate-700 dark:text-slate-200">No passkeys yet.</span>{" "}
          </li>
        ) : (
          passkeys.map((passkey) => (
            <li key={passkey.id} role="presentation" className="flex flex-col gap-3 px-4 py-3 text-sm text-slate-700 dark:text-slate-200 md:flex-row md:items-center md:justify-between">
              <div className="flex-1">
                {editingPasskeyId === passkey.id ? (
                  <div className="flex flex-col gap-2">
                    <input
                      autoFocus
                      type="text"
                      value={renameValue}
                      onChange={(e) => setRenameValue(e.target.value)}
                      onKeyDown={(e) => onRenameKeyDown(e, passkey)}
                      className="w-full rounded-lg border border-slate-300 bg-white px-3 py-2 text-sm font-semibold text-slate-800 shadow-sm outline-none transition focus:border-blue-500 focus:ring-2 focus:ring-blue-200 dark:border-slate-700 dark:bg-slate-900 dark:text-slate-100"
                      disabled={renameFetcher.state !== "idle"}
                    />
                    <p className="text-xs text-slate-500 dark:text-slate-400">{`Created ${formatTimestamp(passkey.created_at)} • Last used ${formatTimestamp(passkey.last_used_at)}`}</p>
                  </div>
                ) : (
                  <>
                    <p className="font-semibold text-slate-800 dark:text-slate-100">{passkey.key_name}</p>
                    <p className="text-xs text-slate-500 dark:text-slate-400">{`Created ${formatTimestamp(passkey.created_at)} • Last used ${formatTimestamp(passkey.last_used_at)}`}</p>
                  </>
                )}
              </div>

              <div className="flex gap-2 md:items-center">
                {editingPasskeyId === passkey.id ? (
                  <>
                    <button
                      type="button"
                      onClick={() => submitRename(passkey)}
                      className="inline-flex items-center justify-center rounded-lg bg-blue-600 px-3 py-2 text-xs font-semibold text-white shadow-sm transition hover:bg-blue-700 disabled:cursor-not-allowed disabled:opacity-60"
                      disabled={renameFetcher.state !== "idle"}
                    >
                      Save
                    </button>
                    <button
                      type="button"
                      onClick={cancelRename}
                      className="inline-flex items-center justify-center rounded-lg border border-slate-300 px-3 py-2 text-xs font-semibold text-slate-700 transition hover:border-slate-400 hover:text-slate-900 disabled:cursor-not-allowed disabled:opacity-60 dark:border-slate-600 dark:text-slate-200 dark:hover:border-slate-500"
                      disabled={renameFetcher.state !== "idle"}
                    >
                      Cancel
                    </button>
                  </>
                ) : (
                  <button
                    type="button"
                    onClick={() => beginRename(passkey)}
                    className="inline-flex items-center justify-center rounded-lg border border-slate-300 px-3 py-2 text-xs font-semibold text-slate-700 transition hover:border-slate-400 hover:text-slate-900 disabled:cursor-not-allowed disabled:opacity-60 dark:border-slate-600 dark:text-slate-200 dark:hover:border-slate-500"
                    disabled={isSubmitting || renameFetcher.state !== "idle"}
                  >
                    Rename
                  </button>
                )}
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
              </div>
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

function serializeAttestationResponse(credential: PublicKeyCredential) {
  const response = credential.response as AuthenticatorAttestationResponse;
  const clientExtensionResults =
    typeof credential.getClientExtensionResults === "function"
      ? credential.getClientExtensionResults()
      : {};

  return {
    id: credential.id,
    rawId: bufferToBase64URL(credential.rawId),
    type: credential.type,
    authenticatorAttachment: credential.authenticatorAttachment,
    clientExtensionResults,
    response: {
      clientDataJSON: bufferToBase64URL(response.clientDataJSON),
      attestationObject: bufferToBase64URL(response.attestationObject),
      transports: typeof response.getTransports === "function" ? response.getTransports() : undefined,
      publicKey: typeof response.getPublicKey === "function" ? bufferSourceToBase64URL(response.getPublicKey()) : undefined,
      publicKeyAlgorithm:
        typeof response.getPublicKeyAlgorithm === "function" ? response.getPublicKeyAlgorithm() : undefined,
    },
  };
}

function normalizePublicKeyCreationOptions(publicKey: PublicKeyCredentialCreationOptions) {
  const normalized: any = { ...publicKey };

  if (typeof normalized.challenge === "string") {
    normalized.challenge = decodeBase64URL(normalized.challenge);
  }

  if (normalized.user?.id && typeof normalized.user.id === "string") {
    normalized.user = { ...normalized.user, id: decodeBase64URL(normalized.user.id) };
  }

  if (Array.isArray(normalized.excludeCredentials)) {
    normalized.excludeCredentials = normalized.excludeCredentials.map((cred: any) => ({
      ...cred,
      id: typeof cred.id === "string" ? decodeBase64URL(cred.id) : cred.id,
    }));
  }

  return normalized as PublicKeyCredentialCreationOptions;
}

function decodeBase64URL(value: string) {
  const pad = value.length % 4 ? 4 - (value.length % 4) : 0;
  const base64 = (value + "=".repeat(pad)).replace(/-/g, "+").replace(/_/g, "/");
  return Uint8Array.from(atob(base64), c => c.charCodeAt(0));
}

function bufferToBase64URL(buffer: ArrayBuffer | ArrayBufferLike) {
  const bytes = new Uint8Array(buffer);
  let binary = "";
  for (let i = 0; i < bytes.byteLength; i++) {
    binary += String.fromCharCode(bytes[i]);
  }
  const base64 = btoa(binary);
  return base64.replace(/\+/g, "-").replace(/\//g, "_").replace(/=+$/, "");
}

function bufferSourceToBase64URL(buffer?: ArrayBuffer | ArrayBufferView | null) {
  if (!buffer) return undefined;
  if (buffer instanceof ArrayBuffer) {
    return bufferToBase64URL(buffer);
  }
  return bufferToBase64URL(buffer.buffer);
}
