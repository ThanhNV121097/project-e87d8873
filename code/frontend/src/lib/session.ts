import { useSyncExternalStore } from "react";

/**
 * Who is signed in, on this browser — asked of the server.
 *
 * The owner signs in with an email and a password the SERVER checks
 * (POST /v1/sessions) and the server's cookie is the session; nothing here
 * ever holds a password past the one request. What is kept in the browser is
 * a HINT: the owner's email, in localStorage under "agentteam:session". It
 * exists only so that `useSession()` — which the management area's gate reads
 * synchronously — can answer on the first frame of a reload or a deep link
 * instead of sending a signed-in owner to the sign-in page. It is trusted for
 * nothing else: the first page that reads the session asks the server who
 * this is (GET /v1/me), once per page load, and
 *
 *   200   the server's email replaces the hint — the session is real;
 *   401, any other answer, or none   the hint is cleared — nobody is signed in.
 *
 * A browser that holds a valid cookie but no hint (storage cleared) is sent
 * to the sign-in page once, where the same answer shows "Signed in as …".
 *
 *   useSession()          the owner signed in, or null. Every component that
 *                         reads it draws again the moment it changes, in this
 *                         tab or another.
 *   signIn(email, password)   resolves when the server accepted them; rejects
 *                         with the sentence to show when it did not.
 *   signOut()             ends the session on the server first, and only
 *                         when the server says so (204) forgets it here.
 *   useRole()             "" — this site has one account and no roles.
 *
 * A write the server answered 401 calls endSession(): the hint goes, the gate
 * sends the owner to /admin/sign-in, and useSessionEnded() carries the
 * sentence the sign-in form shows until the next attempt.
 */

export interface Session {
  email: string;
  name: string;
  /** Absent: the site has one kind of visitor. */
  role?: string;
}

const KEY = "agentteam:session";

export const SESSION_ENDED = "Your session has ended. Sign in again.";
const SIGN_IN_FAILED = "Could not sign in. Please try again.";

/** Where the hint is kept when the browser refuses storage: this page only. */
let memory: string | null = null;

function load(): string | null {
  try {
    return localStorage.getItem(KEY);
  } catch {
    return memory;
  }
}

function store(raw: string | null) {
  memory = raw;
  try {
    if (raw === null) localStorage.removeItem(KEY);
    else localStorage.setItem(KEY, raw);
  } catch {
    // Storage refused (a private window, a sandbox): the hint lasts this page.
  }
}

/** "lan.tran@example.com" → "Lan Tran": a name to greet with. */
function nameFrom(email: string): string {
  const words = email.split("@")[0].split(/[._-]+/).filter(Boolean);
  return words.map((w) => w[0].toUpperCase() + w.slice(1)).join(" ") || email;
}

function parse(raw: string | null): Session | null {
  if (!raw) return null;
  try {
    const v = JSON.parse(raw) as { email?: unknown; name?: unknown; role?: unknown } | null;
    const email = typeof v?.email === "string" ? v.email.trim() : "";
    if (!email) return null;
    const name = typeof v?.name === "string" ? v.name.trim() : "";
    const session: Session = { email, name: name || nameFrom(email) };
    const role = typeof v?.role === "string" ? v.role.trim().toLowerCase() : "";
    if (role) session.role = role;
    return session;
  } catch {
    return null;
  }
}

// The session as read last, kept while its text is the same: React asks for
// it on every render and must get the same object back when nothing changed.
let lastRaw: string | null | undefined;
let last: Session | null = null;

function read(): Session | null {
  const raw = load();
  if (raw !== lastRaw) {
    lastRaw = raw;
    last = parse(raw);
  }
  return last;
}

/** The address of the service: the same base the data module reads. */
export function apiUrl(path: string): string {
  const base = (import.meta.env.VITE_API_URL as string | undefined) ?? "/api";
  return `${base.replace(/\/+$/, "")}/v1${path}`;
}

/** The sentence of the service's error envelope, or "" when the answer is not one. */
export async function sentenceOf(res: Response): Promise<string> {
  try {
    const body = (await res.json()) as { error?: unknown } | null;
    return typeof body?.error === "string" ? body.error.trim() : "";
  } catch {
    return "";
  }
}

async function ownerOf(res: Response): Promise<string> {
  try {
    const body = (await res.json()) as { email?: unknown } | null;
    return typeof body?.email === "string" ? body.email.trim() : "";
  } catch {
    return "";
  }
}

const listeners = new Set<() => void>();

function changed() {
  listeners.forEach((l) => l());
}

/** Counts every change of who is signed in: an answer to a question asked
 *  before the last change is older than it, and is dropped. */
let epoch = 0;
let started = false;
let ended = "";

function remember(email: string) {
  store(JSON.stringify({ email }));
  changed();
}

function forget() {
  epoch += 1;
  store(null);
  changed();
}

/** The one question asked of the server per page load: who is this? */
async function verify() {
  const asked = epoch;
  let owner = "";
  try {
    const res = await fetch(apiUrl("/me"));
    if (res.ok) owner = await ownerOf(res);
  } catch {
    owner = "";
  }
  if (asked !== epoch) return;
  if (owner) remember(owner);
  else if (load() !== null) forget();
}

function subscribe(onChange: () => void) {
  listeners.add(onChange);
  // Another tab signing in or out, or the Studio's frame setting the visitor.
  const onStorage = (e: StorageEvent) => {
    if (e.key === KEY || e.key === null) onChange();
  };
  window.addEventListener("storage", onStorage);
  if (!started) {
    started = true;
    void verify();
  }
  return () => {
    listeners.delete(onChange);
    window.removeEventListener("storage", onStorage);
  };
}

/** The owner signed in, or null. */
export function useSession(): Session | null {
  return useSyncExternalStore(subscribe, read, () => null);
}

/** The sentence a session that ended under a write leaves for the sign-in form, or "". */
export function useSessionEnded(): string {
  return useSyncExternalStore(subscribe, () => ended, () => "");
}

/**
 * Signs the owner in. Resolves when the server accepted the email and
 * password; rejects with the sentence to show when it did not — the server's
 * own ("That email and password do not match.", "Too many attempts. …") or,
 * with no usable answer at all, "Could not sign in. Please try again."
 */
export async function signIn(email: string, password: string): Promise<void> {
  ended = "";
  epoch += 1;
  changed();
  let res: Response;
  try {
    res = await fetch(apiUrl("/sessions"), {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ email: email.trim(), password }),
    });
  } catch {
    throw new Error(SIGN_IN_FAILED);
  }
  if (!res.ok) throw new Error((await sentenceOf(res)) || SIGN_IN_FAILED);
  const owner = await ownerOf(res);
  if (!owner) throw new Error(SIGN_IN_FAILED);
  epoch += 1;
  remember(owner);
}

let leaving = false;

/**
 * Signs the owner out. The session is forgotten here only after the server
 * answered 204: when the request fails the owner stays signed in and can
 * press the button again — the server never says "signed out" while the
 * session still works.
 */
export async function signOut(): Promise<void> {
  if (leaving) return;
  leaving = true;
  try {
    const res = await fetch(apiUrl("/sessions/current"), { method: "DELETE" });
    if (res.status === 204) {
      ended = "";
      forget();
    }
  } catch {
    // No answer: nothing changes on screen.
  } finally {
    leaving = false;
  }
}

/** The server refused a write for want of a session: drop it, and say why on the sign-in form. */
export function endSession(): void {
  ended = SESSION_ENDED;
  forget();
}

/** The role the owner acts as: the empty text — the site has one account and no roles. */
export function useRole(): string {
  return useSession()?.role ?? "";
}

/**
 * Whether the visitor may enter an area open to `roles`.
 *
 * An empty list is open to anyone signed in. Nobody signed in is never
 * allowed: a role check is not a sign-in check, and a page open to
 * ["supplier"] must refuse a visitor with no role exactly as it refuses a
 * buyer.
 */
export function allowed(session: Session | null, roles?: readonly string[]): boolean {
  if (!session) return false;
  if (!roles || roles.length === 0) return true;
  const role = session.role ?? "";
  return roles.some((r) => r.trim().toLowerCase() === role);
}
