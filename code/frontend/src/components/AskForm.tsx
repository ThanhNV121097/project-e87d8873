import { useState, type FormEvent } from "react";
import { T, useContent } from "../editable";
import { canSend, sendEnquiry } from "../site";
import Select from "./Select";

type State = "idle" | "sending" | "sent" | "failed";

/**
 * "LEAVE YOUR DETAILS AND WE CALL YOU": the second way to buy that the brief
 * asks for, on a bike's own page. It goes to the owner through the same line
 * as the contact form (src/site.ts), with the bike, the size and the rider's
 * height attached so the shop can call back with an answer, not a question.
 */
export default function AskForm({ bike, sizes }: { bike: string; sizes: readonly string[] }) {
  const [state, setState] = useState<State>("idle");
  const [problem, setProblem] = useState("");
  const [size, setSize] = useState("");
  const offline = useContent("contact.form.offline");
  const unsure = useContent("ask.unsure");

  async function submit(e: FormEvent<HTMLFormElement>) {
    e.preventDefault();
    const data = new FormData(e.currentTarget);
    const value = (k: string) => String(data.get(k) ?? "").trim();
    setState("sending");
    try {
      await sendEnquiry({
        name: value("name"), reply_to: value("reply_to"), website: value("website"),
        message: value("message") || bike,
        details: { bike, size: size || unsure, height: value("height") },
      });
      setState("sent");
    } catch (err) {
      setProblem(err instanceof Error ? err.message : "");
      setState("failed");
    }
  }

  if (state === "sent") {
    return <T k="ask.sent" as="p" className="text-lg text-ink" />;
  }
  const input = "w-full appearance-none rounded-[var(--radius-sm)] border border-field bg-ground px-4 py-3 text-ink";
  return (
    <form onSubmit={(e) => void submit(e)} data-testid="ask-form" className="grid gap-4 sm:grid-cols-2">
      <label className="flex flex-col gap-2 text-sm"><T k="contact.form.name" className="text-ink-soft" />
        <input id="ask-name" name="name" autoComplete="name" required className={input} /></label>
      <label className="flex flex-col gap-2 text-sm"><T k="contact.form.reply_to" className="text-ink-soft" />
        <input id="ask-reply" name="reply_to" autoComplete="tel" required className={input} /></label>
      <label className="flex flex-col gap-2 text-sm"><T k="ask.height" className="text-ink-soft" />
        <input id="ask-height" name="height" type="number" inputMode="numeric" min={120} max={220} className={input} /></label>
      <div className="flex flex-col gap-2 text-sm"><T k="ask.size" className="text-ink-soft" />
        <Select id="ask-size" aria-label={unsure} value={size} onChange={setSize} placeholder={unsure}
          options={sizes.map((s) => ({ value: s, label: s }))} /></div>
      <label className="flex flex-col gap-2 text-sm sm:col-span-2"><T k="ask.message" className="text-ink-soft" />
        <textarea id="ask-message" name="message" rows={3} className={input} /></label>
      <input name="website" tabIndex={-1} autoComplete="off" aria-hidden="true" className="absolute -left-[9999px] h-px w-px opacity-0" />
      {state === "failed" && <p role="alert" className="text-sm text-danger sm:col-span-2">{problem}</p>}
      {!canSend && <p className="text-sm text-ink-soft sm:col-span-2">{offline}</p>}
      <button type="submit" disabled={state === "sending"} data-testid="ask-send"
        className="justify-self-start whitespace-nowrap rounded-full bg-accent px-7 py-3 text-sm font-medium text-accent-ink disabled:opacity-60">
        <T k={state === "sending" ? "contact.form.sending" : "ask.send"} />
      </button>
    </form>
  );
}
